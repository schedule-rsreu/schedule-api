FROM golang:1.27-alpine3.24 AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
RUN GOBIN=/out go install github.com/swaggo/swag/cmd/swag@v1.16.6
RUN CGO_ENABLED=0 GOBIN=/out go install \
      -tags='no_clickhouse no_libsql no_mssql no_mysql no_sqlite3 no_vertica no_ydb' \
      github.com/pressly/goose/v3/cmd/goose@v3.29.0

COPY . .
RUN /out/swag init --parseDependency --parseInternal -g ./internal/http/handlers/router.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/main cmd/main.go

FROM alpine:3.24 AS certificates
RUN apk upgrade --no-cache && apk add --no-cache ca-certificates-bundle

FROM scratch AS runner
COPY --from=certificates /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
WORKDIR /app
COPY --from=builder /out/main /bin/main
COPY --from=builder /out/goose /usr/local/bin/goose
COPY migrations/postgres /app/migrations/postgres

ENTRYPOINT ["/bin/main"]
