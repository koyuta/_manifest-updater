package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/koyuta/manifest-updater/pkg/repository"
	"github.com/koyuta/manifest-updater/updater"
	"github.com/sirupsen/logrus"

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

	baseDir, err := ioutil.TempDir(os.TempDir(), "manifest-updater")
	if err != nil {
		return err
	}
	repository.DefaultCloneDir = baseDir
	defer func() {
		fmt.Println("err:", os.RemoveAll(baseDir))
	}()

	var shutdown = make(chan struct{})
	go func() {
		sigch := make(chan os.Signal, 1)
		signal.Notify(sigch, syscall.SIGTERM, syscall.SIGINT)
		<-sigch
		shutdown <- struct{}{}
	}()

	var logger = logrus.New()
	logger.Out = os.Stdout

	var (
		queue         = make(chan *updater.Entry, 1)
		checkInterval = time.Duration(c.Int64(intervalFlag.Name)) * time.Second
		key           = c.String(keyFlag.Name)
	)
	looper := updater.NewUpdateLooper(queue, checkInterval, logger, key)

	var stoploop = make(chan struct{})
	go func() {
		if err := looper.Loop(stoploop); err != nil {
			logger.Fatalf("Loop: %v", err)
		}
	}()

	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", c.Uint(portFlag.Name)),
		Handler: BuildRouter(queue),
	}
	go func() {
		if err := srv.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("ListenAndServe: %v", err)
		}
	}()

	<-shutdown
	logger.Info("Shutting down...")

	// Cleanup a base directory
	os.RemoveAll(baseDir)

	srctx, srcancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer srcancel()
	if err := srv.Shutdown(srctx); err != nil {
		logger.Fatalf("Shutdown server: %v", err)
	}
	loctx, locancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer locancel()
	if err := looper.Shutdown(loctx); err != nil {
		logger.Fatalf("Shutdown loop: %v", err)
	}

	logger.Info("Shutdown")
	return nil
}
