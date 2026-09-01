# notifier

The tail of the pipeline. Consumes `processed-orders` and records a customer notification for each.

## Responsibility

- Consume `processed-orders`.
- Log a friendly notification and persist a `NotifiedMessage` to DynamoDB, stamped with the
  moment of notification (`time.Now()`).

## Stored record

`NotifiedMessage` keeps: order id, client id, order status, current delivery location, final
destination, and the notification timestamp. The table is keyed by `OrderId` (partition) +
`Timestamp` (sort), so a single order keeps a full notification history.

## Ports & adapters

- **Input port:** `ProcessedOrderReader` (`Read`/`Close`) — `KafkaReadProcessedOrders` (local).
- **Output port:** `NotifiedMessageRepository` (`Add`/`RunWorker`/`StopWorker`) —
  `DynamoNotifiedMessageRepository`, a buffered worker.

The notifier consumes the processor's `ProcessedOrderDTO` (shared within the module) as the wire
contract.

## Runtime

`APP_RUNTIME=local` (default) runs a Kafka consumer loop with graceful shutdown;
`APP_RUNTIME=lambda` runs an SNS handler that persists one notification per record and flushes the
repo before returning.

## Environment

| Variable | Mode | Purpose |
|---|---|---|
| `KAFKA_BROKER` | local | Kafka bootstrap address |
| `DYNAMODB_NOTIFIED_MESSAGES_TABLE_NAME` | both | notifications table |
| `APP_RUNTIME` | both | `local` (default) or `lambda` |
