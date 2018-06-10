# dfs
Another filesystem experiment


## to mount:
go run main.go 147.75.101.31:10105 /tmp/mymnt -d

## Win fixes
- Find a way to mask the real uid and show the one of the user
- find out why a rename to an existing file works (is this also on osx/linux?)
- correct btim vs birthtim key in node structure

## TODO
- Propertly: Transform Error codes from server platform to client platform
- Find out why: Test Apple Finder crashing with its extended attr
- Implement garbage collection for chunks
- Add CoW Mechanism and snapshotting
- Add offsite backup/restore mechanism
- Add collaboration(locking) mechanism
- Add docker build client
- Add Docker run client
- Move Bazil cas code into our codebase

## Usolved intermittend windows error:

$ go test -v
=== RUN   TestEnd2End
Accepting connections on: 127.0.0.1:50973
Accepted conn from: 127.0.0.1:50974
Mounting...
The service dfs.test has been started.
=== RUN   TestEnd2End/basic_file_writing
create_test............................ OK 1.65s
create_related_test.................... OK 0.19s
create_sd_test......................... OK 0.37s
create_notraverse_test................. OK 0.38s
create_backup_test..................... OK 0.00s
create_restore_test.................... OK 0.00s
create_share_test...................... OK 0.40s
create_curdir_test..................... OK 0.07s
create_namelen_test.................... OK 0.14s
getfileinfo_test....................... OK 0.07s
delete_test............................ OK 0.29s
delete_pending_test.................... OK 0.16s
delete_mmap_test....................... OK 0.07s
delete_standby_test.................... OK 1.12s
rename_test............................ OK 1.62s
rename_open_test....................... OK 0.24s
rename_caseins_test.................... OK 0.56s
rename_flipflop_test................... OK 3.65s
rename_mmap_test....................... OK 0.41s
rename_standby_test.................... OK 2.38s
getvolinfo_test........................ OK 0.11s
setvolinfo_test........................ OK 0.00s
getsecurity_test....................... OK 0.10s
rdwr_noncached_test.................... OK 0.43s
rdwr_noncached_overlapped_test......... OK 0.40s
rdwr_cached_test....................... OK 0.43s
rdwr_cached_append_test................ OK 0.27s
rdwr_cached_overlapped_test............ OK 0.43s
rdwr_writethru_test.................... OK 0.43s
rdwr_writethru_append_test............. OK 0.25s
rdwr_writethru_overlapped_test......... OK 0.41s
rdwr_mmap_test......................... OK 2.64s
rdwr_mixed_test........................ OK 0.52s
flush_test............................. OK 1.25s
flush_volume_test...................... OK 0.00s
lock_noncached_test.................... OK 0.39s
lock_noncached_overlapped_test......... OK 0.45s
lock_cached_test....................... OK 0.40s
lock_cached_overlapped_test............ OK 0.43s
querydir_test.......................... Exception 0xc0000005 0x0 0x0 0x7ffe280f4b                                             3d
PC=0x7ffe280f4b3d

