package storeproto

//go:generate protoc -I. -I ../../../../ --go_out=plugins=grpc,module=github.com/bounoable/postdog/plugin/store/api/storeproto:. storeletter.proto query.proto
