//nolint
package progress

import (
	"context"

	"github.com/containerd/console"
	"github.com/moby/buildkit/client"
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
	opts ...progressui.DisplayOpt,
) (progresswriter.Writer, error) {
	statusCh := make(chan *client.SolveStatus)
	doneCh := make(chan struct{})

	pw := &printer{
		status: statusCh,
		done:   doneCh,
	}

	d, err := progressui.NewDisplay(out, progressui.DisplayMode(mode), opts...)
	if err != nil {
		return nil, err
	}

	go func() {
		// not using shared context to not disrupt display but let it finish reporting errors
		_, pw.err = d.UpdateFrom(ctx, statusCh)
		close(doneCh)
	}()
	return pw, nil
}