runtime: unknown pc 0x7ffe280f4b3d
stack: frame={sp:0x432fdd0, fp:0x0} stack=[0x4131ed0,0x432fed0)
000000000432fcd0:  00000000001d7440  00007ffe280f955b
000000000432fce0:  0000000005150600  0000000000417564 <runtime.wakefing+116>
000000000432fcf0:  0000000000ad3d00  000000c04204c180
000000000432fd00:  0000000005c418e0  00007ffe2828cccc
000000000432fd10:  00000000052ca230  0000000005c57870
000000000432fd20:  0000000000000040  00000000011b0000
000000000432fd30:  0000000000000049  000000c04202f500
000000000432fd40:  0000000000001ff8  0000000000000001
000000000432fd50:  000000c042083e98  00007ffe283b4038
000000000432fd60:  0000000005c57870  00007ffe283b4038
000000000432fd70:  fffffffffffffffe  00000018000012b1
000000000432fd80:  0000000005c57870  00007ffe2812b610
000000000432fd90:  0000000000000001  00007ffe280f6d50
000000000432fda0:  000000c04202f500  000000000432fdf0
000000000432fdb0:  fffffffffffffffe  000000c04204c180
000000000432fdc0:  000000c042083ed8  00007ffe280f4b43
000000000432fdd0: <000000c042083ed8  000000c04202e900
000000000432fde0:  000000c04202f500  0000000000000000
000000000432fdf0:  fffffffffffffffe  0000000000436e25 <runtime.goschedImpl+261>
000000000432fe00:  0000000000ab65d0  000000000045b643 <runtime.asmcgocall+115>
000000000432fe10:  000000000432fe40  000000c04202f500
000000000432fe20:  000000000432fe40  000000000432fe40
000000000432fe30:  0000000000436ffd <runtime.gosched_m+61>  0000000000000198
000000000432fe40:  000000c04204c180  0000000000459d4e <runtime.mcall+94>
000000000432fe50:  000000c04202e900  0000000000000000
000000000432fe60:  0000000000000000  0000000000000000
000000000432fe70:  0000000000a6ff3c  0000000000000000
000000000432fe80:  0000000000000000  0000000000000000
000000000432fe90:  0000000000000000  0000000000000000
000000000432fea0:  00000000012e8f90  0000000000a6feb0
000000000432feb0:  0000000000000000  00007ffe576cf0a0
000000000432fec0:  0000000000000000  00007ffe5767a553
runtime: unknown pc 0x7ffe280f4b3d
stack: frame={sp:0x432fdd0, fp:0x0} stack=[0x4131ed0,0x432fed0)
000000000432fcd0:  00000000001d7440  00007ffe280f955b
000000000432fce0:  0000000005150600  0000000000417564 <runtime.wakefing+116>
000000000432fcf0:  0000000000ad3d00  000000c04204c180
000000000432fd00:  0000000005c418e0  00007ffe2828cccc
000000000432fd10:  00000000052ca230  0000000005c57870
000000000432fd20:  0000000000000040  00000000011b0000
000000000432fd30:  0000000000000049  000000c04202f500
000000000432fd40:  0000000000001ff8  0000000000000001
000000000432fd50:  000000c042083e98  00007ffe283b4038
000000000432fd60:  0000000005c57870  00007ffe283b4038
000000000432fd70:  fffffffffffffffe  00000018000012b1
000000000432fd80:  0000000005c57870  00007ffe2812b610
000000000432fd90:  0000000000000001  00007ffe280f6d50
000000000432fda0:  000000c04202f500  000000000432fdf0
000000000432fdb0:  fffffffffffffffe  000000c04204c180
000000000432fdc0:  000000c042083ed8  00007ffe280f4b43
000000000432fdd0: <000000c042083ed8  000000c04202e900
000000000432fde0:  000000c04202f500  0000000000000000
000000000432fdf0:  fffffffffffffffe  0000000000436e25 <runtime.goschedImpl+261>
000000000432fe00:  0000000000ab65d0  000000000045b643 <runtime.asmcgocall+115>
000000000432fe10:  000000000432fe40  000000c04202f500
000000000432fe20:  000000000432fe40  000000000432fe40
000000000432fe30:  0000000000436ffd <runtime.gosched_m+61>  0000000000000198
000000000432fe40:  000000c04204c180  0000000000459d4e <runtime.mcall+94>
000000000432fe50:  000000c04202e900  0000000000000000
000000000432fe60:  0000000000000000  0000000000000000
000000000432fe70:  0000000000a6ff3c  0000000000000000
000000000432fe80:  0000000000000000  0000000000000000
000000000432fe90:  0000000000000000  0000000000000000
000000000432fea0:  00000000012e8f90  0000000000a6feb0
000000000432feb0:  0000000000000000  00007ffe576cf0a0
000000000432fec0:  0000000000000000  00007ffe5767a553

goroutine 18 [syscall]:
github.com/advanderveer/dfs/vendor/github.com/apple/foundationdb/bindings/go/src/                                             fdb._Cfunc_fdb_future_destroy(0x0)
        _cgo_gotypes.go:215 +0x48
github.com/advanderveer/dfs/vendor/github.com/apple/foundationdb/bindings/go/src/                                             fdb.newFuture.func1.1(0x0)
        Z:/Projects/go/src/github.com/advanderveer/dfs/vendor/github.com/apple/fo                                             undationdb/bindings/go/src/fdb/futures.go:81 +0x5d
