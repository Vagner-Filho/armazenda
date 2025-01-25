package model_error

type ModelError struct {
	IsServerErr bool
	Message     string
}

func (e *ModelError) Error() string {
	return e.Message
}

func (e *ModelError) FromServer() bool {
	return e.IsServerErr
}
