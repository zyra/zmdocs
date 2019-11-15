package zmdocs

import (
	"github.com/russross/blackfriday"
	"html/template"
	"io/ioutil"
)

var blackFridayExtensions = blackfriday.WithExtensions(blackfriday.CommonExtensions | blackfriday.AutoHeadingIDs | blackfriday.Autolink | blackfriday.Footnotes)
var blackFridayRndOpts = blackfriday.HTMLRendererParameters{
	Flags: blackfriday.CommonHTMLFlags | blackfriday.FootnoteReturnLinks,
}

type File struct {
	BasePage
	Title string
}

func (f *File) RenderContext(p *Parser) (*RenderContext, error) {
	var fc []byte
	var err error

	if fc, err = ioutil.ReadFile(f.SourceFile); err != nil {
		return nil, err
	}

	rnd := blackfriday.NewHTMLRenderer(blackFridayRndOpts)
	o := blackfriday.Run(fc, blackFridayExtensions, blackfriday.WithRenderer(rnd))

	if f.Title == "" {
		mp := blackfriday.New(blackFridayExtensions, blackfriday.WithRenderer(rnd))

		a := mp.Parse(fc)
		it := a.FirstChild
		for it != nil && f.Title == "" {
			if it.Type == blackfriday.Heading && it.Level == 1 && it.FirstChild != nil && it.FirstChild.Type == blackfriday.Text {
				f.Title = string(it.FirstChild.Literal)
				break
			}

			it = it.Next
		}
	}

	ctx := NewRenderContext(f, p.Config, template.HTML(o))

	if f.Path == "" || f.Path == "/" {
		if p.Config.BaseURL != "" {
			ctx.Link = p.Config.BaseURL
		}
	}

	return ctx, nil
}

func (f *File) MenuItem() *MenuItem {
	return &MenuItem{
		Name:  f.Name,
		Title: f.Title,
		Link:  f.Path,
	}
}

func (f *File) AppendToMenu(menuItems []*MenuItem) {
	menuItem := f.MenuItem()

	if f.MenuGroup != "" {
		for _, mit := range menuItems {
			if mit.Group && mit.Name == f.MenuGroup {
				if mit.Items == nil {
					mit.Items = make([]*MenuItem, 0)
				}

				mit.Items = append(mit.Items, menuItem)
			}
		}
	} else {
		menuItems = append(menuItems, menuItem)
	}
}
