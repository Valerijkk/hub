package mail_test

import (
	"os"
	"strings"
	"testing"
)

func TestMailboxProductionDNSOverlaySourceContract(t *testing.T) {
	compose := mailboxProductionDNSOverlaySource(t, "../../../deployments/docker-compose.mail.prod.yml")
	resolver := mailboxProductionDNSServiceBlock(t, compose, "mail-dns-resolver", "\n  backend:\n")

	for _, required := range []string{
		"image: coredns/coredns:1.14.7@sha256:7efd3c635b03efd68c4e8398fc45f0d993d0e9ab016f72c1cefb0fd6d01aa286",
		"restart: unless-stopped",
		"user: \"65532:65532\"",
		"read_only: true",
		"command: [\"-conf\", \"/etc/coredns/Corefile\"]",
		"- ./coredns/mail-public.Corefile:/etc/coredns/Corefile:ro",
		"- \"53/tcp\"",
		"- \"53/udp\"",
		"- \"no-new-privileges:true\"",
		"- ALL",
		"- NET_BIND_SERVICE",
		"mem_limit: 128m",
		"cpus: 0.25",
		"pids_limit: 64",
		"memory: 128M",
		"cpus: \"0.25\"",
		"driver: local",
		"max-size: \"10m\"",
		"max-file: \"3\"",
		"- mail_dns_client",
		"- mail_dns_egress",
	} {
		if !strings.Contains(resolver, required) {
			t.Fatalf("mail DNS resolver is missing %q", required)
		}
	}
	for _, forbidden := range []string{"\n    ports:", "\n    healthcheck:"} {
		if strings.Contains(resolver, forbidden) {
			t.Fatalf("mail DNS resolver contains forbidden %q", strings.TrimSpace(forbidden))
		}
	}

	backend := mailboxProductionDNSServiceBlock(t, compose, "backend", "\n\nnetworks:\n")
	for _, required := range []string{
		"MAILBOX_DOMAIN_VERIFICATION_MODE: \"${MAILBOX_DOMAIN_VERIFICATION_MODE:-}\"",
		"mail-dns-resolver:\n        condition: service_started",
		"- mail_dns_client",
	} {
		if !strings.Contains(backend, required) {
			t.Fatalf("backend overlay is missing %q", required)
		}
	}
	if strings.Contains(backend, "mail_dns_egress") {
		t.Fatal("backend must not attach to mail_dns_egress")
	}
	if strings.Count(compose, "- mail_dns_client") != 2 || strings.Count(compose, "- mail_dns_egress") != 1 {
		t.Fatal("mail DNS networks are not shared only by resolver and backend as required")
	}
	for _, required := range []string{
		"mail_dns_client:\n    internal: true",
		"mail_dns_egress:\n    driver: bridge",
	} {
		if !strings.Contains(compose, required) {
			t.Fatalf("mail DNS network topology is missing %q", required)
		}
	}

	envExample := mailboxProductionDNSOverlaySource(t, "../../../deployments/.env.example")
	for _, required := range []string{
		"# Required only when MAILBOX_ENGINE_ENABLED=true; only exact public_dns is accepted.",
		"MAILBOX_DOMAIN_VERIFICATION_MODE=",
	} {
		if !strings.Contains(envExample, required) {
			t.Fatalf("environment example is missing %q", required)
		}
	}

	corefile := mailboxProductionDNSOverlaySource(t, "../../../deployments/coredns/mail-public.Corefile")
	for _, required := range []string{
		".:53 {",
		"cache 30",
		"forward . tls://1.1.1.1 tls://1.0.0.1 {",
		"tls_servername cloudflare-dns.com",
		"policy sequential",
		"health_check 5s",
		"max_fails 2",
		"max_concurrent 16",
		"failfast_all_unhealthy_upstreams",
	} {
		if !strings.Contains(corefile, required) {
			t.Fatalf("mail public Corefile is missing %q", required)
		}
	}
	for _, forbidden := range []string{"/etc/resolv.conf", "log"} {
		if strings.Contains(corefile, forbidden) {
			t.Fatalf("mail public Corefile contains forbidden %q", forbidden)
		}
	}
}

func mailboxProductionDNSOverlaySource(t *testing.T, path string) string {
	t.Helper()

	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read mail production DNS overlay %q: %v", path, err)
	}
	return string(source)
}

func mailboxProductionDNSServiceBlock(t *testing.T, compose, service, endMarker string) string {
	t.Helper()

	start := strings.Index(compose, "\n  "+service+":\n")
	if start < 0 {
		t.Fatalf("compose overlay is missing service %q", service)
	}
	end := strings.Index(compose[start+1:], endMarker)
	if end < 0 {
		t.Fatalf("compose overlay does not terminate service %q", service)
	}
	return compose[start : start+1+end]
}
