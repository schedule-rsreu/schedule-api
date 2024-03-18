FROM golang:1.21

WORKDIR /app


COPY go.mod go.sum ./

RUN go mod download

COPY . /app

RUN CGO_ENABLED=0 GOOS=linux go build -o /schedule-api

CMD ["/schedule-api"]



