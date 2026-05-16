FROM golang:1.26.3-alpine AS builder

WORKDIR /app/src

COPY src/go.mod src/go.sum ./
RUN go mod download

COPY src .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/server .

FROM alpine:3.22

WORKDIR /app/src

RUN addgroup -S app && adduser -S app -G app

COPY --from=builder /app/server /app/server
COPY resources /app/resources

USER app

EXPOSE 8080

ENTRYPOINT ["/app/server"]
