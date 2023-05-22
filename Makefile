.PHONY: integration-test
integration-test:
	@echo "Running integration tests..."
	@go test -v -tags=integration ./...

.PHONY: test
test:
	@echo "Running unit tests..."
	@go test -v ./...

MICROSERVICES= \
	cmd/block_processor/block_processor \
	cmd/tx_processor/tx_processor \
	cmd/scanner/scanner \
	cmd/validator/validator \
	cmd/api/api

.PHONY: $(MICROSERVICES)

cmd/block_processor/block_processor:
	@echo "Building block_processor..."
	@go build -o build/$@ ./cmd/block_processor

cmd/tx_processor/tx_processor:
	@echo "Building tx_processor..."
	@go build -o build/$@ ./cmd/tx_processor

cmd/scanner/scanner:
	@echo "Building scanner..."
	@go build -o build/$@ ./cmd/scanner

cmd/validator/validator:
	@echo "Building validator..."
	@go build -o build/$@ ./cmd/validator

cmd/api/api:
	@echo "Building api..."
	@go build -o build/$@ ./cmd/api
