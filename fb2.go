package gofb2

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
)

// TODO marshal

// Body https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L31
// Main content of the book, multiple bodies are used for additional information,
// like footnotes, that do not appear in the main book flow (extended from this class).
// The first body is presented to the reader by default, and content in the other
// bodies should be accessible by hyperlinks.
type Body struct {
	XMLName xml.Name `xml:"body"`

	// Image to be displayed at the top of this section
	Image *Image `xml:"image,omitempty"`

	// A fancy title for the entire book, should be used if the simple text version
	// in "description"; is not adequate, e.g. the book title has multiple paragraphs
	// and/or character styles
	Title *Title `xml:"title,omitempty"`

	// Epigraph(s) for the entire book, if any
	Epigraphs []Epigraph `xml:"epigraph,omitempty"`

	Sections []Section `xml:"section"`

	Lang string `xml:"http://www.w3.org/XML/1998/namespace lang,attr,omitempty"`
}

// NotesBody https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L55
// Body for footnotes, content is mostly similar to base type and may (!) be
// rendered in the pure environment "as is". Advanced reader should treat
// section[2]/section as endnotes, all other stuff as footnotes
type NotesBody struct {
	Body
	Name string `xml:"name,attr"`
}

// FictionBook describe book scheme based on
// https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L71
type FictionBook struct {
	XMLName     xml.Name `xml:"FictionBook"`
	Stylesheet  []*Stylesheet
	Description *Description
	Body        *Body
	NotesBody   *NotesBody
	Binary      []*Binary
}

// UnmarshalXML unmarshal xml to FictionBook struct
func (f *FictionBook) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var parseErr parseErrors
	f.XMLName = start.Name
	for {
		token, err := d.Token()
		if err != nil {
			if err != io.EOF {
				parseErr = append(parseErr, err)
			}
			break
		}
		e, ok := token.(xml.StartElement)
		if !ok {
			continue
		}
		switch e.Name.Local {
		case "stylesheet":
			s := &Stylesheet{}
			if err := d.DecodeElement(s, &e); err != nil {
				parseErr = append(parseErr, err)
			} else {
				f.Stylesheet = append(f.Stylesheet, s)
			}
		case "description":
			f.Description = &Description{}
			if err := d.DecodeElement(f.Description, &e); err != nil {
				parseErr = append(parseErr, err)
			}
		case "body":
			if len(e.Attr) == 0 {
				f.Body = &Body{}
				if err := d.DecodeElement(f.Body, &e); err != nil {
					parseErr = append(parseErr, err)
				}
			} else if e.Attr[0].Name.Local == "name" && e.Attr[0].Value == "notes" {
				f.NotesBody = &NotesBody{}
				if err := d.DecodeElement(f.NotesBody, &e); err != nil {
					parseErr = append(parseErr, err)
				}
			} else {
				parseErr = append(parseErr, fmt.Errorf("Unknown body with attr %v", e.Attr[0]))
			}
		case "binary":
			b := &Binary{}
			if err := d.DecodeElement(b, &e); err != nil {
				parseErr = append(parseErr, err)
			}
			f.Binary = append(f.Binary, b)
		}
	}
	if parseErr != nil {
		return fmt.Errorf("error while parsing %s: %s", start.Name.Local, parseErr)
	}
	return nil
}

// Stylesheet https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L77
// This element contains an arbitrary stylesheet that is intepreted by a some
// processing programs, e.g. text/css stylesheets can be used by XSLT
// stylesheets to generate better looking html
type Stylesheet struct {
	XMLName xml.Name `xml:"stylesheet"`
	Type    string   `xml:"type,attr"`
	Value   []byte   `xml:",innerxml"`
}

// Description https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L90
type Description struct {
	// Generic information about the book
	TitleInfo *TitleInfo `xml:"title-info"`

	// Generic information about the original book (for translations)
	SrcTitleInfo *TitleInfo `xml:"src-title-info,omitempty"`

	// Information about this particular (xml) document
	DocumentInfo struct {
		// Author(s) of this particular document
		Authors []Author `xml:"author"`

		// Any software used in preparation of this document, in free format
		ProgramUsed *TextField `xml:"program-used"`

		// Date this document was created, same guidelines as in the &lt;title-info&gt; section apply
		Date Date

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
		Publishers []Author `xml:"publisher"`
	} `xml:"document-info"`

	// Information about some paper/outher published document,
	// that was used as a source of this xml document
	PublishInfo *struct {
		// Original (paper) book name
		BookName *TextField `xml:"book-name,omitempty"`
		// Original (paper) book publisher
		Publisher *TextField `xml:"publisher,omitempty"`
		// City where the original (paper) book was published
		City *TextField `xml:"city,omitempty"`
		// Year of the original (paper) publication
		Year string `xml:"year,omitempty"`

		ISBN      *TextField `xml:"isbn,omitempty"`
		Sequences []Sequence `xml:"sequence,omitempty"`
	} `xml:"publish-info,omitempty"`

	// Any other information about the book/document
	// that didn't fit in the above groups
	CustomInfo []struct {
		TextField
		InfoType string `xml:"info-type,attr"`
	} `xml:"custom-info,omitempty"`

	// Describes, how the document should be presented to end-user, what parts
	// are free, what parts should be sold and what price should be used
	Output []ShareInstruction `xml:"output,omitempty"`
}

