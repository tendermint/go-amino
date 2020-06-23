The "proto" folder is where proto3 schema files will be written to by main.go.
Then, run `protoc -I=. -I=./proto --go_out=./pb types.proto`.