github.com/advanderveer/dfs/vendor/github.com/apple/foundationdb/bindings/go/src/                                             fdb.newFuture.func1(0xc04292a000)
        Z:/Projects/go/src/github.com/advanderveer/dfs/vendor/github.com/apple/fo                                             undationdb/bindings/go/src/fdb/futures.go:81 +0x35

goroutine 1 [chan receive]:
testing.(*T).Run(0xc0421380f0, 0x855ae8, 0xb, 0x86f6f8, 0x48a43d)
        C:/Go/src/testing/testing.go:825 +0x308
testing.runTests.func1(0xc042138000)
        C:/Go/src/testing/testing.go:1063 +0x6b
testing.tRunner(0xc042138000, 0xc042077df8)
        C:/Go/src/testing/testing.go:777 +0xd7
testing.runTests(0xc0420f6f00, 0xaa3220, 0x1, 0x1, 0x4120a3)
        C:/Go/src/testing/testing.go:1061 +0x2cb
testing.(*M).Run(0xc042132080, 0x0)
        C:/Go/src/testing/testing.go:978 +0x178
main.main()
        _testmain.go:42 +0x158

goroutine 17 [chan receive, locked to thread]:
net/rpc.(*Client).Call(0xc042045080, 0x852e33, 0x7, 0x79e860, 0xc0424cf660, 0x79e                                             8a0, 0xc0424cf640, 0xc, 0xc)
        C:/Go/src/net/rpc/client.go:317 +0xc3
github.com/advanderveer/dfs/ffs/fsrpc.(*Sender).Open(0xc04212f740, 0xc0424cbe70,                                              0xc, 0x0, 0xc04202e300, 0xc042af4080)
        Z:/Projects/go/src/github.com/advanderveer/dfs/ffs/fsrpc/server_rpc.go:59                                             7 +0xec
github.com/advanderveer/dfs/vendor/github.com/billziss-gh/cgofuse/fuse.hostOpen(0                                             x17d750, 0x4a4fd30, 0xc000000000)
        Z:/Projects/go/src/github.com/advanderveer/dfs/vendor/github.com/billziss                                             -gh/cgofuse/fuse/host.go:782 +0xb8
github.com/advanderveer/dfs/vendor/github.com/billziss-gh/cgofuse/fuse._cgoexpwra                                             p_301f34628f8d_hostOpen(0x17d750, 0x4a4fd30, 0x0)
        _cgo_gotypes.go:693 +0x3c

goroutine 19 [syscall]:
os/signal.signal_recv(0x0)
        C:/Go/src/runtime/sigqueue.go:139 +0xad
os/signal.loop()
        C:/Go/src/os/signal/signal_unix.go:22 +0x29
created by os/signal.init.0
        C:/Go/src/os/signal/signal_unix.go:28 +0x48

goroutine 21 [syscall]:
github.com/advanderveer/dfs/vendor/github.com/billziss-gh/cgofuse/fuse._Cfunc_hos                                             tMount(0x3, 0xc042154900, 0x12e8650, 0x0)
        _cgo_gotypes.go:462 +0x54
github.com/advanderveer/dfs/vendor/github.com/billziss-gh/cgofuse/fuse.(*FileSyst                                             emHost).Mount.func9(0xc042154900, 0x4, 0x4, 0xc000000003, 0xc042154900, 0x12e8650                                             , 0xc04203dee8)
        Z:/Projects/go/src/github.com/advanderveer/dfs/vendor/github.com/billziss                                             -gh/cgofuse/fuse/host.go:1232 +0x124
github.com/advanderveer/dfs/vendor/github.com/billziss-gh/cgofuse/fuse.(*FileSyst                                             emHost).Mount(0xc04212f770, 0x850640, 0x2, 0xc04203df30, 0x0, 0x0, 0xc04212f700)
        Z:/Projects/go/src/github.com/advanderveer/dfs/vendor/github.com/billziss                                             -gh/cgofuse/fuse/host.go:1232 +0x37d
github.com/advanderveer/dfs.TestEnd2End(0xc0421380f0)
        Z:/Projects/go/src/github.com/advanderveer/dfs/main_test.go:83 +0x317
testing.tRunner(0xc0421380f0, 0x86f6f8)
        C:/Go/src/testing/testing.go:777 +0xd7
