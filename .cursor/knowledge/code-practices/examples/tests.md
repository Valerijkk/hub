# Test examples

## One behavioral concept, not one assertion

```python
response = client.confirm_order(order.id, actor=owner)

assert response.status_code == 201
assert response.json()["order_id"] == str(order.id)
assert repository.get(order.id).status == "confirmed"
```

Три assertions подтверждают один claim: successful confirmation returns and persists confirmed order. Artificial split может утроить setup без diagnostic gain. Evidence: `REC`; [S25, S26](../sources/registry.md).

## DAMP authorization scenarios

**BAD**

```python
@pytest.mark.parametrize("role", ["admin", "viewer"])
def test_access(role, client, user_factory, order):
    expected = 200 if role == "admin" else 403
    assert client.get_order(order.id, user_factory(role)).status_code == expected
```

**GOOD**

```python
def test_admin_can_read_order(client, admin, order):
    assert client.get_order(order.id, admin).status_code == 200


def test_viewer_cannot_read_order(client, viewer, order):
    assert client.get_order(order.id, viewer).status_code == 403
```

Some duplication сохраняет различие authorization behaviors. Parametrize остаётся хорошим для одной behavior/equivalence family.

## State vs interaction

**BAD**

```python
service.create_order(command)
repo.get.assert_called_once()
repo.save.assert_called_once()
clock.now.assert_called_once()
```

**GOOD**

```python
created = service.create_order(command)
assert created.customer_id == command.customer_id
assert repo.load(created.id) == created
```

Interaction assertion допустим, если interaction — contract, например «one payment authorization with this idempotency key». Mocks не mandatory и не forbidden. Evidence: `REC`, `DISPUTED`; [S23](../sources/registry.md).

## UI abstraction ladder

**BAD** — giant `AppPage` знает весь продукт и содержит specific assertions.

**GOOD**

```ts
class OrderPage {
  constructor(private readonly page: Page) {}

  async cancel() {
    await this.page.getByRole("button", { name: "Cancel order" }).click();
  }

  status() {
    return this.page.getByTestId("order-status");
  }
}

await orderPage.cancel();
await expect(orderPage.status()).toHaveText("Cancelled");
```

Start with semantic locators. Add Component/Page Object only when repeated UI knowledge justifies maintenance layer. Evidence: `REC`; [S31, S32](../sources/registry.md).

## Async wait

**BAD:** `sleep(5)`.

**GOOD:** wait for observable semantic condition with bounded timeout. Built-in web-first assertions reduce the need for custom wait wrappers.

## Suite mix

Do not gate on 70/20/10, Trophy or Honeycomb. Choose the cheapest scope that detects the risk, then add real boundary/user-flow evidence where the cheap test cannot. Evidence: `REC`, `OPINION`; [S25, S28](../sources/registry.md).
