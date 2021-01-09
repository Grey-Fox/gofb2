package gofb2

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
)

// Node define the basic interface for XML nodes
type Node interface {
	SetXMLName(xml.Name)
	GetXMLName() xml.Name
	tagCallback(xml.StartElement) (Node, error)
	attrCallback(xml.Attr) error
	charDataCallback(xml.CharData) error
}

type baseNode struct {
	XMLName xml.Name
}

func (n *baseNode) tagCallback(start xml.StartElement) (Node, error) {
	return nil, fmt.Errorf("unexpected tag %s", start.Name)
}

func (n *baseNode) attrCallback(attr xml.Attr) error {
	return fmt.Errorf("unexpected attr %s", attr.Name)
}

func (n *baseNode) charDataCallback(xml.CharData) error {
	return nil
}

func (n *baseNode) SetXMLName(name xml.Name) {
	n.XMLName = name
}

func (n *baseNode) GetXMLName() xml.Name {
	return n.XMLName
}

type stringNode struct {
	baseNode
	s *string
}

func (s *stringNode) charDataCallback(cd xml.CharData) error {
	*s.s = string(cd)
	return nil
}

type floatNode struct {
	baseNode
	f *float64
}

func (f *floatNode) charDataCallback(cd xml.CharData) error {
	fl, err := strconv.ParseFloat(string(cd), 64)
	*f.f = fl
	return err
}

type stringArrayNode struct {
	s *[]string
	baseNode
}

func (s *stringArrayNode) charDataCallback(cd xml.CharData) error {
	*s.s = append(*s.s, string(cd))
	return nil
}

// Parser parse xml document
type Parser struct {
	stack []Node
	last  Node
	first Node
}

// NewParser return new parser
func NewParser(n Node) *Parser {
	return &Parser{first: n}
}

// ParseToken parse one xml.Token.
// StartElement, EndElement or CharData.
func (p *Parser) ParseToken(token xml.Token) error {
	switch e := token.(type) {
	case xml.StartElement:
		if p.last != nil {
			nt, err := p.last.tagCallback(e)
			if err != nil {
				return err
			}
			p.stack = append(p.stack, p.last)
			p.last = nt
		} else {
			p.last = p.first
		}
		p.last.SetXMLName(e.Name)

		for _, attr := range e.Attr {
			err := p.last.attrCallback(attr)
			if err != nil {
				return err
			}
		}
	case xml.EndElement:
		if p.last.GetXMLName() != e.Name {
			return fmt.Errorf("unexpected close tag %s", e.Name)
		}
		if len(p.stack) > 0 {
			p.last = p.stack[len(p.stack)-1]
			p.stack = p.stack[:len(p.stack)-1]
		} else {
			p.last = nil
		}
	case xml.CharData:
		if p.last != nil {
			err := p.last.charDataCallback(e)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Parse xml document
func (p *Parser) Parse(d *xml.Decoder, start xml.StartElement) error {
	var parseErr parseErrors
	p.ParseToken(start)
	for {
		token, err := d.Token()
		if err != nil {
			if err != io.EOF {
				parseErr = append(parseErr, err)
			}
			break
		}
		err = p.ParseToken(token)
	}
	if parseErr != nil {
		return fmt.Errorf("error while parsing %s: %s", start.Name.Local, parseErr)
	}
	return nil
}
