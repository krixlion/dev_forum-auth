# dev_forum-auth

The authorization and authentication service that dev_forum relies on to provide user identities and permissions.

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
make build-image version=latest` 
``` 

```shell
docker run -p 50051:50051 -p 2223:2223 krixlion/dev_forum-auth:0.1.0
```

### On Kubernetes (recommended)
You need a working [Kubernetes environment](https://kubernetes.io/docs/setup) with [kustomize](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization).

Use `make` to apply manifests for dev_forum-auth and needed DBs for either dev or stage environment.
```shell
make k8s-run overlay=<dev/stage>
```

## Testing
All tests are written as Go tests.

Run unit and integration tests using Go command.
```
go test ./... -race
```
Include `-short` flag to skip integration tests.
