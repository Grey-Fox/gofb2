package gofb2

import (
	"encoding/xml"
	"fmt"
	"io"
	"reflect"
	"strconv"
)

// Contenter provide interface for tag content
type Contenter interface {
	GetXMLName() string
	GetContent() []Contenter
	GetText() []byte
}

// Title https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L273
// A title, used in sections, poems and body elements
type Title struct {
	Lang string `xml:"http://www.w3.org/XML/1998/namespace lang,attr,omitempty"`
	contentBase
}

// UnmarshalXML unmarshal XML to Title
func (t *Title) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	tagCallback := func(name string, e xml.StartElement) (Contenter, error) {
		switch name {
		case "p":
			return &P{}, nil
		case "empty-line":
			return &EmptyLine{}, nil
		default:
			return nil, fmt.Errorf("unknown tag %s", name)
		}
	}
	attrCallback := func(attr xml.Attr) error {
		if attr.Name.Local == "lang" {
			t.Lang = attr.Value
			return nil
		}
		return t.contentBase.attrCallback(attr)
	}
	return t.unmarshalHelper(d, start, tagCallback, attrCallback, false)
}

// Image https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L283
// An empty element with an image name as an attribute
type Image struct {
	XMLName   xml.Name `xml:"image"`
	XlinkType string   `xml:"http://www.w3.org/1999/xlink type,attr,omitempty"`
	XlinkHref string   `xml:"http://www.w3.org/1999/xlink href,attr,omitempty"`
	Alt       string   `xml:"alt,attr,omitempty"`
	Title     string   `xml:"title,attr,omitempty"`
	ID        string   `xml:"id,attr,omitempty"`

	emptyContent
}

// GetXMLName for Contenter interface
func (i Image) GetXMLName() string { return i.XMLName.Local }

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
	return p.unmarshalHelper(d, start, p.tagCallback, p.attrCallback, true)
}

// Cite https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L304
// A citation with an optional citation author at the end
type Cite struct {
	ID         string `xml:"id,omitempty"`
	Lang       string `xml:"http://www.w3.org/XML/1998/namespace lang,attr,omitempty"`
	TextAuthor []P    `xml:"text-author,omitempty"`
	contentBase
}

// UnmarshalXML unmarshal XML to Cite
func (c *Cite) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	tagCallback := func(name string, e xml.StartElement) (Contenter, error) {
		switch name {
		case "text-author":
			p := P{}
			err := d.DecodeElement(&p, &e)
			c.TextAuthor = append(c.TextAuthor, p)
			return nil, err
		case "p":
			return &P{}, nil
		case "poem":
			return &Poem{}, nil
		case "subtitle":
			return &P{}, nil
		case "table":
			return &Table{}, nil
		case "empty-line":
			return &EmptyLine{}, nil
		default:
			return nil, fmt.Errorf("unknown tag %s", name)
		}
	}
	attrCallback := func(attr xml.Attr) error {
		if attr.Name.Local == "lang" {
			c.Lang = attr.Value
		} else if attr.Name.Local == "id" {
			c.ID = attr.Value
		} else {
			return fmt.Errorf("unknown attr %s", attr.Name.Local)
		}
		return nil
	}
	return c.unmarshalHelper(d, start, tagCallback, attrCallback, false)
}

// Poem https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L321
// A poem
type Poem struct {
	// Poem title
	XMLName xml.Name `xml:"poem"`
	Title   *Title   `xml:"title,omitempty"`

	// Poem epigraph(s), if any
	Epigraphs []Epigraph `xml:"epigraph,omitempty"`

	// subtitle and stanza
	contentBase
}

// UnmarshalXML unmarshal XML to StyleType
func (p *Poem) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	tagCallback := func(name string, e xml.StartElement) (Contenter, error) {
		switch name {
		case "title":
			p.Title = &Title{}
			err := d.DecodeElement(p.Title, &e)
			return nil, err
		case "epigraph":
			ep := Epigraph{}
			err := d.DecodeElement(&ep, &e)
			if err == nil {
				p.Epigraphs = append(p.Epigraphs, ep)
			}
			return nil, err
		case "subtitle":
			return &P{}, nil
		case "stanza":
			return &Stanza{}, nil
		default:
			return nil, fmt.Errorf("unknown tag %s", name)
		}
	}
	return p.unmarshalHelper(d, start, tagCallback, nil, false)
}

