FROM golang:1.20-alpine3.18 AS builder

WORKDIR /app

RUN apk add --update --no-cache make git

COPY go.mod vendor* ./
RUN [ ! -d "vendor" ] && go mod download all || echo "skipping..."

COPY . .

RUN make cmd/block_processor/block_processor

FROM alpine:3.18

COPY --from=builder /app/build/cmd/block_processor/block_processor /
COPY --from=builder /app/.env /

ENTRYPOINT ["/block_processor"]
