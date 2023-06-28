FROM golang:1.20.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build proximity/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .

CMD ["./main"]