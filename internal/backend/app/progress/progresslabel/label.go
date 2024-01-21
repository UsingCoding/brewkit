package progresslabel

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

const (
	// HiddenLabel - internal label used to exclude logs of specific stages
	HiddenLabel      = "[HIDDEN]"
	CatchOutputLabel = "[CATCH]"

	InternalLabel = "[internal]"
)

const (
	separator        = " "
	payloadSeparator = " | "
)

func MakeLabelf(l, format string, a ...any) string {
	return l + separator + fmt.Sprintf(format, a...)
}

func MakePayloadLabel(payload map[string]string, text string) (string, error) {
	p, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	return string(p) + payloadSeparator + text, nil
}

func ParseLabel(s string) (label, text string) {
	parts := strings.SplitN(s, separator, 2)
	if len(parts) != 2 {
		return "", ""
	}

	return parts[0], parts[1]
}

func ParsePayloadLabel(s string) (p map[string]string, text string, err error) {
	parts := strings.SplitN(s, payloadSeparator, 2)
	if len(parts) != 2 {
		return nil, "", nil
	}

	err = json.Unmarshal([]byte(parts[0]), &p)
	if err != nil {
		return nil, "", errors.Wrapf(err, "failed to unmarshall label payload %s", parts[0])
	}

	return p, parts[1], nil
}