created by testing.(*T).Run
        C:/Go/src/testing/testing.go:824 +0x2e7

goroutine 22 [syscall]:
github.com/advanderveer/dfs/vendor/github.com/apple/foundationdb/bindings/go/src/                                             fdb._Cfunc_fdb_run_network(0x0)
        _cgo_gotypes.go:411 +0x50
github.com/advanderveer/dfs/vendor/github.com/apple/foundationdb/bindings/go/src/                                             fdb.startNetwork.func1()
        Z:/Projects/go/src/github.com/advanderveer/dfs/vendor/github.com/apple/fo                                             undationdb/bindings/go/src/fdb/fdb.go:182 +0x2d
created by github.com/advanderveer/dfs/vendor/github.com/apple/foundationdb/bindi                                             ngs/go/src/fdb.startNetwork
        Z:/Projects/go/src/github.com/advanderveer/dfs/vendor/github.com/apple/fo                                             undationdb/bindings/go/src/fdb/fdb.go:181 +0x8c

goroutine 6 [IO wait]:
internal/poll.runtime_pollWait(0x3b30f58, 0x72, 0x8a0000)
        C:/Go/src/runtime/netpoll.go:173 +0x5e
internal/poll.(*pollDesc).wait(0xc0421501c8, 0x72, 0xa75200, 0x0, 0x0)
        C:/Go/src/internal/poll/fd_poll_runtime.go:85 +0xa2
internal/poll.(*ioSrv).ExecIO(0xab41f8, 0xc042150018, 0xc042154880, 0x1, 0x0, 0x3                                             4c)
        C:/Go/src/internal/poll/fd_windows.go:223 +0x13a
internal/poll.(*FD).acceptOne(0xc042150000, 0x34c, 0xc0421620e0, 0x2, 0x2, 0xc042                                             150018, 0x7d7640, 0xc04207fd68, 0x41290f, 0x10)
        C:/Go/src/internal/poll/fd_windows.go:793 +0xae
internal/poll.(*FD).Accept(0xc042150000, 0xc04201f150, 0x0, 0x0, 0x0, 0x0, 0xc000                                             000000, 0x0, 0x0, 0x0, ...)
        C:/Go/src/internal/poll/fd_windows.go:827 +0x142
net.(*netFD).accept(0xc042150000, 0x3b31088, 0x89f580, 0xc04207a008)
        C:/Go/src/net/fd_windows.go:192 +0x86
net.(*TCPListener).accept(0xc042004398, 0x459460, 0xc04207ff18, 0xc04207ff20)
        C:/Go/src/net/tcpsock_posix.go:136 +0x35
net.(*TCPListener).Accept(0xc042004398, 0x86fd40, 0xc042130320, 0x3b31088, 0xc042                                             004818)
        C:/Go/src/net/tcpsock.go:259 +0x50
github.com/advanderveer/dfs/ffs/fsrpc.(*Svr).ListenAndServe(0xc04212e9f0, 0xa, 0x                                             c04212e9f0)
        Z:/Projects/go/src/github.com/advanderveer/dfs/ffs/fsrpc/server_net.go:38                                              +0x15b
created by github.com/advanderveer/dfs.TestEnd2End
        Z:/Projects/go/src/github.com/advanderveer/dfs/main_test.go:33 +0x174

goroutine 8 [IO wait]:
internal/poll.runtime_pollWait(0x3b30db8, 0x72, 0x8a0000)
        C:/Go/src/runtime/netpoll.go:173 +0x5e
internal/poll.(*pollDesc).wait(0xc042150488, 0x72, 0xa75200, 0x0, 0x0)
        C:/Go/src/internal/poll/fd_poll_runtime.go:85 +0xa2
internal/poll.(*ioSrv).ExecIO(0xab41f8, 0xc0421502d8, 0x86fa00, 0x40242f, 0xc0421                                             73930, 0xc042173938)
        C:/Go/src/internal/poll/fd_windows.go:223 +0x13a
internal/poll.(*FD).Read(0xc0421502c0, 0xc04216d000, 0x1000, 0x1000, 0x0, 0x0, 0x                                             0)
        C:/Go/src/internal/poll/fd_windows.go:484 +0x248
