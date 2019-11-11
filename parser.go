package zmdocs

import (
	"bytes"
	"fmt"
	"github.com/russross/blackfriday"
	"github.com/sirupsen/logrus"
	"github.com/zyra/zmdocs/templates"
	"gopkg.in/yaml.v2"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
)

type basePage struct {
	Name         string `yaml:"name"`
	Path         string `yaml:"path"`
	SourceFile   string `yaml:"source"`
	Template     string `yaml:"template"`
	AddToMenu    bool   `yaml:"addToMenu"`
	MenuGroup    string `yaml:"menuGroup"`
	EditOnGithub bool   `yaml:"editOnGithub"`
}

type Page struct {
	basePage `yaml:",inline"`
	Title    string `yaml:"title,omitempty"`
}

type PagePattern struct {
	basePage   `yaml:",inline"`
	SourceGlob string `yaml:"sourceGlob"`
	Pattern    string `yaml:"pattern"`
}

type Template struct {
	Name       string `yaml:"name"`
	SourceFile string `yaml:"source"`
}

type MenuItem struct {
	Name   string      `yaml:"name"`  // name is required if group == true
	Title  string      `yaml:"title"` // menu item title
	Link   string      `yaml:"link"`  // menu item link (optional)
	Items  []*MenuItem `yaml:"items"` // sub menu items (optional)
	Group  bool        `yaml:"group"` // whether this is a group heading
	Active bool        `yaml:"-"`
}

type ParserConfig struct {
	RootDir     string         `yaml:"rootDir"` // root directory of project, defaults to the config file directory if initialized with NewParserFromConfig
	OutDir      string         `yaml:"outDir"`
	Pages       []*Page        `yaml:"pages"`        // list of pages to render
	AutoPages   []*PagePattern `yaml:"pagePatterns"` // list of patterns to derive pages from
	Templates   []*Template    `yaml:"templates"`    // list of template files
	MenuItems   []*MenuItem    `yaml:"menuItems"`    // menu items
	SiteTitle   string         `yaml:"siteTitle"`
	Description string         `yaml:"description"`
	Repo        string         `yaml:"repo"`
	BaseURL     string         `yaml:"baseUrl"`
}

type File struct {
	basePage
	Title string
}

type Parser struct {
	Config *ParserConfig
	Files  []*File
	Logger *logrus.Logger
}

func NewParser(config *ParserConfig) *Parser {
	l := logrus.New()

	l.SetFormatter(&logrus.TextFormatter{
		ForceColors:               true,
		EnvironmentOverrideColors: true,
		FullTimestamp:             true,
		QuoteEmptyFields:          true,
	})

	if os.Getenv("ZMDOCS_DEBUG") == "true" {
		l.SetLevel(logrus.DebugLevel)
	} else {
		l.SetLevel(logrus.InfoLevel)
	}

	p := Parser{
		Config: config,
		Files:  make([]*File, 0),
		Logger: l,
	}

	l.WithFields(logrus.Fields{
		"rootDir":   config.RootDir,
		"outDir":    config.OutDir,
		"baseUrl":   config.BaseURL,
		"repo":      config.Repo,
		"siteTitle": config.SiteTitle,
	}).Debug("creating new parser")

	return &p
}

func NewParserFromConfig(configPath string) (*Parser, error) {
	var config ParserConfig

	if data, err := ioutil.ReadFile(configPath); err != nil {
		return nil, fmt.Errorf("unabel to read config file: %s", err.Error())
	} else if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("unable to parse config: %s", err.Error())
	}

	config.RootDir = filepath.Dir(configPath)
	config.OutDir = filepath.Join(config.RootDir, config.OutDir)

	return NewParser(&config), nil
}

func (p *Parser) LoadSourceFiles() error {
	p.Logger.Info("loading source files")

	p.Logger.Debug("Loading static files")
	if err := p.loadStaticFiles(); err != nil {
		return err
	}
	p.Logger.Debug("Done loading static files")

	p.Logger.Debug("Loading glob files")
	if err := p.loadGlobFiles(); err != nil {
		return err
	}
	p.Logger.Debug("Done loading glob files")

	return nil
}

