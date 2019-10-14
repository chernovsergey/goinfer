how to generate *.pb.go? 
from goinfer folder!

1) look at 'go env' and find GOPATH
2) export GOPATH="/Users/sergey/go"
3) export PATH=$PATH:$GOPATH/bin
4) protoc -I ./api/ ./api/api.proto --go_out=plugins=grpc:api