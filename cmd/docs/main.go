package main

import (
	"github.com/urfave/cli"
	"github.com/zyra/zmdocs/cmd/docs/actions"
	"log"
	"os"
)

var AppVersion = "0.0.1"

func main() {
	app := cli.NewApp()
	app.Name = "ZM Docs"
	app.Version = AppVersion

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:   "verbose, V",
			Usage:  "verbose logging",
			EnvVar: "ZMDOCS_DEBUG",
		},
	}

	app.Before = func(ctx *cli.Context) error {
		if ctx.Bool("verbose") {
			_ = os.Setenv("ZMDOCS_DEBUG", "true")
		}

		return nil
	}

	app.Commands = []cli.Command{
		{
			Name:    "generate",
			Aliases: []string{"g"},
			Usage:   "Generate documentation",
			Action:  actions.Generate,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "config, c",
					Usage:  "Config file",
					EnvVar: "ZMDOC_CONFIG",
					Value:  "./.docs.yaml",
				},
			},
		},
		{
			Name:    "serve",
			Aliases: []string{"s"},
			Usage:   "Run a webserver with livereload",
			Action:  actions.Serve,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "config, c",
					Usage:  "Config file",
					EnvVar: "ZMDOC_CONFIG",
					Value:  "./.docs.yaml",
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