// Stanza https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L338
// Each poem should have at least one stanza.
// Stanzas are usually separated with empty lines by user agents.
type Stanza struct {
	XMLName  xml.Name `xml:"stanza"`
	Lang     string   `xml:"http://www.w3.org/XML/1998/namespace lang,attr,omitempty"`
	Title    *Title   `xml:"title,omitempty"`
	Subtitle P        `xml:"subtitle,omitempty"`
	// An individual line in a stanza
	V []P `xml:"v"`

	emptyText
}

// GetXMLName for Contenter interface
func (s Stanza) GetXMLName() string {
	return s.XMLName.Local
}

// GetContent for Contenter interface
func (s Stanza) GetContent() []Contenter {
	c := make([]Contenter, len(s.V))
	for i, v := range s.V {
		c[i] = v
	}
	return c
}

// Epigraph https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L366
// An epigraph
type Epigraph struct {
	ID         string `xml:"id,omitempty"`
	TextAuthor []P    `xml:"text-author,omitempty"`
	contentBase
}

// UnmarshalXML unmarshal XML to Epigraph
func (ep *Epigraph) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	tagCallback := func(name string, e xml.StartElement) (Contenter, error) {
		switch name {
		case "text-author":
			p := P{}
			err := d.DecodeElement(&p, &e)
			ep.TextAuthor = append(ep.TextAuthor, p)
			return nil, err
		case "p":
			return &P{}, nil
		case "poem":
			return &Poem{}, nil
		case "cite":
			return &Cite{}, nil
		case "empty-line":
			return &EmptyLine{}, nil
		default:
			return nil, fmt.Errorf("unknown tag %s", name)
		}
	}
	attrCallback := func(attr xml.Attr) error {
		if attr.Name.Local == "id" {
			ep.ID = attr.Value
		} else {
			return fmt.Errorf("unknown attr %s", attr.Name.Local)
		}
		return nil
	}
	return ep.unmarshalHelper(d, start, tagCallback, attrCallback, false)
}

// Annotation https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L381
// A cut-down version of "section" used in annotations
type Annotation struct {
	contentBase
	ID   string `xml:"id,omitempty"`
	Lang string `xml:"http://www.w3.org/XML/1998/namespace lang,attr,omitempty"`
}

// UnmarshalXML unmarshal xml to FictionBook struct
func (a *Annotation) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	tagCallback := func(name string, e xml.StartElement) (Contenter, error) {
		switch name {
		case "p":
			return &P{}, nil
		case "poem":
			return &Poem{}, nil
		case "cite":
			return &Cite{}, nil
		case "subtitle":
			return &P{}, nil
		case "table":
			return &Table{}, nil
		case "empty-line":
			return &EmptyLine{}, nil
		default:
			return nil, fmt.Errorf("unknown tag %s", name)
		}
	}
	attrCallback := func(attr xml.Attr) error {
		if attr.Name.Local == "lang" {
			a.Lang = attr.Value
		} else if attr.Name.Local == "id" {
			a.ID = attr.Value
		} else {
			return fmt.Errorf("unknown attr %s", attr.Name.Local)
		}
		return nil
	}
	return a.unmarshalHelper(d, start, tagCallback, attrCallback, false)
}

// Section https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L396
// A basic block of a book, can contain more child sections or textual content
type Section struct {
	// Section's title
	Title *Title `xml:"title,omitempty"`

	// Epigraph(s) for this section
	Epigraphs []Epigraph `xml:"epigraph,omitempty"`

	// Image to be displayed at the top of this section
	Image *Image `xml:"image,omitempty"`

	// Annotation for this section, if any
	Annotation *Annotation `xml:"annotation,omitempty"`

	// or child Sections
	Sections []Section `xml:"section,omitempty"`
	// or other content
	contentBase

	ID   string `xml:"id,omitempty"`
	Lang string `xml:"http://www.w3.org/XML/1998/namespace lang,attr,omitempty"`
}

