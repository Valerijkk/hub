# Backend: безопасность

> Часть плейбука. Канон и оглавление — [`../backend.md`](../backend.md).

Аутентификация, авторизация, секреты, валидация входа.

---

## 7. Безопасность

Security-инварианты Hub — не рекомендации, а жёсткие требования TZ (`docs/tz_constraints.md`) и результат аудита (12/12 CRITICAL исправлены). Их нарушение — блокер мержа. `be-ver` и `security-auditor` грепают именно по этим паттернам.

### 7.1. SSRF: safehttp для user-controlled URL

Любой исходящий запрос на URL, который может задать пользователь (GitLab/GitHub интеграции, WebPush endpoint, OpenGraph link preview, webhook), идёт через `safehttp.NewSafeHTTPClient` + предварительную проверку `ValidateOutboundURL`. **Запрещён** `http.DefaultClient`.

<good-example>

Реальная защита (`platform/http/safedialer.go`):

```go
// ValidateOutboundURL: только http/https, без userinfo, непустой host.
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
    return nil
}

// SafeDialContext резолвит хост и отклоняет приватные/loopback/link-local IP
// (закрывает DNS-rebinding — резолв прямо перед подключением).
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
```

Заблокированы диапазоны: `10.0.0.0/8`, `127.0.0.0/8`, `169.254.0.0/16` (метаданные AWS `169.254.169.254`!), `172.16.0.0/12`, `192.168.0.0/16`, IPv6 ULA/link-local и др.

Использование в модуле:

```go
client := safehttp.NewSafeHTTPClient(10 * time.Second)
if err := safehttp.ValidateOutboundURL(userURL); err != nil {
    return apperrors.NewBadRequest("недопустимый URL")
}
resp, err := client.Get(userURL) // и dialer повторно проверит IP после резолва
```
</good-example>

<bad-example>

```go
// SSRF: пользователь укажет http://169.254.169.254/latest/meta-data/ и утащит креды инстанса.
resp, err := http.DefaultClient.Get(userURL)
```
</bad-example>

### 7.2. DoS: io.LimitReader вместо io.ReadAll

Чтение тела/ответа без лимита — memory DoS. Оборачивайте reader в `io.LimitReader`. Максимум для тела запроса align с Fiber `BodyLimit` (25 МБ); для файлов/бэкапов — свой cap.

<good-example>

```go
const maxBody = 25 << 20 // 25 MiB
limited := io.LimitReader(r.Body, maxBody)
data, err := io.ReadAll(limited) // теперь безопасно
```

Реальные места применения (`security_guidelines.md`): `backup/service.go` (`LimitReader` 512 MiB cap), `apigateway/webhook.go` (`io.Copy(io.LimitReader)` вместо `ReadFrom`), `platform/http/opengraph.go` (`LimitReader` на превью).
</good-example>

<bad-example>

```go
data, err := io.ReadAll(r.Body) // атакующий шлёт 10 ГБ → OOM
```
</bad-example>

### 7.3. Timing attack: subtle.ConstantTimeCompare

Сравнение секретов (magic-link коды, refresh-хеши, webhook-токены) — только `subtle.ConstantTimeCompare` (или `hmac.Equal`). **Запрещён** `!=` / `==` для секретов.

<good-example>

Реально (`auth/magiclink.go`, SEC-07):

```go
stored, err := s.redis.Get(ctx, codeKey).Result()
// …
// SEC-07: constant-time compare — защита от timing side-channel при подборе кода.
if subtle.ConstantTimeCompare([]byte(stored), []byte(code)) != 1 {
    s.recordFailure(ctx, failsKey)
    return ErrMagicLinkInvalid
}
```

И для refresh-хеша (`auth/service_session.go`):

```go
oldHash := hashRefreshToken(oldRefreshToken)
if subtle.ConstantTimeCompare([]byte(sess.RefreshToken), []byte(oldHash)) != 1 {
    return fmt.Errorf("refresh token: hash mismatch")
}
```
</good-example>

<bad-example>

```go
if stored != code { // timing side-channel: время сравнения зависит от совпавшего префикса
    return ErrMagicLinkInvalid
}
```
</bad-example>

### 7.4. Webhook: подпись обязательна, skip запрещён

