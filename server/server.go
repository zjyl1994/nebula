package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/template/infra/vars"
	"example.com/template/webui"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/iancoleman/strcase"
	"github.com/sirupsen/logrus"
)

func Run(listen string) error {
	appName := strcase.ToCamel(vars.APP_NAME)
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ServerHeader:          appName,
		AppName:               appName,
	})

	RegisterRoutes(app)

	// Remove if not using SPA //
	app.Use("/", compress.New(compress.Config{
		Level: compress.LevelDefault,
	}), filesystem.New(filesystem.Config{
		Root:         http.FS(webui.WebUI),
		PathPrefix:   "dist",
		NotFoundFile: "dist/index.html",
	}))
	/////////////////////////////

	srvErr := make(chan error, 1)
	go func() {
		logrus.Infof("%s is running on %s", appName, listen)
		srvErr <- app.Listen(listen)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-srvErr:
		return err
	case sig := <-sigCh:
		logrus.Infoln("Received signal:", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := app.ShutdownWithContext(ctx); err != nil {
			logrus.Errorln("Shutdown error:", err)
		}
		return nil
	}
}
