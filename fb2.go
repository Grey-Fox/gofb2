package gofb2

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"strconv"
	"time"
)

// TODO marshal

// Body https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L31
// Main content of the book, multiple bodies are used for additional information,
// like footnotes, that do not appear in the main book flow (extended from this class).
// The first body is presented to the reader by default, and content in the other
// bodies should be accessible by hyperlinks.
type Body struct {
	baseNode

	// Image to be displayed at the top of this section
	Image *Image `xml:"image,omitempty"`

	// A fancy title for the entire book, should be used if the simple text version
	// in "description"; is not adequate, e.g. the book title has multiple paragraphs
	// and/or character styles
	Title *Title `xml:"title,omitempty"`

	// Epigraph(s) for the entire book, if any
	Epigraphs []*Epigraph `xml:"epigraph,omitempty"`

	Sections []*Section `xml:"section"`

	Lang string `xml:"http://www.w3.org/XML/1998/namespace lang,attr,omitempty"`
}

func (b *Body) tagCallback(start xml.StartElement) (Node, error) {
	switch start.Name.Local {
	case "image":
		b.Image = &Image{}
		return b.Image, nil
	case "title":
		b.Title = &Title{}
		return b.Title, nil
	case "epigraph":
		ep := &Epigraph{}
		b.Epigraphs = append(b.Epigraphs, ep)
		return ep, nil
	case "section":
		cs := &Section{}
		b.Sections = append(b.Sections, cs)
		return cs, nil
	default:
		return b.baseNode.tagCallback(start)
	}
}

func (b *Body) attrCallback(attr xml.Attr) error {
	if attr.Name.Local == "lang" {
		b.Lang = attr.Value
		return nil
	}
	return b.baseNode.attrCallback(attr)
}

// UnmarshalXML unmarshal XML to Body
func (b *Body) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return NewParser(b).Parse(d, start)
}

// NotesBody https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L55
// Body for footnotes, content is mostly similar to base type and may (!) be
// rendered in the pure environment "as is". Advanced reader should treat
// section[2]/section as endnotes, all other stuff as footnotes
type NotesBody struct {
	Body
	Name string `xml:"name,attr"`
}

func (b *NotesBody) attrCallback(attr xml.Attr) error {
	if attr.Name.Local == "name" {
		b.Name = attr.Value
		return nil
	}
	return b.Body.attrCallback(attr)
}

// UnmarshalXML unmarshal XML to NotesBody
func (b *NotesBody) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return NewParser(b).Parse(d, start)
}

// FictionBook describe book scheme based on
// https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L71
type FictionBook struct {
	baseNode

	Stylesheet  []*Stylesheet
	Description *Description
	Body        *Body
	NotesBody   *NotesBody
	Binary      []*Binary
}

func (f *FictionBook) tagCallback(start xml.StartElement) (Node, error) {
	switch start.Name.Local {
	case "stylesheet":
		s := &Stylesheet{}
		f.Stylesheet = append(f.Stylesheet, s)
		return s, nil
	case "description":
		f.Description = &Description{}
		return f.Description, nil
	case "body":
		if len(start.Attr) == 0 {
			f.Body = &Body{}
			return f.Body, nil
		} else if start.Attr[0].Name.Local == "name" && start.Attr[0].Value == "notes" {
			f.NotesBody = &NotesBody{}
			return f.NotesBody, nil
		} else {
			return nil, fmt.Errorf("Unknown body with attr %v", start.Attr[0])
		}
	case "binary":
		b := &Binary{}
		f.Binary = append(f.Binary, b)
		return b, nil
	default:
		return f.baseNode.tagCallback(start)
	}
}

func (f *FictionBook) attrCallback(attr xml.Attr) error {
	// TODO
	return nil
}

// UnmarshalXML unmarshal XML
func (f *FictionBook) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return NewParser(f).Parse(d, start)
}

