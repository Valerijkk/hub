package main

import (
	"context"
	"net"

	mailmod "github.com/valerii/hub/backend/internal/mail"
)

const mailboxPublicDNSResolverAddress = "mail-dns-resolver:53"

func newMailboxPublicDNSLookup() (mailmod.PublicDNSLookup, error) {
	resolver := &net.Resolver{
		PreferGo:     true,
		StrictErrors: true,
		Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
			if ctx == nil || ctx.Err() != nil {
				return nil, mailmod.ErrMailboxEngineConfigInvalid
			}
			switch network {
			case "udp", "tcp":
			default:
				return nil, mailmod.ErrMailboxEngineConfigInvalid
			}
			return (&net.Dialer{}).DialContext(ctx, network, mailboxPublicDNSResolverAddress)
		},
	}
	lookup, err := mailmod.NewNetResolverPublicDNSLookup(resolver)
	if err != nil {
		return nil, mailmod.ErrMailboxEngineConfigInvalid
	}
	return lookup, nil
}
