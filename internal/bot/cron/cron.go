package cron

import (
	"famoria/internal/bot/cron/tasks"
	"famoria/internal/bot/idle/item"
	"famoria/internal/database/mongo/repositories/brak"

	"github.com/go-co-op/gocron/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Opts struct {
	fx.In
	Log      *zap.Logger
	BrakRepo brak.Repository
	Manager  *item.Manager
	S        gocron.Scheduler
}

func Start(opts Opts) {
	tasks.StartMining(tasks.MiningOpts{
		Log:      opts.Log,
		BrakRepo: opts.BrakRepo,
		Manager:  opts.Manager,
		S:        opts.S,
	})
	opts.S.Start()
	opts.Log.Info("Cron scheduler is started.")
	//err = s.Shutdown()
	//if err != nil {
	//	// handle error
	//}
}

func New() gocron.Scheduler {
	s, err := gocron.NewScheduler()
	if err != nil {
		panic(err)
	}
	return s
}
