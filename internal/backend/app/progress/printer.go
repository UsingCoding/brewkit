//nolint
package progress

import (
	"context"
	"os"

	"github.com/containerd/console"
	"github.com/moby/buildkit/client"
	"github.com/pkg/errors"

	`github.com/moby/buildkit/util/progress/progresswriter`

	`github.com/ispringtech/brewkit/internal/backend/app/progress/progressui`
)

type printer struct {
	status chan *client.SolveStatus
	done   <-chan struct{}
	err    error
}

func (p *printer) Done() <-chan struct{} {
	return p.done
}

func (p *printer) Err() error {
	return p.err
}

func (p *printer) Status() chan *client.SolveStatus {
	if p == nil {
		return nil
	}
	return p.status
}

func NewPrinter(
	ctx context.Context,
	out console.File,
	mode string,
	opts ...progressui.DisplaySolveStatusOpt,
) (progresswriter.Writer, error) {
	statusCh := make(chan *client.SolveStatus)
	doneCh := make(chan struct{})

	pw := &printer{
		status: statusCh,
		done:   doneCh,
	}

	if v := os.Getenv("BUILDKIT_PROGRESS"); v != "" && mode == "auto" {
		mode = v
	}

	var c console.Console
	switch mode {
	case "auto", "tty", "":
		if cons, err := console.ConsoleFromFile(out); err == nil {
			c = cons
		} else {
			if mode == "tty" {
				return nil, errors.Wrap(err, "failed to get console")
			}
		}
	case "plain":
	default:
		return nil, errors.Errorf("invalid progress mode %s", mode)
	}

	go func() {
		_, pw.err = progressui.DisplaySolveStatus(ctx, c, out, statusCh, opts...)
		close(doneCh)
	}()
	return pw, nil
}