// Stylesheet https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L77
// This element contains an arbitrary stylesheet that is intepreted by a some
// processing programs, e.g. text/css stylesheets can be used by XSLT
// stylesheets to generate better looking html
type Stylesheet struct {
	Type  string `xml:"type,attr"`
	Value []byte `xml:",innerxml"`

	baseNode
}

func (s *Stylesheet) attrCallback(attr xml.Attr) error {
	if attr.Name.Local == "type" {
		s.Type = attr.Value
		return nil
	}
	return s.baseNode.attrCallback(attr)
}

func (s *Stylesheet) charDataCallback(cd xml.CharData) error {
	s.Value = cd
	return nil
}

// UnmarshalXML unmarshal XML
func (s *Stylesheet) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return NewParser(s).Parse(d, start)
}

// Description https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L90
type Description struct {
	baseNode

	// Generic information about the book
	TitleInfo *TitleInfo `xml:"title-info"`

	// Generic information about the original book (for translations)
	SrcTitleInfo *TitleInfo `xml:"src-title-info,omitempty"`

	// Information about this particular (xml) document
	DocumentInfo *DocumentInfo `xml:"document-info"`

	// Information about some paper/outher published document,
	// that was used as a source of this xml document
	PublishInfo *PublishInfo `xml:"publish-info,omitempty"`

	// Any other information about the book/document
	// that didn't fit in the above groups
	CustomInfo []*CustomInfo `xml:"custom-info,omitempty"`

	// Describes, how the document should be presented to end-user, what parts
	// are free, what parts should be sold and what price should be used
	Output []*ShareInstruction `xml:"output,omitempty"`
}

func (d *Description) tagCallback(start xml.StartElement) (Node, error) {
	switch start.Name.Local {
	case "title-info":
		d.TitleInfo = &TitleInfo{}
		return d.TitleInfo, nil
	case "src-title-info":
		d.SrcTitleInfo = &TitleInfo{}
		return d.SrcTitleInfo, nil
	case "document-info":
		d.DocumentInfo = &DocumentInfo{}
		return d.DocumentInfo, nil
	case "publish-info":
		d.PublishInfo = &PublishInfo{}
		return d.PublishInfo, nil
	case "custom-info":
		ci := &CustomInfo{}
		d.CustomInfo = append(d.CustomInfo, ci)
		return ci, nil
	case "output":
		o := &ShareInstruction{}
		d.Output = append(d.Output, o)
		return o, nil
	default:
		return d.baseNode.tagCallback(start)
	}
}

// UnmarshalXML unmarshal XML
func (d *Description) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	return NewParser(d).Parse(dec, start)
}

// DocumentInfo https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L102
type DocumentInfo struct {
	baseNode

	// Author(s) of this particular document
	Authors []*Author `xml:"author"`

	// Any software used in preparation of this document, in free format
	ProgramUsed *TextField `xml:"program-used"`

	// Date this document was created, same guidelines as in the &lt;title-info&gt; section apply
	Date *Date `xml:"date"`

	// Source URL if this document is a conversion of some other (online) document
	SrcURLs []string `xml:"src-url"`

	// Author of the original (online) document, if this is a conversion
	SrcOcr *TextField `xml:"src-ocr"`

	// this is a unique identifier for a document. this must not change
	ID string `xml:"id"`

	// Document version, in free format, should be incremented if the document is changed and re-released to the public
	Version float64 `xml:"version"`

	// Short description for all changes made to this document, like "Added missing chapter 6", in free form.
	History *Annotation `xml:"history"`

	// Owner of the fb2 document copyrights
	Publishers []*Author `xml:"publisher"`
}

