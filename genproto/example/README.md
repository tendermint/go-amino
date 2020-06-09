The "proto" folder is where proto3 schema files will be written to by main.go.
In the future, genproto will help generate these project-level "proto" folders
appropriately.

For now, we just add a symlink from
proto/github.com/tendermint/go-amino/genproto/example to the example directory
(recursive).  go.mod is required for this hack.
