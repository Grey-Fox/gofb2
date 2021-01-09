# GoFB2

GoFB2 is golang structures for parse `.fb2` book format. It's based on
XML schema https://github.com/gribuser/fb2.

Usage example:
```go
package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"strings"

	fb2 "github.com/Grey-Fox/gofb2"
)

func printContent(cont ...fb2.Contenter) {
	for _, c := range cont {
		switch e := c.(type) {
		case fb2.CharData:
			fmt.Printf(strings.ReplaceAll(string(e.GetText()), "\t", ""))
		case fb2.P:
			printContent(e.Content...)
			fmt.Println()
		case fb2.EmptyLine:
			fmt.Println()
		case fb2.Poem:
			if e.Title != nil {
				printContent(e.Title)
			}
			printContent(e.GetContent()...)
		case fb2.Cite:
			fmt.Printf("> ")
			printContent(e.GetContent()...)
			for _, a := range e.TextAuthor {
				fmt.Printf("(c) ")
				printContent(a)
			}
			fmt.Println()
		default:
			printContent(c.GetContent()...)
		}
	}
}

func printSection(s *fb2.Section) {
	for _, s := range s.Sections {
		printSection(&s)
	}
	printContent(s.Content...)
}

func main() {
	data, err := ioutil.ReadFile("example.fb2")
	check(err)

	v := fb2.FictionBook{}
	check(xml.Unmarshal([]byte(data), &v))
	printContent(v.Description.TitleInfo.Annotation)
	printSection(&v.Body.Sections[0])
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
```

Parse only description:
```go
package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"

	fb2 "github.com/Grey-Fox/gofb2"
	"golang.org/x/net/html/charset"
)

func parseDescription(r io.Reader) (*fb2.FictionBook, error) {
	v := &fb2.FictionBook{}
	p := fb2.NewParser(v)
	decoder := xml.NewDecoder(r)
	decoder.CharsetReader = charset.NewReaderLabel

	for {
		token, _ := decoder.Token()
		p.ParseToken(token)
		if e, ok := token.(xml.EndElement); ok && e.Name.Local == "description" {
			break
		}
	}

	return v, nil
}

func main() {
	data, _ := ioutil.ReadFile("example.fb2")

	reader := bytes.NewReader(data)
	v, _ := parseDescription(reader)
	fmt.Println(
		v.Description.TitleInfo.Authors[0].FirstName.Value,
		v.Description.TitleInfo.Authors[0].MiddleName.Value,
		v.Description.TitleInfo.Authors[0].LastName.Value,
	)
}
```

Use libxml2 for parse:
```go
package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"

	fb2 "github.com/Grey-Fox/gofb2"
	"github.com/lestrrat-go/libxml2/parser"
	"github.com/lestrrat-go/libxml2/types"
)

func parse(doc types.Document) (*fb2.FictionBook, error) {
	var walk func(types.NodeList)

	v := &fb2.FictionBook{}
	p := fb2.NewParser(v)
	walk = func(nodes types.NodeList) {
		for _, n := range nodes {
			if e, ok := n.(types.Element); ok {
				attrs, _ := e.Attributes()
				xmlAttrs := make([]xml.Attr, len(attrs))
				for i, a := range attrs {
					xmlAttrs[i] = xml.Attr{
						Name:  xml.Name{Local: a.NodeName()},
						Value: a.Value(),
					}
				}
				st := xml.StartElement{
					Name: xml.Name{Local: e.NodeName()},
					Attr: xmlAttrs,
				}
				p.ParseToken(st)

				cn, _ := e.ChildNodes()
				walk(cn)

				p.ParseToken(xml.EndElement{Name: xml.Name{Local: n.NodeName()}})
			} else {
				p.ParseToken(xml.CharData(n.NodeValue()))
			}
		}
	}

	cn, _ := doc.ChildNodes()
	walk(cn)

	return v, nil
}

func main() {
	data, _ := ioutil.ReadFile("example.fb2")

	reader := bytes.NewReader(data)
	p := parser.New()
	doc, _ := p.ParseReader(reader)
	v, _ := parse(doc)

	fmt.Println(
		v.Description.TitleInfo.Authors[0].FirstName.Value,
		v.Description.TitleInfo.Authors[0].MiddleName.Value,
		v.Description.TitleInfo.Authors[0].LastName.Value,
	)
}
```
