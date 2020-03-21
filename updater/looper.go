package updater

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"
)

var timeout = 20 * time.Second

type UpdateLooper struct {
	entries       []*Entry
	checkInterval time.Duration
	logger        *logrus.Logger

	token string

	queue <-chan *Entry

	shutdown     chan struct{}
	done         chan struct{}
	shuttingDown *atomic.Value
}

func NewUpdateLooper(queue <-chan *Entry, c time.Duration, logger *logrus.Logger, token string) *UpdateLooper {
	return &UpdateLooper{
		queue:         queue,
		checkInterval: c,
		logger:        logger,
		token:         token,
		shutdown:      make(chan struct{}),
		done:          make(chan struct{}),
		shuttingDown:  &atomic.Value{},
	}
}

func (u *UpdateLooper) Loop(stop <-chan struct{}) error {
	if v := u.shuttingDown.Load(); v != nil {
		return errors.New("Looper is shutting down")
	}

	ticker := time.NewTicker(u.checkInterval)
	defer ticker.Stop()

	var (
		wg         = sync.WaitGroup{}
		sem        = semaphore.NewWeighted(10)
		repoLocker = map[string]sync.Locker{}
	)

	for {
		select {
		case entry, ok := <-u.queue:
			if !ok {
				return errors.New("Queue was closed")
			}
			j, _ := json.Marshal(entry)
			u.logger.Infof("Recieved a entry: %s", string(j))
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
				updater := NewUpdater(entry, u.token)

				var errch = make(chan error, 1)
				mux := repoLocker[entry.Git]
				sem.Acquire(context.Background(), 1)
				wg.Add(1)
				go func() {
					defer func() {
						mux.Unlock()
						sem.Release(1)
						wg.Done()
					}()

					mux.Lock()

					ctx, cancel := context.WithTimeout(context.Background(), timeout)
					defer cancel()

					errch <- updater.Run(ctx)

					select {
					case <-ctx.Done():
						u.logger.Error(fmt.Errorf("Updater: %w", ctx.Err()))
					case err := <-errch:
						if err != nil {
							u.logger.Error(fmt.Errorf("Updater: %w", err))
						} else {
							u.logger.Info("Pull request was created")
						}
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