func (p *Parser) loadStaticFiles() error {
	for _, pg := range p.Config.Pages {
		g := filepath.Join(p.Config.RootDir, pg.SourceFile)

		file := File{
			basePage: *&pg.basePage,
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

		p.Logger.WithFields(logrus.Fields{
			"pattern": ap.Pattern,
			"absPath": g,
		}).Debug("processing page pattern")

		if patternMatches, err := GetPatternMatches(g, ap.Pattern); err != nil {
			return fmt.Errorf("unable to process page pattern #%d: %s", i, err.Error())
		} else {
			p.Logger.WithFields(logrus.Fields{
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

type RenderContext struct {
	Title       string
	SiteTitle   string
	Description string
	MenuItems   []*MenuItem
	Content     template.HTML
	OutDir      string
	OutFile     string
	BaseURL     string
	Link        string
}

func (p *Parser) handleHomePage(it *MenuItem) {
	if it.Group {
		for _, iit := range it.Items {
			p.handleHomePage(iit)
		}
		return
	}

	if it.Link == "" || it.Link == "/" {
		p.Logger.WithFields(logrus.Fields{
			"title": it.Title,
		}).Debug("setting page link to baseURL")
		it.Link = p.Config.BaseURL
	}
}

func (p *Parser) Render() error {
	menuItems := p.Config.MenuItems
	rndCtxs := make([]*RenderContext, 0)

	if p.Config.BaseURL != "" {
		for _, it := range menuItems {
			p.handleHomePage(it)
		}
	}

	for _, f := range p.Files {
		if fc, err := ioutil.ReadFile(f.SourceFile); err != nil {
			return err
		} else {
			exts := blackfriday.WithExtensions(blackfriday.CommonExtensions | blackfriday.AutoHeadingIDs | blackfriday.Autolink | blackfriday.Footnotes)
			rndOpts := blackfriday.HTMLRendererParameters{
				Flags: blackfriday.CommonHTMLFlags | blackfriday.FootnoteReturnLinks,
			}
			rnd := blackfriday.NewHTMLRenderer(rndOpts)

			o := blackfriday.Run(fc, exts, blackfriday.WithRenderer(rnd))

			if f.Title == "" {
				mp := blackfriday.New(exts, blackfriday.WithRenderer(rnd))

				a := mp.Parse(fc)
				it := a.FirstChild
				for it != nil && f.Title == "" {
					if it.Type == blackfriday.Heading && it.Level == 1 && it.FirstChild != nil && it.FirstChild.Type == blackfriday.Text {
						f.Title = string(it.FirstChild.Literal)
					}

					it = it.Next
				}
			}

			outDir := filepath.Join(p.Config.RootDir, "docs", f.Path)
			outFile := filepath.Join(outDir, "index.html")

			ctx := RenderContext{
				Title:   f.Title,
				Content: template.HTML(o),
				OutDir:  outDir,
				OutFile: outFile,
				Link:    f.Path,
			}

			if f.Path == "" || f.Path == "/" {
				if p.Config.BaseURL != "" {
					ctx.Link = p.Config.BaseURL
				}
			}

			rndCtxs = append(rndCtxs, &ctx)

			if f.AddToMenu {
				menuItem := MenuItem{
					Name:  f.Name,
					Title: f.Title,
					Link:  f.Path,
				}

				if f.MenuGroup != "" {
					for _, mit := range menuItems {
						if mit.Group && mit.Name == f.MenuGroup {
							if mit.Items == nil {
								mit.Items = make([]*MenuItem, 0)
							}

							mit.Items = append(mit.Items, &menuItem)
						}
					}
				} else {
					menuItems = append(menuItems, &menuItem)
				}
			}
		}
	}

	var baseTemplateStr string

	for _, t := range p.Config.Templates {
		if t.Name == "base" {
			baseTemplatePath := filepath.Join(p.Config.RootDir, t.SourceFile)
			baseTemplateData, err := ioutil.ReadFile(baseTemplatePath)

			if err != nil {
				return err
			}

			baseTemplateStr = string(baseTemplateData)
		}
	}

	if baseTemplateStr == "" {
		baseTemplateStr = templates.BaseTemplate
	}

	baseTemplate, err := template.New("").Parse(baseTemplateStr)

	if err != nil {
		return err
	}

	for _, ctx := range rndCtxs {
		ctx.SiteTitle = p.Config.SiteTitle
		ctx.Description = p.Config.Description
		ctx.MenuItems = menuItems
		ctx.BaseURL = p.Config.BaseURL

		p.Logger.WithFields(logrus.Fields{
			"title": ctx.Title,
			"link":  ctx.Link,
		}).Debug("rendering page")

		for _, mit := range menuItems {
			if mit.Group {
				for _, cmit := range mit.Items {
					cmit.Active = cmit.Link == ctx.Link
				}
			} else if mit.Link == ctx.Link {
				mit.Active = true
			} else {
				mit.Active = false
			}
		}

		buff := bytes.NewBuffer(make([]byte, 0))

		if err := baseTemplate.Execute(buff, ctx); err != nil {
			return err
		}

		ctx.Content = template.HTML(buff.String())

		if err := os.MkdirAll(ctx.OutDir, 0755); err != nil {
			return err
		}

		if err := ioutil.WriteFile(ctx.OutFile, []byte(ctx.Content), 0644); err != nil {
			return err
		}
	}

	return nil
}
