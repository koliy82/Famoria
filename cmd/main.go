package main

import (
	"famoria/internal/app"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		app.App,
	).Run()
}
