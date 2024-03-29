FROM golang:1.18.2 AS builder
WORKDIR /go/src/github.com/labbcb/rnnr
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o rnnr .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
COPY --from=0 /go/src/github.com/labbcb/rnnr/rnnr /usr/local/bin/rnnr
ENTRYPOINT ["rnnr"]
