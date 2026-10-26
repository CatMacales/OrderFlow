# Order acceptance examples

[English](order-acceptance.md) | [Русский](order-acceptance.ru.md)

Acceptance criteria accepted by the user on 2026-10-27. These specify future tests, not executed test results.

Rules: [requirements](order-requirements.md). Architecture: [ADR 001](adr/001-monorepo.md).

## Creation and retrieval

Given customer_id `customer-42` and two lines: product_id `SKU-MUG`, quantity 2; product_id `catalog:notebook:7`, quantity 1.

When creating an order, return a new UUID order_id. Retrieving that ID returns the same customer_id, both product lines with their original quantities, and created state. No external service calls are required.

## Transitions

Each cell is a separate example. Given an existing order in the row state, when executing the column action, expect the specified result and verify state through a separate retrieval. Customer and lines remain unchanged.

| Initial state | Store confirmation | Complete | Cancel |
| --- | --- | --- | --- |
| created | Success, confirmed | INVALID_ORDER_STATE, created | Success, cancelled |
| confirmed | Success, confirmed | Success, completed | Success, cancelled |
| completed | Success, completed | Success, completed | INVALID_ORDER_STATE, completed |
| cancelled | INVALID_ORDER_STATE, cancelled | INVALID_ORDER_STATE, cancelled | Success, cancelled |

Retrieval in each of the four states returns the order unchanged. Successful state actions return no order body.

## Invalid input

Check each case separately; expect INVALID_ARGUMENT and no newly stored order:

- Empty line list.
- quantity = 0.
- quantity = -1.
- Two lines with the same product_id, even with different quantities: no automatic merging.

For get, confirm, complete, and cancel with malformed order_id, expect INVALID_ARGUMENT; existing orders remain unchanged.

## Missing order

Given a valid UUID absent from storage, get, confirm, complete, and cancel each return ORDER_NOT_FOUND.

## Storage lifetime

First-version storage is in memory and does not survive restart. Persistent storage and restart durability checks belong to the PostgreSQL stage.

## Agreed verification scope

Lines contain only product_id and quantity, without names. Authentication and authorization are outside the first version; business roles do not imply permission enforcement. No arbitrary business limits are introduced. Numeric representation belongs to implementation; request-size bounds to the API stage. Persistent storage, concurrent update checks, and keyed creation idempotency belong to later stages.

## Order identifiers

Creation with customer-42, SKU-MUG, and catalog:notebook:7 is valid: external IDs need not be UUIDs. Generated order_id is UUID v4. Get, confirm, complete, and cancel with empty order_id or nil UUID return INVALID_ARGUMENT without changing orders.

## External string-ID checks

- Empty customer_id: INVALID_ARGUMENT, no order created.
- Empty product_id in any line: INVALID_ARGUMENT, no order created.
- After creation and retrieval, nonempty external IDs exactly match input, including case and whitespace.
- An exactly repeated product_id across two lines: INVALID_ARGUMENT. Distinct strings are not merged through normalization.
