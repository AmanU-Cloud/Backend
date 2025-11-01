FROM golang:1.25-alpine AS builder

LABEL authors="whoami"

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY  . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /caritas ./cmd/reviewer

CMD ["/caritas"]