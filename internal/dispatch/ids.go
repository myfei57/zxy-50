package dispatch

import (
	"time"

	"github.com/google/uuid"
)

func newID() string {
	return uuid.NewString()
}

func now() time.Time {
	return time.Now()
}
