FROM golang:1.25.10-alpine AS build

RUN apk add --no-cache build-base

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o /out/finance-tracker ./cmd/server

FROM alpine:3.22

RUN apk add --no-cache ca-certificates \
    && adduser -D -u 10001 app \
    && mkdir -p /data \
    && chown app:app /data

USER app
WORKDIR /app

COPY --from=build /out/finance-tracker /app/finance-tracker

ENV ADDR=:8080
ENV DATABASE_URL=file:/data/finance-tracker.db?_foreign_keys=on&_busy_timeout=5000

EXPOSE 8080

CMD ["/app/finance-tracker"]
