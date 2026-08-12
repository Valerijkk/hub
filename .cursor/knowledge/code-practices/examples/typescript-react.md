# TypeScript / React examples

## Model finite states

- Принцип: invalid combinations лучше исключать type model-ю.
- Evidence: `NORM`; high; [S44](../sources/registry.md).
- Режим: strict type check — `HARD` candidate; конкретная union — `REVIEW`.

**BAD**

```ts
type LoadState = {
  loading: boolean;
  data?: Order;
  error?: Error;
};
```

**GOOD**

```ts
type LoadState =
  | { status: "idle" }
  | { status: "loading" }
  | { status: "ready"; data: Order }
  | { status: "failed"; error: Error };
```

Не создавайте giant universal union для open-ended plugin data.

## Derived state

**BAD**

```tsx
const [fullName, setFullName] = useState("");
useEffect(() => setFullName(`${first} ${last}`), [first, last]);
```

**GOOD**

```tsx
const fullName = `${first} ${last}`;
```

Effect синхронизирует внешний system; обычное derived value вычисляется during render. Evidence: `REC`; high; [S42](../sources/registry.md).

## Composition at real variation seam

**BAD**

```ts
class RetryableEmailSender extends SmtpSender {
  // inherits transport lifetime merely to alter retries
}
```

**GOOD**

```ts
class RetryingNotifier implements Notifier {
  constructor(
    private readonly inner: Notifier,
    private readonly retry: RetryPolicy,
  ) {}

  send(message: Message) {
    return this.retry.run(() => this.inner.send(message));
  }
}
```

Wrapper оправдан, когда retry и transport реально меняются независимо. Если SDK уже владеет correct retry semantics, wrapper может быть лишним.

## Semantic control

```tsx
<button type="button" onClick={cancelOrder}>
  Cancel order
</button>
```

Native semantics предпочтительнее recreated ARIA role. Evidence: `NORM`; [S45, S46](../sources/registry.md).
