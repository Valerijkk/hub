# Go examples

## Consumer-side interface

- Принцип: interface принадлежит потребителю capability, а не каждой struct.
- Evidence: `NORM`, `REC`; high.
- Режим: `REVIEW`.

```go
type BlobReader interface {
	Get(ctx context.Context, key string) ([]byte, error)
}
```

Не добавляйте interface, если concrete internal implementation проще и variation/testing seam отсутствует.

## Error identity

**BAD**

```go
if err != nil {
	return fmt.Errorf("failed")
}
```

**GOOD**

```go
if err != nil {
	return fmt.Errorf("authorize payment: %w", err)
}
```

Typed/sentinel error нужен только если caller действительно выбирает другую recovery. Иначе достаточно cause + useful context.

## Request cancellation

```go
func (s *Service) Load(ctx context.Context, id string) (User, error) {
	return s.repo.Load(ctx, id)
}
```

`context.Context` проходит первым аргументом request-scoped call chain; не хранится произвольно в struct. Evidence: `NORM`, `REC`; [S39, S40](../sources/registry.md).

## Explicit ownership

**BAD**

```go
for event := range incoming {
	go process(event)
}
```

**GOOD shape**

```go
limit := make(chan struct{}, 32)
for event := range incoming {
	select {
	case limit <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}

	go func(e Event) {
		defer func() { <-limit }()
		process(ctx, e)
	}(event)
}
```

Production variant также определяет error propagation и shutdown. Channel не обязателен: для shared map с ясным ownership `sync.RWMutex` может быть проще. Evidence: `NORM`, `REC`; [S38–S40](../sources/registry.md).
