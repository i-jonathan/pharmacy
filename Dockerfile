FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main .

FROM alpine:latest

RUN adduser -D appuser

COPY --from=builder /app/main /main
USER appuser

ENTRYPOINT [ "/main" ]