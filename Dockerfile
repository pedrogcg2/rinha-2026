FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

FROM alpine:3.22

WORKDIR /app

RUN addgroup -S app && adduser -S app -G app

COPY --from=builder /app/server /app/server
COPY --from=builder /app/mcc_risk.json /app/mcc_risk.json
COPY --from=builder /app/normalization.json /app/normalization.json
COPY --from=builder /app/references.json /app/references.json

USER app

EXPOSE 9999

ENTRYPOINT ["/app/server"]
