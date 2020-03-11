package main

import (
	"context"
	"fmt"
	"time"
)

type UpdateLooper struct {
	Updater       *Updater
	CheckInterval time.Duration
}

func NewUpdateLooper(u *Updater, c time.Duration) *UpdateLooper {
	return &UpdateLooper{Updater: u, CheckInterval: c}
}

func (u *UpdateLooper) Loop(ctx context.Context, stop <-chan struct{}) {
	var errch = make(chan error, 1)
L:
	for {
		select {
		case <-stop:
			break L
		case <-ctx.Done():
			fmt.Println(ctx.Err())
			time.Sleep(u.CheckInterval)
		case err := <-errch:
			fmt.Println(err)
			time.Sleep(u.CheckInterval)
		default:
			if err := u.Updater.Run(ctx); err != nil {
				errch <- err
			}
			time.Sleep(u.CheckInterval)
		}
	}
	fmt.Println("Loop end")
}
