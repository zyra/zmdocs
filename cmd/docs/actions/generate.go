package actions

import (
	"fmt"
	"github.com/urfave/cli"
	"github.com/zyra/zmdocs"
	"os"
	"path/filepath"
)

func Generate(ctx *cli.Context) error {
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
}
