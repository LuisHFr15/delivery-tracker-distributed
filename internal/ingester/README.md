# ingester

The entry point of the pipeline. Accepts HTTP requests, validates them, and publishes order and
location events to the messaging backend. It is write-only (publish-only) — it never reads events.

## Responsibility

- Expose the public HTTP API.
- Validate that events carry the identifiers the downstream services require.
- Publish `order-events` and `location-events`, responding `202 Accepted` (fire-and-forget).

## HTTP API

| Method | Path | Body | Response |
|---|---|---|---|
| `POST` | `/api/order` | `OrderEventDTO` | `202` |
| `POST` | `/api/order/{id}/location` | `LocationEventDTO` | `202` |
| `GET` | `/api/health` | — | `200` |

The order id for a location update comes from the JSON body (`order_id`), not the path segment.

## Ports & adapters

- **Port:** `EventPublisher` (`PublishOrder`, `PublishLocation`) in `app/services`.
- **Adapters:**
  - `queues.KafkaPublisher` — buffered async writer to Kafka (local mode).
  - `queues.SnsPublisher` — synchronous publish to SNS (Lambda mode).

## Runtime

`APP_RUNTIME=local` (default) runs a Gin HTTP server with graceful shutdown; `APP_RUNTIME=lambda`
runs an API Gateway (HTTP API) handler that reuses the same `IngesterService`.

## Environment

| Variable | Mode | Purpose |
|---|---|---|
| `KAFKA_BROKER` | local | Kafka bootstrap address (default `localhost:9092`) |
| `SNS_ORDER_TOPIC_ARN` | lambda | order-events topic ARN |
| `SNS_LOCATION_TOPIC_ARN` | lambda | location-events topic ARN |
| `APP_RUNTIME` | both | `local` (default) or `lambda` |
