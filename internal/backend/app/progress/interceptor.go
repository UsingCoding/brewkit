package progress

import (
	stderrors "errors"

	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/util/progress/progresswriter"
)

type interceptor struct {
	status chan *client.SolveStatus
	done   <-chan struct{}
	err    error
}

func (i *interceptor) Done() <-chan struct{} {
	return i.done
}

func (i *interceptor) Err() error {
	return i.err
}

func (i *interceptor) Status() chan *client.SolveStatus {
	return i.status
}

type InterceptFunc func(s *client.SolveStatus) error

func Intercept(w progresswriter.Writer, f InterceptFunc) progresswriter.Writer {
	statusCh := make(chan *client.SolveStatus)
	doneCh := make(chan struct{})

	pw := &interceptor{
		status: statusCh,
		done:   doneCh,
	}

	go func() {
		for {
			select {
			case s, ok := <-statusCh:
				if !ok {
					close(w.Status())

					<-w.Done()
					close(doneCh)
					pw.err = stderrors.Join(pw.err, w.Err())
					return
				}

				err := f(s)
				if err != nil {
					pw.err = err
					close(statusCh)
					continue
				}
				w.Status() <- s
			case <-w.Done():
				close(doneCh)
				pw.err = stderrors.Join(pw.err, w.Err())
				return
			}
		}
	}()

	return pw
}
