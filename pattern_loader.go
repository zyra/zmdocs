package zmdocs

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"text/template"
)

// Matched file with the file path and the pattern matches
type PatternMatch struct {
	Path        string
	PathMatches [][]string
}

// Finds files that match a glob (source) and finds pattern matches on each file
func GetPatternMatches(source, pattern string) ([]*PatternMatch, error) {
	if source == "" {
		return nil, errors.New("source glob is required")
	}

	if pattern == "" {
		return nil, errors.New("pattern is required")
	}

	rgx, err := regexp.Compile(pattern)

	if err != nil {
		return nil, fmt.Errorf("unable to compile regex pattern: %s", pattern)
	}

	matches, err := filepath.Glob(source)

	if err != nil {
		return nil, fmt.Errorf("unable to find files matching glob (%s): %s", source, err.Error())
	}

	mLen := len(matches)

	if mLen == 0 {
		return nil, fmt.Errorf("no files were found for glob (%s)", source)
	}

	pms := make([]*PatternMatch, mLen, mLen)

	for i, path := range matches {
		pathMatches := rgx.FindAllStringSubmatch(path, -1)

		pm := PatternMatch{
			Path:        path,
			PathMatches: pathMatches,
		}

		pms[i] = &pm
	}

	return pms, nil
}

// Constructs a File instance from a pattern match
func GetFileForPatternMatch(ap *PagePattern, pm *PatternMatch) (*File, error) {
	if ap == nil {
		return nil, errors.New("page pattern object cannot be nil")
	}

	if pm == nil {
		return nil, errors.New("pattern match cannot be nil")
	}

	file := File{}
	file.EditOnGithub = ap.EditOnGithub
	file.AddToMenu = ap.AddToMenu
	file.MenuGroup = ap.MenuGroup
	file.Template = ap.Template
	file.SourceFile = pm.Path

	var err error

	if file.Name, err = stringFromTemplate(ap.Name, pm); err != nil {
		return nil, fmt.Errorf("unable to parse name template: \n\t%s", err.Error())
	} else if file.Path, err = stringFromTemplate(ap.Path, pm); err != nil {
		return nil, fmt.Errorf("unable to parse path template: \n\t%s", err.Error())
	}

	return &file, nil
}

// Utility function to generate a string from a provided Go template
func stringFromTemplate(t string, pm *PatternMatch) (string, error) {
	if tmpl, err := template.New("").Parse(t); err != nil {
		return "", fmt.Errorf("unable to parse template: \n\t%s", err.Error())
	} else {
		buff := bytes.NewBuffer(make([]byte, 0))

		if err := tmpl.Execute(buff, pm); err != nil {
			return "", fmt.Errorf("unable to execute template: \n\t%s", err.Error())
		}

		return buff.String(), nil
	}
}
