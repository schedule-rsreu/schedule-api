FROM golang:1.27-alpine3.24 AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
RUN GOBIN=/out go install github.com/swaggo/swag/cmd/swag@v1.16.6
RUN GOBIN=/out go install github.com/pressly/goose/v3/cmd/goose@v3.27.3

COPY . .
RUN /out/swag init --parseDependency --parseInternal -g ./internal/http/handlers/router.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/main cmd/main.go

FROM alpine:3.24 AS runner
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /out/main /bin/main
COPY --from=builder /out/goose /usr/local/bin/goose
COPY migrations/postgres /app/migrations/postgres

ENTRYPOINT ["/bin/main"]
