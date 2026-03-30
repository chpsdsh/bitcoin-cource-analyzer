FROM golang:1.25

WORKDIR /app

COPY . .

RUN go build -o app ./cmd/newsgateway

CMD ["./app"]