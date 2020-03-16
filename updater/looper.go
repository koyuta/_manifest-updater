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
	entries       []*Entry
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

	var wg sync.WaitGroup

	var repoLocker = map[string]sync.Locker{}

	for {
		select {
		case entry, ok := <-u.queue:
			if !ok {
				return errors.New("queue was closed")
			}
			u.entries = append(u.entries, entry)
			repoLocker[entry.Git] = &sync.Mutex{}
		case <-u.shutdown:
			wg.Wait()
			close(u.done)
			return nil
		case <-stop:
			wg.Wait()
			return nil
		case <-ticker.C:
			for i := range u.entries {
				entry := u.entries[i]

				updater := NewUpdater(entry, u.keyFilePath)

				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()

				mux := repoLocker[entry.Git]

				var errch = make(chan error, 1)
				wg.Add(1)
				go func() {
					mux.Lock()
					defer mux.Unlock()
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
