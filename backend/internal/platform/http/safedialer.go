package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ErrBlockedAddress возвращается, когда DNS-резолюция URL даёт адрес из
// приватного диапазона (RFC1918, loopback, link-local) — это блокирует
// классические SSRF-пайвоты во внутреннюю сеть.
var ErrBlockedAddress = errors.New("blocked private address")

// ErrBlockedScheme возвращается для запрещённых схем URL (file://, gopher://,
// данных и т.п.). Разрешены только http и https.
var ErrBlockedScheme = errors.New("blocked URL scheme")

// ErrUserInfoNotAllowed возвращается, когда URL содержит userinfo
// (user:pass@host) — это потенциальный вектор для SSRF и кражи кредов.
var ErrUserInfoNotAllowed = errors.New("URL with userinfo not allowed")

// blockedCIDRs — список диапазонов IP, к которым запрещено подключаться
// серверным HTTP-клиентом. Это базовая защита от SSRF в облачных средах
// (метаданные AWS/GCP, внутренние сервисы), при пайвотинге в RFC1918 и
// link-local-адреса.
var blockedCIDRs = func() []*net.IPNet {
	raw := []string{
		"0.0.0.0/8",       // «эта сеть»
		"10.0.0.0/8",      // RFC1918
		"100.64.0.0/10",   // CGNAT
		"127.0.0.0/8",     // локальная петля
		"169.254.0.0/16",  // link-local (метаданные AWS: 169.254.169.254)
		"172.16.0.0/12",   // RFC1918
		"192.0.0.0/24",    // назначения протоколов IETF
		"192.0.2.0/24",    // TEST-NET-1
		"192.168.0.0/16",  // RFC1918
		"198.18.0.0/15",   // бенчмарк
		"198.51.100.0/24", // TEST-NET-2
		"203.0.113.0/24",  // TEST-NET-3
		"224.0.0.0/4",     // мультикаст
		"240.0.0.0/4",     // зарезервировано
		"::1/128",         // локальная петля IPv6
		"fc00::/7",        // IPv6 ULA
		"fe80::/10",       // link-local IPv6
		"ff00::/8",        // мультикаст IPv6
	}
	nets := make([]*net.IPNet, 0, len(raw))
	for _, c := range raw {
		_, n, err := net.ParseCIDR(c)
		if err != nil {
			continue
		}
		nets = append(nets, n)
	}
	return nets
}()

// normalizeIPForBlocklist приводит IPv4-mapped IPv6 (::ffff:a.b.c.d) к 4 байтам.
// P1-BE-23: без To4() Contains на IPv4-CIDR может пропускать mapped (и на Go,
// где Contains уже матчит, мутация «drop To4» должна краснеть по len==4).
func normalizeIPForBlocklist(ip net.IP) net.IP {
	if ip == nil {
		return nil
	}
	if v4 := ip.To4(); v4 != nil {
		out := make(net.IP, net.IPv4len)
		copy(out, v4)
		return out
	}
	return ip
}

// isBlockedIP возвращает true, если IP принадлежит одному из запрещённых
// диапазонов (приватная сеть, loopback, link-local и т.п.).
//
// P1-BE-23: IPv4-mapped IPv6 (::ffff:a.b.c.d) нормализуем через To4() —
// иначе net.IPNet.Contains для IPv4-CIDR (4 байта) не матчит 16-байтный
// mapped-адрес и SSRF обходит blocklist (метаданные/loopback/RFC1918).
func isBlockedIP(ip net.IP) bool {
	ip = normalizeIPForBlocklist(ip)
	if ip == nil {
		return true
	}
	for _, n := range blockedCIDRs {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// ValidateOutboundURL проверяет, что URL безопасен для серверного fetch:
// схема http/https, отсутствует userinfo и хост не пуст.
// W4-A1-009 / P3-BE-11: IP-литералы из blockedCIDRs режем на parse
// (ранний fail до dial; контракт «Validate до Do» достаточен один).
// DNS-rebinding для hostname по-прежнему закрывает SafeDialContext.
func ValidateOutboundURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("%w: %s", ErrBlockedScheme, u.Scheme)
	}
	if u.User != nil {
		return ErrUserInfoNotAllowed
	}
	if u.Host == "" {
		return fmt.Errorf("invalid URL: empty host")
	}
	// Hostname() без порта; ParseIP ловит IPv4/IPv6/mapped литералы.
	if ip := net.ParseIP(u.Hostname()); ip != nil && isBlockedIP(ip) {
		return fmt.Errorf("%w: %s", ErrBlockedAddress, u.Hostname())
	}
	return nil
}

// SafeDialContext возвращает DialContext, который резолвит хост и отклоняет
// подключения к приватным/loopback/link-local IP. Применяется как обработчик
// для http.Transport, чтобы блокировать SSRF-попытки даже при DNS-rebinding
// (резолв происходит непосредственно перед подключением).
func SafeDialContext() func(ctx context.Context, network, addr string) (net.Conn, error) {
	d := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, fmt.Errorf("invalid addr %q: %w", addr, err)
		}

		// Резолвим все IP хоста и валидируем — это закрывает DNS-rebinding.
		ips, err := (&net.Resolver{}).LookupIPAddr(ctx, host)
		if err != nil {
			return nil, fmt.Errorf("resolve %s: %w", host, err)
		}
		if len(ips) == 0 {
			return nil, fmt.Errorf("resolve %s: no addresses", host)
		}
		for _, ip := range ips {
			if isBlockedIP(ip.IP) {
				return nil, fmt.Errorf("%w: %s -> %s", ErrBlockedAddress, host, ip.IP.String())
			}
		}

		// Подключаемся к первому подходящему адресу.
		return d.DialContext(ctx, network, net.JoinHostPort(ips[0].IP.String(), port))
	}
}

// NewSafeHTTPClient возвращает http.Client с подключённым SafeDialContext
// и пользовательским таймаутом. Используется для исходящих запросов на
// произвольные пользовательские URL (вебхуки, link previews).
func NewSafeHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			DialContext:           SafeDialContext(),
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			IdleConnTimeout:       60 * time.Second,
		},
	}
}