net.(*netFD).Read(0xc0421502c0, 0xc04216d000, 0x1000, 0x1000, 0xc0421739b0, 0x447                                             2b6, 0xc0421739b0)
        C:/Go/src/net/fd_windows.go:151 +0x56
net.(*conn).Read(0xc042004818, 0xc04216d000, 0x1000, 0x1000, 0x0, 0x0, 0x0)
        C:/Go/src/net/net.go:176 +0x71
bufio.(*Reader).Read(0xc042044f00, 0xc042009550, 0x1, 0x9, 0xc0423a92e8, 0x83aa20                                             0c69c7ea01, 0x830000c04202f080)
        C:/Go/src/bufio/bufio.go:216 +0x23f
io.ReadAtLeast(0x89ef40, 0xc042044f00, 0xc042009550, 0x1, 0x9, 0x1, 0x198, 0xc042                                             4cbe80, 0x3)
        C:/Go/src/io/io.go:309 +0x8d
io.ReadFull(0x89ef40, 0xc042044f00, 0xc042009550, 0x1, 0x9, 0xc042173b48, 0x42df8                                             b, 0x86f1b8)
        C:/Go/src/io/io.go:327 +0x5f
encoding/gob.decodeUintReader(0x89ef40, 0xc042044f00, 0xc042009550, 0x9, 0x9, 0xc                                             042042760, 0x7b3ae0, 0xc042173ba0, 0x42d660)
        C:/Go/src/encoding/gob/decode.go:120 +0x6a
encoding/gob.(*Decoder).recvMessage(0xc042137280, 0xc042173bb8)
        C:/Go/src/encoding/gob/decoder.go:80 +0x5e
encoding/gob.(*Decoder).decodeTypeSequence(0xc042137280, 0x870200, 0xc042137280)
        C:/Go/src/encoding/gob/decoder.go:142 +0x13d
encoding/gob.(*Decoder).DecodeValue(0xc042137280, 0x7a43a0, 0xc0421548a0, 0x16, 0                                             x0, 0x0)
        C:/Go/src/encoding/gob/decoder.go:210 +0xe3
encoding/gob.(*Decoder).Decode(0xc042137280, 0x7a43a0, 0xc0421548a0, 0xc0421548a0                                             , 0xc042130348)
        C:/Go/src/encoding/gob/decoder.go:187 +0x156
net/rpc.(*gobServerCodec).ReadRequestHeader(0xc04212f590, 0xc0421548a0, 0x79e8a0,                                              0x11b06a8)
        C:/Go/src/net/rpc/server.go:404 +0x4c
net/rpc.(*Server).readRequestHeader(0xc042130320, 0x8a2720, 0xc04212f590, 0xc0424                                             cf6a0, 0x16, 0xc0421b5280, 0x100000000000001, 0x0, 0x0)
        C:/Go/src/net/rpc/server.go:589 +0x6e
net/rpc.(*Server).readRequest(0xc042130320, 0x8a2720, 0xc04212f590, 0xc042130320,                                              0xc042009560, 0xc042009570, 0xc042136900, 0xc0421b5280, 0x79e860, 0xc0424cf680,                                              ...)
        C:/Go/src/net/rpc/server.go:549 +0x61
net/rpc.(*Server).ServeCodec(0xc042130320, 0x8a2720, 0xc04212f590)
        C:/Go/src/net/rpc/server.go:464 +0x9e
net/rpc.(*Server).ServeConn(0xc042130320, 0x3b31088, 0xc042004818)
        C:/Go/src/net/rpc/server.go:455 +0x188
created by github.com/advanderveer/dfs/ffs/fsrpc.(*Svr).ListenAndServe
        Z:/Projects/go/src/github.com/advanderveer/dfs/ffs/fsrpc/server_net.go:45                                              +0x297

goroutine 9 [IO wait]:
internal/poll.runtime_pollWait(0x3b30e88, 0x72, 0x8a0000)
        C:/Go/src/runtime/netpoll.go:173 +0x5e
internal/poll.(*pollDesc).wait(0xc0420baf88, 0x72, 0xa75200, 0x0, 0x0)
        C:/Go/src/internal/poll/fd_poll_runtime.go:85 +0xa2
