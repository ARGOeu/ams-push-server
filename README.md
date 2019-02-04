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
  "trust_unknown_cas": false,
  "log_level": "INFO",
  "skip_subs_load": false
}
 ```
 - `service_port:` The port that the service will bind to.  
 
 - `certificate:` The certificate file which the service will use.
 
 - `certificate_key:` The key to the respective certificate
 
 - `certificate_authroties_dir:` Directory containing `.pem` files that the service will use in order to build the trusted CA pool.
 The CA pool will be used to validate the certificates from incoming client requests.
 
 - `ams_token:` THe argo messaging token that the service will use in order to communicate with ams.`NOTE` that the
 token `MUST` correspond to a push worker user in ams.
 
 - `ams_host:` The ams http endpoint.
 
 - `ams_port:` The ams http endpoint port.
 
 - `verify_ssl:` Whether or not it should verify the various `https` endpoint it targets.
 
 - `tls_enabled:` Enable or disable tls support between client and server
 
 - `trust_unknown_cas:` Whether or not the service should accept certificates from CAs not found in its trusted CA pool.
 (Mainly used for development purposes)
 
 - `log_level:` DEBUG,INFO,WARNING,ERROR
 
 - `skip_subs_load:`  The service will try by default to contact the ams in order to retrieve all active push subscriptions   
 that are assigned to it and start their push cycles`(consume->send->ack)`. This will be done through the its user profile in ams(which is the profile associated with the)
 `ams_token`). You can control this behavior and decide whether or not to pre-load any already active subscriptions.
  
You can find the configuration template at `conf/ams-push-server-config.template`.
## Managing the protocol buffers and gRPC definitions

In order to modify any `.proto` file you will need the following

 - Read on how to install the protoc compiler on your platform [here.](https://github.com/protocolbuffers/protobuf)

 -  Install the go plugin. `go get -u github.com/golang/protobuf/protoc-gen-go`

 - install the go gRPC package. `go get -u google.golang.org/grpc`

 - Inside `api/<version>/grpc` compile. `protoc -I proto/ proto/ams.proto --go_out=plugins=grpc:proto`
