// Code automtically generated. DO NOT EDIT.
package server

import(
	"github.com/billziss-gh/cgofuse/fuse"
)

//Receiver receives RPC requests and returns results
type Receiver struct {
	fs FS
}

type AccessArgs struct {
	Path string
	Mask uint32
	
}

type AccessReply struct {
	Args *AccessArgs
	R0 int
	
}
func (rcvr *Receiver) Access(a *AccessArgs, r *AccessReply) (err error) {
	r.R0  = rcvr.fs.Access(a.Path , a.Mask )
	r.Args = a
	return
}

type ChflagsArgs struct {
	Path string
	Flags uint32
	
}

type ChflagsReply struct {
	Args *ChflagsArgs
	R0 int
	
}
func (rcvr *Receiver) Chflags(a *ChflagsArgs, r *ChflagsReply) (err error) {
	r.R0  = rcvr.fs.Chflags(a.Path , a.Flags )
	r.Args = a
	return
}

type ChmodArgs struct {
	Path string
	Mode uint32
	
}

type ChmodReply struct {
	Args *ChmodArgs
	R0 int
	
}
func (rcvr *Receiver) Chmod(a *ChmodArgs, r *ChmodReply) (err error) {
	r.R0  = rcvr.fs.Chmod(a.Path , a.Mode )
	r.Args = a
	return
}

type ChownArgs struct {
	Path string
	Uid uint32
	Gid uint32
	
}

type ChownReply struct {
	Args *ChownArgs
	R0 int
	
}
func (rcvr *Receiver) Chown(a *ChownArgs, r *ChownReply) (err error) {
	r.R0  = rcvr.fs.Chown(a.Path , a.Uid , a.Gid )
	r.Args = a
	return
}

type CreateArgs struct {
	Path string
	Flags int
	Mode uint32
	
}

type CreateReply struct {
	Args *CreateArgs
	R0 int
	R1 uint64
	
}
func (rcvr *Receiver) Create(a *CreateArgs, r *CreateReply) (err error) {
	r.R0 ,r.R1  = rcvr.fs.Create(a.Path , a.Flags , a.Mode )
	r.Args = a
	return
}

type DestroyArgs struct {
	
}

type DestroyReply struct {
	Args *DestroyArgs
	
}
func (rcvr *Receiver) Destroy(a *DestroyArgs, r *DestroyReply) (err error) {
	rcvr.fs.Destroy()
	r.Args = a
	return
}

type FlushArgs struct {
	Path string
	Fh uint64
	
}

type FlushReply struct {
	Args *FlushArgs
	R0 int
	
}
func (rcvr *Receiver) Flush(a *FlushArgs, r *FlushReply) (err error) {
	r.R0  = rcvr.fs.Flush(a.Path , a.Fh )
	r.Args = a
	return
}

type FsyncArgs struct {
	Path string
	Datasync bool
	Fh uint64
	
}

type FsyncReply struct {
	Args *FsyncArgs
	R0 int
	
}
func (rcvr *Receiver) Fsync(a *FsyncArgs, r *FsyncReply) (err error) {
	r.R0  = rcvr.fs.Fsync(a.Path , a.Datasync , a.Fh )
	r.Args = a
	return
}

type FsyncdirArgs struct {
	Path string
	Datasync bool
	Fh uint64
	
}

type FsyncdirReply struct {
	Args *FsyncdirArgs
	R0 int
	
}
func (rcvr *Receiver) Fsyncdir(a *FsyncdirArgs, r *FsyncdirReply) (err error) {
	r.R0  = rcvr.fs.Fsyncdir(a.Path , a.Datasync , a.Fh )
	r.Args = a
	return
}

type GetattrArgs struct {
	Path string
	Stat *fuse.Stat_t
	Fh uint64
	
}

type GetattrReply struct {
	Args *GetattrArgs
	R0 int
	
}
func (rcvr *Receiver) Getattr(a *GetattrArgs, r *GetattrReply) (err error) {
	r.R0  = rcvr.fs.Getattr(a.Path , a.Stat , a.Fh )
	r.Args = a
	return
}

type GetxattrArgs struct {
	Path string
	Name string
	
}

type GetxattrReply struct {
	Args *GetxattrArgs
	R0 int
	R1 []byte
	
}
func (rcvr *Receiver) Getxattr(a *GetxattrArgs, r *GetxattrReply) (err error) {
	r.R0 ,r.R1  = rcvr.fs.Getxattr(a.Path , a.Name )
	r.Args = a
	return
}

type InitArgs struct {
	
}

type InitReply struct {
	Args *InitArgs
	
}
func (rcvr *Receiver) Init(a *InitArgs, r *InitReply) (err error) {
	rcvr.fs.Init()
	r.Args = a
	return
}

