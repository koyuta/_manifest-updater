package updater

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var timeout = 20 * time.Second

type UpdateLooper struct {
	updaters      []*Updater
	checkInterval time.Duration
	// logger Logger

	keyFilePath string

	queue <-chan *Entry

	shutdown     chan struct{}
	done         chan struct{}
	shuttingDown *atomic.Value
}

func NewUpdateLooper(queue <-chan *Entry, c time.Duration, keyFilePath string) *UpdateLooper {
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

	wg := &sync.WaitGroup{}

	for {
		select {
		case entry, ok := <-u.queue:
			if !ok {
				return errors.New("queue was closed")
			}
			u.updaters = append(u.updaters, NewUpdater(entry, u.keyFilePath))
		case <-u.shutdown:
			wg.Wait()
			close(u.done)
			return nil
		case <-stop:
			wg.Wait()
			return nil
		case <-ticker.C:
			for i := range u.updaters {
				updater := u.updaters[i]

				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()

				var errch = make(chan error, 1)
				wg.Add(1)
				go func() {
					defer wg.Done()
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
