FROM golang:1.23.4 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o go-kweb-lang ./cmd

FROM alpine:latest

RUN apk update &&\
    apk add --no-cache git

WORKDIR /app

COPY --from=builder /app/go-kweb-lang .

EXPOSE 8080

CMD ["/app/go-kweb-lang"]
