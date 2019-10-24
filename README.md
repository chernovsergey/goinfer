# goinfer
Inference service 

# Dependencies
 - go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
 - go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
 - go get -u github.com/golang/protobuf/protoc-gen-go
 - go get -u github.com/chapsuk/wait
 - go get -u golang.org/x/sync

# How to generate *.pb.go? 

Enter *goinfer/api* folder!

1) look at 'go env' and find GOPATH
2) export GOPATH=<XXX> 
3) export PATH=$PATH:$GOPATH/bin
4) protoc -I/usr/local/include -I. -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --go_out=plugins=grpc:. api.proto
5) protoc -I/usr/local/include -I. -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --grpc-gateway_out=logtostderr=true:. api.proto


# How to profile performance

1) go tool pprof localhost:8080/debug/pprof/profile?seconds=<NUM_SECONDS>
2) go tool pprof -http=:9090 /path/to/profile/pprof.pb.gz