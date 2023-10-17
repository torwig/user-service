FROM golang:1.21 AS builder

WORKDIR /app

COPY . .

RUN make build

FROM alpine:3.18

RUN adduser -D -H -u 10001 appuser
USER appuser

WORKDIR /app

COPY --from=builder /app/users.bin .

CMD [ "/app/users.bin" ]
