package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	slogfiber "github.com/samber/slog-fiber"
	hive "gitlab-service/hive/Web"
	"os"

	"github.com/caarlos0/env/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"gitlab-service/config"
	"gitlab-service/server"
	"log/slog"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	conf := &config.Config{}
	opts := env.Options{UseFieldNameByDefault: true}

	err := env.ParseWithOptions(conf, opts)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	serv, err := server.NewServer(conf, logger, hive.New(conf.HiveURL))
	if err != nil {
		logger.Error(err.Error())
		return
	}

	app := fiber.New(fiber.Config{
		AppName: conf.ApplicationName,
	})

	app.Use(slogfiber.New(logger))
	app.Use(requestid.New())
	app.Use(recover.New())

	app.Post("/", serv.OnNewAssignment)
	app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.SendString(fmt.Sprintf("Welcome to %s", conf.ApplicationName))
	})

	logger.Error(app.Listen(fmt.Sprintf(":%s", conf.Port)).Error())
}
