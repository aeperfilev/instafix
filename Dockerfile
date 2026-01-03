# Stage 1: Build
FROM golang:1.25-alpine AS builder
WORKDIR /app

ENV GO111MODULE=on
ENV GOWORK=off

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o instafix-server ./cmd/service

FROM alpine:3.18
WORKDIR /app

# Copy binary, config, and assets for runtime.
COPY --from=builder /app/instafix-server .
COPY --from=builder /app/config ./config
COPY --from=builder /app/assets ./assets

EXPOSE 8080
CMD ["./instafix-server"]
