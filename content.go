package gofb2

import (
	"encoding/xml"
	"fmt"
	"strconv"
)

// Contenter provide interface for tag content
type Contenter interface {
	GetXMLName() xml.Name
	GetContent() []Contenter
	GetText() []byte
}

// Title https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L273
// A title, used in sections, poems and body elements
type Title struct {
	Lang string `xml:"http://www.w3.org/XML/1998/namespace lang,attr,omitempty"`
	contentBase
}

func (t *Title) tagCallback(start xml.StartElement) (Node, error) {
	switch start.Name.Local {
	case "p":
		p := &P{}
		t.appendContent(p)
		return p, nil
	case "empty-line":
		el := &EmptyLine{}
		t.appendContent(el)
		return el, nil
	default:
		return t.contentBase.tagCallback(start)
	}
}

func (t *Title) attrCallback(attr xml.Attr) error {
	if attr.Name.Local == "lang" {
		t.Lang = attr.Value
		return nil
	}
	return t.contentBase.attrCallback(attr)
}

// UnmarshalXML unmarshal XML to Title
func (t *Title) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return NewParser(t).Parse(d, start)
}

// Image https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L283
// An empty element with an image name as an attribute
type Image struct {
	XlinkType string `xml:"http://www.w3.org/1999/xlink type,attr,omitempty"`
	XlinkHref string `xml:"http://www.w3.org/1999/xlink href,attr,omitempty"`
	Alt       string `xml:"alt,attr,omitempty"`
	Title     string `xml:"title,attr,omitempty"`
	ID        string `xml:"id,attr,omitempty"`

	emptyContent
}

// P https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L293
// A basic paragraph, may include simple formatting inside
type P struct {
	StyleType
	ID    string `xml:"id,attr,omitempty"`
	Style string `xml:"style,attr,omitempty"`
}

func (p *P) attrCallback(attr xml.Attr) error {
	if attr.Name.Local == "style" {
		p.Style = attr.Value
	} else if attr.Name.Local == "id" {
		p.ID = attr.Value
	} else {
		return p.StyleType.attrCallback(attr)
	}
	return nil
}

// UnmarshalXML unmarshal XML to P
func (p *P) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return NewParser(p).Parse(d, start)
}

// Cite https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L304
// A citation with an optional citation author at the end
type Cite struct {
	ID         string `xml:"id,omitempty"`
	Lang       string `xml:"http://www.w3.org/XML/1998/namespace lang,attr,omitempty"`
	TextAuthor []*P   `xml:"text-author,omitempty"`
	contentBase
}

func (c *Cite) tagCallback(start xml.StartElement) (Node, error) {
	switch start.Name.Local {
	case "text-author":
		p := &P{}
		c.TextAuthor = append(c.TextAuthor, p)
		return p, nil
	case "p":
		p := &P{}
		c.appendContent(p)
		return p, nil
	case "poem":
		p := &Poem{}
		c.appendContent(p)
		return p, nil
	case "subtitle":
		p := &P{}
		c.appendContent(p)
		return p, nil
	case "table":
		t := &Table{}
		c.appendContent(t)
		return t, nil
	case "empty-line":
		el := &EmptyLine{}
		c.appendContent(el)
		return el, nil
	default:
		return c.contentBase.tagCallback(start)
	}
}

func (c *Cite) attrCallback(attr xml.Attr) error {
	if attr.Name.Local == "lang" {
		c.Lang = attr.Value
	} else if attr.Name.Local == "id" {
		c.ID = attr.Value
	} else {
		return c.contentBase.attrCallback(attr)
	}
	return nil
}

// UnmarshalXML unmarshal XML to Cite
func (c *Cite) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return NewParser(c).Parse(d, start)
}

// Poem https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L321
// A poem
type Poem struct {
	// Poem title
	XMLName xml.Name `xml:"poem"`
	Title   *Title   `xml:"title,omitempty"`

	// Poem epigraph(s), if any
	Epigraphs []*Epigraph `xml:"epigraph,omitempty"`

	// subtitle and stanza
	contentBase
}

