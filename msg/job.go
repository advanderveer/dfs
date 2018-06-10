package msg

type Data struct {
	Dest string
}

type Task struct {
	Image   string
	Command []string
	Data    map[string]Data `hcl:"data"`
}

type Job struct {
	Workspace string
	Tasks     map[string]Task `hcl:"task"`
}
