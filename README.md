# goinfer
Inference service 

# Package dependencies
 - go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
 - go get -u github.com/grpc-ecosystem/go-grpc-prometheus
 - go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
 - go get -u github.com/golang/protobuf/protoc-gen-go
 - go get -u github.com/chapsuk/wait
 - go get -u golang.org/x/sync

# Required tools
 - Prometheus
 - Grafana

# How to generate *.pb.go? 

Enter *goinfer/api* folder!

1) `export GOPATH=$(go env GOPATH)`
2) `export PATH=$PATH:$GOPATH/bin`
3) `protoc -I/usr/local/include -I. -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --go_out=plugins=grpc:. api.proto`
4) `protoc -I/usr/local/include -I. -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --grpc-gateway_out=logtostderr=true:. api.proto`


# How to profile performance

 - `go tool pprof localhost:8080/debug/pprof/profile?seconds=<NUM_SECONDS>`
 - `go tool pprof -http=:9090 /path/to/profile/pprof.pb.gz`

# How to configure monitoring
 - Run prometheus server `prometheus --config.file ./config/prometheus.yaml`
 - Run grafana (default address `localhost:3000`)
 - Enter grafana, create prometheus date source, create dashboard

# TODO
 - logging
 - dockerization