// UnmarshalXML unmarshal XML to Section
func (s *Section) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	parsingContent := false
	tagCallback := func(name string, e xml.StartElement) (Contenter, error) {
		switch name {
		case "title":
			s.Title = &Title{}
			return nil, d.DecodeElement(s.Title, &e)
		case "epigraph":
			ep := Epigraph{}
			err := d.DecodeElement(&ep, &e)
			if err == nil {
				s.Epigraphs = append(s.Epigraphs, ep)
			}
			return nil, err
		case "image":
			if parsingContent {
				return &Image{}, nil
			}
			s.Image = &Image{}
			return nil, d.DecodeElement(s.Image, &e)
		case "annotation":
			s.Annotation = &Annotation{}
			return nil, d.DecodeElement(s.Annotation, &e)
		case "section":
			cs := Section{}
			err := d.DecodeElement(&cs, &e)
			if err == nil {
				s.Sections = append(s.Sections, cs)
			}
			return nil, err
		case "p":
			parsingContent = true
			return &P{}, nil
		case "poem":
			parsingContent = true
			return &Poem{}, nil
		case "subtitle":
			parsingContent = true
			return &P{}, nil
		case "cite":
			parsingContent = true
			return &Cite{}, nil
		case "empty-line":
			parsingContent = true
			return &EmptyLine{}, nil
		case "table":
			parsingContent = true
			return &Table{}, nil
		default:
			return nil, fmt.Errorf("unknown tag %s", name)
		}
	}
	attrCallback := func(attr xml.Attr) error {
		if attr.Name.Local == "lang" {
			s.Lang = attr.Value
		} else if attr.Name.Local == "id" {
			s.ID = attr.Value
		} else {
			return fmt.Errorf("unknown attr %s", attr.Name.Local)
		}
		return nil
	}
	return s.unmarshalHelper(d, start, tagCallback, attrCallback, false)
}

// StyleType https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L453
// Markup
type StyleType struct {
	contentBase
	XMLName xml.Name
	Lang    string `xml:"http://www.w3.org/XML/1998/namespace lang,attr,omitempty"`
}

func (s *StyleType) tagCallback(name string, e xml.StartElement) (Contenter, error) {
	switch name {
	case "style":
		return &NamedStyleType{}, nil
	case "a":
		return &Link{}, nil
	case "image":
		return &InlineImage{}, nil
	case "strong", "emphasis", "strikethrough", "sub", "sup", "code":
		return &StyleType{}, nil
	default:
		return nil, fmt.Errorf("unknown tag %s", name)
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
	return s.unmarshalHelper(d, start, s.tagCallback, s.attrCallback, true)
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
	return s.unmarshalHelper(d, start, s.tagCallback, s.attrCallback, true)
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
		l.XlinkType = attr.Value
	default:
		return fmt.Errorf("unknown attr %s", attr.Name.Local)
	}
	return nil
}

// UnmarshalXML unmarshal XML to Link
func (l *Link) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return l.unmarshalHelper(d, start, l.tagCallback, l.attrCallback, true)
}

// StyleLinkType https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L506
// Markup
type StyleLinkType struct {
	contentBase
}

func (s *StyleLinkType) tagCallback(name string, e xml.StartElement) (Contenter, error) {
	switch name {
	case "image":
		return &InlineImage{}, nil
	case "style", "strong", "emphasis", "strikethrough", "sub", "sup", "code":
		return &StyleLinkType{}, nil
	default:
		return nil, fmt.Errorf("unknown tag %s", name)
	}
}

// UnmarshalXML unmarshal XML to StyleType
func (s *StyleLinkType) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return s.unmarshalHelper(d, start, s.tagCallback, nil, true)
}

