package main

import (
	"fmt"
	"github.com/urfave/cli"
	"github.com/zyra/zmdocs"
	"log"
	"os"
	"path/filepath"
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
			Action: func(ctx *cli.Context) error {
				configPath := ctx.String("config")

				if !filepath.IsAbs(configPath) {
					if pwd, err := os.Getwd(); err != nil {
						return err
					} else {
						configPath = filepath.Join(pwd, configPath)
					}
				}

				if p, e := zmdocs.NewParserFromConfigFile(configPath); e != nil {
					return fmt.Errorf("unable to parse config: %s", e)
				} else if e := p.LoadSourceFiles(); e != nil {
					return fmt.Errorf("unable to load files: %s", e)
				} else if rnd, e := p.Renderer(); e != nil {
					return fmt.Errorf("unable to create renderer: %s", e)
				} else if e := rnd.Render(); e != nil {
					return fmt.Errorf("unable to render files: %s", e)
				}
				return nil
			},
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
