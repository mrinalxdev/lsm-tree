FROM golang:1.21-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git build-base
COPY go.mod go.sum ./
RUN go mod download


COPY . .


RUN CGO_ENABLED=1 GOOS=linux go build -o lsm-store ./cmd/server

FROM alpine:3.18

WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /app/lsm-store .
COPY --from=builder /app/web ./web
RUN mkdir -p /app/data

EXPOSE 8080

CMD ["./lsm-store", "-data-dir=/app/data"]
