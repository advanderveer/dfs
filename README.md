# dfs
Another filesystem experiment

## to mount:
go run main.go 147.75.101.31:10105 /tmp/mymnt -d

## TODO
- Reduce nr of RPC that we generate and include the base fs
- Transform GID/UID from client
- Transform Error codes from server to client
- Test Remount (ino count)
- Test Apple Finder crasching with its extended attr
