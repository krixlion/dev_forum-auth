# dev_forum-auth

The authorization and authentication service that dev_forum relies on to provide user identities and permissions.

## Set up

Before any further steps copy and rename `.env.example` to `.env`. \
Fill in the missing values as needed.

### Locally
You need a working [Go environment](https://go.dev/doc/install).

To build the executable simply download dependencies with and compile using Go command.

Run from project root:
```shell
go mod tidy 
go mod vendor

go build cmd/main.go
```
Make sure to place the executable in the same directory as the `.env` file.

### On Docker
You need a working [Docker environment](https://docs.docker.com/engine).

You can also use the Dockerfile to build and run the service on a docker container. \

```shell
make build-image version=latest` 
``` 

### On Kubernetes
You need a working [Kubernetes environment](https://kubernetes.io/docs/setup) with [kustomize](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization).

Use `make` to apply manifests for dev_forum-auth and needed DBs for either dev or stage environment.
```shell
make k8s-run version=<dev/stage>
```

## Testing
All tests are written as Go tests.

Run unit and integration tests using Go command.
```
go test ./... -race
```
Include `-short` flag to skip integration tests.