func (p *Poem) tagCallback(start xml.StartElement) (Node, error) {
	switch start.Name.Local {
	case "title":
		p.Title = &Title{}
		return p.Title, nil
	case "epigraph":
		ep := &Epigraph{}
		p.Epigraphs = append(p.Epigraphs, ep)
		return ep, nil
	case "subtitle":
		pp := &P{}
		p.appendContent(pp)
		return pp, nil
	case "stanza":
		s := &Stanza{}
		p.appendContent(s)
		return s, nil
	default:
		return p.contentBase.tagCallback(start)
	}
}

// UnmarshalXML unmarshal XML to StyleType
func (p *Poem) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return NewParser(p).Parse(d, start)
}

// Stanza https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L338
// Each poem should have at least one stanza.
// Stanzas are usually separated with empty lines by user agents.
type Stanza struct {
	Lang     string `xml:"http://www.w3.org/XML/1998/namespace lang,attr,omitempty"`
	Title    *Title `xml:"title,omitempty"`
	Subtitle *P     `xml:"subtitle,omitempty"`
	// An individual line in a stanza
	V []*P `xml:"v"`

	emptyText
}

// GetContent for Contenter interface
func (s Stanza) GetContent() []Contenter {
	c := make([]Contenter, len(s.V))
	for i, v := range s.V {
		c[i] = v
	}
	return c
}

func (s *Stanza) tagCallback(start xml.StartElement) (Node, error) {
	switch start.Name.Local {
	case "title":
		s.Title = &Title{}
		return s.Title, nil
	case "subtitle":
		s.Subtitle = &P{}
		return s.Subtitle, nil
	case "v":
		p := &P{}
		s.V = append(s.V, p)
		return p, nil
	default:
		return s.emptyText.tagCallback(start)
	}
}

func (s *Stanza) attrCallback(attr xml.Attr) error {
	if attr.Name.Local == "lang" {
		s.Lang = attr.Value
		return nil
	}
	return s.emptyText.attrCallback(attr)
}

// Epigraph https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L366
// An epigraph
type Epigraph struct {
	ID         string `xml:"id,omitempty"`
	TextAuthor []*P   `xml:"text-author,omitempty"`
	contentBase
}

func (ep *Epigraph) tagCallback(start xml.StartElement) (Node, error) {
	switch start.Name.Local {
	case "text-author":
		p := &P{}
		ep.TextAuthor = append(ep.TextAuthor, p)
		return p, nil
	case "p":
		p := &P{}
		ep.appendContent(p)
		return p, nil
	case "poem":
		p := &Poem{}
		ep.appendContent(p)
		return p, nil
	case "cite":
		c := &Cite{}
		ep.appendContent(c)
		return c, nil
	case "empty-line":
		el := &EmptyLine{}
		ep.appendContent(el)
		return el, nil
	default:
		return ep.contentBase.tagCallback(start)
	}
}

func (ep *Epigraph) attrCallback(attr xml.Attr) error {
	if attr.Name.Local == "id" {
		ep.ID = attr.Value
	} else {
		return fmt.Errorf("unknown attr %s", attr.Name.Local)
	}
	return nil
}

// UnmarshalXML unmarshal XML to Epigraph
func (ep *Epigraph) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return NewParser(ep).Parse(d, start)
}

// Annotation https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L381
// A cut-down version of "section" used in annotations
type Annotation struct {
	contentBase
	ID   string `xml:"id,omitempty"`
	Lang string `xml:"http://www.w3.org/XML/1998/namespace lang,attr,omitempty"`
}

