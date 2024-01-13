package progresscatcher

import (
	"github.com/ispringtech/brewkit/internal/backend/app/progress/progresslabel"
)

func MakeCatchLabelf(key, format string, a ...any) (string, error) {
	return progresslabel.MakePayloadLabel(
		map[string]string{
			progresslabel.CatchOutputLabel: "",
			logKey:                         key,
		},
		format,
		a...,
	)
}
