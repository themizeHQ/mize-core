package app_errors

import (
	"fmt"
)

type RequestError struct {
	StatusCode int
	Err        error
}

func (err RequestError) Error() string {
	return fmt.Sprintf("statusCode %d : error %v", err.StatusCode, err.Err)
}
