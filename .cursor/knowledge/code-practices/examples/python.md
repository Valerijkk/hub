# Python examples

## Consumer capability, not interface quota

- Принцип: policy зависит от минимального capability.
- Evidence: `NORM`, `REC`; [S35](../sources/registry.md).
- Режим: `REVIEW`; Protocol оправдан boundary/несколькими implementations/testing seam.

**BAD** — скрытая dependency:

```python
class InvoiceService:
    def send(self, invoice: Invoice) -> None:
        smtp = GlobalContainer.smtp()
        smtp.send(invoice.render())
```

**GOOD** — capability видим:

```python
class Mailer(Protocol):
    def send(self, message: Message) -> None: ...


class InvoiceService:
    def __init__(self, mailer: Mailer) -> None:
        self._mailer = mailer
```

Не создавайте `Protocol` для каждого trivial helper с одной очевидной implementation.

## Immutable value, not universal immutability

```python
@dataclass(frozen=True)
class Money:
    currency: str
    minor_units: int
```

`frozen=True` даёт shallow value semantics, но не deep immutability и не domain validation. Evidence: `NORM`; [S36](../sources/registry.md).

## Classified retry

**BAD**

```python
while True:
    try:
        return gateway.charge(amount)
    except Exception:
        time.sleep(1)
```

**GOOD**

```python
for delay in retry_policy.delays(deadline):
    try:
        return gateway.charge(amount, idempotency_key=command.id)
    except PaymentUnavailable:
        sleep_with_jitter(delay)

raise PaymentTimeout(command.id)
```

Validation/decline не retry; deadline конечен; idempotency boundary известна. Evidence: `CONTROL`, `REC`; [S49, S50](../sources/registry.md).

## Authorization boundary

```python
def get_invoice(invoice_id: InvoiceId, actor: Principal) -> Invoice:
    invoice = invoices.require(invoice_id)
    if not policy.can_read_invoice(actor, invoice):
        raise Forbidden()
    return invoice
```

BAD-вариант доверяет UI или `tenant_id` request-а. Режим: `HARD`; [S56–S58](../sources/registry.md).
