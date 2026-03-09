FROM docker.io/golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /app/bot ./cmd/bot/main.go

FROM docker.io/alpine:latest
WORKDIR /root/
COPY --from=builder /app/bot .

CMD ["./bot"]
