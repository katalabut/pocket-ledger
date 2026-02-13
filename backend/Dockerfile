FROM golang:1.23-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /pocket-ledger ./cmd/api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /pocket-ledger /app/pocket-ledger
COPY migrations/ /app/migrations/
EXPOSE 8080
CMD ["/app/pocket-ledger"]
