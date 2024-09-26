package error

type ClientAlreadyExists struct{}

func NewClientAlreadyExists() *ClientAlreadyExists {
	return &ClientAlreadyExists{}
}

func (e *ClientAlreadyExists) Error() string {
	return "client already exists"
}
