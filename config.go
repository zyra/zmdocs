package zmdocs

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
)

// Base page properties that can be found in pages and page patterns
type BasePage struct {
	Name         string `yaml:"name"`         // Page name
	Path         string `yaml:"path"`         // Relative path of generated page. This is only required if "AddToMenu" is to to true
	SourceFile   string `yaml:"source"`       // Relative path to source file. Required only if this is a static page.
	Template     string `yaml:"template"`     // Template name. Defaults to `"base"`.
	AddToMenu    bool   `yaml:"addToMenu"`    // Whether to add this to the menu automatically
	MenuGroup    string `yaml:"menuGroup"`    // Name of menu group to automatically add this entry to
	EditOnGithub bool   `yaml:"editOnGithub"` // Whether to show "Edit on Github" button on this page. Link will be automatically generated.
}

// Static page configuration
type Page struct {
	BasePage `yaml:",inline"`
	Title    string `yaml:"title,omitempty"` // Page title to be used in the html `<title>` tag.
}

// Auto / Pattern page configuration
type PagePattern struct {
	BasePage   `yaml:",inline"`
	SourceGlob string `yaml:"sourceGlob"` // Glob to find files to process under this rule
	Pattern    string `yaml:"pattern"`    // Pattern to use to extract relevant information that can be used in other page properties
}

// Go template used to render pages
type Template struct {
	Name       string `yaml:"name"`   // Template name
	SourceFile string `yaml:"source"` // Source file containing Go template
}

// Menu item config
type MenuItem struct {
	Name   string      `yaml:"name"`  // name is required if group == true
	Title  string      `yaml:"title"` // menu item title
	Link   string      `yaml:"link"`  // menu item link (optional)
	Items  []*MenuItem `yaml:"items"` // sub menu items (optional)
	Group  bool        `yaml:"group"` // whether this is a group heading
	Active bool        `yaml:"-"`     // This property is populated and used in the rendering stage to determine whether it should have an `.active` class added.
}

// Main parser config
type ParserConfig struct {
	RootDir     string         `yaml:"rootDir"`      // root directory of project, defaults to the config file directory if initialized with NewParserFromConfigFile
	OutDir      string         `yaml:"outDir"`       // output directory for generated docs
	Pages       []*Page        `yaml:"pages"`        // list of pages to render
	AutoPages   []*PagePattern `yaml:"pagePatterns"` // list of patterns to derive pages from
	Templates   []*Template    `yaml:"templates"`    // list of template files
	MenuItems   []*MenuItem    `yaml:"menuItems"`    // menu items
	SiteTitle   string         `yaml:"siteTitle"`    // Site title
	Description string         `yaml:"description"`  // Site description to be used in the `<meta name="description" value"...">` HTML tag
	Repo        string         `yaml:"repo"`         // Project repo URL
	BaseURL     string         `yaml:"baseUrl"`      // Base public URL for the generated docs
}

// Loads configuration from a .yml / .yaml file
func NewConfigFromFile(path string) (*ParserConfig, error) {
	var config ParserConfig
	var err error
	var data []byte

	if data, err = ioutil.ReadFile(path); err != nil {
		return nil, fmt.Errorf("unabel to read config file: %s", err.Error())
	} else if err = yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("unable to parse config: %s", err.Error())
	}

	config.RootDir = filepath.Dir(path)
	config.OutDir = filepath.Join(config.RootDir, config.OutDir)

	for _, t := range config.Templates {
		if !filepath.IsAbs(t.SourceFile) {
			t.SourceFile = filepath.Join(config.RootDir, t.SourceFile)
		}
	}

	return &config, nil
}
