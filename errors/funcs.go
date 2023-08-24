package errors

func Is(err error, target *Error) bool {
	if err == nil {
		return false
	}
	switch e := err.(type) {
	case *Error:
		return e.Code == target.Code
	default:
		return err.Error() == target.Error()
	}
}