func (di *DocumentInfo) tagCallback(start xml.StartElement) (Node, error) {
	switch start.Name.Local {
	case "author":
		a := &Author{}
		di.Authors = append(di.Authors, a)
		return a, nil
	case "program-used":
		di.ProgramUsed = &TextField{}
		return di.ProgramUsed, nil
	case "date":
		di.Date = &Date{}
		return di.Date, nil
	case "src-url":
		return &stringArrayNode{s: &di.SrcURLs}, nil
	case "src-ocr":
		di.SrcOcr = &TextField{}
		return di.SrcOcr, nil
	case "id":
		return &stringNode{s: &di.ID}, nil
	case "version":
		return &floatNode{f: &di.Version}, nil
	case "history":
		di.History = &Annotation{}
		return di.History, nil
	case "publisher":
		a := &Author{}
		di.Publishers = append(di.Publishers, a)
		return a, nil
	default:
		return di.baseNode.tagCallback(start)
	}
}

// UnmarshalXML unmarshal XML
func (di *DocumentInfo) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return NewParser(di).Parse(d, start)
}

// PublishInfo https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L160
type PublishInfo struct {
	baseNode

	// Original (paper) book name
	BookName *TextField `xml:"book-name,omitempty"`
	// Original (paper) book publisher
	Publisher *TextField `xml:"publisher,omitempty"`
	// City where the original (paper) book was published
	City *TextField `xml:"city,omitempty"`
	// Year of the original (paper) publication
	Year string `xml:"year,omitempty"`

	ISBN      *TextField  `xml:"isbn,omitempty"`
	Sequences []*Sequence `xml:"sequence,omitempty"`
}

func (pi *PublishInfo) tagCallback(start xml.StartElement) (Node, error) {
	switch start.Name.Local {
	case "book-name":
		pi.BookName = &TextField{}
		return pi.BookName, nil
	case "publisher":
		pi.Publisher = &TextField{}
		return pi.Publisher, nil
	case "city":
		pi.City = &TextField{}
		return pi.City, nil
	case "year":
		return &stringNode{s: &pi.Year}, nil
	case "isbn":
		pi.ISBN = &TextField{}
		return pi.ISBN, nil
	case "sequence":
		s := &Sequence{}
		pi.Sequences = append(pi.Sequences, s)
		return s, nil
	default:
		return pi.baseNode.tagCallback(start)
	}
}

// UnmarshalXML unmarshal XML
func (pi *PublishInfo) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return NewParser(pi).Parse(d, start)
}

// CustomInfo https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L191
type CustomInfo struct {
	TextField
	InfoType string `xml:"info-type,attr"`
}

func (ci *CustomInfo) attrCallback(attr xml.Attr) error {
	if attr.Name.Local == "info-type" {
		ci.InfoType = attr.Value
		return nil
	}
	return ci.TextField.attrCallback(attr)
}

// UnmarshalXML unmarshal XML
func (ci *CustomInfo) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return NewParser(ci).Parse(d, start)
}

// Binary https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L217
// Any binary data that is required for the presentation of this book in base64
// format. Currently only images are used
type Binary struct {
	baseNode

	ID          string `xml:"id,attr"`
	ContentType string `xml:"content-type,attr"`
	Value       []byte `xml:",chardata"`
}

func (b *Binary) attrCallback(attr xml.Attr) error {
	if attr.Name.Local == "content-type" {
		b.ContentType = attr.Value
	} else if attr.Name.Local == "id" {
		b.ID = attr.Value
	} else {
		return b.baseNode.attrCallback(attr)
	}
	return nil
}

func (b *Binary) charDataCallback(cd xml.CharData) error {
	b.Value = make([]byte, len(cd))
	n, err := base64.StdEncoding.Decode(b.Value, cd)
	if err != nil {
		return err
	}
	b.Value = b.Value[:n]
	return nil
}

// UnmarshalXML decode from base64
func (b *Binary) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return NewParser(b).Parse(d, start)
}

