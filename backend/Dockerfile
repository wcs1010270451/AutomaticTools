FROM registry-1.docker.io/library/golang:1.24-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -buildvcs=false \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/automatictools-api \
    ./cmd/api

FROM registry-1.docker.io/library/alpine:3.22

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S automatictools \
    && adduser -S -G automatictools automatictools

WORKDIR /app
COPY --from=builder /out/automatictools-api /app/automatictools-api

USER automatictools
EXPOSE 8088

ENTRYPOINT ["/app/automatictools-api"]