internal/poll.(*ioSrv).ExecIO(0xab41f8, 0xc0420badd8, 0x86fa00, 0xc04202f200, 0x8                                             6f1b8, 0xc042ae6780)
        C:/Go/src/internal/poll/fd_windows.go:223 +0x13a
internal/poll.(*FD).Read(0xc0420badc0, 0xc042171000, 0x1000, 0x1000, 0x0, 0x0, 0x                                             0)
        C:/Go/src/internal/poll/fd_windows.go:484 +0x248
net.(*netFD).Read(0xc0420badc0, 0xc042171000, 0x1000, 0x1000, 0x10, 0xc04202f200,                                              0x86f1b8)
        C:/Go/src/net/fd_windows.go:151 +0x56
net.(*conn).Read(0xc042004820, 0xc042171000, 0x1000, 0x1000, 0x0, 0x0, 0x0)
        C:/Go/src/net/net.go:176 +0x71
bufio.(*Reader).Read(0xc042044fc0, 0xc042009590, 0x1, 0x9, 0xc042167c98, 0x10, 0x                                             c04202f200)
        C:/Go/src/bufio/bufio.go:216 +0x23f
io.ReadAtLeast(0x89ef40, 0xc042044fc0, 0xc042009590, 0x1, 0x9, 0x1, 0xc042167d10,                                              0x42df8b, 0x86f1c8)
        C:/Go/src/io/io.go:309 +0x8d
io.ReadFull(0x89ef40, 0xc042044fc0, 0xc042009590, 0x1, 0x9, 0xc042167d48, 0x40536                                             c, 0x430bc2)
        C:/Go/src/io/io.go:327 +0x5f
encoding/gob.decodeUintReader(0x89ef40, 0xc042044fc0, 0xc042009590, 0x9, 0x9, 0xc                                             0421b8620, 0xc042167e08, 0xc042167db8, 0x42d660)
        C:/Go/src/encoding/gob/decode.go:120 +0x6a
encoding/gob.(*Decoder).recvMessage(0xc042137300, 0xc042167dd0)
        C:/Go/src/encoding/gob/decoder.go:80 +0x5e
encoding/gob.(*Decoder).decodeTypeSequence(0xc042137300, 0x870200, 0xc042137300)
        C:/Go/src/encoding/gob/decoder.go:142 +0x13d
encoding/gob.(*Decoder).DecodeValue(0xc042137300, 0x7a43e0, 0xc04206fe30, 0x16, 0                                             x0, 0x0)
        C:/Go/src/encoding/gob/decoder.go:210 +0xe3
encoding/gob.(*Decoder).Decode(0xc042137300, 0x7a43e0, 0xc04206fe30, 0x0, 0x0)
        C:/Go/src/encoding/gob/decoder.go:187 +0x156
net/rpc.(*gobClientCodec).ReadResponseHeader(0xc04212f6e0, 0xc04206fe30, 0xc0424d                                             0eb0, 0x0)
        C:/Go/src/net/rpc/client.go:223 +0x4c
net/rpc.(*Client).input(0xc042045080)
        C:/Go/src/net/rpc/client.go:109 +0xae
created by net/rpc.NewClientWithCodec
        C:/Go/src/net/rpc/client.go:201 +0x99

goroutine 10 [syscall, locked to thread]:
syscall.Syscall(0x7ffe577e08c0, 0x2, 0x3cc, 0xffffffff, 0x0, 0x0, 0x0, 0x0)
        C:/Go/src/runtime/syscall_windows.go:171 +0xf9
syscall.WaitForSingleObject(0x3cc, 0xc0ffffffff, 0x0, 0x0, 0xc)
        C:/Go/src/syscall/zsyscall_windows.go:737 +0x6b
os.(*Process).wait(0xc042228510, 0x0, 0x0, 0x0)
        C:/Go/src/os/exec_windows.go:18 +0x6f
os.(*Process).Wait(0xc042228510, 0x86fd88, 0x86fd90, 0x86fd80)
        C:/Go/src/os/exec.go:123 +0x32
os/exec.(*Cmd).Wait(0xc042182000, 0x0, 0x0)
        C:/Go/src/os/exec/exec.go:461 +0x63
os/exec.(*Cmd).Run(0xc042182000, 0x14, 0xc042177eb8)
        C:/Go/src/os/exec/exec.go:305 +0x63
