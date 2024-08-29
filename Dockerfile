FROM golang:1.21

WORKDIR /app

RUN go install github.com/swaggo/swag/cmd/swag@v1.16.3

COPY go.mod go.sum ./

RUN go mod download

COPY . /app

RUN swag init -g api/v1/routers.go

RUN CGO_ENABLED=0 GOOS=linux go build -o /schedule-api

CMD ["/schedule-api"]



