# GoFB2

GoFB2 is a golang structures for parse `.fb2` book format. It's based on
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
