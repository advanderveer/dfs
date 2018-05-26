package fsrpc

// 38 fuse.ENOSYS
// 95 fuse.ENOTSUP
// 61 fuse.ENOATTR
// 28 fuse.ENOSPC
// 2 fuse.ENOENT
// 17 fuse.EEXIST
// 2 fuse.ENOENT
// 22 fuse.EINVAL
// 34 fuse.ERANGE
// 21 fuse.EISDIR
// 20 fuse.ENOTDIR
// 39 fuse.ENOTEMPTY
// 36 fuse.ENAMETOOLONG

//ercc will change the errcode to a platform specific version
//hack for making the error codes portable, probably not really the way to go
//@TODO how can this work for local codes, are they mangled?
func errc(in int) (out int) {
	// iin := in
	// defer func() {
	// 	fmt.Println("IN", iin, "OUT", out)
	// }()

	if in >= 0 {
		return in //not an error code, probably read/write result
	}

	out = in
	// switch in {
	// case -38:
	// 	out = -fuse.ENOSYS
	// case -95:
	// 	out = -fuse.ENOTSUP
	// case -61: //ENODATA or ENOATTR
	// 	out = -fuse.ENOATTR
	// case -28:
	// 	out = -fuse.ENOSPC
	// case -2:
	// 	out = -fuse.ENOENT
	// case -17:
	// 	out = -fuse.EEXIST
	// case -22:
	// 	out = -fuse.EINVAL
	// case -34:
	// 	out = -fuse.ERANGE
	// case -21:
	// 	out = -fuse.EISDIR
	// case -20:
	// 	out = -fuse.ENOTDIR
	// case -39:
	// 	out = -fuse.ENOTEMPTY
	// case -36:
	// 	out = -fuse.ENAMETOOLONG
	// }

	return out

	//
	// //how about: On Mac: Operation not supported on socket
	//
	// switch in {
	// case -1:
	// 	out = -fuse.EPERM /* Operation not permitted */
	// case -2:
	// 	out = -fuse.ENOENT /* No such file or directory */
	// case -3:
	// 	out = -fuse.ESRCH /* No such process */
	// case -4:
	// 	out = -fuse.EINTR /* Interrupted system call */
	// case -5:
	// 	out = -fuse.EIO /* I/O error */
	// case -6:
	// 	out = -fuse.ENXIO /* No such device or address */
	// case -7:
	// 	out = -fuse.E2BIG /* Argument list too long */
	// case -8:
	// 	out = -fuse.ENOEXEC /* Exec format error */
	// case -9:
	// 	out = -fuse.EBADF /* Bad file number */
	// case -10:
	// 	out = -fuse.ECHILD /* No child processes */
	// case -11:
	// 	out = -fuse.EAGAIN /* Try again */
	// case -12:
	// 	out = -fuse.ENOMEM /* Out of memory */
	// case -13:
	// 	out = -fuse.EACCES /* Permission denied */
	// case -14:
	// 	out = -fuse.EFAULT /* Bad address */
	// 	// case -15:
	// 	// 	out = -fuse.ENOTBLK /* Block device required */
	// case -16:
	// 	out = -fuse.EBUSY /* Device or resource busy */
	// case -17:
	// 	out = -fuse.EEXIST /* File exists */
	// case -18:
	// 	out = -fuse.EXDEV /* Cross-device link */
	// case -19:
	// 	out = -fuse.ENODEV /* No such device */
	// case -20:
	// 	out = -fuse.ENOTDIR /* Not a directory */
	// case -21:
	// 	out = -fuse.EISDIR /* Is a directory */
	// case -22:
	// 	out = -fuse.EINVAL /* Invalid argument */
	// case -23:
	// 	out = -fuse.ENFILE /* File table overflow */
	// case -24:
	// 	out = -fuse.EMFILE /* Too many open files */
	// case -25:
	// 	out = -fuse.ENOTTY /* Not a typewriter */
	// case -26:
	// 	out = -fuse.ETXTBSY /* Text file busy */
	// case -27:
	// 	out = -fuse.EFBIG /* File too large */
	// case -28:
	// 	out = -fuse.ENOSPC /* No space left on device */
	// case -29:
	// 	out = -fuse.ESPIPE /* Illegal seek */
	// case -30:
	// 	out = -fuse.EROFS /* Read-only file system */
	// case -31:
	// 	out = -fuse.EMLINK /* Too many links */
	// case -32:
	// 	out = -fuse.EPIPE /* Broken pipe */
	// case -33:
	// 	out = -fuse.EDOM /* Math argument out of domain of func */
	// case -34:
	// 	out = -fuse.ERANGE /* Math result not representable */
	// case -35:
	// 	out = -fuse.EDEADLK /* Resource deadlock would occur */
	// case -36:
	// 	out = -fuse.ENAMETOOLONG /* File name too long */
	// case -37:
	// 	out = -fuse.ENOLCK /* No record locks available */
	// case -38:
	// 	out = -fuse.ENOSYS /* Function not implemented */
	// case -39:
	// 	out = -fuse.ENOTEMPTY /* Directory not empty */
	// case -40:
	// 	out = -fuse.ELOOP /* Too many symbolic links encountered */
	// 	// case -EAGAIN:
	// 	// 	out = -fuse.EWOULDBLOCK /* Operation would block */
	// case -42:
	// 	out = -fuse.ENOMSG /* No message of desired type */
	// case -43:
	// 	out = -fuse.EIDRM /* Identifier removed */
	// 	// case -44:
	// 	// 	out = -fuse.ECHRNG /* Channel number out of range */
	// 	// case -45:
	// 	// 	out = -fuse.EL2NSYNC /* Level 2 not synchronized */
	// 	// case -46:
	// 	// 	out = -fuse.EL3HLT /* Level 3 halted */
	// 	// case -47:
	// 	// 	out = -fuse.EL3RST /* Level 3 reset */
	// 	// case -48:
	// 	// 	out = -fuse.ELNRNG /* Link number out of range */
	// 	// case -49:
	// 	// 	out = -fuse.EUNATCH /* Protocol driver not attached */
	// 	// case -50:
	// 	// 	out = -fuse.ENOCSI /* No CSI structure available */
	// 	// case -51:
	// 	// 	out = -fuse.EL2HLT /* Level 2 halted */
	// 	// case -52:
	// 	// 	out = -fuse.EBADE /* Invalid exchange */
	// 	// case -53:
	// 	// 	out = -fuse.EBADR /* Invalid request descriptor */
	// 	// case -54:
	// 	// 	out = -fuse.EXFULL /* Exchange full */
	// 	// case -55:
	// 	// 	out = -fuse.ENOANO /* No anode */
	// 	// case -56:
	// 	// 	out = -fuse.EBADRQC /* Invalid request code */
	// 	// case -57:
	// 	// 	out = -fuse.EBADSLT /* Invalid slot */
	// 	// case -EDEADLK:
	// 	// 	out = -fuse.EDEADLOCK
	// 	// case -59:
	// 	// 	out = -fuse.EBFONT /* Bad font file format */
	// case -60:
	// 	out = -fuse.ENOSTR /* Device not a stream */
	// case -61:
	// 	out = -fuse.ENODATA /* No data available */
	// case -62:
	// 	out = -fuse.ETIME /* Timer expired */
	// case -63:
	// 	out = -fuse.ENOSR /* Out of streams resources */
	// 	// case -64:
	// 	// 	out = -fuse.ENONET /* Machine is not on the network */
	// 	// case -65:
	// 	// 	out = -fuse.ENOPKG /* Package not installed */
	// 	// case -66:
	// 	// 	out = -fuse.EREMOTE /* Object is remote */
	// 	// case -67:
	// 	out = -fuse.ENOLINK /* Link has been severed */
	// 	// case -68:
	// 	// 	out = -fuse.EADV /* Advertise error */
	// 	// case -69:
	// 	// 	out = -fuse.ESRMNT /* Srmount error */
	// 	// case -70:
	// 	// 	out = -fuse.ECOMM /* Communication error on send */
	// case -71:
	// 	out = -fuse.EPROTO /* Protocol error */
	// 	// case -72:
	// 	// 	out = -fuse.EMULTIHOP /* Multihop attempted */
	// 	// case -73:
	// 	// 	out = -fuse.EDOTDOT /* RFS specific error */
	// case -74:
	// 	out = -fuse.EBADMSG /* Not a data message */
	// case -75:
	// 	out = -fuse.EOVERFLOW /* Value too large for defined data type */
	// 	// case -76:
	// 	// 	out = -fuse.ENOTUNIQ /* Name not unique on network */
	// 	// case -77:
	// 	// 	out = -fuse.EBADFD /* File descriptor in bad state */
	// 	// case -78:
	// 	// 	out = -fuse.EREMCHG /* Remote address changed */
	// 	// case -79:
	// 	// 	out = -fuse.ELIBACC /* Can not access a needed shared library */
	// 	// case -80:
	// 	// 	out = -fuse.ELIBBAD /* Accessing a corrupted shared library */
	// 	// case -81:
	// 	// 	out = -fuse.ELIBSCN /* .lib section in a.out corrupted */
	// 	// case -82:
	// 	// 	out = -fuse.ELIBMAX /* Attempting to link in too many shared libraries */
	// 	// case -83:
	// 	// 	out = -fuse.ELIBEXEC /* Cannot exec a shared library directly */
	// case -84:
	// 	out = -fuse.EILSEQ /* Illegal byte sequence */
	// 	// case -85:
	// 	// 	out = -fuse.ERESTART /* Interrupted system call should be restarted */
	// 	// case -86:
	// 	// 	out = -fuse.ESTRPIPE /* Streams pipe error */
	// 	// case -87:
	// 	// 	out = -fuse.EUSERS /* Too many users */
	// case -88:
	// 	out = -fuse.ENOTSOCK /* Socket operation on non-socket */
	// case -89:
	// 	out = -fuse.EDESTADDRREQ /* Destination address required */
	// case -90:
	// 	out = -fuse.EMSGSIZE /* Message too long */
	// case -91:
	// 	out = -fuse.EPROTOTYPE /* Protocol wrong type for socket */
	// case -92:
	// 	out = -fuse.ENOPROTOOPT /* Protocol not available */
	// case -93:
	// 	out = -fuse.EPROTONOSUPPORT /* Protocol not supported */
	// 	// case -94:
	// 	// 	out = -fuse.ESOCKTNOSUPPORT /* Socket type not supported */
	// case -95:
	// 	out = -fuse.ENOTSUP //@TODO fixes the resourcefork being crashy
	// 	// out = -fuse.EOPNOTSUPP /* Operation not supported on transport endpoint */
	// 	// case -96:
	// 	// 	out = -fuse.EPFNOSUPPORT /* Protocol family not supported */
	// case -97:
	// 	out = -fuse.EAFNOSUPPORT /* Address family not supported by protocol */
	// case -98:
	// 	out = -fuse.EADDRINUSE /* Address already in use */
	// case -99:
	// 	out = -fuse.EADDRNOTAVAIL /* Cannot assign requested address */
	// case -100:
	// 	out = -fuse.ENETDOWN /* Network is down */
	// case -101:
	// 	out = -fuse.ENETUNREACH /* Network is unreachable */
	// case -102:
	// 	out = -fuse.ENETRESET /* Network dropped connection because of reset */
	// case -103:
	// 	out = -fuse.ECONNABORTED /* Software caused connection abort */
	// case -104:
	// 	out = -fuse.ECONNRESET /* Connection reset by peer */
	// case -105:
	// 	out = -fuse.ENOBUFS /* No buffer space available */
	// case -106:
	// 	out = -fuse.EISCONN /* Transport endpoint is already connected */
	// case -107:
	// 	out = -fuse.ENOTCONN /* Transport endpoint is not connected */
	// 	// case -108:
	// 	// 	out = -fuse.ESHUTDOWN /* Cannot send after transport endpoint shutdown */
	// 	// case -109:
	// 	// 	out = -fuse.ETOOMANYREFS /* Too many references: cannot splice */
	// case -110:
	// 	out = -fuse.ETIMEDOUT /* Connection timed out */
	// case -111:
	// 	out = -fuse.ECONNREFUSED /* Connection refused */
	// 	// case -112:
	// 	// 	out = -fuse.EHOSTDOWN /* Host is down */
	// case -113:
	// 	out = -fuse.EHOSTUNREACH /* No route to host */
	// case -114:
	// 	out = -fuse.EALREADY /* Operation already in progress */
	// case -115:
	// 	out = -fuse.EINPROGRESS /* Operation now in progress */
	// 	// case -116:
	// 	// 	out = -fuse.ESTALE /* Stale NFS file handle */
	// 	// case -117:
	// 	// 	out = -fuse.EUCLEAN /* Structure needs cleaning */
	// 	// case -118:
	// 	// 	out = -fuse.ENOTNAM /* Not a XENIX named type file */
	// 	// case -119:
	// 	// 	out = -fuse.ENAVAIL /* No XENIX semaphores available */
	// 	// case -120:
	// 	// 	out = -fuse.EISNAM /* Is a named type file */
	// 	// case -121:
	// 	// 	out = -fuse.EREMOTEIO /* Remote I/O error */
	// 	// case -122:
	// 	// 	out = -fuse.EDQUOT /* Quota exceeded */
	// 	// case -123:
	// 	// 	out = -fuse.ENOMEDIUM /* No medium found */
	// 	// case -124:
	// 	// 	out = -fuse.EMEDIUMTYPE /* Wrong medium type */
	// case -125:
	// 	out = -fuse.ECANCELED /* Operation Canceled */
	// 	// case -126:
	// 	// 	out = -fuse.ENOKEY /* Required key not available */
	// 	// case -127:
	// 	// 	out = -fuse.EKEYEXPIRED /* Key has expired */
	// 	// case -128:
	// 	// 	out = -fuse.EKEYREVOKED /* Key has been revoked */
	// 	// case -129:
	// 	// 	out = -fuse.EKEYREJECTED /* Key was rejected by service */
	// case -130:
	// 	out = -fuse.EOWNERDEAD /* Owner died */
	// case -131:
	// 	out = -fuse.ENOTRECOVERABLE /* State not recoverable */
	// default:
	// 	out = in
	// }
	// return out
}
