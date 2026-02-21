FROM golang:1.24.2-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/server ./cmd/server

FROM alpine:3.22

WORKDIR /app
RUN adduser -D -H -u 10001 app

COPY --from=builder /bin/server /app/server

USER app
EXPOSE 8080

ENTRYPOINT ["/app/server"]
