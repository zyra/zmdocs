package zmdocs

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/zyra/zmdocs/templates"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Renderer contains all the relevant data to render the docs
type Renderer struct {
	MenuItems []*MenuItem
	Contexts  []*RenderContext
	Templates []*Template
}

// Renders all all render contexts and outputs their files
func (r *Renderer) Render() error {
	log.WithFields(logrus.Fields{
		"menuItems": len(r.MenuItems),
		"pages":     len(r.Contexts),
		"templates": len(r.Templates),
	}).Infof("rendering")

	var baseTemplateStr string

	for _, t := range r.Templates {
		if t.Name == "base" {
			baseTemplateData, err := ioutil.ReadFile(t.SourceFile)

			if err != nil {
				return err
			}

			baseTemplateStr = string(baseTemplateData)
		}
	}

	if baseTemplateStr == "" {
		log.Info("no base template was provided, using the default one")
		baseTemplateStr = templates.BaseTemplate
	}

	baseTemplate, err := template.New("").Parse(baseTemplateStr)

	if err != nil {
		return err
	}

	log.Debug("starting render process")

	for _, ctx := range r.Contexts {
		if err := ctx.Render(baseTemplate); err != nil {
			return fmt.Errorf("unable to render page: %s", err.Error())
		}
	}

	log.Infof("rendered %d pages", len(r.Contexts))

	return nil
}

// A render context contains all required information to render a single page
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

	l *logrus.Entry
}

// Returns a new render context from the provided file, parser config, and HTML content
func NewRenderContext(f *File, c *ParserConfig, content template.HTML) *RenderContext {
	outDir := filepath.Join(c.RootDir, "docs", f.Path)
	outFile := filepath.Join(outDir, "index.html")

	ctx := RenderContext{
		Title:       f.Title,
		SiteTitle:   c.SiteTitle,
		Description: c.Description,
		MenuItems:   c.MenuItems,
		Content:     content,
		OutDir:      outDir,
		OutFile:     outFile,
		BaseURL:     c.BaseURL,
		Link:        f.Path,
	}

	ctx.l = log.WithFields(logrus.Fields{
		"title": ctx.Title,
		"link":  ctx.Link,
	})

	return &ctx
}

// Renders and outputs the page
func (c *RenderContext) Render(tmpl *template.Template) error {
	c.l.Debug("rendering page")

	for _, mit := range c.MenuItems {
		if mit.Group {
			for _, cmit := range mit.Items {
				cmit.Active = cmit.Link == c.Link
			}
		} else if mit.Link == c.Link {
			mit.Active = true
		} else {
			mit.Active = false
		}
	}

	if tmpl != nil {
		buff := bytes.NewBuffer(make([]byte, 0))

		if err := tmpl.Execute(buff, c); err != nil {
			return fmt.Errorf("unable eto execute template: %s", err.Error())
		}

		c.Content = template.HTML(buff.String())
	}

	if err := os.MkdirAll(c.OutDir, 0755); err != nil {
		return fmt.Errorf("unable to create directory: %s", err.Error())
	}

	c.l.Debugf("writing file to %s", c.OutFile)

	if err := ioutil.WriteFile(c.OutFile, []byte(c.Content), 0644); err != nil {
		return fmt.Errorf("unable to write file: %s", err.Error())
	}

	return nil
}
