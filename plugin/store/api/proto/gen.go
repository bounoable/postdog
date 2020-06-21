package proto

//go:generate protoc -I. -I ../../../../ --go_out=plugins=grpc,module=github.com/bounoable/postdog/plugin/store/api/proto:. letter.proto query.proto
