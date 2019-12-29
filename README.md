# Distributed task executor for genomics research

## Development

Generate Go code from ProtoBuffer file

```bash
protoc -I pb --go_out=plugins=grpc:pb pb/worker.proto 
```