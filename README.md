# This is a simple kv database using Raft 

## Usage

- `proto` file is in the ./rpc. If you edit it, pleause use the command below to generate the new gprc file and then move the files into ./raftkv

```shell
protoc --proto_path=./ --go_out=./ --go-grpc_out=./ ./rpc/kvrpc.proto
```