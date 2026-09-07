.PHONY: build test vet fmt compose-up compose-down deploy deploy-ecr deploy-ecr-lambda

# ---------------------------------------------------------------------------
# Local development
# ---------------------------------------------------------------------------
build:
	go build ./...

test:
	go test ./...

test-integration:
	go test -tags=integration ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

compose-up:
	docker compose up --build

compose-down:
	docker compose down

# ---------------------------------------------------------------------------
# Serverless deploy (AWS SAM)
# ---------------------------------------------------------------------------
deploy:
	sam build && sam deploy --guided

# SAM invokes one target per function (BuildMethod: makefile). Each compiles a
# static linux/amd64 `bootstrap` binary into the artifacts directory SAM hands us.
GO_LAMBDA_BUILD = GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags lambda.norpc -o $(ARTIFACTS_DIR)/bootstrap

build-IngesterFunction:
	$(GO_LAMBDA_BUILD) ./cmd/ingester

build-ProcessorFunction:
	$(GO_LAMBDA_BUILD) ./cmd/processor

build-NotifierFunction:
	$(GO_LAMBDA_BUILD) ./cmd/notifier

# ---------------------------------------------------------------------------
# ECR & Lambda Container Deployment
# ---------------------------------------------------------------------------
deploy-ecr:
	./scripts/deploy-ecr.sh $(SERVICE) $(TAG)

deploy-ecr-lambda:
	./scripts/deploy-ecr.sh $(or $(SERVICE),all) $(TAG) -l
