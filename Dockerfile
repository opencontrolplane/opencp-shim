FROM golang:latest as builder

ARG ACCESS_TOKEN_USR
ARG ACCESS_TOKEN_PWD

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main .

FROM debian:stable-slim

WORKDIR /app/

COPY --from=builder /app/main .
COPY --from=builder /app/config.yaml .
ADD ssl ./ssl

EXPOSE 4000

CMD ["./main"]
