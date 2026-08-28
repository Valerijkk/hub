package mail_test

import (
	"os"
	"strings"
	"testing"
)

func TestMailboxPublicDNSResolverWiringSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../cmd/server/wiring_mailbox_public_dns.go")
	if err != nil {
		t.Fatalf("read public DNS resolver wiring: %v", err)
	}
	source := string(sourceBytes)
	compact := strings.NewReplacer(" ", "", "\t", "", "\r", "", "\n", "").Replace(source)

	for _, required := range []string{
		"constmailboxPublicDNSResolverAddress=\"mail-dns-resolver:53\"",
		"funcnewMailboxPublicDNSLookup()(mailmod.PublicDNSLookup,error)",
		"resolver:=&net.Resolver{",
		"PreferGo:true,",
		"StrictErrors:true,",
		"Dial:func(ctxcontext.Context,network,_string)(net.Conn,error){",
		"ifctx==nil||ctx.Err()!=nil{returnnil,mailmod.ErrMailboxEngineConfigInvalid}",
		"switchnetwork{case\"udp\",\"tcp\":default:returnnil,mailmod.ErrMailboxEngineConfigInvalid}",
		"(&net.Dialer{}).DialContext(ctx,network,mailboxPublicDNSResolverAddress)",
		"mailmod.NewNetResolverPublicDNSLookup(resolver)",
	} {
		if !strings.Contains(compact, required) {
			t.Fatalf("public DNS resolver wiring is missing %q", required)
		}
	}

	for _, forbidden := range []string{
		"os.Getenv",
		"os.LookupEnv",
		"net.DefaultResolver",
		"net/http",
		"http.",
		"DevMode",
		"DialContext(ctx, network, address)",
		"return nil, ctx.Err()",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("public DNS resolver wiring contains forbidden %q", forbidden)
		}
	}
}
