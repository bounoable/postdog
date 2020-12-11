FROM golang:alpine

WORKDIR /postdog
COPY go.mod go.sum /postdog/
RUN go mod download
COPY . /postdog

CMD CGO_ENABLED=0 go test -v ./...
