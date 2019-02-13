# Ams Push Server
The ams push server is a component that handles the push functionality for the respective 
push enabled subscriptions in the [argo-messaging-service.](https://github.com/ARGOeu/argo-messaging-service)

## Perquisites

Before you start, you need to issue a valid certificate.

## Set Up

1. Install Golang 1.10
2. Create a new work space:

      `mkdir ~/go-workspace`

      `export GOPATH=~/go-workspace`

      `export PATH=$PATH:$GOPATH/bin`

     You may add the `export` lines into the `~/.bashrc`, `/.zshrc` or the `~/.bash_profile` file to have `GOPATH` environment variable properly setup upon every login.

3. Get the latest version

      `go get github.com/ARGOeu/ams-push-server`

4. Get dependencies(If you plan on contributing to the project else skip this step):

   Ams-push-server uses the dep tool for dependency handling.

    - Install the dep tool. You can find instructions depending on your platform at [Dep](https://github.com/golang/dep).

5. To build the service use the following command:

      `go build`

6. To run the service use the following command:

      `./ams-push-server` (This assumes that there is a valid configuration file at
       `/etc/ams-push-server/conf.d/ams-push-server-config.json`).

      Else

      `./ams-push-server --config /path/to/a/json/config/file`

7. To run the unit-tests:

    Inside the project's folder issue the command:

      `go test $(go list ./... | grep -v /vendor/)`

 ## Configuration
    
 The service depends on a configuration file in order to be able to run.This file contains the following information:

 ```json
{
  "service_port": 9000,
  "certificate": "/path/cert.pem",
  "certificate_key": "/path/certkey.pem",
  "certificate_authorities_dir": "/path/to/cas",
  "ams_token": "sometoken",
  "ams_host": "localhost",
  "ams_port": 8080,
  "verify_ssl": true,
  "tls_enabled": true,
  "trust_unknown_cas": false 
}
 ```
You can find the configuration template at `conf/ams-push-server-config.template`.
## Managing the protocol buffers and gRPC definitions

In order to modify any `.proto` file you will need the following

 - Read on how to install the protoc compiler on your platform [here.](https://github.com/protocolbuffers/protobuf)
 
 -  Install the go plugin. `go get -u github.com/golang/protobuf/protoc-gen-go`
 
 - install the go gRPC package. `go get -u google.golang.org/grpc`
 
 - Inside `api/<version>/grpc` compile. `protoc -I proto/ proto/ams.proto --go_out=plugins=grpc:proto`

