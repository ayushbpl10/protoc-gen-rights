protoc -I /usr/local/include -I ./  --go_out=plugins=grpc:./example  ./example/example.proto
protoc -I /usr/local/include -I ./  --go_out=plugins=grpc:./  ./rights/rights.proto

packr && go build && protoc -I /usr/local/include -I  ./  --plugin=protoc-gen-rights=protoc-gen-rights  --zap_out=:./example  ./example/example.proto