func (a *Annotation) tagCallback(start xml.StartElement) (Node, error) {
	switch start.Name.Local {
	case "p":
		p := &P{}
		a.appendContent(p)
		return p, nil
	case "poem":
		p := &Poem{}
		a.appendContent(p)
		return p, nil
	case "cite":
		c := &Cite{}
		a.appendContent(c)
		return c, nil
	case "subtitle":
		p := &P{}
		a.appendContent(p)
		return p, nil
	case "table":
		t := &Table{}
		a.appendContent(t)
		return t, nil
	case "empty-line":
		el := &EmptyLine{}
		a.appendContent(el)
		return el, nil
	default:
		return a.contentBase.tagCallback(start)
	}
}

func (a *Annotation) attrCallback(attr xml.Attr) error {
	if attr.Name.Local == "lang" {
		a.Lang = attr.Value
	} else if attr.Name.Local == "id" {
		a.ID = attr.Value
	} else {
		return fmt.Errorf("unknown attr %s", attr.Name.Local)
	}
	return nil
}

// UnmarshalXML unmarshal xml to FictionBook struct
func (a *Annotation) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return NewParser(a).Parse(d, start)
}

// Section https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L396
// A basic block of a book, can contain more child sections or textual content
type Section struct {
	// Section's title
	Title *Title `xml:"title,omitempty"`

	// Epigraph(s) for this section
	Epigraphs []*Epigraph `xml:"epigraph,omitempty"`

	// Image to be displayed at the top of this section
	Image *Image `xml:"image,omitempty"`

	// Annotation for this section, if any
	Annotation *Annotation `xml:"annotation,omitempty"`

	// or child Sections
	Sections []*Section `xml:"section,omitempty"`
	// or other content
	contentBase

	ID   string `xml:"id,omitempty"`
	Lang string `xml:"http://www.w3.org/XML/1998/namespace lang,attr,omitempty"`
}

func (s *Section) tagCallback(start xml.StartElement) (Node, error) {
	switch start.Name.Local {
	case "title":
		s.Title = &Title{}
		return s.Title, nil
	case "epigraph":
		ep := &Epigraph{}
		s.Epigraphs = append(s.Epigraphs, ep)
		return ep, nil
	case "image":
		i := &Image{}
		if len(s.Content) > 0 {
			s.appendContent(i)
		} else {
			s.Image = i
		}
		return i, nil
	case "annotation":
		s.Annotation = &Annotation{}
		return s.Annotation, nil
	case "section":
		cs := &Section{}
		s.Sections = append(s.Sections, cs)
		return cs, nil
	case "p":
		p := &P{}
		s.appendContent(p)
		return p, nil
	case "poem":
		p := &Poem{}
		s.appendContent(p)
		return p, nil
	case "subtitle":
		p := &P{}
		s.appendContent(p)
		return p, nil
	case "cite":
		c := &Cite{}
		s.appendContent(c)
		return c, nil
	case "empty-line":
		el := &EmptyLine{}
		s.appendContent(el)
		return el, nil
	case "table":
		t := &Table{}
		s.appendContent(t)
		return t, nil
	default:
		return s.contentBase.tagCallback(start)
	}
}

func (s *Section) attrCallback(attr xml.Attr) error {
	if attr.Name.Local == "lang" {
		s.Lang = attr.Value
	} else if attr.Name.Local == "id" {
		s.ID = attr.Value
	} else {
		return s.contentBase.attrCallback(attr)
	}
	return nil
}

// UnmarshalXML unmarshal XML to Section
func (s *Section) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return NewParser(s).Parse(d, start)
}

// StyleType https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L453
// Markup
type StyleType struct {
	mixed
	Lang string `xml:"http://www.w3.org/XML/1998/namespace lang,attr,omitempty"`
}

func (s *StyleType) tagCallback(start xml.StartElement) (Node, error) {
	switch start.Name.Local {
	case "style":
		nst := &NamedStyleType{}
		s.appendContent(nst)
		return nst, nil
	case "a":
		link := &Link{}
		s.appendContent(link)
		return link, nil
	case "image":
		i := &InlineImage{}
		s.appendContent(i)
		return i, nil
	case "strong", "emphasis", "strikethrough", "sub", "sup", "code":
		st := &StyleType{}
		s.appendContent(st)
		return st, nil
	default:
		return s.contentBase.tagCallback(start)
	}
}

