FROM golang:latest as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app/

COPY --from=builder /app/main .

EXPOSE 4000

CMD ["./main"]
