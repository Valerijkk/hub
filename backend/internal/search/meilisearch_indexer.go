// Пакет search — индексация Meilisearch (SEARCH-01).
package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/issue"
	safehttp "github.com/valerii/hub/backend/internal/platform/http"
)

const issuesIndex = "hub_issues"

// MeiliIndexer отправляет документы в Meilisearch при наличии MEILI_HOST.
type MeiliIndexer struct {
	host   string
	apiKey string
	client *http.Client
}

// NewMeiliIndexerFromEnv возвращает nil без MEILI_HOST — Meili опционален.
// W5-A4-009 / P2-BE-18: ValidateOutboundURL + блок metadata/link-local SSRF.
// RFC1918 (Docker Meili) разрешён — host задаёт оператор, не пользователь.
func NewMeiliIndexerFromEnv() *MeiliIndexer {
	host := strings.TrimSpace(os.Getenv("MEILI_HOST"))
	if host == "" {
		return nil
	}
	if err := validateMeiliHost(host); err != nil {
		slog.Warn("MEILI_HOST rejected (SSRF guard)", "host", host, "error", err)
		return nil
	}
	return &MeiliIndexer{
		host:   strings.TrimRight(host, "/"),
		apiKey: os.Getenv("MEILI_MASTER_KEY"),
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// validateMeiliHost — схема/userinfo через SafeURL + запрет cloud-metadata pivots.
func validateMeiliHost(host string) error {
	if err := safehttp.ValidateOutboundURL(host); err != nil {
		return err
	}
	u, err := url.Parse(host)
	if err != nil {
		return err
	}
	h := strings.ToLower(u.Hostname())
	switch h {
	case "169.254.169.254", "metadata.google.internal", "metadata":
		return fmt.Errorf("%w: %s", safehttp.ErrBlockedAddress, h)
	}
	if ip := net.ParseIP(h); ip != nil {
		if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return fmt.Errorf("%w: %s", safehttp.ErrBlockedAddress, h)
		}
		// IPv4 link-local / AWS metadata
		if v4 := ip.To4(); v4 != nil && v4[0] == 169 && v4[1] == 254 {
			return fmt.Errorf("%w: %s", safehttp.ErrBlockedAddress, h)
		}
	}
	return nil
}

type meiliIssueDoc struct {
	ID        string `json:"id"`
	IssueKey  string `json:"issue_key"`
	Title     string `json:"title"`
	ProjectID string `json:"project_id"`
}

// IndexAllIssues загружает активные issues из Ent и upsert в Meili.
func (m *MeiliIndexer) IndexAllIssues(ctx context.Context, db *ent.Client) (int, error) {
	if m == nil {
		return 0, fmt.Errorf("meilisearch not configured")
	}
	issues, err := db.Issue.Query().
		Where(issue.DeletedAtIsNil()).
		WithProject().
		Limit(5000).
		All(ctx)
	if err != nil {
		return 0, err
	}
	docs := make([]meiliIssueDoc, 0, len(issues))
	for _, iss := range issues {
		pid := ""
		if iss.Edges.Project != nil {
			pid = iss.Edges.Project.ID.String()
		}
		docs = append(docs, meiliIssueDoc{
			ID:        iss.ID.String(),
			IssueKey:  iss.IssueKey,
			Title:     iss.Title,
			ProjectID: pid,
		})
	}
	if err := m.upsertDocuments(ctx, issuesIndex, docs); err != nil {
		return 0, err
	}
	return len(docs), nil
}

func (m *MeiliIndexer) upsertDocuments(ctx context.Context, index string, docs any) error {
	body, err := json.Marshal(docs)
	if err != nil {
		return err
	}
	reqURL := fmt.Sprintf("%s/indexes/%s/documents", m.host, index)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if m.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+m.apiKey)
	}
	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("meilisearch %d: %s", resp.StatusCode, string(b))
	}
	return nil
}