github.com/advanderveer/dfs.WindowsEnd2End(0x850640, 0x2, 0xc0421380f0)
        Z:/Projects/go/src/github.com/advanderveer/dfs/main_test.go:130 +0x110
github.com/advanderveer/dfs.TestEnd2End.func1(0x850640, 0x2, 0xc0421380f0, 0xc042                                             12f740, 0xc04212f770)
        Z:/Projects/go/src/github.com/advanderveer/dfs/main_test.go:69 +0x13f
created by github.com/advanderveer/dfs.TestEnd2End
        Z:/Projects/go/src/github.com/advanderveer/dfs/main_test.go:53 +0x2a2

goroutine 4772 [runnable]:
sync.runtime_SemacquireMutex(0xc0421b024c, 0x561c00)
        C:/Go/src/runtime/sema.go:71 +0x44
sync.(*Mutex).Lock(0xc0421b0280)
        C:/Go/src/sync/mutex.go:134 +0x10f
github.com/advanderveer/dfs/vendor/github.com/apple/foundationdb/bindings/go/src/                                             fdb.fdb_future_block_until_ready(0x5cb52f0)
        Z:/Projects/go/src/github.com/advanderveer/dfs/vendor/github.com/apple/fo                                             undationdb/bindings/go/src/fdb/futures.go:93 +0x91
github.com/advanderveer/dfs/vendor/github.com/apple/foundationdb/bindings/go/src/                                             fdb.future.BlockUntilReady(0x5cb52f0)
        Z:/Projects/go/src/github.com/advanderveer/dfs/vendor/github.com/apple/fo                                             undationdb/bindings/go/src/fdb/futures.go:97 +0x32
github.com/advanderveer/dfs/vendor/github.com/apple/foundationdb/bindings/go/src/                                             fdb.(*futureByteSlice).Get.func1()
        Z:/Projects/go/src/github.com/advanderveer/dfs/vendor/github.com/apple/fo                                             undationdb/bindings/go/src/fdb/futures.go:140 +0x93
sync.(*Once).Do(0xc0423020b0, 0xc0421bd458)
        C:/Go/src/sync/once.go:44 +0xc5
github.com/advanderveer/dfs/vendor/github.com/apple/foundationdb/bindings/go/src/                                             fdb.(*futureByteSlice).Get(0xc042302080, 0xb, 0xb, 0x0, 0x8a2c60, 0xc042302080)
        Z:/Projects/go/src/github.com/advanderveer/dfs/vendor/github.com/apple/fo                                             undationdb/bindings/go/src/fdb/futures.go:135 +0x53
github.com/advanderveer/dfs/vendor/github.com/apple/foundationdb/bindings/go/src/                                             fdb.(*futureByteSlice).MustGet(0xc042302080, 0x89fb00, 0xc042276040, 0x8a2c60)
        Z:/Projects/go/src/github.com/advanderveer/dfs/vendor/github.com/apple/fo                                             undationdb/bindings/go/src/fdb/futures.go:157 +0x32
github.com/advanderveer/dfs/ffs/nodes.(*Node).getUint32At(0xc0424cf820, 0xc0424d1                                             1c0, 0x85178c, 0x5, 0x1)
        Z:/Projects/go/src/github.com/advanderveer/dfs/ffs/nodes/node.go:55 +0x13                                             d
github.com/advanderveer/dfs/ffs/nodes.(*Node).Stat(0xc0424cf820, 0xc0424d11c0, 0x                                             0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, ...)
        Z:/Projects/go/src/github.com/advanderveer/dfs/ffs/nodes/node_stat.go:17                                              +0x1ca
github.com/advanderveer/dfs/ffs.(*Memfs).openNode(0xc04212e990, 0xc0424d11c0, 0xc                                             0424cbe80, 0xc, 0x0, 0xc042241990, 0x4120a3)
        Z:/Projects/go/src/github.com/advanderveer/dfs/ffs/ffs.go:529 +0x289
github.com/advanderveer/dfs/ffs.(*Memfs).Open.func1(0xc0424d11c0, 0x8, 0xc042ae67                                             80)
        Z:/Projects/go/src/github.com/advanderveer/dfs/ffs/ffs.go:219 +0x52
