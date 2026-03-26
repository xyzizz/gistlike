FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/gistlike ./cmd/server

FROM alpine:3.21

WORKDIR /app

RUN adduser -D -h /app appuser

COPY --from=builder /out/gistlike /app/gistlike
COPY web /app/web
COPY migrations /app/migrations

RUN mkdir -p /app/data && chown -R appuser:appuser /app

USER appuser

ENV APP_ADDR=:8080
ENV APP_DB_PATH=data/snippets.db
ENV APP_NAME=GistLike

EXPOSE 8080

CMD ["/app/gistlike"]
