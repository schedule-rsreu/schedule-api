FROM golang:1.26 AS builder

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download


RUN go install github.com/swaggo/swag/cmd/swag@v1.16.6

COPY . .

RUN swag init --parseDependency --parseInternal -g ./internal/http/handlers/router.go


RUN CGO_ENABLED=0 GOOS=linux go build -o /main cmd/main.go


FROM alpine AS runner

COPY --from=builder main /bin/main

ENTRYPOINT ["/bin/main"]