// Binary https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L217
// Any binary data that is required for the presentation of this book in base64
// format. Currently only images are used
type Binary struct {
	ID          string `xml:"id,attr"`
	ContentType string `xml:"content-type,attr"`
	Value       []byte `xml:",chardata"`
}

// UnmarshalXML decode from base64
func (b *Binary) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type T Binary
	var overlay struct {
		*T
	}
	overlay.T = (*T)(b)

	if err := d.DecodeElement(&overlay, &start); err != nil {
		return err
	}
	_, err := base64.StdEncoding.Decode(b.Value, overlay.T.Value)
	return err
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
}

// TextField https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L255
type TextField struct {
	Lang  string `xml:"http://www.w3.org/XML/1998/namespace lang,attr,omitempty"`
	Value string `xml:",chardata"`
}

// Date https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L262
// A human readable date, maybe not exact, with an optional computer readable variant
type Date struct {
	Value    *XMLDate `xml:"value,attr,omitempty"`
	Lang     string   `xml:"http://www.w3.org/XML/1998/namespace lang,attr,omitempty"`
	StrValue string   `xml:",chardata"`
}

// Sequence https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L521
// Book sequences
type Sequence struct {
	Sequences []Sequence `xml:"sequence,omitempty"`
	Name      string     `xml:"name,attr"`
	Number    int        `xml:"number,attr"`
}

// TitleInfo https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L570
// Book (as a book opposite a document) description
type TitleInfo struct {
	// Genre of this book, with the optional match percentage
	Genres []struct {
		Genre `xml:",chardata"`
		Match *int `xml:"match,attr,omitempty"`
	} `xml:"genre"`

	// Author(s) of this book
	Authors []Author `xml:"author"`

	//Book title
	BookTitle TextField `xml:"book-title"`

	// Annotation for this book
	Annotation *Annotation `xml:"annotation"`

	// Any keywords for this book, intended for use in search engines
	Keywords *TextField `xml:"keywords"`

	// Date this book was written, can be not exact, e.g. 1863-1867.
	// If an optional attribute is present, then it should contain some
	// computer-readable date from the interval for use by search and indexingengines
	Date *Date `xml:"date"`

	// Any coverpage items, currently only images
	Coverpage *struct {
		Image InlineImage `xml:"image"`
	} `xml:"coverpage"`

	// Book's language
	Lang string `xml:"lang"`

	// Book's source language if this is a translation
	SrcLang string `xml:"src-lang,omitempty"`

	// Translators if this is a translation
	Translators []Author `xml:"translator"`

	// Any sequences this book might be part of
	Sequences []Sequence `xml:"sequence"`
}

// ShareInstruction https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L649
// In-document instruction for generating output free and payed documents
type ShareInstruction struct {
	Mode                ShareMode                `xml:"mode,attr"`
	IncludeAll          DocGenerationInstruction `xml:"include-all,attr"`
	Price               float64                  `xml:"price,attr,omitempty"`
	Currency            string                   `xml:"currency,attr,omitempty"`
	Parts               []PartShareInstruction   `xml:"part,omitempty"`
	OutputDocumentClass []OutPutDocument         `xml:"output-document-class,omitempty"`
}

// ShareMode https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L662
// Modes for document sharing (free|paid for now)
type ShareMode string

// DocGenerationInstruction https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#671
// List of instructions to process sections (allow|deny|require)
type DocGenerationInstruction string

// PartShareInstruction https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#681
// Pointer to specific document section, explaining how to deal with it
type PartShareInstruction struct {
	XlinkType string                   `xml:"http://www.w3.org/1999/xlink type,attr,omitempty"`
	XlinkHref string                   `xml:"http://www.w3.org/1999/xlink href,attr"`
	Include   DocGenerationInstruction `xml:"include,attr"`
}

// OutPutDocument https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#689
// Selector for output documents. Defines, which rule to apply to any specific output documents
type OutPutDocument struct {
	Name   string                   `xml:"name,attr"`
	Create DocGenerationInstruction `xml:"create,attr,omitempty"`
	Price  float64                  `xml:"price,attr,omitempty"`
	Parts  []PartShareInstruction   `xml:"part,omitempty"`
}

// InlineImage https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBook.xsd#L712
// It's Contenter, but has no text or "child" content
type InlineImage struct {
	XMLName   xml.Name `xml:"image"`
	XlinkType string   `xml:"http://www.w3.org/1999/xlink type,attr,omitempty"`
	XlinkHref string   `xml:"http://www.w3.org/1999/xlink href,attr,omitempty"`
	Alt       string   `xml:"alt,attr,omitempty"`
	emptyContent
}

// GetXMLName for Contenter interface
func (i *InlineImage) GetXMLName() string {
	return i.XMLName.Local
}

// Genre https://github.com/gribuser/fb2/blob/14b5fcc6/FictionBookGenres.xsd#L4
type Genre string
