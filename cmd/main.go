package main

import (
	"go.uber.org/fx"
	"go_tg_bot/internal/app"
)

func main() {
	fx.New(
		app.App,
	).Run()
}
