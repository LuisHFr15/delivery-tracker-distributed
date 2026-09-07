# processor

The brain of the pipeline. Consumes order and location events, applies the domain rules, persists
state to DynamoDB, and publishes `processed-orders`.

## Responsibility

- **Order events:** convert to an audit record and persist the order.
- **Location events:** load the order, advance its delivery state, compute the ETA (haversine at a
  fixed cruising speed), persist audit + order + processed-order records, and publish a
  `ProcessedOrderDTO`.
- Reject events without an `event_id`.

## Domain layer

Pure, infrastructure-free (`domain/`): the `Order` aggregate and its delivery state machine,
`DeliveryCalculator` (ETA), `DeliveryUpdater`, and the event converters. This layer is unchanged
between local and Lambda runtimes.

## Ports & adapters

- **Input ports:** `OrderReader`, `LocationReader` (`Read`/`Close`) — used by the Kafka loop.
- **Output ports:**
  - `NotificationWriter` (`Write`/`RunWorker`/`StopWorker`) — `KafkaWriteNotification` or
    `SnsNotificationWriter`.
  - Repository ports (`AuditingEventRepository`, `OrderRepository`, `ProcessedOrderRepository`) —
    DynamoDB adapters using a buffered worker.

## Runtime

`APP_RUNTIME=local` (default) runs two Kafka consumer loops with graceful shutdown;
`APP_RUNTIME=lambda` runs an SNS handler that routes each record (by topic ARN) to the same app
services and flushes the buffered writers via `StopWorker` before returning.

## Events

- **In:** `order-events`, `location-events`
- **Out:** `processed-orders`

## Environment

| Variable | Mode | Purpose |
|---|---|---|
| `KAFKA_BROKER` | local | Kafka bootstrap address |
| `DYNAMODB_AUDITING_TABLE_ARN` | both | auditing table ARN (or name) |
| `DYNAMODB_ORDER_TABLE_ARN` | both | orders table ARN (or name) |
| `DYNAMODB_PROCESSED_ORDER_TABLE_ARN` | both | processed-orders table ARN (or name) |
| `SNS_ORDER_TOPIC_ARN` / `SNS_LOCATION_TOPIC_ARN` | lambda | topic routing |
| `SNS_PROCESSED_ORDERS_TOPIC_ARN` | lambda | processed-orders topic |
| `APP_RUNTIME` | both | `local` (default) or `lambda` |

## Tests

Unit tests cover the domain and application layers. Integration tests (DynamoDB via
`testcontainers-go`) are gated behind the `integration` build tag:

```bash
go test ./...                     # unit
go test -tags=integration ./...   # + integration (needs Docker)
```
