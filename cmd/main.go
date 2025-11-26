package main

import (
	"context"
	"github.com/AndreySirin/newProject-28-11/internal/client"
	"github.com/AndreySirin/newProject-28-11/internal/server"
	"github.com/AndreySirin/newProject-28-11/internal/storage"
	"github.com/AndreySirin/newProject-28-11/internal/taskManager"
	"log/slog"
	"os"
	"os/signal"
	"sync"
)

const (
	pathDB       = "myDB"
	addr         = ":8080"
	countWorkers = 3
)

func main() {
	lg := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	db, err := storage.New(pathDB)
	if err != nil {
		lg.Error("error init db", err.Error())
		os.Exit(1)
	}
	Client := client.New()
	manager := taskManager.New(lg, db, Client)

	err = manager.Init()
	if err != nil {
		lg.Error("error get all pending ids", err.Error())
	}

	srv := server.New(addr, manager, lg)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	wg := new(sync.WaitGroup)

	for i := 0; i < countWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			manager.RunWorker(ctx)
		}()
	}
	wg.Add(1)
	go func() {

		defer wg.Done()
		srv.Run()
	}()
	wg.Add(1)
	go func() {

		defer wg.Done()
		<-ctx.Done()
		srv.ShutDown()
	}()
	wg.Wait()
}
