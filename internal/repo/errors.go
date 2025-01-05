package repo

import (
	"errors"
	"fmt"
)

var ErrNoResults = errors.New("no results")

type NoScheduleGroupError struct {
	ParamName string
}

// Error implements the error interface.
func (e NoScheduleGroupError) Error() string {
	return fmt.Sprintf("schedule for group %v not found", e.ParamName)
}
