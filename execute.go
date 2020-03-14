package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/koyuta/manifest-updater/updater"

	cli "github.com/urfave/cli/v2"
)

func execute(c *cli.Context) error {
	if c.NArg() != 0 {
		if err := cli.ShowAppHelp(c); err != nil {
			return err
		}
		// Return empty error to set 1 to exit status.
		return errors.New("")
	}

	var shutdown = make(chan struct{})
	go func() {
		sigch := make(chan os.Signal, 1)
		signal.Notify(sigch, syscall.SIGTERM)
		<-sigch
		shutdown <- struct{}{}
	}()

	checkInterval := time.Duration(c.Int64(intervalFlag.Name)) * time.Second

	var queue = make(chan *updater.Updater, 1)

	var stoploop = make(chan struct{})
	looper := updater.NewUpdateLooper(queue, checkInterval)
	go func() {
		if err := looper.Loop(stoploop); err != nil {
			log.Fatalf("Loop: %v", err)
		}
	}()

	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", c.Uint(portFlag.Name)),
		Handler: BuildRouter(),
	}
	go func() {
		if err := srv.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("ListenAndServe: %v", err)
		}
	}()

	<-shutdown
	fmt.Println("shutting down...")

	srctx, srcancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer srcancel()
	if err := srv.Shutdown(srctx); err != nil {
		log.Fatalf("server shutdown: %v", err)
	}
	loctx, locancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer locancel()
	if err := looper.Shutdown(loctx); err != nil {
		log.Fatalf("loop shutdown: %v", err)
	}

	fmt.Println("shutdown")
	return nil
}