// Author https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L233
// Information about a single author
type Author struct {
	FirstName  *TextField `xml:"first-name,omitempty"`
	MiddleName *TextField `xml:"middle-name,omitempty"`
	LastName   *TextField `xml:"last-name,omitempty"`
	Nickname   *TextField `xml:"nickname,omitempty"`
	HomePages  []string   `xml:"home-page,omitempty"`
	Emails     []string   `xml:"email,omitempty"`
	ID         string     `xml:"id,omitempty"`

	baseNode
}

func (a *Author) tagCallback(start xml.StartElement) (Node, error) {
	switch start.Name.Local {
	case "first-name":
		a.FirstName = &TextField{}
		return a.FirstName, nil
	case "middle-name":
		a.MiddleName = &TextField{}
		return a.MiddleName, nil
	case "last-name":
		a.LastName = &TextField{}
		return a.LastName, nil
	case "nickname":
		a.Nickname = &TextField{}
		return a.Nickname, nil
	case "home-page":
		return &stringArrayNode{s: &a.HomePages}, nil
	case "email":
		return &stringArrayNode{s: &a.Emails}, nil
	case "id":
		return &stringNode{s: &a.ID}, nil
	default:
		return a.baseNode.tagCallback(start)
	}
}

// UnmarshalXML unmarshal XML
func (a *Author) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return NewParser(a).Parse(d, start)
}

// TextField https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L255
type TextField struct {
	baseNode

	Lang  string `xml:"http://www.w3.org/XML/1998/namespace lang,attr,omitempty"`
	Value string `xml:",chardata"`
}

func (t *TextField) attrCallback(attr xml.Attr) error {
	if attr.Name.Local == "lang" {
		t.Lang = attr.Value
		return nil
	}
	return t.baseNode.attrCallback(attr)
}

func (t *TextField) charDataCallback(cd xml.CharData) error {
	t.Value = string(cd)
	return nil
}

// UnmarshalXML unmarshal XML to TextField
func (t *TextField) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return NewParser(t).Parse(d, start)
}

// Date https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L262
// A human readable date, maybe not exact, with an optional computer readable variant
type Date struct {
	Value    *XMLDate `xml:"value,attr,omitempty"`
	Lang     string   `xml:"http://www.w3.org/XML/1998/namespace lang,attr,omitempty"`
	StrValue string   `xml:",chardata"`

	baseNode
}

func (d *Date) attrCallback(attr xml.Attr) error {
	if attr.Name.Local == "value" {
		parse, err := time.Parse(dateFormat, attr.Value)
		if err != nil {
			return err
		}
		d.Value = &XMLDate{parse}
	} else if attr.Name.Local == "lang" {
		d.Lang = attr.Value
	} else {
		return d.baseNode.attrCallback(attr)
	}
	return nil
}

func (d *Date) charDataCallback(cd xml.CharData) error {
	d.StrValue = string(cd)
	return nil
}

// UnmarshalXML unmarshal XML
func (d *Date) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	return NewParser(d).Parse(dec, start)
}

// Sequence https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L521
// Book sequences
type Sequence struct {
	baseNode

	Sequences []*Sequence `xml:"sequence,omitempty"`
	Name      string      `xml:"name,attr"`
	Number    int         `xml:"number,attr"`
}

func (s *Sequence) tagCallback(start xml.StartElement) (Node, error) {
	if start.Name.Local == "sequence" {
		s.Sequences = append(s.Sequences, &Sequence{})
		return s.Sequences[len(s.Sequences)-1], nil
	}
	return s.baseNode.tagCallback(start)
}

func (s *Sequence) attrCallback(attr xml.Attr) error {
	if attr.Name.Local == "name" {
		s.Name = attr.Value
	} else if attr.Name.Local == "number" {
		n, err := strconv.Atoi(attr.Value)
		if err != nil {
			return err
		}
		s.Number = n
	} else {
		return s.baseNode.attrCallback(attr)
	}
	return nil
}

// UnmarshalXML unmarshal XML
func (s *Sequence) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return NewParser(s).Parse(d, start)
}

