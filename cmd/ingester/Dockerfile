FROM golang:alpine AS builder
ENV GOPROXY=https://goproxy.io,direct
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /main cmd/ingester/main.go

FROM alpine:latest
WORKDIR /
COPY --from=builder /main /main
EXPOSE 8080
ENTRYPOINT ["/main"]