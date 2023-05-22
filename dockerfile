FROM golang:1.18-alpine

WORKDIR /go/src/app

RUN apk add --no-cache bash

COPY . .

RUN go mod download
# backend
RUN go build -o transbot main.go

EXPOSE 8080

# start up frontend
CMD ["./transbot"]