// TitleInfo https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L570
// Book (as a book opposite a document) description
type TitleInfo struct {
	baseNode

	// Genre of this book, with the optional match percentage
	Genres []*Genre `xml:"genre"`

	// Author(s) of this book
	Authors []*Author `xml:"author"`

	//Book title
	BookTitle *TextField `xml:"book-title"`

	// Annotation for this book
	Annotation *Annotation `xml:"annotation"`

	// Any keywords for this book, intended for use in search engines
	Keywords *TextField `xml:"keywords"`

	// Date this book was written, can be not exact, e.g. 1863-1867.
	// If an optional attribute is present, then it should contain some
	// computer-readable date from the interval for use by search and indexingengines
	Date *Date `xml:"date"`

	// Any coverpage items, currently only images
	Coverpage *Coverpage `xml:"coverpage"`

	// Book's language
	Lang string `xml:"lang"`

	// Book's source language if this is a translation
	SrcLang string `xml:"src-lang,omitempty"`

	// Translators if this is a translation
	Translators []*Author `xml:"translator"`

	// Any sequences this book might be part of
	Sequences []*Sequence `xml:"sequence"`
}

func (ti *TitleInfo) tagCallback(start xml.StartElement) (Node, error) {
	switch start.Name.Local {
	case "genre":
		g := &Genre{}
		ti.Genres = append(ti.Genres, g)
		return g, nil
	case "author":
		a := &Author{}
		ti.Authors = append(ti.Authors, a)
		return a, nil
	case "book-title":
		ti.BookTitle = &TextField{}
		return ti.BookTitle, nil
	case "annotation":
		ti.Annotation = &Annotation{}
		return ti.Annotation, nil
	case "keywords":
		ti.Keywords = &TextField{}
		return ti.Keywords, nil
	case "date":
		ti.Date = &Date{}
		return ti.Date, nil
	case "coverpage":
		ti.Coverpage = &Coverpage{}
		return ti.Coverpage, nil
	case "lang":
		return &stringNode{s: &ti.Lang}, nil
	case "src-lang":
		return &stringNode{s: &ti.SrcLang}, nil
	case "translator":
		a := &Author{}
		ti.Translators = append(ti.Translators, a)
		return a, nil
	case "sequence":
		s := &Sequence{}
		ti.Sequences = append(ti.Sequences, s)
		return s, nil
	default:
		return ti.baseNode.tagCallback(start)
	}
}

// UnmarshalXML unmarshal XML
func (ti *TitleInfo) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return NewParser(ti).Parse(d, start)
}

// Genre https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L581
type Genre struct {
	baseNode

	Match *int   `xml:"match,attr,omitempty"`
	Genre string `xml:",chardata"`
}

func (g *Genre) attrCallback(attr xml.Attr) error {
	if attr.Name.Local == "match" {
		match, err := strconv.Atoi(attr.Value)
		if err != nil {
			return err
		}
		g.Match = &match
		return nil
	}
	return g.baseNode.attrCallback(attr)
}

func (g *Genre) charDataCallback(cd xml.CharData) error {
	g.Genre = string(cd)
	return nil
}

// UnmarshalXML unmarshal XML
func (g *Genre) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return NewParser(g).Parse(d, start)
}

// Coverpage https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L621
type Coverpage struct {
	baseNode
	Image *InlineImage `xml:"image"`
}

func (c *Coverpage) tagCallback(start xml.StartElement) (Node, error) {
	if start.Name.Local == "image" {
		c.Image = &InlineImage{}
		return c.Image, nil
	}
	return c.baseNode.tagCallback(start)
}

// UnmarshalXML unmarshal XML
func (c *Coverpage) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return NewParser(c).Parse(d, start)
}

