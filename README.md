# dfs
Another filesystem experiment

fuse.ENOSYS is 78 on Darwin
fuse.ENOSYS is 38 on Linux


ON mac side
```
unique: 26, opcode: GETATTR (3), nodeid: 1, insize: 56, pid: 12151
getattr /
   unique: 26, success, outsize: 136
unique: 18, opcode: GETXATTR (22), nodeid: 1, insize: 77, pid: 12151
getxattr / com.apple.FinderInfo 32 0
GETXATTR 0
   unique: 18, success, outsize: 48
unique: 24, opcode: GETXATTR (22), nodeid: 1, insize: 92, pid: 12151
getxattr / com.apple.metadata:_kMDItemUserTags 0 0
GETXATTR 0
   unique: 24, success, outsize: 24
unique: 22, opcode: GETXATTR (22), nodeid: 1, insize: 77, pid: 12151
getxattr / com.apple.FinderInfo 32 0
GETXATTR 0
   unique: 22, success, outsize: 48
unique: 28, opcode: OPENDIR (27), nodeid: 1, insize: 48, pid: 12151
opendir flags: 0x0 /
   opendir[0] flags: 0x0 /
   unique: 28, success, outsize: 32
unique: 29, opcode: READDIR (28), nodeid: 1, insize: 80, pid: 12151
readdir[0] from 0
   unique: 29, success, outsize: 128
unique: 21, opcode: GETATTR (3), nodeid: 3, insize: 56, pid: 12151
getattr /rancher-images.txt
   unique: 21, success, outsize: 136
unique: 27, opcode: GETXATTR (22), nodeid: 3, insize: 77, pid: 12151
getxattr /rancher-images.txt com.apple.FinderInfo 32 0
GETXATTR 0
   unique: 27, success, outsize: 48
unique: 25, opcode: GETXATTR (22), nodeid: 3, insize: 79, pid: 12151
getxattr /rancher-images.txt com.apple.ResourceFork 0 0
GETXATTR -95
   unique: 25, error: -45 (Operation not supported), outsize: 16
unique: 19, opcode: LOOKUP (1), nodeid: 1, insize: 61, pid: 12151
LOOKUP /._rancher-images.txt
getattr /._rancher-images.txt
   unique: 19, error: -2 (No such file or directory), outsize: 16
unique: 2, opcode: READDIR (28), nodeid: 1, insize: 80, pid: 12151
   unique: 2, success, outsize: 16
unique: 3, opcode: READDIR (28), nodeid: 1, insize: 80, pid: 12151
   unique: 3, success, outsize: 16
unique: 5, opcode: READDIR (28), nodeid: 1, insize: 80, pid: 12151
   unique: 5, success, outsize: 16
unique: 4, opcode: RELEASEDIR (29), nodeid: 1, insize: 64, pid: 12151
releasedir[0] flags: 0x0
   unique: 4, success, outsize: 16
unique: 10, opcode: GETXATTR (22), nodeid: 3, insize: 90, pid: 12151
getxattr /rancher-images.txt com.apple.LaunchServices.OpenWith 0 0
GETXATTR 0
   unique: 10, success, outsize: 24
unique: 8, opcode: GETATTR (3), nodeid: 1, insize: 56, pid: 12151
getattr /
   unique: 8, success, outsize: 136
unique: 7, opcode: STATFS (17), nodeid: 1, insize: 40, pid: 12151
statfs /
   unique: 7, success, outsize: 96
unique: 6, opcode: GETXATTR (22), nodeid: 3, insize: 92, pid: 12151
getxattr /rancher-images.txt com.apple.metadata:_kMDItemUserTags 0 0
GETXATTR 0
   unique: 6, success, outsize: 24
unique: 20, opcode: GETXATTR (22), nodeid: 3, insize: 77, pid: 12151
getxattr /rancher-images.txt com.apple.FinderInfo 32 0
GETXATTR 0
   unique: 20, success, outsize: 48
unique: 9, opcode: OPEN (14), nodeid: 3, insize: 48, pid: 12151
open flags: 0x0 /rancher-images.txt
   open[2] flags: 0x0 /rancher-images.txt
   unique: 9, success, outsize: 32
unique: 15, opcode: GETXATTR (22), nodeid: 3, insize: 79, pid: 12151
getxattr /rancher-images.txt com.apple.TextEncoding 1000 0
GETXATTR 0
   unique: 15, success, outsize: 16
unique: 11, opcode: RELEASE (18), nodeid: 3, insize: 64, pid: 12151
release[2] flags: 0x0
   unique: 11, success, outsize: 16
unique: 14, opcode: GETATTR (3), nodeid: 1, insize: 56, pid: 12294
getattr /
   unique: 14, success, outsize: 136
unique: 13, opcode: GETXATTR (22), nodeid: 1, insize: 77, pid: 12294
getxattr / com.apple.FinderInfo 32 0
GETXATTR 0
   unique: 13, success, outsize: 48
unique: 16, opcode: STATFS (17), nodeid: 1, insize: 40, pid: 12294
statfs /
   unique: 16, success, outsize: 96
unique: 17, opcode: STATFS (17), nodeid: 1, insize: 40, pid: 12294
statfs /
   unique: 17, success, outsize: 96
unique: 23, opcode: STATFS (17), nodeid: 1, insize: 40, pid: 12294
statfs /
   unique: 23, success, outsize: 96
unique: 26, opcode: STATFS (17), nodeid: 1, insize: 40, pid: 12294
statfs /
   unique: 26, success, outsize: 96
unique: 18, opcode: STATFS (17), nodeid: 1, insize: 40, pid: 12294
statfs /
   unique: 18, success, outsize: 96
unique: 24, opcode: GETXATTR (22), nodeid: 1, insize: 77, pid: 12294
getxattr / com.apple.FinderInfo 32 0
GETXATTR 0
   unique: 24, success, outsize: 48
unique: 22, opcode: LOOKUP (1), nodeid: 1, insize: 51, pid: 12294
LOOKUP /.localized
getattr /.localized
   unique: 22, error: -2 (No such file or directory), outsize: 16
unique: 28, opcode: LOOKUP (1), nodeid: 1, insize: 50, pid: 12294
LOOKUP /.DS_Store
getattr /.DS_Store
   unique: 28, error: -2 (No such file or directory), outsize: 16
unique: 29, opcode: LOOKUP (1), nodeid: 1, insize: 50, pid: 12294
LOOKUP /.DS_Store
getattr /.DS_Store
   unique: 29, error: -2 (No such file or directory), outsize: 16
unique: 21, opcode: LOOKUP (1), nodeid: 1, insize: 51, pid: 12294
LOOKUP /.localized
getattr /.localized
   unique: 21, error: -2 (No such file or directory), outsize: 16
```
