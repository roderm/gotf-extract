package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/roderm/gotf-extract/pkg/server"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
)

func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		&cli.PathFlag{
			Name:  "config",
			Value: "./config.yaml",
		},
		altsrc.NewIntFlag(&cli.IntFlag{
			Name:  "port",
			Value: 8080,
			Usage: "Port number",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "log.level",
			Value: "info",
			Usage: "Log level",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "address",
			Value: "0.0.0.0",
			Usage: "address to listen on - if empty lookup the outbound address",
		}),
	}
	app.Before = func(c *cli.Context) error {
		if _, err := os.Stat(c.Path("config")); err == nil {
			err := altsrc.InitInputSourceWithContext(app.Flags, altsrc.NewYamlSourceFromFlagFunc("config"))(c)
			if err != nil {
				return err
			}
		}
		return nil
	}
	app.Action = func(c *cli.Context) error {
		srv := http.Server{
			Addr:    fmt.Sprintf("%s:%d", c.String("address"), c.Int("port")),
			Handler: server.Handler(),
		}
		logrus.WithField("address", srv.Addr).Info("start server")
		return srv.ListenAndServe()
	}
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
