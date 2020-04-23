# Distributed task executor for genomics research

## Getting started

Run RNNR master server. Requires MongoDB server.

```bash
rnnr master
```

Endpoint is http://localhost:8080/tasks

Run RNNR worker server. Requires Docker server.

```bash
rnnr worker
```

Add worker node.

```bash
rnnr add localhost
```

Run Cromwell with RNNR.
Cromwell utilizes TES backend to submit jobs.
For more information see <https://cromwell.readthedocs.io/en/stable/backends/TES/>.
Get latest Cromwell at <https://github.com/broadinstitute/cromwell/releases>.

```
java -Dconfig.file=examples/cromwell.conf -jar cromwell-49.jar run examples/hello.wdl
```

## Development

Generate Go code from ProtoBuffer file

```bash
go get -u google.golang.org/grpc
go get -u github.com/golang/protobuf/protoc-gen-go
protoc -I pb --go_out=plugins=grpc:pb pb/worker.proto 
```

Build Docker image and publish to Docker Hub

```bash
docker build -t welliton/rnnr:latest .
docker push welliton/rnnr:latest
```

## Internals

[Canonical error codes](https://pkg.go.dev/google.golang.org/grpc/codes?tab=doc) are used to differentiate gRPC network communication error from other errors.
**Unavailable (14)** always return a `NetworkError`.