func (s *StyleType) attrCallback(attr xml.Attr) error {
	if attr.Name.Local == "lang" {
		s.Lang = attr.Value
		return nil
	}
	return fmt.Errorf("unknown attr %s", attr.Name.Local)
}

// UnmarshalXML unmarshal XML to StyleType
func (s *StyleType) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return NewParser(s).Parse(d, start)
}

// NamedStyleType https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L470
// Markup
type NamedStyleType struct {
	StyleType
	Name string `xml:"name,attr"`
}

func (s *NamedStyleType) attrCallback(attr xml.Attr) error {
	if attr.Name.Local == "name" {
		s.Name = attr.Value
	} else {
		return s.StyleType.attrCallback(attr)
	}
	return nil
}

// UnmarshalXML unmarshal XML to NamedStyleType
func (s *NamedStyleType) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return NewParser(s).Parse(d, start)
}

// Link https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L488
// Generic hyperlinks. Cannot be nested. Footnotes should be implemented by
// links referring to additional bodies in the same document
type Link struct {
	StyleLinkType
	XlinkType string `xml:"http://www.w3.org/1999/xlink type,attr,omitempty"`
	XlinkHref string `xml:"http://www.w3.org/1999/xlink href,attr"`
	Type      string `xml:"type,attr,omitempty"`
}

func (l *Link) attrCallback(attr xml.Attr) error {
	switch attr.Name.Local {
	case "type":
		if attr.Name.Space == "" {
			l.Type = attr.Value
		} else {
			l.XlinkType = attr.Value
		}
	case "href":
		l.XlinkHref = attr.Value
	default:
		return fmt.Errorf("unknown attr %s", attr.Name.Local)
	}
	return nil
}

// UnmarshalXML unmarshal XML to Link
func (l *Link) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return NewParser(l).Parse(d, start)
}

// StyleLinkType https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L506
// Markup
type StyleLinkType struct {
	mixed
}

func (s *StyleLinkType) tagCallback(start xml.StartElement) (Node, error) {
	switch start.Name.Local {
	case "image":
		i := &InlineImage{}
		s.appendContent(i)
		return i, nil
	case "style", "strong", "emphasis", "strikethrough", "sub", "sup", "code":
		c := &StyleLinkType{}
		s.appendContent(c)
		return c, nil
	default:
		return s.contentBase.tagCallback(start)
	}
}

// UnmarshalXML unmarshal XML to StyleType
func (s *StyleLinkType) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return NewParser(s).Parse(d, start)
}

// Table https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L532
// Basic html-like tables
type Table struct {
	XMLName xml.Name `xml:"table"`
	TR      []*TR    `xml:"tr"`
	ID      string   `xml:"id,omitempty"`
	Style   string   `xml:"style,attr,omitempty"`
	emptyText
}

func (t *Table) tagCallback(start xml.StartElement) (Node, error) {
	switch start.Name.Local {
	case "id":
		s := &stringNode{s: &t.ID}
		return s, nil
	case "tr":
		tr := &TR{}
		t.TR = append(t.TR, tr)
		return tr, nil
	default:
		return t.emptyText.tagCallback(start)
	}
}

func (t *Table) attrCallback(attr xml.Attr) error {
	if attr.Name.Local == "style" {
		t.Style = attr.Value
		return nil
	}
	return t.emptyText.attrCallback(attr)
}

// GetContent for Contenter interface
func (t *Table) GetContent() []Contenter {
	c := make([]Contenter, len(t.TR))
	for i, v := range t.TR {
		c[i] = v
	}
	return c
}

// TR https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L538
type TR struct {
	Align string `xml:"align,attr,omitempty"`
	// td or th
	contentBase
}

