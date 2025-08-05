FROM golang:1.24.5 AS builder

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download


RUN go install github.com/swaggo/swag/cmd/swag@v1.16.3

COPY . .

RUN swag init --parseDependency --parseInternal -g ./internal/http/handlers/router.go


RUN CGO_ENABLED=0 GOOS=linux go build -o /main cmd/main.go


FROM alpine AS runner

COPY --from=builder main /bin/main

ENTRYPOINT ["/bin/main"]
