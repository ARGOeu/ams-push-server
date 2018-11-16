# Ams Push Server
The ams push server is a component that handles the push functionality for the respective 
push enabled subscriptions in the [argo-messaging-service.](https://github.com/ARGOeu/argo-messaging-service)

## Managing the protocol buffers and gRPC definitions

In order to modify the any `.proto` file you will need the following

 - Read on how to install the protoc compiler on your platform [here.](https://github.com/protocolbuffers/protobuf)
 
 -  Install the go plugin. `go get -u github.com/golang/protobuf/protoc-gen-go`
 
 - install the go gRPC package. `go get -u google.golang.org/grpc`
 
 - Inside `api/<version>/grpc` compile. `protoc -I proto/ proto/ams.proto --go_out=plugins=grpc:proto`

