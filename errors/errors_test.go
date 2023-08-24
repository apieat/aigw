package errors

import (
	"errors"
	"testing"
)

func TestErrorEqual(t *testing.T) {
	var err = errors.New("[1001]invalid response")
	if !Is(err, ErrorInvalidResponse) {
		t.Error("error should be equal")
	}
}
