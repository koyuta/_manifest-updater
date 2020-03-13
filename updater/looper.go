package updater

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"
)

var timeout = 60 * time.Second

type UpdateLooper struct {
	updaters      []Updater
	checkInterval time.Duration
	// logger Logger

	queue <-chan *Updater

	shutdown     chan struct{}
	done         chan struct{}
	shuttingDown *atomic.Value
}

func NewUpdateLooper(queue <-chan *Updater, c time.Duration) *UpdateLooper {
	return &UpdateLooper{
		queue:         queue,
		checkInterval: c,
		shutdown:      make(chan struct{}),
		done:          make(chan struct{}),
		shuttingDown:  &atomic.Value{},
	}
}

func (u *UpdateLooper) Loop(stop <-chan struct{}) error {
	if v := u.shuttingDown.Load(); v != nil {
		return errors.New("Shuting down error")
	}

	ticker := time.NewTicker(u.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case updater := <-u.queue:
			u.updaters = append(u.updaters, *updater)
		case <-u.shutdown:
			close(u.done)
			return nil
		case <-stop:
			return nil
		case <-ticker.C:
			for i := range u.updaters {
				updater := u.updaters[i]
				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), timeout)
					defer cancel()

					var errch = make(chan error, 1)
					errch <- updater.Run(ctx)

					select {
					case <-ctx.Done():
						fmt.Println(ctx.Err())
					case err := <-errch:
						fmt.Println(err)
					}
				}()
			}
		}
	}
}

func (u *UpdateLooper) Shutdown(ctx context.Context) error {
	u.shuttingDown.Store(struct{}{})
	close(u.shutdown)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-u.done:
		u.shuttingDown = &atomic.Value{}
		return nil
	}
}
