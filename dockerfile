FROM golang:1.18-alpine

WORKDIR /go/src/app

RUN apk add --no-cache bash

COPY . .

RUN go mod download
# backend
RUN go build -o transbot main.go
# frontend
RUN go build -o transfrontend ./frontend/httpserver.go
RUN cp ./frontend/index.html ./

EXPOSE 8080
EXPOSE 8081

# start up frontend
CMD ["/bin/sh", "-c", "./run.sh"]

