package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	mailmod "github.com/valerii/hub/backend/internal/mail"
)

const (
	mailboxLocalProfileCAFile = "/run/caddy-ca/pki/authorities/local/root.crt"
	mailboxLocalProfileCAMax  = 64 * 1024
)

func newMailboxHTTPClient(lookupEnv func(string) (string, bool)) (*http.Client, error) {
	tlsConfig, err := mailboxLocalTLSConfig(lookupEnv)
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Transport: &http.Transport{
			Proxy:                 nil,
			DialContext:           (&net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
			MaxIdleConns:          8,
			MaxIdleConnsPerHost:   4,
			MaxConnsPerHost:       8,
			IdleConnTimeout:       30 * time.Second,
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: time.Second,
			TLSClientConfig:       tlsConfig,
		},
		Timeout: 15 * time.Second,
	}, nil
}

func mailboxLocalTLSConfig(lookupEnv func(string) (string, bool)) (*tls.Config, error) {
	if lookupEnv == nil {
		return nil, mailmod.ErrMailboxEngineConfigInvalid
	}

	localProfile := false
	if value, present := lookupEnv("MAILBOX_LOCAL_PROFILE"); present {
		switch value {
		case "false":
		case "true":
			localProfile = true
		default:
			return nil, mailmod.ErrMailboxEngineConfigInvalid
		}
	}

	caFile, caFilePresent := lookupEnv("MAILBOX_ENGINE_CA_FILE")
	if !localProfile {
		if caFilePresent {
			return nil, mailmod.ErrMailboxEngineConfigInvalid
		}
		return nil, nil
	}
	if !caFilePresent || caFile != mailboxLocalProfileCAFile {
		return nil, mailmod.ErrMailboxEngineConfigInvalid
	}

	rootCAs, err := mailboxLocalRootCAs(caFile)
	if err != nil {
		return nil, mailmod.ErrMailboxEngineConfigInvalid
	}
	return &tls.Config{
		MinVersion: tls.VersionTLS12,
		RootCAs:    rootCAs,
	}, nil
}

func mailboxLocalRootCAs(caFile string) (*x509.CertPool, error) {
	file, err := os.Open(caFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() < 1 || info.Size() > mailboxLocalProfileCAMax {
		return nil, mailmod.ErrMailboxEngineConfigInvalid
	}

	encoded, err := io.ReadAll(io.LimitReader(file, mailboxLocalProfileCAMax+1))
	if err != nil {
		return nil, err
	}
	if len(encoded) < 1 || len(encoded) > mailboxLocalProfileCAMax {
		return nil, mailmod.ErrMailboxEngineConfigInvalid
	}

	rest := encoded
	certificates := make([]*x509.Certificate, 0, 1)
	for {
		rest = bytes.TrimSpace(rest)
		if len(rest) == 0 {
			break
		}
		if !bytes.HasPrefix(rest, []byte("-----BEGIN ")) {
			return nil, mailmod.ErrMailboxEngineConfigInvalid
		}
		block, remaining := pem.Decode(rest)
		if block == nil || block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
			return nil, mailmod.ErrMailboxEngineConfigInvalid
		}
		certificate, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}
		certificates = append(certificates, certificate)
		rest = remaining
	}
	if len(certificates) != 1 || !certificates[0].IsCA {
		return nil, mailmod.ErrMailboxEngineConfigInvalid
	}

	rootCAs, err := x509.SystemCertPool()
	if err != nil || rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}
	rootCAs.AddCert(certificates[0])
	return rootCAs, nil
}
