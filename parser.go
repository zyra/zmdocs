package zmdocs

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

// A parser loads files from the provided config
type Parser struct {
	Config *ParserConfig
	Files  []*File
}

// Returns a new Parser instance from the provided config
func NewParser(config *ParserConfig) *Parser {
	log.SetFormatter(&logrus.TextFormatter{
		ForceColors:               true,
		EnvironmentOverrideColors: true,
		FullTimestamp:             true,
		QuoteEmptyFields:          true,
	})

	if os.Getenv("ZMDOCS_DEBUG") == "true" {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}

	p := Parser{
		Config: config,
		Files:  make([]*File, 0),
	}

	log.WithFields(logrus.Fields{
		"rootDir":   config.RootDir,
		"outDir":    config.OutDir,
		"baseUrl":   config.BaseURL,
		"repo":      config.Repo,
		"siteTitle": config.SiteTitle,
	}).Debug("creating new parser")

	return &p
}

// Load configuration from a file and creates a new parser with it
func NewParserFromConfigFile(configPath string) (*Parser, error) {
	if config, err := NewConfigFromFile(configPath); err != nil {
		return nil, err
	} else {
		return NewParser(config), nil
	}
}

// Returns a Renderer instance from the parsed files
// This must be ran after a successful LoadSourceFiles so there are files to render
func (p *Parser) Renderer() (*Renderer, error) {
	if p.Config.BaseURL != "" {
		for _, it := range p.Config.MenuItems {
			p.handleHomePage(it)
		}
	}

	rndCtxs := make([]*RenderContext, 0)

	for _, f := range p.Files {
		if ctx, err := f.RenderContext(p); err != nil {
			return nil, err
		} else {
			rndCtxs = append(rndCtxs, ctx)
		}

		if f.AddToMenu {
			f.AppendToMenu(p.Config.MenuItems)
		}
	}

	rnd := &Renderer{
		MenuItems: p.Config.MenuItems,
		Contexts:  rndCtxs,
	}

	return rnd, nil
}

// Load all source files
func (p *Parser) LoadSourceFiles() error {
	log.Info("loading source files")

	log.Debug("Loading static files")
	if err := p.loadStaticFiles(); err != nil {
		return err
	}
	log.Debug("Done loading static files")

	log.Debug("Loading glob files")
	if err := p.loadGlobFiles(); err != nil {
		return err
	}
	log.Debug("Done loading glob files")

	log.Infof("loaded %d files", len(p.Files))

	return nil
}

func (p *Parser) loadStaticFiles() error {
	for _, pg := range p.Config.Pages {
		g := filepath.Join(p.Config.RootDir, pg.SourceFile)

		file := File{
			BasePage: *&pg.BasePage,
			Title:    pg.Title,
		}

		file.SourceFile = g

		p.Files = append(p.Files, &file)
	}

	return nil
}

func (p *Parser) loadGlobFiles() error {
	for i, ap := range p.Config.AutoPages {
		g := filepath.Join(p.Config.RootDir, ap.SourceGlob)

		log.WithFields(logrus.Fields{
			"pattern": ap.Pattern,
			"absPath": g,
		}).Debug("processing page pattern")

		if patternMatches, err := GetPatternMatches(g, ap.Pattern); err != nil {
			return fmt.Errorf("unable to process page pattern #%d: %s", i, err.Error())
		} else {
			log.WithFields(logrus.Fields{
				"pattern": ap.Pattern,
			}).Debugf("found %d matches", len(patternMatches))

			for _, pm := range patternMatches {
				if file, err := GetFileForPatternMatch(ap, pm); err != nil {
					return fmt.Errorf("unable to process file %s: \n\t%s", pm.Path, err.Error())
				} else {
					p.Files = append(p.Files, file)
				}
			}
		}
	}

	return nil
}

func (p *Parser) handleHomePage(it *MenuItem) {
	if it.Group {
		for _, iit := range it.Items {
			p.handleHomePage(iit)
		}
		return
	}

	if it.Link == "" || it.Link == "/" {
		log.WithFields(logrus.Fields{
			"title": it.Title,
		}).Debug("setting page link to baseURL")
		it.Link = p.Config.BaseURL
	}
}