github.com/advanderveer/dfs/ffs/nodes.(*Store).TxWithErrcUint64.func1(0xc0424d11c                                             0, 0x86f748, 0xc0422419a0, 0x11b06a8, 0x0)
        Z:/Projects/go/src/github.com/advanderveer/dfs/ffs/nodes/store.go:103 +0x                                             4b
github.com/advanderveer/dfs/vendor/github.com/apple/foundationdb/bindings/go/src/                                             fdb.Database.Transact.func1(0x0, 0x0, 0x0, 0x0)
        Z:/Projects/go/src/github.com/advanderveer/dfs/vendor/github.com/apple/fo                                             undationdb/bindings/go/src/fdb/database.go:136 +0x84
github.com/advanderveer/dfs/vendor/github.com/apple/foundationdb/bindings/go/src/                                             fdb.retryable(0xc0424cf720, 0xc0421bda08, 0x0, 0x0, 0x41290f, 0x20)
        Z:/Projects/go/src/github.com/advanderveer/dfs/vendor/github.com/apple/fo                                             undationdb/bindings/go/src/fdb/database.go:87 +0x32
github.com/advanderveer/dfs/vendor/github.com/apple/foundationdb/bindings/go/src/                                             fdb.Database.Transact(0xc042004040, 0xc0424cf700, 0xc04212e928, 0xc042241a70, 0x4                                             1290f, 0x20)
        Z:/Projects/go/src/github.com/advanderveer/dfs/vendor/github.com/apple/fo                                             undationdb/bindings/go/src/fdb/database.go:145 +0xba
github.com/advanderveer/dfs/ffs/nodes.(*Store).TxWithErrcUint64(0xc04212e900, 0xc                                             0424cf6e0, 0xc0424cf6c0, 0x2)
        Z:/Projects/go/src/github.com/advanderveer/dfs/ffs/nodes/store.go:102 +0x                                             116
github.com/advanderveer/dfs/ffs.(*Memfs).Open(0xc04212e990, 0xc0424cbe80, 0xc, 0x                                             0, 0xffffffffffffffff, 0xc0421503f8)
        Z:/Projects/go/src/github.com/advanderveer/dfs/ffs/ffs.go:218 +0x22b
github.com/advanderveer/dfs/ffs/fsrpc.(*Receiver).Open(0xc04201e930, 0xc0424cf680                                             , 0xc0424cf6a0, 0x0, 0x0)
        Z:/Projects/go/src/github.com/advanderveer/dfs/ffs/fsrpc/server_rpc.go:58                                             4 +0x60
reflect.Value.call(0xc0420447e0, 0xc0420045c8, 0x13, 0x850b66, 0x4, 0xc042241f18,                                              0x3, 0x3, 0xc0421ffdc0, 0xc042009570, ...)
        C:/Go/src/reflect/value.go:447 +0x970
reflect.Value.Call(0xc0420447e0, 0xc0420045c8, 0x13, 0xc042241f18, 0x3, 0x3, 0xc0                                             4212f590, 0x0, 0x0)
        C:/Go/src/reflect/value.go:308 +0xab
net/rpc.(*service).call(0xc042042880, 0xc042130320, 0xc042009560, 0xc042009570, 0                                             xc042136900, 0xc0421b5280, 0x79e860, 0xc0424cf680, 0x16, 0x79e8a0, ...)
        C:/Go/src/net/rpc/server.go:384 +0x155
created by net/rpc.(*Server).ServeCodec
        C:/Go/src/net/rpc/server.go:480 +0x441
rax     0xa6f1f0
rbx     0xc042083ed8
rcx     0x0
rdi     0xc042083ed8
rsi     0xc04202f500
rbp     0xc042083e98
rsp     0x432fdd0
r8      0xc04204c180
r9      0x0
r10     0xc04292a000
r11     0x173ee0
r12     0x1ff8
r13     0x49
r14     0x11b0000
r15     0x40
rip     0x7ffe280f4b3d
rflags  0x10202
cs      0x33
fs      0x53
gs      0x2b
KO
    ASSERT(Success) failed at dirctl-test.c:220:querydir_dotest
exit status 2
FAIL    github.com/advanderveer/dfs     44.598s
