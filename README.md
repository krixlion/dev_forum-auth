# Status
🚧 **Under Development** 🚧

This repository is a part of an ongoing project and is currently under active development. I'm continuously working on adding features, fixing bugs, and improving documentation. 
Although this is a one-man project, contributions are welcome.
Please feel free to open issues or submit pull requests.

# dev_forum-auth
[![GoDoc](https://godoc.org/github.com/krixlion/dev_forum-auth?status.svg)](https://godoc.org/github.com/krixlion/dev_forum-auth)
[![Coverage Status](https://coveralls.io/repos/github/krixlion/dev_forum-user/badge.svg?branch=dev)](https://coveralls.io/github/krixlion/dev_forum-user?branch=dev)
[![Go Report Card](https://goreportcard.com/badge/github.com/krixlion/dev_forum-auth)](https://goreportcard.com/report/github.com/krixlion/dev_forum-auth)
[![GitHub License](https://img.shields.io/github/license/krixlion/dev_forum-auth)](LICENSE)

The authorization and authentication service that dev_forum relies on to provide user identities and authorization mechanisms.

It's dependent on:
  - [HashiCorp Vault](https://developer.hashicorp.com/vault/docs?product_intent=vault) for private key storage,
  - [MongoDB](https://www.mongodb.com/docs/manual/introduction/) for token storage,
  - [RabbitMQ](https://www.rabbitmq.com/#getstarted) for asynchronous communication with other components in the domain,
  - [OtelCollector](https://opentelemetry.io/docs/collector) for receiving and forwarding telemetry data.

## Set up
Rename `.env.example` to `.env` and fill in missing values.

### Locally
You need a working [Go environment](https://go.dev/doc/install).

To build the executable simply download dependencies with and compile using the Go command.

```shell
go mod tidy
go mod vendor
go build cmd/main.go 
```

### On Docker
You need a working [Docker environment](https://docs.docker.com/engine).

You can use the Dockerfile located in `deployment/` to build and run the service on a docker container.

```shell
make build-image version=<version>
``` 

```shell
docker run -p 50051:50051 -p 2223:2223 krixlion/dev_forum-auth:<version>
```

### On Kubernetes (recommended)
You need a working [Kubernetes environment](https://kubernetes.io/docs/setup) with [kustomize](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization).

Kubernetes resources are defined in `deployment/k8s` and deployed using [Kustomize](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/).
Currently there are `stage`, `dev` and `test` overlays available and include any needed resources and configs.

Use `make` to apply manifests for dev_forum-auth and needed DBs for either dev or stage environment.
Every `make` rule that depends on k8s accepts an `overlay` param which indicates the namespace for the rule.
```shell
make k8s-run overlay=<dev/stage/...>
```
```shell
# To delete
make k8s-stop overlay=<dev/stage/...>
```

## Testing
All tests are written as Go tests.

Run unit and integration tests using Go command.
```shell
# Include `-short` flag to skip integration tests.
go test ./... -race
```

Generate coverage report using `go tool cover`.
```
go test -coverprofile  cover.out ./...
go tool cover -html cover.out -o cover.html
```

If the service is deployed on kubernetes you can use `make`.
```shell
make k8s-integration-test overlay=<dev/stage/...>
```
or
```shell
make k8s-unit-test overlay=<dev/stage/...>
```

## Documentation 
For in-detail documentation refer to the [Wiki](https://github.com/krixlion/dev_forum-auth/wiki).

## API
Service is exposing a [gRPC](https://grpc.io/docs/what-is-grpc/introduction) API.

Regenerate `pb` packages after making changes to any of the `.proto` files located in `api/`.
You can use [go-grpc-gen](https://github.com/krixlion/go-grpc-gen), a containerized tool for generating gRPC bindings, with `make grpc-gen`.
