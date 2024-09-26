package health

type Status string

const (
	StatusUp      Status = "up"
	StatusDown    Status = "down"
	StatusTimeout Status = "timeout"
)