// Table https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L532
// Basic html-like tables
type Table struct {
	XMLName xml.Name `xml:"table"`
	TR      []TR     `xml:"tr"`
	ID      string   `xml:"id,omitempty"`
	Style   string   `xml:"style,attr,omitempty"`
	emptyText
}

// GetXMLName for Contenter interface
func (t Table) GetXMLName() string { return t.XMLName.Local }

// GetContent for Contenter interface
func (t Table) GetContent() []Contenter {
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

// UnmarshalXML unmarshal XML to Table
func (t *TR) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	tagCallback := func(name string, e xml.StartElement) (Contenter, error) {
		switch name {
		case "th", "td":
			return &TD{}, nil
		default:
			return nil, fmt.Errorf("unknown tag %s", name)
		}
	}
	attrCallback := func(attr xml.Attr) error {
		if attr.Name.Local == "align" {
			t.Align = attr.Value
			return nil
		}
		return t.contentBase.attrCallback(attr)
	}
	return t.unmarshalHelper(d, start, tagCallback, attrCallback, false)
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
		return fmt.Errorf("unknown attr %s", attr.Name.Local)
	}
	return nil
}

// UnmarshalXML unmarshal XML to TD
func (t *TD) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return t.unmarshalHelper(d, start, t.tagCallback, t.attrCallback, false)
}

type emptyText struct{}

func (e emptyText) GetText() []byte { return nil }

type emptyContent struct{ emptyText }

func (e emptyContent) GetContent() []Contenter { return nil }

// EmptyLine -- dummy struct for <empty-line/>
type EmptyLine struct {
	XMLName xml.Name `xml:"empty-line"`
	emptyContent
}

// GetXMLName for Contenter interface
func (e EmptyLine) GetXMLName() string { return e.XMLName.Local }

// A CharData represents raw text
type CharData xml.CharData

// GetXMLName return empty string for raw text
func (c CharData) GetXMLName() string {
	return ""
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
	XMLName xml.Name
	Content []Contenter
}

// GetXMLName return type of content (p, strong, strikethrough, ...)
func (c contentBase) GetXMLName() string {
	return c.XMLName.Local
}

// GetText return nil
// All text in CharData type
func (c contentBase) GetText() []byte {
	return nil
}

// GetContent return nil, because CharData contains only raw text
func (c contentBase) GetContent() []Contenter {
	return c.Content
}

func (c *contentBase) attrCallback(attr xml.Attr) error {
	return fmt.Errorf("unknown attr: %s", attr.Name.Local)
}

func (c *contentBase) unmarshalHelper(
	d *xml.Decoder, start xml.StartElement,
	tagCallback func(string, xml.StartElement) (Contenter, error),
	attrCallback func(xml.Attr) error,
	mixed bool,
) error {
	var parseErr parseErrors
	c.XMLName = start.Name
	for {
		token, err := d.Token()
		if err != nil {
			if err != io.EOF {
				parseErr = append(parseErr, err)
			}
			break
		}
		switch e := token.(type) {
		case xml.StartElement:
			o, err := tagCallback(e.Name.Local, e)
			if err != nil {
				parseErr = append(parseErr, err)
			} else if o != nil {
				if err := d.DecodeElement(o, &e); err != nil {
					parseErr = append(parseErr, err)
				}
				val := reflect.ValueOf(o)
				c.Content = append(c.Content, val.Elem().Interface().(Contenter))
			}
		case xml.CharData:
			if mixed {
				tmp := make(CharData, len(e))
				// copy buffer, golang reuse byte array
				copy(tmp, e)
				c.Content = append(c.Content, tmp)
			}
		}
	}
	if attrCallback == nil {
		attrCallback = c.attrCallback
	}
	for _, attr := range start.Attr {
		if err := attrCallback(attr); err != nil {
			parseErr = append(parseErr, err)
		}
	}
	if parseErr != nil {
		return fmt.Errorf("error while parsing %s: %s", start.Name.Local, parseErr)
	}
	return nil
}
