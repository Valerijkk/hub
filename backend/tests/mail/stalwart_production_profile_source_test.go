package mail_test

import (
	"os"
	"strings"
	"testing"
)

func TestStalwartProductionProfileSourceContract(t *testing.T) {
	privatePlan := stalwartProductionProfilePlanLines(t, "../../../deployments/stalwart/steady-private.ndjson")
	smtpPlan := stalwartProductionProfilePlanLines(t, "../../../deployments/stalwart/steady-smtp.ndjson")

	allowedEndpoints := `{"match":{"0":{"if":"listener == 'hub-jmap' && ((path == '/jmap' && method == 'POST') || (path == '/jmap/session' && method == 'GET') || (path == '/healthz/live' && method == 'GET') || (path == '/healthz/ready' && method == 'GET'))","then":"200"}},"else":"404"}`
	systemSettings := `{"@type":"update","object":"SystemSettings","value":{"maxConnections":512,"proxyTrustedNetworks":{}}}`
	http := `{"@type":"update","object":"Http","value":{"allowedEndpoints":` + allowedEndpoints + `,"usePermissiveCors":false,"useXForwarded":false}}`
	privateListener := `{"@type":"reconcile","object":"NetworkListener","matchOn":["name"],"value":{"hub-jmap":{"name":"hub-jmap","protocol":"http","bind":{"[::]:8081":true},"useTls":false,"tlsImplicit":false,"maxConnections":256,"socketBacklog":128,"overrideProxyTrustedNetworks":{}}}}`
	smtpListener := `{"@type":"reconcile","object":"NetworkListener","matchOn":["name"],"value":{"hub-jmap":{"name":"hub-jmap","protocol":"http","bind":{"[::]:8081":true},"useTls":false,"tlsImplicit":false,"maxConnections":256,"socketBacklog":128,"overrideProxyTrustedNetworks":{}},"smtp-inbound":{"name":"smtp-inbound","protocol":"smtp","bind":{"[::]:25":true},"useTls":true,"tlsImplicit":false,"maxConnections":128,"socketBacklog":128,"overrideProxyTrustedNetworks":{}}}}`

	stalwartProductionProfileRequireExactPlan(t, privatePlan, []string{systemSettings, http, privateListener})
	stalwartProductionProfileRequireExactPlan(t, smtpPlan, []string{systemSettings, http, smtpListener})
	if privatePlan[1] != smtpPlan[1] {
		t.Fatal("private and SMTP plans must use byte-identical Http records")
	}

	composeBytes, err := os.ReadFile("../../../deployments/docker-compose.mail.prod.yml")
	if err != nil {
		t.Fatalf("read mail production compose source: %v", err)
	}
	compose := string(composeBytes)
	if !strings.Contains(compose, "# Registry-owned private steady-state Stalwart service.") {
		t.Fatal("compose source must declare Registry ownership for the steady profile")
	}
	stalwart := stalwartProductionProfileServiceBlock(t, compose, "stalwart")
	for _, required := range []string{
		"image: stalwartlabs/stalwart:v0.16.18@sha256:0df5900cab389a8ec47b7521ef0681ec93598caf72a09097685845211861f6c2",
		"STALWART_HEALTHCHECK_URL: \"http://127.0.0.1:8081/healthz/ready\"",
		"read_only: true",
		"- \"no-new-privileges:true\"",
		"- mail_engine_internal",
		"- ./stalwart/config.json:/etc/stalwart/config.json:ro",
		"- stalwart_data:/var/lib/stalwart",
		"driver: local",
		"max-size: \"10m\"",
		"max-file: \"3\"",
		"mem_limit: 1g",
		"cpus: 1.0",
		"pids_limit: 256",
		"memory: 1G",
		"cpus: \"1.0\"",
	} {
		if !strings.Contains(stalwart, required) {
			t.Fatalf("Stalwart steady service is missing %q", required)
		}
	}
	if expose := stalwartProductionProfileMappingBlock(t, stalwart, "expose", 4); expose != "    expose:\n      - \"8081\"\n" {
		t.Fatalf("Stalwart steady service expose block = %q, want only internal 8081", expose)
	}
	for _, forbiddenField := range []string{"ports", "healthcheck", "network_mode", "extends", "<<"} {
		if stalwartProductionProfileHasDirectField(stalwart, forbiddenField, 4) {
			t.Fatalf("Stalwart steady service must not declare %q", forbiddenField)
		}
	}
	for _, forbidden := range []string{
		"STALWART_RECOVERY", "STALWART_RECOVERY_ADMIN",
		"oauth", "8080", "\"25\"", "\"465\"", "\"587\"", "\"4190\"",
		"imap", "pop3", "lmtp", "sieve",
	} {
		if strings.Contains(strings.ToLower(stalwart), strings.ToLower(forbidden)) {
			t.Fatalf("Stalwart steady service contains forbidden %q", forbidden)
		}
	}

	backend := stalwartProductionProfileServiceBlock(t, compose, "backend")
	backendDependsOn := stalwartProductionProfileMappingBlock(t, backend, "depends_on", 4)
	stalwartDependency := stalwartProductionProfileMappingBlock(t, backendDependsOn, "stalwart", 6)
	if stalwartDependency != "      stalwart:\n        condition: service_healthy\n" {
		t.Fatalf("backend Stalwart dependency = %q, want only service_healthy", stalwartDependency)
	}
}

