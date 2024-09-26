package error

type ClientNotFound struct{}

func NewClientNotFound() *ClientNotFound {
	return &ClientNotFound{}
}

func (e *ClientNotFound) Error() string {
	return "client not found"
}
