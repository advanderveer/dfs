package fsrpc

import (
	"runtime"

	"github.com/billziss-gh/cgofuse/fuse"
)

//ercc will change the errcode to a platform specific version
//hack for making the error codes portable, probably not really the way to go
//@TODO how can this work for local codes, are they mangled?
func errc(in int) (out int) {
	if in >= 0 {
		return in //not an error code, probably read/write result
	}

	out = in

	//@TODO Find way to generate actual fuse error codes on the client side
	//the fix below only works if server is linux and client is not
	if runtime.GOOS != "linux" {
		switch in {
		case -38:
			out = -fuse.ENOSYS
		case -95:
			out = -fuse.ENOTSUP
		case -61: //ENODATA or ENOATTR
			out = -fuse.ENOATTR
		case -28:
			out = -fuse.ENOSPC
		case -2:
			out = -fuse.ENOENT
		case -17:
			out = -fuse.EEXIST
		case -22:
			out = -fuse.EINVAL
		case -34:
			out = -fuse.ERANGE
		case -21:
			out = -fuse.EISDIR
		case -20:
			out = -fuse.ENOTDIR
		case -39:
			out = -fuse.ENOTEMPTY
		case -36:
			out = -fuse.ENAMETOOLONG
		}
	}

	return out
}