func (t *TR) tagCallback(start xml.StartElement) (Node, error) {
	switch start.Name.Local {
	case "th", "td":
		td := &TD{}
		t.appendContent(td)
		return td, nil
	default:
		return t.contentBase.tagCallback(start)
	}
}

func (t *TR) attrCallback(attr xml.Attr) error {
	if attr.Name.Local == "align" {
		t.Align = attr.Value
		return nil
	}
	return t.contentBase.attrCallback(attr)
}

// UnmarshalXML unmarshal XML to Table
func (t *TR) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return NewParser(t).Parse(d, start)
}

// TD https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L700
type TD struct {
	StyleType
	ID      string `xml:"id,omitempty"`
	Style   string `xml:"style,attr,omitempty"`
	Colspan int    `xml:"colspan,attr,omitempty"`
	Rowspan int    `xml:"rowspan,attr,omitempty"`
	Align   string `xml:"align,attr,omitempty"`
	Valign  string `xml:"valign,attr,omitempty"`
}

func (t *TD) attrCallback(attr xml.Attr) error {
	switch attr.Name.Local {
	case "id":
		t.ID = attr.Value
	case "style":
		t.Style = attr.Value
	case "colspan":
		c, err := strconv.Atoi(attr.Value)
		t.Colspan = c
		return err
	case "rowspan":
		r, err := strconv.Atoi(attr.Value)
		t.Rowspan = r
		return err
	case "align":
		t.Align = attr.Value
	case "valign":
		t.Valign = attr.Value
	default:
		return t.contentBase.attrCallback(attr)
	}
	return nil
}

// UnmarshalXML unmarshal XML to TD
func (t *TD) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return NewParser(t).Parse(d, start)
}

// InlineImage https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L712
// It's Contenter, but has no text or "child" content
type InlineImage struct {
	XlinkType string `xml:"http://www.w3.org/1999/xlink type,attr,omitempty"`
	XlinkHref string `xml:"http://www.w3.org/1999/xlink href,attr,omitempty"`
	Alt       string `xml:"alt,attr,omitempty"`
	emptyContent
}

func (i *InlineImage) attrCallback(attr xml.Attr) error {
	switch attr.Name.Local {
	case "type":
		i.XlinkType = attr.Value
	case "href":
		i.XlinkHref = attr.Value
	case "alt":
		i.Alt = attr.Value
	default:
		return i.emptyContent.attrCallback(attr)
	}
	return nil
}

// UnmarshalXML unmarshal XML to InlineImage
func (i *InlineImage) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return NewParser(i).Parse(d, start)
}

type emptyText struct{ baseNode }

func (e emptyText) GetXMLName() xml.Name {
	return e.baseNode.GetXMLName()
}

func (e emptyText) GetText() []byte { return nil }

type emptyContent struct{ emptyText }

func (e emptyContent) GetContent() []Contenter { return nil }

// EmptyLine -- dummy struct for <empty-line/>
type EmptyLine struct {
	XMLName xml.Name `xml:"empty-line"`
	emptyContent
}

// A CharData represents raw text
type CharData xml.CharData

// GetXMLName return empty string for raw text
func (c CharData) GetXMLName() xml.Name {
	return xml.Name{}
}

// GetText return text
func (c CharData) GetText() []byte {
	return []byte(c)
}

// GetContent return nil, because CharData contains only raw text
func (c CharData) GetContent() []Contenter {
	return nil
}

type contentBase struct {
	baseNode
	Content []Contenter
}

// GetText return nil
// All text in CharData type
func (c contentBase) GetText() []byte {
	return nil
}

// GetContent return content
func (c contentBase) GetContent() []Contenter {
	return c.Content
}

func (c *contentBase) appendContent(cont Contenter) {
	c.Content = append(c.Content, cont)
}

type mixed struct {
	contentBase
}

func (m *mixed) charDataCallback(e xml.CharData) error {
	tmp := make(CharData, len(e))
	// copy buffer, golang reuse byte array
	copy(tmp, e)
	m.Content = append(m.Content, tmp)
	return nil
}
