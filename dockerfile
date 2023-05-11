FROM golang:1.18-alpine

WORKDIR /go/src/app

COPY . .

RUN go mod download
# backend
RUN go build -o transbot main.go
# frontend
RUN go build -o transfrontend ./frontend/httpserver.go

EXPOSE 8080
EXPOSE 8081

# start up frontend
CMD ["./transfrontend"]
# startup the server
CMD ["./transbot"]