type LinkArgs struct {
	Oldpath string
	Newpath string
	
}

type LinkReply struct {
	Args *LinkArgs
	R0 int
	
}
func (rcvr *Receiver) Link(a *LinkArgs, r *LinkReply) (err error) {
	r.R0  = rcvr.fs.Link(a.Oldpath , a.Newpath )
	r.Args = a
	return
}

type ListxattrArgs struct {
	Path string
	Fill func(name string) bool
	
}

type ListxattrReply struct {
	Args *ListxattrArgs
	R0 int
	
}
func (rcvr *Receiver) Listxattr(a *ListxattrArgs, r *ListxattrReply) (err error) {
	r.R0  = rcvr.fs.Listxattr(a.Path , a.Fill )
	r.Args = a
	return
}

type MkdirArgs struct {
	Path string
	Mode uint32
	
}

type MkdirReply struct {
	Args *MkdirArgs
	R0 int
	
}
func (rcvr *Receiver) Mkdir(a *MkdirArgs, r *MkdirReply) (err error) {
	r.R0  = rcvr.fs.Mkdir(a.Path , a.Mode )
	r.Args = a
	return
}

type MknodArgs struct {
	Path string
	Mode uint32
	Dev uint64
	
}

type MknodReply struct {
	Args *MknodArgs
	R0 int
	
}
func (rcvr *Receiver) Mknod(a *MknodArgs, r *MknodReply) (err error) {
	r.R0  = rcvr.fs.Mknod(a.Path , a.Mode , a.Dev )
	r.Args = a
	return
}

type OpenArgs struct {
	Path string
	Flags int
	
}

type OpenReply struct {
	Args *OpenArgs
	R0 int
	R1 uint64
	
}
func (rcvr *Receiver) Open(a *OpenArgs, r *OpenReply) (err error) {
	r.R0 ,r.R1  = rcvr.fs.Open(a.Path , a.Flags )
	r.Args = a
	return
}

type OpendirArgs struct {
	Path string
	
}

type OpendirReply struct {
	Args *OpendirArgs
	R0 int
	R1 uint64
	
}
func (rcvr *Receiver) Opendir(a *OpendirArgs, r *OpendirReply) (err error) {
	r.R0 ,r.R1  = rcvr.fs.Opendir(a.Path )
	r.Args = a
	return
}

type ReadArgs struct {
	Path string
	Buff []byte
	Ofst int64
	Fh uint64
	
}

type ReadReply struct {
	Args *ReadArgs
	R0 int
	
}
func (rcvr *Receiver) Read(a *ReadArgs, r *ReadReply) (err error) {
	r.R0  = rcvr.fs.Read(a.Path , a.Buff , a.Ofst , a.Fh )
	r.Args = a
	return
}

type ReaddirArgs struct {
	Path string
	Fill func(name string, stat *fuse.Stat_t, ofst int64) bool
	Ofst int64
	Fh uint64
	
}

type ReaddirReply struct {
	Args *ReaddirArgs
	R0 int
	
}
func (rcvr *Receiver) Readdir(a *ReaddirArgs, r *ReaddirReply) (err error) {
	r.R0  = rcvr.fs.Readdir(a.Path , a.Fill , a.Ofst , a.Fh )
	r.Args = a
	return
}

type ReadlinkArgs struct {
	Path string
	
}

type ReadlinkReply struct {
	Args *ReadlinkArgs
	R0 int
	R1 string
	
}
func (rcvr *Receiver) Readlink(a *ReadlinkArgs, r *ReadlinkReply) (err error) {
	r.R0 ,r.R1  = rcvr.fs.Readlink(a.Path )
	r.Args = a
	return
}

type ReleaseArgs struct {
	Path string
	Fh uint64
	
}

type ReleaseReply struct {
	Args *ReleaseArgs
	R0 int
	
}
func (rcvr *Receiver) Release(a *ReleaseArgs, r *ReleaseReply) (err error) {
	r.R0  = rcvr.fs.Release(a.Path , a.Fh )
	r.Args = a
	return
}

type ReleasedirArgs struct {
	Path string
	Fh uint64
	
}

type ReleasedirReply struct {
	Args *ReleasedirArgs
	R0 int
	
}
func (rcvr *Receiver) Releasedir(a *ReleasedirArgs, r *ReleasedirReply) (err error) {
	r.R0  = rcvr.fs.Releasedir(a.Path , a.Fh )
	r.Args = a
	return
}

type RemovexattrArgs struct {
	Path string
	Name string
	
}

