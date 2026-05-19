FROM golang:1.26.3-alpine AS builder

WORKDIR /app/src

COPY src/go.mod src/go.sum ./
RUN go mod download

COPY src .
COPY resources /app/resources

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/preprocessor ./preprocessor
WORKDIR /app/src/preprocessor
RUN /app/preprocessor

WORKDIR /app/src
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/server .

FROM alpine:3.22

WORKDIR /app/src

RUN addgroup -S app && adduser -S app -G app


COPY --from=builder /app/server /app/server
COPY --from=builder /app/resources/idx.bin /app/resources/idx.bin
COPY --from=builder /app/resources/normalization.json /app/resources/normalization.json
COPY --from=builder /app/resources/mcc_risk.json /app/resources/mcc_risk.json

USER app

EXPOSE 8080

ENTRYPOINT ["/app/server"]
