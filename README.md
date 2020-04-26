# Distributed task executor for genomics research

*RNNR* is a bioinformatics task processing system for distributed computing environments. It was designed to work with a workflow management system, such as [Cromwell](https://github.com/broadinstitute/cromwell) and [Nextflow](https://www.nextflow.io/), through the [Task Execution Service (TES) API](https://github.com/ga4gh/task-execution-schemas). The workflow manager submits tasks to RNNR that distributes the processing load among computational nodes connected in a local network. The system is composed of the *master* instance and one or more *worker* instances. A distributed file system (NFS for example) is required.

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