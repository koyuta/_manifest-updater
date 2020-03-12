package main

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"
)

var timeout = 60 * time.Second

type UpdateLooper struct {
	updater       *Updater
	checkInterval time.Duration
	// logger Logger

	shutdown chan struct{}
	done     chan struct{}

	shuttingDown atomic.Value
}

func NewUpdateLooper(u *Updater, c time.Duration) *UpdateLooper {
	return &UpdateLooper{updater: u, checkInterval: c}
}

func (u *UpdateLooper) Loop(stop <-chan struct{}) error {
	if v := u.shuttingDown.Load(); v != nil {
		return errors.New("Shuting down error")
	}

	ticker := time.NewTicker(u.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-u.shutdown:
			close(u.done)
			return nil
		case <-stop:
			return nil
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			var errch = make(chan error, 1)
			go func() { errch <- u.updater.Run(ctx) }()

			select {
			case <-ctx.Done():
				fmt.Println(ctx.Err())
			case err := <-errch:
				fmt.Println(err)
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
		u.shuttingDown = atomic.Value{}
		return nil
	}
}