func stalwartProductionProfilePlanLines(t *testing.T, path string) []string {
	t.Helper()

	planBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read Stalwart plan %q: %v", path, err)
	}
	lines := strings.Split(string(planBytes), "\n")
	if len(lines) != 4 || lines[3] != "" {
		t.Fatalf("Stalwart plan %q must contain exactly three newline-terminated NDJSON records", path)
	}
	return lines[:3]
}

func stalwartProductionProfileRequireExactPlan(t *testing.T, got, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("Stalwart plan line count = %d, want %d", len(got), len(want))
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("Stalwart plan line %d = %q, want %q", index+1, got[index], want[index])
		}
	}
}

func stalwartProductionProfileServiceBlock(t *testing.T, compose, service string) string {
	t.Helper()

	services := stalwartProductionProfileMappingBlock(t, compose, "services", 0)
	return stalwartProductionProfileMappingBlock(t, services, service, 2)
}

func stalwartProductionProfileMappingBlock(t *testing.T, source, name string, indent int) string {
	t.Helper()

	header := strings.Repeat(" ", indent) + name + ":"
	start := stalwartProductionProfileLineOffset(source, header)
	if start < 0 {
		t.Fatalf("compose source is missing mapping %q at indentation %d", name, indent)
	}

	for offset := start + len(header) + 1; offset < len(source); {
		lineEnd := strings.IndexByte(source[offset:], '\n')
		if lineEnd < 0 {
			break
		}
		lineEnd += offset
		line := source[offset:lineEnd]
		if stalwartProductionProfileLineAtIndent(line, indent) && strings.HasSuffix(line, ":") {
			return source[start:offset]
		}
		offset = lineEnd + 1
	}
	return source[start:]
}

func stalwartProductionProfileHasDirectField(source, name string, indent int) bool {
	_, found := stalwartProductionProfileDirectFieldValue(source, name, indent)
	return found
}

func stalwartProductionProfileDirectFieldValue(source, name string, indent int) (string, bool) {
	prefix := strings.Repeat(" ", indent) + name + ":"
	for offset := 0; offset < len(source); {
		lineEnd := strings.IndexByte(source[offset:], '\n')
		if lineEnd < 0 {
			lineEnd = len(source)
		} else {
			lineEnd += offset
		}
		line := source[offset:lineEnd]
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix)), true
		}
		if lineEnd == len(source) {
			break
		}
		offset = lineEnd + 1
	}
	return "", false
}

func stalwartProductionProfileLineOffset(source, wanted string) int {
	for offset := 0; offset < len(source); {
		lineEnd := strings.IndexByte(source[offset:], '\n')
		if lineEnd < 0 {
			lineEnd = len(source)
		} else {
			lineEnd += offset
		}
		if source[offset:lineEnd] == wanted {
			return offset
		}
		if lineEnd == len(source) {
			break
		}
		offset = lineEnd + 1
	}
	return -1
}

func stalwartProductionProfileLineAtIndent(line string, indent int) bool {
	prefix := strings.Repeat(" ", indent)
	return strings.HasPrefix(line, prefix) && (len(line) == indent || line[indent] != ' ')
}
