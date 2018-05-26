// Code automtically generated. DO NOT EDIT.
package fsrpc

import (
	"fmt"
	"github.com/billziss-gh/cgofuse/fuse"
)

type ReaddirCall struct {
	Name string
	Stat *fuse.Stat_t
	Ofst int64
}

type ListxattrCall struct {
	Name string
}

type AccessArgs struct {
	Path string
	Mask uint32
}

type AccessReply struct {
	Args *AccessArgs
	R0   int
}

func (rcvr *Receiver) Access(a *AccessArgs, r *AccessReply) (err error) {

	r.R0 = rcvr.fs.Access(a.Path, a.Mask)
	r.Args = a
	return
}

func (sndr *Sender) Access(path string, mask uint32) int {
	r := &AccessReply{}
	a := &AccessArgs{
		Path: path,
		Mask: mask,
	}

	sndr.LastErr = sndr.rpc.Call("FS.Access", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	return errc(r.R0)
}

type ChflagsArgs struct {
	Path  string
	Flags uint32
}

type ChflagsReply struct {
	Args *ChflagsArgs
	R0   int
}

func (rcvr *Receiver) Chflags(a *ChflagsArgs, r *ChflagsReply) (err error) {

	r.R0 = rcvr.fs.Chflags(a.Path, a.Flags)
	r.Args = a
	return
}

func (sndr *Sender) Chflags(path string, flags uint32) int {
	r := &ChflagsReply{}
	a := &ChflagsArgs{
		Path:  path,
		Flags: flags,
	}

	sndr.LastErr = sndr.rpc.Call("FS.Chflags", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	return errc(r.R0)
}

type ChmodArgs struct {
	Path string
	Mode uint32
}

type ChmodReply struct {
	Args *ChmodArgs
	R0   int
}

func (rcvr *Receiver) Chmod(a *ChmodArgs, r *ChmodReply) (err error) {

	r.R0 = rcvr.fs.Chmod(a.Path, a.Mode)
	r.Args = a
	return
}

func (sndr *Sender) Chmod(path string, mode uint32) int {
	r := &ChmodReply{}
	a := &ChmodArgs{
		Path: path,
		Mode: mode,
	}

	sndr.LastErr = sndr.rpc.Call("FS.Chmod", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	return errc(r.R0)
}

type ChownArgs struct {
	Path string
	Uid  uint32
	Gid  uint32
}

type ChownReply struct {
	Args *ChownArgs
	R0   int
}

func (rcvr *Receiver) Chown(a *ChownArgs, r *ChownReply) (err error) {

	r.R0 = rcvr.fs.Chown(a.Path, a.Uid, a.Gid)
	r.Args = a
	return
}

func (sndr *Sender) Chown(path string, uid uint32, gid uint32) int {
	r := &ChownReply{}
	a := &ChownArgs{
		Path: path,
		Uid:  uid,
		Gid:  gid,
	}

	sndr.LastErr = sndr.rpc.Call("FS.Chown", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	return errc(r.R0)
}

type CreateArgs struct {
	Path  string
	Flags int
	Mode  uint32
}

type CreateReply struct {
	Args *CreateArgs
	R0   int
	R1   uint64
}

func (rcvr *Receiver) Create(a *CreateArgs, r *CreateReply) (err error) {

	r.R0, r.R1 = rcvr.fs.Create(a.Path, a.Flags, a.Mode)
	r.Args = a
	return
}

func (sndr *Sender) Create(path string, flags int, mode uint32) (int, uint64) {
	r := &CreateReply{}
	a := &CreateArgs{
		Path:  path,
		Flags: flags,
		Mode:  mode,
	}

	sndr.LastErr = sndr.rpc.Call("FS.Create", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	return errc(r.R0), r.R1
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

func (sndr *Sender) Destroy() {
	r := &struct{}{}
	a := &DestroyArgs{}

	sndr.LastErr = sndr.rpc.Call("FS.Destroy", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	return
}

type FlushArgs struct {
	Path string
	Fh   uint64
}

type FlushReply struct {
	Args *FlushArgs
	R0   int
}

func (rcvr *Receiver) Flush(a *FlushArgs, r *FlushReply) (err error) {

	r.R0 = rcvr.fs.Flush(a.Path, a.Fh)
	r.Args = a
	return
}

func (sndr *Sender) Flush(path string, fh uint64) int {
	r := &FlushReply{}
	a := &FlushArgs{
		Path: path,
		Fh:   fh,
	}

	sndr.LastErr = sndr.rpc.Call("FS.Flush", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	return errc(r.R0)
}

type FsyncArgs struct {
	Path     string
	Datasync bool
	Fh       uint64
}

type FsyncReply struct {
	Args *FsyncArgs
	R0   int
}

func (rcvr *Receiver) Fsync(a *FsyncArgs, r *FsyncReply) (err error) {

	r.R0 = rcvr.fs.Fsync(a.Path, a.Datasync, a.Fh)
	r.Args = a
	return
}

func (sndr *Sender) Fsync(path string, datasync bool, fh uint64) int {
	r := &FsyncReply{}
	a := &FsyncArgs{
		Path:     path,
		Datasync: datasync,
		Fh:       fh,
	}

	sndr.LastErr = sndr.rpc.Call("FS.Fsync", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	return errc(r.R0)
}

type FsyncdirArgs struct {
	Path     string
	Datasync bool
	Fh       uint64
}

type FsyncdirReply struct {
	Args *FsyncdirArgs
	R0   int
}

func (rcvr *Receiver) Fsyncdir(a *FsyncdirArgs, r *FsyncdirReply) (err error) {

	r.R0 = rcvr.fs.Fsyncdir(a.Path, a.Datasync, a.Fh)
	r.Args = a
	return
}

func (sndr *Sender) Fsyncdir(path string, datasync bool, fh uint64) int {
	r := &FsyncdirReply{}
	a := &FsyncdirArgs{
		Path:     path,
		Datasync: datasync,
		Fh:       fh,
	}

	sndr.LastErr = sndr.rpc.Call("FS.Fsyncdir", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	return errc(r.R0)
}

type GetattrArgs struct {
	Path string
	Stat *fuse.Stat_t
	Fh   uint64
}

type GetattrReply struct {
	Args *GetattrArgs
	R0   int
}

func (rcvr *Receiver) Getattr(a *GetattrArgs, r *GetattrReply) (err error) {

	r.R0 = rcvr.fs.Getattr(a.Path, a.Stat, a.Fh)
	r.Args = a
	return
}

func (sndr *Sender) Getattr(path string, stat *fuse.Stat_t, fh uint64) int {
	r := &GetattrReply{}
	a := &GetattrArgs{
		Path: path,
		Stat: stat,
		Fh:   fh,
	}

	sndr.LastErr = sndr.rpc.Call("FS.Getattr", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	*stat = *r.Args.Stat

	stat.Uid = sndr.uid
	stat.Gid = sndr.gid

	return errc(r.R0)
}

type GetxattrArgs struct {
	Path string
	Name string
}

type GetxattrReply struct {
	Args *GetxattrArgs
	R0   int
	R1   []byte
}

func (rcvr *Receiver) Getxattr(a *GetxattrArgs, r *GetxattrReply) (err error) {

	r.R0, r.R1 = rcvr.fs.Getxattr(a.Path, a.Name)
	r.Args = a
	return
}

func (sndr *Sender) Getxattr(path string, name string) (int, []byte) {
	r := &GetxattrReply{}
	a := &GetxattrArgs{
		Path: path,
		Name: name,
	}

	sndr.LastErr = sndr.rpc.Call("FS.Getxattr", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	return errc(r.R0), r.R1
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

func (sndr *Sender) Init() {
	r := &struct{}{}
	a := &InitArgs{}

	sndr.LastErr = sndr.rpc.Call("FS.Init", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	return
}

type LinkArgs struct {
	Oldpath string
	Newpath string
}

type LinkReply struct {
	Args *LinkArgs
	R0   int
}

func (rcvr *Receiver) Link(a *LinkArgs, r *LinkReply) (err error) {

	r.R0 = rcvr.fs.Link(a.Oldpath, a.Newpath)
	r.Args = a
	return
}

func (sndr *Sender) Link(oldpath string, newpath string) int {
	r := &LinkReply{}
	a := &LinkArgs{
		Oldpath: oldpath,
		Newpath: newpath,
	}

	sndr.LastErr = sndr.rpc.Call("FS.Link", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	return errc(r.R0)
}

type ListxattrArgs struct {
	Path string
	Fill func(name string) bool
}

type ListxattrReply struct {
	Args *ListxattrArgs
	R0   int

	Fills []ListxattrCall
}

func (rcvr *Receiver) Listxattr(a *ListxattrArgs, r *ListxattrReply) (err error) {

	a.Fill = func(name string) bool {
		r.Fills = append(r.Fills, ListxattrCall{Name: name})
		return true
	}
	r.R0 = rcvr.fs.Listxattr(a.Path, a.Fill)
	r.Args = a
	return
}

func (sndr *Sender) Listxattr(path string, fill func(name string) bool) int {
	r := &ListxattrReply{}
	a := &ListxattrArgs{
		Path: path,
		Fill: fill,
	}

	sndr.LastErr = sndr.rpc.Call("FS.Listxattr", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	for _, c := range r.Fills {
		if !fill(c.Name) {
			break
		}
	}

	return errc(r.R0)
}

type MkdirArgs struct {
	Path string
	Mode uint32
}

type MkdirReply struct {
	Args *MkdirArgs
	R0   int
}

func (rcvr *Receiver) Mkdir(a *MkdirArgs, r *MkdirReply) (err error) {

	r.R0 = rcvr.fs.Mkdir(a.Path, a.Mode)
	r.Args = a
	return
}

func (sndr *Sender) Mkdir(path string, mode uint32) int {
	r := &MkdirReply{}
	a := &MkdirArgs{
		Path: path,
		Mode: mode,
	}

	sndr.LastErr = sndr.rpc.Call("FS.Mkdir", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	return errc(r.R0)
}

type MknodArgs struct {
	Path string
	Mode uint32
	Dev  uint64
}

type MknodReply struct {
	Args *MknodArgs
	R0   int
}

func (rcvr *Receiver) Mknod(a *MknodArgs, r *MknodReply) (err error) {

	r.R0 = rcvr.fs.Mknod(a.Path, a.Mode, a.Dev)
	r.Args = a
	return
}

func (sndr *Sender) Mknod(path string, mode uint32, dev uint64) int {
	r := &MknodReply{}
	a := &MknodArgs{
		Path: path,
		Mode: mode,
		Dev:  dev,
	}

	sndr.LastErr = sndr.rpc.Call("FS.Mknod", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	return errc(r.R0)
}

type OpenArgs struct {
	Path  string
	Flags int
}

type OpenReply struct {
	Args *OpenArgs
	R0   int
	R1   uint64
}

func (rcvr *Receiver) Open(a *OpenArgs, r *OpenReply) (err error) {

	r.R0, r.R1 = rcvr.fs.Open(a.Path, a.Flags)
	r.Args = a
	return
}

func (sndr *Sender) Open(path string, flags int) (int, uint64) {
	r := &OpenReply{}
	a := &OpenArgs{
		Path:  path,
		Flags: flags,
	}

	sndr.LastErr = sndr.rpc.Call("FS.Open", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	return errc(r.R0), r.R1
}

type OpendirArgs struct {
	Path string
}

type OpendirReply struct {
	Args *OpendirArgs
	R0   int
	R1   uint64
}

func (rcvr *Receiver) Opendir(a *OpendirArgs, r *OpendirReply) (err error) {

	r.R0, r.R1 = rcvr.fs.Opendir(a.Path)
	r.Args = a
	return
}

func (sndr *Sender) Opendir(path string) (int, uint64) {
	r := &OpendirReply{}
	a := &OpendirArgs{
		Path: path,
	}

	sndr.LastErr = sndr.rpc.Call("FS.Opendir", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	return errc(r.R0), r.R1
}

type ReadArgs struct {
	Path string
	Buff []byte
	Ofst int64
	Fh   uint64
}

type ReadReply struct {
	Args *ReadArgs
	R0   int
}

func (rcvr *Receiver) Read(a *ReadArgs, r *ReadReply) (err error) {

	r.R0 = rcvr.fs.Read(a.Path, a.Buff, a.Ofst, a.Fh)
	r.Args = a
	return
}

func (sndr *Sender) Read(path string, buff []byte, ofst int64, fh uint64) int {
	r := &ReadReply{}
	a := &ReadArgs{
		Path: path,
		Buff: buff,
		Ofst: ofst,
		Fh:   fh,
	}

	sndr.LastErr = sndr.rpc.Call("FS.Read", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	copy(buff, r.Args.Buff)

	return errc(r.R0)
}

type ReaddirArgs struct {
	Path string
	Fill func(name string, stat *fuse.Stat_t, ofst int64) bool
	Ofst int64
	Fh   uint64
}

type ReaddirReply struct {
	Args *ReaddirArgs
	R0   int

	Fills []ReaddirCall
}

func (rcvr *Receiver) Readdir(a *ReaddirArgs, r *ReaddirReply) (err error) {
	a.Fill = func(name string, stat *fuse.Stat_t, ofst int64) bool {
		r.Fills = append(r.Fills, ReaddirCall{Name: name, Stat: stat, Ofst: ofst})
		return true
	}
	r.R0 = rcvr.fs.Readdir(a.Path, a.Fill, a.Ofst, a.Fh)
	r.Args = a
	return
}

func (sndr *Sender) Readdir(path string, fill func(name string, stat *fuse.Stat_t, ofst int64) bool, ofst int64, fh uint64) int {
	r := &ReaddirReply{}
	a := &ReaddirArgs{
		Path: path,
		Fill: fill,
		Ofst: ofst,
		Fh:   fh,
	}

	sndr.LastErr = sndr.rpc.Call("FS.Readdir", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	for _, c := range r.Fills {
		if c.Stat != nil {
			c.Stat.Uid = sndr.uid
			c.Stat.Gid = sndr.gid
		}
		if !fill(c.Name, c.Stat, c.Ofst) {
			break
		}
	}

	return errc(r.R0)
}

type ReadlinkArgs struct {
	Path string
}

type ReadlinkReply struct {
	Args *ReadlinkArgs
	R0   int
	R1   string
}

func (rcvr *Receiver) Readlink(a *ReadlinkArgs, r *ReadlinkReply) (err error) {

	r.R0, r.R1 = rcvr.fs.Readlink(a.Path)
	r.Args = a
	return
}

func (sndr *Sender) Readlink(path string) (int, string) {
	r := &ReadlinkReply{}
	a := &ReadlinkArgs{
		Path: path,
	}

	sndr.LastErr = sndr.rpc.Call("FS.Readlink", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	return errc(r.R0), r.R1
}

type ReleaseArgs struct {
	Path string
	Fh   uint64
}

type ReleaseReply struct {
	Args *ReleaseArgs
	R0   int
}

func (rcvr *Receiver) Release(a *ReleaseArgs, r *ReleaseReply) (err error) {

	r.R0 = rcvr.fs.Release(a.Path, a.Fh)
	r.Args = a
	return
}

func (sndr *Sender) Release(path string, fh uint64) int {
	r := &ReleaseReply{}
	a := &ReleaseArgs{
		Path: path,
		Fh:   fh,
	}

	sndr.LastErr = sndr.rpc.Call("FS.Release", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	return errc(r.R0)
}

type ReleasedirArgs struct {
	Path string
	Fh   uint64
}

type ReleasedirReply struct {
	Args *ReleasedirArgs
	R0   int
}

func (rcvr *Receiver) Releasedir(a *ReleasedirArgs, r *ReleasedirReply) (err error) {

	r.R0 = rcvr.fs.Releasedir(a.Path, a.Fh)
	r.Args = a
	return
}

func (sndr *Sender) Releasedir(path string, fh uint64) int {
	r := &ReleasedirReply{}
	a := &ReleasedirArgs{
		Path: path,
		Fh:   fh,
	}

	sndr.LastErr = sndr.rpc.Call("FS.Releasedir", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	return errc(r.R0)
}

type RemovexattrArgs struct {
	Path string
	Name string
}

type RemovexattrReply struct {
	Args *RemovexattrArgs
	R0   int
}

func (rcvr *Receiver) Removexattr(a *RemovexattrArgs, r *RemovexattrReply) (err error) {

	r.R0 = rcvr.fs.Removexattr(a.Path, a.Name)
	r.Args = a
	return
}

func (sndr *Sender) Removexattr(path string, name string) int {
	r := &RemovexattrReply{}
	a := &RemovexattrArgs{
		Path: path,
		Name: name,
	}

	sndr.LastErr = sndr.rpc.Call("FS.Removexattr", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	return errc(r.R0)
}

type RenameArgs struct {
	Oldpath string
	Newpath string
}

type RenameReply struct {
	Args *RenameArgs
	R0   int
}

func (rcvr *Receiver) Rename(a *RenameArgs, r *RenameReply) (err error) {

	r.R0 = rcvr.fs.Rename(a.Oldpath, a.Newpath)
	r.Args = a
	return
}

func (sndr *Sender) Rename(oldpath string, newpath string) int {
	r := &RenameReply{}
	a := &RenameArgs{
		Oldpath: oldpath,
		Newpath: newpath,
	}

	sndr.LastErr = sndr.rpc.Call("FS.Rename", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	return errc(r.R0)
}

type RmdirArgs struct {
	Path string
}

type RmdirReply struct {
	Args *RmdirArgs
	R0   int
}

func (rcvr *Receiver) Rmdir(a *RmdirArgs, r *RmdirReply) (err error) {

	r.R0 = rcvr.fs.Rmdir(a.Path)
	r.Args = a
	return
}

func (sndr *Sender) Rmdir(path string) int {
	r := &RmdirReply{}
	a := &RmdirArgs{
		Path: path,
	}

	sndr.LastErr = sndr.rpc.Call("FS.Rmdir", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	return errc(r.R0)
}

type SetchgtimeArgs struct {
	Path string
	Tmsp fuse.Timespec
}

type SetchgtimeReply struct {
	Args *SetchgtimeArgs
	R0   int
}

func (rcvr *Receiver) Setchgtime(a *SetchgtimeArgs, r *SetchgtimeReply) (err error) {

	r.R0 = rcvr.fs.Setchgtime(a.Path, a.Tmsp)
	r.Args = a
	return
}

func (sndr *Sender) Setchgtime(path string, tmsp fuse.Timespec) int {
	r := &SetchgtimeReply{}
	a := &SetchgtimeArgs{
		Path: path,
		Tmsp: tmsp,
	}

	sndr.LastErr = sndr.rpc.Call("FS.Setchgtime", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	return errc(r.R0)
}

type SetcrtimeArgs struct {
	Path string
	Tmsp fuse.Timespec
}

type SetcrtimeReply struct {
	Args *SetcrtimeArgs
	R0   int
}

func (rcvr *Receiver) Setcrtime(a *SetcrtimeArgs, r *SetcrtimeReply) (err error) {

	r.R0 = rcvr.fs.Setcrtime(a.Path, a.Tmsp)
	r.Args = a
	return
}

func (sndr *Sender) Setcrtime(path string, tmsp fuse.Timespec) int {
	r := &SetcrtimeReply{}
	a := &SetcrtimeArgs{
		Path: path,
		Tmsp: tmsp,
	}

	sndr.LastErr = sndr.rpc.Call("FS.Setcrtime", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	return errc(r.R0)
}

type SetxattrArgs struct {
	Path  string
	Name  string
	Value []byte
	Flags int
}

type SetxattrReply struct {
	Args *SetxattrArgs
	R0   int
}

func (rcvr *Receiver) Setxattr(a *SetxattrArgs, r *SetxattrReply) (err error) {

	r.R0 = rcvr.fs.Setxattr(a.Path, a.Name, a.Value, a.Flags)
	r.Args = a
	return
}

func (sndr *Sender) Setxattr(path string, name string, value []byte, flags int) int {
	r := &SetxattrReply{}
	a := &SetxattrArgs{
		Path:  path,
		Name:  name,
		Value: value,
		Flags: flags,
	}

	sndr.LastErr = sndr.rpc.Call("FS.Setxattr", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	return errc(r.R0)
}

type StatfsArgs struct {
	Path string
	Stat *fuse.Statfs_t
}

type StatfsReply struct {
	Args *StatfsArgs
	R0   int
}

func (rcvr *Receiver) Statfs(a *StatfsArgs, r *StatfsReply) (err error) {

	r.R0 = rcvr.fs.Statfs(a.Path, a.Stat)
	r.Args = a
	return
}

func (sndr *Sender) Statfs(path string, stat *fuse.Statfs_t) int {
	r := &StatfsReply{}
	a := &StatfsArgs{
		Path: path,
		Stat: stat,
	}

	sndr.LastErr = sndr.rpc.Call("FS.Statfs", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	*stat = *r.Args.Stat

	return errc(r.R0)
}

type SymlinkArgs struct {
	Target  string
	Newpath string
}

type SymlinkReply struct {
	Args *SymlinkArgs
	R0   int
}

func (rcvr *Receiver) Symlink(a *SymlinkArgs, r *SymlinkReply) (err error) {

	r.R0 = rcvr.fs.Symlink(a.Target, a.Newpath)
	r.Args = a
	return
}

func (sndr *Sender) Symlink(target string, newpath string) int {
	r := &SymlinkReply{}
	a := &SymlinkArgs{
		Target:  target,
		Newpath: newpath,
	}

	sndr.LastErr = sndr.rpc.Call("FS.Symlink", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	return errc(r.R0)
}

type TruncateArgs struct {
	Path string
	Size int64
	Fh   uint64
}

type TruncateReply struct {
	Args *TruncateArgs
	R0   int
}

func (rcvr *Receiver) Truncate(a *TruncateArgs, r *TruncateReply) (err error) {

	r.R0 = rcvr.fs.Truncate(a.Path, a.Size, a.Fh)
	r.Args = a
	return
}

func (sndr *Sender) Truncate(path string, size int64, fh uint64) int {
	r := &TruncateReply{}
	a := &TruncateArgs{
		Path: path,
		Size: size,
		Fh:   fh,
	}

	sndr.LastErr = sndr.rpc.Call("FS.Truncate", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	return errc(r.R0)
}

type UnlinkArgs struct {
	Path string
}

type UnlinkReply struct {
	Args *UnlinkArgs
	R0   int
}

func (rcvr *Receiver) Unlink(a *UnlinkArgs, r *UnlinkReply) (err error) {

	r.R0 = rcvr.fs.Unlink(a.Path)
	r.Args = a
	return
}

func (sndr *Sender) Unlink(path string) int {
	r := &UnlinkReply{}
	a := &UnlinkArgs{
		Path: path,
	}

	sndr.LastErr = sndr.rpc.Call("FS.Unlink", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	return errc(r.R0)
}

type UtimensArgs struct {
	Path string
	Tmsp []fuse.Timespec
}

type UtimensReply struct {
	Args *UtimensArgs
	R0   int
}

func (rcvr *Receiver) Utimens(a *UtimensArgs, r *UtimensReply) (err error) {

	r.R0 = rcvr.fs.Utimens(a.Path, a.Tmsp)
	r.Args = a
	return
}

func (sndr *Sender) Utimens(path string, tmsp []fuse.Timespec) int {
	r := &UtimensReply{}
	a := &UtimensArgs{
		Path: path,
		Tmsp: tmsp,
	}

	sndr.LastErr = sndr.rpc.Call("FS.Utimens", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	return errc(r.R0)
}

type WriteArgs struct {
	Path string
	Buff []byte
	Ofst int64
	Fh   uint64
}

type WriteReply struct {
	Args *WriteArgs
	R0   int
}

func (rcvr *Receiver) Write(a *WriteArgs, r *WriteReply) (err error) {

	r.R0 = rcvr.fs.Write(a.Path, a.Buff, a.Ofst, a.Fh)
	r.Args = a
	return
}

func (sndr *Sender) Write(path string, buff []byte, ofst int64, fh uint64) int {
	r := &WriteReply{}
	a := &WriteArgs{
		Path: path,
		Buff: buff,
		Ofst: ofst,
		Fh:   fh,
	}

	sndr.LastErr = sndr.rpc.Call("FS.Write", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	return errc(r.R0)
}
