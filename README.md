# Neptune-简单快速的grpc & web框架

## Installation

Using go proxy
```shell
go env -w GOPROXY="https://goproxy.cn,direct"
```

Local development
```shell
go env -w  GOPRIVATE="gitlab.com.cn"  ##Optional command，for faster pull dependency locally
```

With [Go module][] support (Go 1.11+), simply add the following import
```shell
import "github.com/no-mole/neptune"
```

To your code, and then `go [build|run|test]` will automatically fetch the
necessary dependencies.

Otherwise, to install the `neptune` package or using `neptune` cli, run the following command:

```console
$ go get github.com/no-mole/neptune
$ neptune help
```

How to generate proto files in different languages
```console
neptune proto-gen -l golang -path /bar/bar.proto

-l Specifies the programming language, currently supported golang, java, python, php; Default golang
-path Specify the address of the proto file

```


## Features

Neptune abstracts away the details of distributed systems. Here are the main features.

- **Dynamic Config** - Load and hot reload dynamic config from anywhere. The config interface provides a way to load application
  level config from any source such as env vars, file, etcd. You can merge the sources and even define fallbacks.

- **Data Storage** - A simple data store interface to read, write and delete records.

- **Service Discovery** - Automatic service registration and name resolution. Service discovery is at the core of micro service
    development. When service A needs to speak to service B it needs the location of that service. The default discovery mechanism is
    multicast DNS (mdns), a zeroconf system.

- **Load Balancing** - Client side load balancing built on service discovery. Once we have the addresses of any number of instances
  of a service we now need a way to decide which node to route to. We use random hashed load balancing to provide even distribution
  across the services and retry a different node if there's a problem. 

- **RPC Client/Server** - RPC based request/response with support for bidirectional streaming. We provide an abstraction for synchronous
  communication. A request made to a service will be automatically resolved, load balanced, dialled and streamed.