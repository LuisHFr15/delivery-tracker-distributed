FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /main cmd/processor/main.go

FROM alpine:latest
WORKDIR /
COPY --from=builder /main /main
ENTRYPOINT ["/main"]