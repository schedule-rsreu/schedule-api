package repo

import (
	"errors"
	"fmt"
)

var ErrNoResults = errors.New("no results")

type ErrNoScheduleGroup struct {
	ParamName string
}

// Error implements the error interface.
func (e ErrNoScheduleGroup) Error() string {
	return fmt.Sprintf("schedule for group %v not found", e.ParamName)
}
