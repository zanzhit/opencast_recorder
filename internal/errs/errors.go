package errs

type ErrNoRecording struct{}

func (e *ErrNoRecording) Error() string {
	return "recording not found"
}

type BadRequst struct {
	Message string
}

func (e *BadRequst) Error() string {
	return e.Message
}
