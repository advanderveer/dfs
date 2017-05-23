package ddfs

func (fs *FS) synchronize() func() {
	fs.lock.Lock()
	return func() {
		fs.lock.Unlock()
	}
}
