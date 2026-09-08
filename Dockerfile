# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder
WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates \
	&& adduser -D -H -u 10001 appuser

WORKDIR /app
COPY --from=builder /out/server /app/server

USER appuser

# Runtime configuration is provided via environment variables.
# Do not bake secrets or .env files into the image.
EXPOSE 8080
ENTRYPOINT ["/app/server"]
CMD ["-serve", "-http-addr", ":8080"]
