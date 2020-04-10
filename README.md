# Distributed task executor for genomics research

## Development

Generate Go code from ProtoBuffer file

```bash
protoc -I pb --go_out=plugins=grpc:pb pb/worker.proto 
```

Build Docker image and publish to Docker Hub

```bash
docker build -t welliton/rnnr:1.0.1 .
docker push
```