Проверка подписи вебхука (LiveKit, GitHub, GitLab) не может быть пропущена — даже при отсутствии claim реджектим (SEC-02). Fail-closed.

<good-example>

```go
// ПРАВИЛЬНО — reject if hash claim missing.
expectedHash, ok := claims["sha256"].(string)
if !ok || expectedHash == "" {
    return fiber.ErrUnauthorized
}
// затем hmac.Equal(computed, expected)
```
</good-example>

<bad-example>

```go
// ЗАПРЕЩЕНО — если подписи нет, пропускаем проверку (обход аутентичности).
if expectedHash != "" {
    // verify…
}
// иначе — принимаем как валидный
```
</bad-example>

### 7.5. Авторизация: RBAC/membership на каждом чувствительном действии

Каждый endpoint, отдающий или меняющий данные, проверяет право пользователя (system role / project role / membership). Особо: WS-подписки и поиск.

**WebSocket (NF-02 / SEC-H1):** подписка на комнату проходит `Authorizer.CanSubscribe` — неизвестный префикс fail-closed.

<good-example>

Реальный `ws/authorizer.go`:

```go
func (a *configAuthorizer) CanSubscribe(ctx context.Context, userIDStr, room string) (bool, error) {
    userID, err := uuid.Parse(userIDStr)
    if err != nil {
        return false, fmt.Errorf("ws authz: invalid user id %q: %w", userIDStr, err)
    }
    prefix, idStr, ok := splitRoom(room) // "channel:{id}" → ("channel", id)
    if !ok {
        return false, fmt.Errorf("ws authz: malformed room %q", room)
    }
    entityID, _ := uuid.Parse(idStr)
    switch prefix {
    case "channel":
        if a.cfg.IsChannelMember == nil {
            return false, nil // fail-closed
        }
        return a.cfg.IsChannelMember(ctx, entityID, userID)
    case "call":
        if a.cfg.IsCallParticipantOrChannelMember == nil {
            return false, nil
        }
        return a.cfg.IsCallParticipantOrChannelMember(ctx, entityID, userID)
    case "project":
        if a.cfg.IsProjectMember == nil {
            return false, nil
        }
        return a.cfg.IsProjectMember(ctx, entityID, userID)
    default:
        return false, nil // неизвестный префикс — запрет
    }
}
```
</good-example>

**Поиск (NF-03 / SEC-H2):** любой поиск фильтруется по данным, доступным пользователю. Никогда не отдавайте глобальную выборку.

<good-example>

Из `search/service.go`: `searchMessages` фильтрует `channel_id ∈ getUserChannelIDs(userID)`; issues/projects/files — по `projectIDs` пользователя.

```go
// Mongo-фильтр поиска сообщений — только каналы пользователя.
filter := bson.M{"channel_id": bson.M{"$in": channelIDs}}
```
</good-example>

<bad-example>

```go
// Утечка: поиск по всем сообщениям без скоупа пользователя.
filter := bson.M{"$text": bson.M{"$search": q}}
```
</bad-example>

### 7.6. Rate limiting

Лимиты (TZ mandatory) — предопределены в `platform/middleware/ratelimit.go`:

| Тип | Лимит | Конфиг |
|-----|-------|--------|
| Auth | 3 / 10 мин | `AuthRateLimit` |
| API general | 100 / сек | `APIRateLimit` |
| File upload | 10 / мин | `FileUploadRateLimit` |
| Messages | 60 / мин | `MessageSendRateLimit` |
| AI | 20 / мин | `AIRateLimit` |

<good-example>

Реально (`ratelimit.go`):

```go
var (
    AuthRateLimit        = RateLimitConfig{Max: 3, Window: 10 * time.Minute}
    APIRateLimit         = RateLimitConfig{Max: 100, Window: 1 * time.Second}
    FileUploadRateLimit  = RateLimitConfig{Max: 10, Window: 1 * time.Minute}
    MessageSendRateLimit = RateLimitConfig{Max: 60, Window: 1 * time.Minute}
)
```

Для чувствительных путей (auth, magic-link verify) — Redis-лимитер `RateLimitRedis` (устойчив к перезапуску и горизонтальному масштабированию, ключ по email через `ExtractEmailFromBody`, а не только по IP). При сбое Redis — fallback на in-memory (fail-safe, не fail-open).
</good-example>

### 7.7. Аутентификация: Magic Link + JWT + refresh

