FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .

RUN go build -o main .

FROM alpine:latest

RUN adduser -D appuser \
    && mkdir /app \
    && chown -R appuser:appuser /app

WORKDIR /

COPY --from=builder /app/main /main

USER appuser

ENTRYPOINT ["/main"]
