FROM golang:latest

WORKDIR /postdog
COPY go.mod go.sum /postdog/
RUN go mod download
COPY . /postdog

CMD ["go", "test", "-v", "./..."]