# RNNR: distributed task execution system for scaling reproducible workflows

*RNNR* is a bioinformatics task processing system for distributed computing environments.
It was designed to work with a workflow management systems, such as [Cromwell](https://github.com/broadinstitute/cromwell) and [Nextflow](https://www.nextflow.io/), through the [Task Execution Service (TES) API](https://github.com/ga4gh/task-execution-schemas).
The workflow manager submits tasks to RNNR that distributes the processing load among computational nodes connected in a local network.
The system is composed of the *Main* instance and one or more *Worker* instances.
A distributed file system (NFS for example) is required.

> Source code is available at <https://github.com/labbcb/rnnr>

## Getting started

Run RNNR main server.
Requires MongoDB server.
The main server and worker instances do not manage any workflow or task files.

```bash
rnnr main
```

RNNR main server endpoint is <http://localhost:8080/v1/tasks>.

Run RNNR worker server. Requires Docker server.

```bash
rnnr worker
```

Add worker node.

```bash
rnnr add localhost
```

## Run Cromwell

Cromwell utilizes TES backend to submit jobs.
For more information see <https://cromwell.readthedocs.io/en/stable/backends/TES/>.
Get latest Cromwell at <https://github.com/broadinstitute/cromwell/releases>.

```bash
java -Dconfig.file=examples/cromwell.conf -jar cromwell-48.jar run examples/hello.wdl
```

Cromwell server is also supported and it is the recommended mode in production settings.

```bash
java -Dconfig.file=examples/cromwell.conf -jar cromwell-48.jar server
``` 

Cromwell server endpoint is <http://localhost:8000>

## Run Nextflow

[Nextflow](https://www.nextflow.io/) is a domain specific language and workflow execution system.
This system supports [GA4GH TES API](https://www.nextflow.io/docs/latest/executor.html#ga4gh-tes).
The following code runs a very simple workflow using Nextflow and RNNR.

```bash
export NXF_WORK=/home/nfs/tmp/nextflow-executions
export NXF_MODE=ga4gh
export NXF_EXECUTOR=tes
export NXF_EXECUTOR_TES_ENDPOINT='http://localhost:8080'
nextflow run examples/tutorial.nf
```

## Deployment

Cromwell and RNNR server instances can be deployed using Docker and Docker Swarm.
In this session we show how we use these services at BCBLab (<https://bcblab.org>).

At BCBLab we have 5 computing server and 1 NFS server in a local network.

NFS server (hostname `nfs`) has `/home/nfs` directory configured as NFS).
All computing server mounts NFS directory using same path (`/home/nfs`).
This is very important for Cromwell to create hard links and generate task command. 

All computing servers have Docker installed.
One compute server runs Cromwell and RNNR main server instances including databases (hostname `main`).
This server will not execute tasks because Cromwell instance uses lot of memory when executing complex workflows.
The other 4 compute servers run RNNR worker server instances (hostname `worker1`, `worker2`, `worker3`, `worker4`).

First we have to deploy RNNR main server.

```bash
docker network create rnnr
docker volume create rnnr-data

docker container run \
    --rm \
    --detach \
    --name rnnr-db \
    --network rnnr \
    --volume rnnr-data:/data/db \
    mongo:4

docker container run \
    --detach \
    --name rnnr \
    --publish 8080:8080 \
    --network rnnr \
    welliton/rnnr:latest \
    main --database mongodb://rnnr-db:27017
```

> It creates `rnnr` network for communication between RNNR and MongoDB;
> and `rnnr-data` volume to store database files.
> It starts `rnnr-db` container mounting the previous volume;
> and `rnnr` container exposing port `8080`.

At this point we need a Cromwell configuration file.
See [examples/cromwell-docker.yml](examples/cromwell-docker.conf) for example.
It set Cromwell workflow logs to `/home/nfs/tmp/cromwell-workflow-logs` and workflow root to `/home/nfs/tmp/cromwell-executions`.
Also set URL of MySQL to `jdbc:mysql://cromwell-db/cromwell?rewriteBatchedStatements=true`, where `cromwell-db` is the name of MySQL container.
The actor factory `cromwell.backend.impl.tes.TesBackendLifecycleActorFactory` tells Cromwell to use TES backend.
Endpoint `http://main:8080/v1/tasks` is the RNNR main endpoint, where `main` is the hostname of the server that runs RNNR main. 
We copy this file to `/etc/crowmell.conf`.

Next deploy Cromwell in server mode.

```bash
docker network create cromwell
docker volume create cromwell-data

docker container run \
    --rm \
    --detach \
    --name cromwell-db \
    --network cromwell \
    --volume cromwell-data:/var/lib/mysql \
    -e MYSQL_DATABASE=cromwell \
    -e MYSQL_ROOT_PASSWORD=secret \
    mysql:5.7

docker run \
    --name cromwell \
    --detach \
    --network cromwell \
    --publish 8000:8000 \
    --volume /etc/cromwell.conf:/application.conf \
    --volume /home/nfs:/home/nfs \
    -e JAVA_OPTS=-Dconfig.file=/application.conf \
    broadinstitute/cromwell:48 server
```

> Create `cromwell` network and `cromwell-data` volume.
> Create `cromwell-db` container setting `MYSQL_DATABASE` and `MYSQL_ROOT_PASSWORD` according to `/etc/cromwell.conf` file.
> Create `cromwell` container exposing port `8000`; mounting `/etc/cromwell.conf` file as `/application.conf` inside container;
> and mounting `/home/nfs` NFS directory as **same path** inside container.
> We have tested Cromwell release version 48.

You may notice that these service don't require Docker at all.
Cromwell delegates tasks to RNNR main which remotely runs them at active worker nodes.
However, deploying these services as Docker container simplifies system management.

Now, at each computing node, start a RNNR worker instance. **Docker is a requirement.**
We don't want to start task-related containers *inside* RNNR worker container (will not work anyway).
To solve this issue we have to mount Docker socket.

```bash
docker container run \
    --detach \
    --name rnnr \
    --publish 50051:50051 \
    --volume /var/run/docker.sock:/var/run/docker.sock \
    welliton/rnnr:latest \
    worker
```

> Create `rnnr` container exposing port `50051` and mounting Docker socket (`/var/run/docker.sock`).

Finally, we add the worker nodes to main server.
For each server, we set -2 CPU cores less memory as maximum computing resources.
Since RNNR is not aware of external process, this avoids any over consumption.  

```bash
rnnr enable --host main worker1 --cpu 14 --ram 180
rnnr enable --host main worker2 --cpu 14 --ram 130
rnnr enable --host main worker3 --cpu 48 --ram 120
rnnr enable --host main worker4 --cpu 48 --ram 70
```

Done. Cromwell will be available to run submitted workflows.

```bash
java -jar cromwell-48.jar submit --host http://main:8000 examples/hello.wdl
```

## Command line

Export all tasks as JSON.

```bash
rnnr ls --all --format json > rnnr_tasks.json
```

Export worker nodes as JSON.

```bash
rnnr nodes --format json > rnnr_nodes.json
```

## Development

Direct dependencies

- [cobra](https://github.com/spf13/cobra) for command line
- [viper](https://github.com/spf13/viper) for configuration files
- [logrus](https://github.com/sirupsen/logrus) for logging
- [mongo](https://github.com/mongodb/mongo-go-driver) for database
- [mux](https://github.com/gorilla/mux) for routing requests
- [uuid](https://github.com/google/uuid) for unique id generation
- [docker](https://pkg.go.dev/github.com/docker/docker/client) for container management
- [grpc](https://pkg.go.dev/mod/google.golang.org/grpc) for main-worker communication

Generate Go code from ProtoBuffer file

```bash
export GO111MODULE=on
go get google.golang.org/protobuf/cmd/protoc-gen-go \
    google.golang.org/grpc/cmd/protoc-gen-go-grpc

protoc --go_out=. \
    --go_opt=paths=source_relative \
    --go-grpc_out=. \
    --go-grpc_opt=paths=source_relative \
    proto/worker.proto
```

Build Docker image and publish to Docker Hub

```bash
docker build -t welliton/rnnr:latest .
docker push welliton/rnnr:latest
```

### Development environment

Start MongoDB inside container exposing 27017 TCP port

```bash
docker container run --rm --publish 27017:27017 mongo:4
```

## Internals

[Canonical error codes](https://pkg.go.dev/google.golang.org/grpc/codes?tab=doc) are used to differentiate gRPC network communication error from other errors.
**Unavailable (14)** always return a `NetworkError`.
