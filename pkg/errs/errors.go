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

// type ErrRecordingAlreadyMoved struct{}

// func (e *ErrRecordingAlreadyMoved) Error() string {
// 	return "recording has already been moved"
// }

// type ErrTooManyArguments struct{}

// func (e *ErrTooManyArguments) Error() string {
// 	return "too many arguments"
// }
