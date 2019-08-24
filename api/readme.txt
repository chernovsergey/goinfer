how to generate *.pb.go?
    - protoc -I api/ api/api.proto --go_out=plugins=grpc:api