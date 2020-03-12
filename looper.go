package main

import (
	"context"
	"fmt"
	"time"
)

var timeout = 60 * time.Second

type UpdateLooper struct {
	updater       *Updater
	checkInterval time.Duration
	// logger Logger
}

func NewUpdateLooper(u *Updater, c time.Duration) *UpdateLooper {
	return &UpdateLooper{updater: u, checkInterval: c}
}

func (u *UpdateLooper) Loop(stop <-chan struct{}) {
	ticker := time.NewTicker(u.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
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