type RemovexattrReply struct {
	Args *RemovexattrArgs
	R0 int
	
}
func (rcvr *Receiver) Removexattr(a *RemovexattrArgs, r *RemovexattrReply) (err error) {
	r.R0  = rcvr.fs.Removexattr(a.Path , a.Name )
	r.Args = a
	return
}

type RenameArgs struct {
	Oldpath string
	Newpath string
	
}

type RenameReply struct {
	Args *RenameArgs
	R0 int
	
}
func (rcvr *Receiver) Rename(a *RenameArgs, r *RenameReply) (err error) {
	r.R0  = rcvr.fs.Rename(a.Oldpath , a.Newpath )
	r.Args = a
	return
}

type RmdirArgs struct {
	Path string
	
}

type RmdirReply struct {
	Args *RmdirArgs
	R0 int
	
}
func (rcvr *Receiver) Rmdir(a *RmdirArgs, r *RmdirReply) (err error) {
	r.R0  = rcvr.fs.Rmdir(a.Path )
	r.Args = a
	return
}

type SetchgtimeArgs struct {
	Path string
	Tmsp fuse.Timespec
	
}

type SetchgtimeReply struct {
	Args *SetchgtimeArgs
	R0 int
	
}
func (rcvr *Receiver) Setchgtime(a *SetchgtimeArgs, r *SetchgtimeReply) (err error) {
	r.R0  = rcvr.fs.Setchgtime(a.Path , a.Tmsp )
	r.Args = a
	return
}

type SetcrtimeArgs struct {
	Path string
	Tmsp fuse.Timespec
	
}

type SetcrtimeReply struct {
	Args *SetcrtimeArgs
	R0 int
	
}
func (rcvr *Receiver) Setcrtime(a *SetcrtimeArgs, r *SetcrtimeReply) (err error) {
	r.R0  = rcvr.fs.Setcrtime(a.Path , a.Tmsp )
	r.Args = a
	return
}

type SetxattrArgs struct {
	Path string
	Name string
	Value []byte
	Flags int
	
}

type SetxattrReply struct {
	Args *SetxattrArgs
	R0 int
	
}
func (rcvr *Receiver) Setxattr(a *SetxattrArgs, r *SetxattrReply) (err error) {
	r.R0  = rcvr.fs.Setxattr(a.Path , a.Name , a.Value , a.Flags )
	r.Args = a
	return
}

type StatfsArgs struct {
	Path string
	Stat *fuse.Statfs_t
	
}

type StatfsReply struct {
	Args *StatfsArgs
	R0 int
	
}
func (rcvr *Receiver) Statfs(a *StatfsArgs, r *StatfsReply) (err error) {
	r.R0  = rcvr.fs.Statfs(a.Path , a.Stat )
	r.Args = a
	return
}

type SymlinkArgs struct {
	Target string
	Newpath string
	
}

type SymlinkReply struct {
	Args *SymlinkArgs
	R0 int
	
}
func (rcvr *Receiver) Symlink(a *SymlinkArgs, r *SymlinkReply) (err error) {
	r.R0  = rcvr.fs.Symlink(a.Target , a.Newpath )
	r.Args = a
	return
}

type TruncateArgs struct {
	Path string
	Size int64
	Fh uint64
	
}

type TruncateReply struct {
	Args *TruncateArgs
	R0 int
	
}
func (rcvr *Receiver) Truncate(a *TruncateArgs, r *TruncateReply) (err error) {
	r.R0  = rcvr.fs.Truncate(a.Path , a.Size , a.Fh )
	r.Args = a
	return
}

type UnlinkArgs struct {
	Path string
	
}

type UnlinkReply struct {
	Args *UnlinkArgs
	R0 int
	
}
func (rcvr *Receiver) Unlink(a *UnlinkArgs, r *UnlinkReply) (err error) {
	r.R0  = rcvr.fs.Unlink(a.Path )
	r.Args = a
	return
}

type UtimensArgs struct {
	Path string
	Tmsp []fuse.Timespec
	
}

type UtimensReply struct {
	Args *UtimensArgs
	R0 int
	
}
func (rcvr *Receiver) Utimens(a *UtimensArgs, r *UtimensReply) (err error) {
	r.R0  = rcvr.fs.Utimens(a.Path , a.Tmsp )
	r.Args = a
	return
}

type WriteArgs struct {
	Path string
	Buff []byte
	Ofst int64
	Fh uint64
	
}

type WriteReply struct {
	Args *WriteArgs
	R0 int
	
}
func (rcvr *Receiver) Write(a *WriteArgs, r *WriteReply) (err error) {
	r.R0  = rcvr.fs.Write(a.Path , a.Buff , a.Ofst , a.Fh )
	r.Args = a
	return
}


