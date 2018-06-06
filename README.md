# dfs
Another filesystem experiment

## log structured file filesystem example in c
https://github.com/sphurti/Log-Structured-Filesystem/blob/master/src/lfs.c

cas: https://github.com/bazil/bazil/tree/7d1f80b37293381ba5a2bab5640ecc6b51157af4/cas/blobs

## to mount:
go run main.go 147.75.101.31:10105 /tmp/mymnt -d

## Win fixes
- Find a way to mask the real uid and show the one of the user
- find out why a rename to an existing file works (is this also on osx/linux?)
- correct btim vs birthtim key in node structure

## TODO
- Reduce nr of RPC that we generate and include the base fs
- Transform GID/UID from client
- Transform Error codes from server to client
- Test Remount (ino count)
- Test Apple Finder crasching with its extended attr


### panic from windows
May 26 11:07:41 remove-me go[1036]: goroutine 392 [running]:
May 26 11:07:41 remove-me go[1036]: github.com/advanderveer/dfs/vendor/github.com/apple/foundationdb/bindings/go/src/fdb.panicToError(0xc4220079b0)
May 26 11:07:41 remove-me go[1036]: /root/go/src/github.com/advanderveer/dfs/vendor/github.com/apple/foundationdb/bindings/go/src/fdb/fdb.go:355 +0
May 26 11:07:41 remove-me go[1036]: panic(0x77a160, 0xbdef20)
May 26 11:07:41 remove-me go[1036]: /usr/local/go/src/runtime/panic.go:502 +0x229
May 26 11:07:41 remove-me go[1036]: github.com/advanderveer/dfs/ffs/nodes.(*Node).GetChld(0x0, 0xc421fad130, 0xc420168687, 0xb, 0x0)
May 26 11:07:41 remove-me go[1036]: /root/go/src/github.com/advanderveer/dfs/ffs/nodes/node_chld.go:22 +0xce
May 26 11:07:41 remove-me go[1036]: github.com/advanderveer/dfs/ffs.(*Memfs).lookupNode(0xc420100b00, 0xc421fad130, 0xc420168680, 0x12, 0x0, 0xc421
May 26 11:07:41 remove-me go[1036]: /root/go/src/github.com/advanderveer/dfs/ffs/ffs.go:597 +0xcf
May 26 11:07:41 remove-me go[1036]: github.com/advanderveer/dfs/ffs.(*Memfs).getNode(0xc420100b00, 0xc421fad130, 0xc420168680, 0x12, 0xffffffffffff
May 26 11:07:41 remove-me go[1036]: /root/go/src/github.com/advanderveer/dfs/ffs/ffs.go:578 +0x61
May 26 11:07:41 remove-me go[1036]: github.com/advanderveer/dfs/ffs.(*Memfs).Getattr.func1(0xc421fad130, 0x754360)
May 26 11:07:41 remove-me go[1036]: /root/go/src/github.com/advanderveer/dfs/ffs/ffs.go:238 +0x73
May 26 11:07:41 remove-me go[1036]: github.com/advanderveer/dfs/ffs/nodes.(*Store).TxWithErrc.func1(0xc421fad130, 0x803950, 0xc4200b39b0, 0x7fcc2a2
May 26 11:07:41 remove-me go[1036]: /root/go/src/github.com/advanderveer/dfs/ffs/nodes/store.go:141 +0x3b
May 26 11:07:41 remove-me go[1036]: github.com/advanderveer/dfs/vendor/github.com/apple/foundationdb/bindings/go/src/fdb.Database.Transact.func1(0x
May 26 11:07:41 remove-me go[1036]: /root/go/src/github.com/advanderveer/dfs/vendor/github.com/apple/foundationdb/bindings/go/src/fdb/database.go:1
May 26 11:07:41 remove-me go[1036]: github.com/advanderveer/dfs/vendor/github.com/apple/foundationdb/bindings/go/src/fdb.retryable(0xc421fc5bc0, 0x
May 26 11:07:41 remove-me go[1036]: /root/go/src/github.com/advanderveer/dfs/vendor/github.com/apple/foundationdb/bindings/go/src/fdb/database.go:8
May 26 11:07:41 remove-me go[1036]: github.com/advanderveer/dfs/vendor/github.com/apple/foundationdb/bindings/go/src/fdb.Database.Transact(0xc42016
May 26 11:07:41 remove-me go[1036]: /root/go/src/github.com/advanderveer/dfs/vendor/github.com/apple/foundationdb/bindings/go/src/fdb/database.go:1
May 26 11:07:41 remove-me go[1036]: github.com/advanderveer/dfs/ffs/nodes.(*Store).TxWithErrc(0xc420156450, 0xc420156de0, 0xc421fc5b80)
May 26 11:07:41 remove-me go[1036]: /root/go/src/github.com/advanderveer/dfs/ffs/nodes/store.go:140 +0xe8
May 26 11:07:41 remove-me go[1036]: github.com/advanderveer/dfs/ffs.(*Memfs).Getattr(0xc420100b00, 0xc420168680, 0x12, 0xc4200a0870, 0xffffffffffff
May 26 11:07:41 remove-me go[1036]: /root/go/src/github.com/advanderveer/dfs/ffs/ffs.go:237 +0x237
May 26 11:07:41 remove-me go[1036]: github.com/advanderveer/dfs/ffs/fsrpc.(*Receiver).Getattr(0xc420152020, 0xc421fc5b60, 0xc421fad100, 0x0, 0x0)
May 26 11:07:41 remove-me go[1036]: /root/go/src/github.com/advanderveer/dfs/ffs/fsrpc/server_rpc.go:324 +0x5f
May 26 11:07:42 remove-me go[1036]: reflect.Value.call(0xc4200ca3c0, 0xc42016e1b0, 0x13, 0x7e73c4, 0x4, 0xc4200b3f18, 0x3, 0x3, 0xc420160b40, 0x756
May 26 11:07:42 remove-me go[1036]: /usr/local/go/src/reflect/value.go:447 +0x969
May 26 11:07:42 remove-me go[1036]: reflect.Value.Call(0xc4200ca3c0, 0xc42016e1b0, 0x13, 0xc421037f18, 0x3, 0x3, 0x7ba880, 0x7cbd01, 0xc420160080)
May 26 11:07:42 remove-me go[1036]: /usr/local/go/src/reflect/value.go:308 +0xa4
May 26 11:07:42 remove-me go[1036]: net/rpc.(*service).call(0xc420018080, 0xc420154230, 0xc420150060, 0xc420150070, 0xc420166580, 0xc4201784c0, 0x7
May 26 11:07:42 remove-me go[1036]: /usr/local/go/src/net/rpc/server.go:384 +0x14e
