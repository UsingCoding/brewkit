package progresscatcher

import (
	"github.com/ispringtech/brewkit/internal/backend/app/progress/progresslabel"
)

func MakeCatchLabelPayload(key string) map[string]string {
	return map[string]string{
		progresslabel.CatchOutputLabel: "",
		logKey:                         key,
	}
}
