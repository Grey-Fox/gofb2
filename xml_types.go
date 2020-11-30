package gofb2

import (
	"encoding/xml"
	"time"
)

const (
	dateFormat = "2006-01-02"
)

// XMLDate is a time.Time wrapper for correct unmarshalling
type XMLDate struct {
	time.Time
}

// UnmarshalXMLAttr unmarshal xml xs:date to golang time.Time
func (d *XMLDate) UnmarshalXMLAttr(attr xml.Attr) error {
	parse, err := time.Parse(dateFormat, attr.Value)

	if err != nil {
		return err
	}

	d.Time = parse
	return nil
}