// ShareInstruction https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L649
// In-document instruction for generating output free and payed documents
type ShareInstruction struct {
	baseNode

	Mode                ShareMode                `xml:"mode,attr"`
	IncludeAll          DocGenerationInstruction `xml:"include-all,attr"`
	Price               float64                  `xml:"price,attr,omitempty"`
	Currency            string                   `xml:"currency,attr,omitempty"`
	Parts               []*PartShareInstruction  `xml:"part,omitempty"`
	OutputDocumentClass []*OutPutDocument        `xml:"output-document-class,omitempty"`
}

// UnmarshalXML unmarshal XML
func (si *ShareInstruction) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return NewParser(si).Parse(d, start)
}
func (si *ShareInstruction) attrCallback(attr xml.Attr) error {
	switch attr.Name.Local {
	case "mode":
		si.Mode = ShareMode(attr.Value)
	case "include-all":
		si.IncludeAll = DocGenerationInstruction(attr.Value)
	case "price":
		p, err := strconv.ParseFloat(attr.Value, 64)
		si.Price = p
		return err
	case "currency":
		si.Currency = attr.Value
	default:
		return si.baseNode.attrCallback(attr)
	}
	return nil
}

func (si *ShareInstruction) tagCallback(start xml.StartElement) (Node, error) {
	switch start.Name.Local {
	case "part":
		p := &PartShareInstruction{}
		si.Parts = append(si.Parts, p)
		return p, nil
	case "output-document-class":
		o := &OutPutDocument{}
		si.OutputDocumentClass = append(si.OutputDocumentClass, o)
		return o, nil
	default:
		return si.baseNode.tagCallback(start)
	}
}

// ShareMode https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L662
// Modes for document sharing (free|paid for now)
type ShareMode string

// DocGenerationInstruction https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L671
// List of instructions to process sections (allow|deny|require)
type DocGenerationInstruction string

// PartShareInstruction https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L681
// Pointer to specific document section, explaining how to deal with it
type PartShareInstruction struct {
	baseNode

	XlinkType string                   `xml:"http://www.w3.org/1999/xlink type,attr,omitempty"`
	XlinkHref string                   `xml:"http://www.w3.org/1999/xlink href,attr"`
	Include   DocGenerationInstruction `xml:"include,attr"`
}

func (psi *PartShareInstruction) attrCallback(attr xml.Attr) error {
	switch attr.Name.Local {
	case "type":
		psi.XlinkType = attr.Value
	case "href":
		psi.XlinkHref = attr.Value
	case "include":
		psi.Include = DocGenerationInstruction(attr.Value)
	default:
		return psi.baseNode.attrCallback(attr)
	}
	return nil
}

// UnmarshalXML unmarshal XML
func (psi *PartShareInstruction) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return NewParser(psi).Parse(d, start)
}

// OutPutDocument https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L689
// Selector for output documents. Defines, which rule to apply to any specific output documents
type OutPutDocument struct {
	baseNode

	Name   string                   `xml:"name,attr"`
	Create DocGenerationInstruction `xml:"create,attr,omitempty"`
	Price  float64                  `xml:"price,attr,omitempty"`
	Parts  []*PartShareInstruction  `xml:"part,omitempty"`
}

func (od *OutPutDocument) attrCallback(attr xml.Attr) error {
	switch attr.Name.Local {
	case "name":
		od.Name = attr.Value
	case "create":
		od.Create = DocGenerationInstruction(attr.Value)
	case "price":
		p, err := strconv.ParseFloat(attr.Value, 64)
		od.Price = p
		return err
	default:
		return od.baseNode.attrCallback(attr)
	}
	return nil
}

func (od *OutPutDocument) tagCallback(start xml.StartElement) (Node, error) {
	switch start.Name.Local {
	case "part":
		p := &PartShareInstruction{}
		od.Parts = append(od.Parts, p)
		return p, nil
	default:
		return od.baseNode.tagCallback(start)
	}
}

// UnmarshalXML unmarshal XML
func (od *OutPutDocument) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return NewParser(od).Parse(d, start)
}
