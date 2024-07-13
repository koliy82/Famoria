package main

import (
	"go_tg_bot/internal/bot"
	"go_tg_bot/internal/config"
	"go_tg_bot/internal/database"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		println(".env ne zapolneno, peredelyvay", err.Error())
		return
	}

	conn, err := database.Connect(cfg)

	if err != nil {
		println(err.Error())
		return
	}

	err = bot.AppendStruct(conn)
	if err != nil {
		println(err.Error())
	}

	println("кукуе")
}