Модель Hub (ADR-AUTH-01, TZ):

- **Только Magic Link (email).** Ни паролей, ни SSO/OAuth/LDAP. Telegram — только notify bind, не login.
- **JWT access — 15 минут TTL**, HS256, claims `uid`/`sid`.
- **Refresh — 30 дней, в Redis**, мгновенный отзыв (удаление ключа), ротация при каждом refresh.
- **Refresh at rest хешируется SHA-256** (NF-12) — в Redis и в записи сессии хранится хеш, не сам токен.
- **Rate limit auth — 3/10мин.**

<good-example>

Хеширование refresh (`auth/service_session.go`, NF-12):

```go
func hashRefreshToken(token string) string {
    sum := sha256.Sum256([]byte(token))
    return hex.EncodeToString(sum[:])
}

func refreshRedisKey(plaintext string) string {
    return "refresh:" + hashRefreshToken(plaintext)
}
```

Генерация access-токена с TTL 15 минут (accessTTL передаётся из config):

```go
func (s *SessionService) generateAccessToken(userID, sessionID uuid.UUID) (string, error) {
    now := time.Now().UTC()
    claims := JWTClaims{
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)), // 15m
            IssuedAt:  jwt.NewNumericDate(now),
            Issuer:    "hub",
        },
        UserID:    userID.String(),
        SessionID: sessionID.String(),
    }
    return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
}
```

Валидация токена проверяет метод подписи (защита от `alg=none`):

```go
token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (any, error) {
    if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
    }
    return s.jwtSecret, nil
})
```

Ротация refresh (порядок ME-02 против TOCTOU): (1) генерируем новый → (2) сохраняем новый в Redis → (3) обновляем хеш в сессии → (4) удаляем старый:

```go
newRefresh := uuid.New().String()
newHash := hashRefreshToken(newRefresh)
accessToken, _ := s.generateAccessToken(userID, sessionID)
if err := s.storeRefreshToken(ctx, newRefresh, *payload); err != nil { /* … */ }
if err := s.repo.UpdateSessionRefreshHash(ctx, sessionID, newHash, time.Now().UTC()); err != nil { /* … */ }
s.deleteRefreshToken(ctx, oldRefreshToken) // мгновенный отзыв старого
```
</good-example>

### 7.8. Секреты и шифрование at rest

- Секреты — только из env (`.env` в `.gitignore`, `.env.example` без значений). Никаких hardcoded creds в репо (INF-03).
- Токены интеграций (GitHub/GitLab) шифруются AES-GCM (`platform/crypto/aesgcm.go`).
- `DEV_MODE=true` запрещён в CI/prod (INF-02).

### 7.9. Валидация и санитизация входа

| Вход | Правило |
|------|---------|
| UUID-параметры | `httputil.ParseUUIDParam` (+ отклонение `uuid.Nil`) |
| Тело запроса | `c.Bind().JSON` + validate-теги, лимит 25 МБ |
| User URLs | `safehttp.ValidateOutboundURL` перед fetch |
| Webhook body | обязательная подпись + хеш |
| Имена файлов | `sanitizeFilename` (path traversal, 200-char cap) |
| Mongo-поиск | фильтр по каналам/проектам пользователя |

### 7.10. Чек-лист security (грепается be-ver)

- [ ] Нет `http.DefaultClient` для user-URL → `safehttp.NewSafeHTTPClient` + `ValidateOutboundURL`.
- [ ] Нет `io.ReadAll` без `io.LimitReader`.
- [ ] Сравнение секретов — `subtle.ConstantTimeCompare` / `hmac.Equal`, не `!=`.
- [ ] Проверка подписи вебхука не пропускается (fail-closed).
- [ ] RBAC/membership на каждом чувствительном endpoint.
- [ ] WS-подписка проходит `CanSubscribe`; неизвестный префикс — запрет.
- [ ] Поиск/выборка скоупятся по данным пользователя.
- [ ] Rate-limit применён (auth 3/10м, api 100/сек, файлы 10/мин, сообщения 60/мин).
- [ ] Refresh хешируется (SHA-256); access TTL 15м; отзыв мгновенный.
- [ ] Нет hardcoded секретов; нет `DEV_MODE=true` в prod-путях.
- [ ] Секреты не логируются.

---
