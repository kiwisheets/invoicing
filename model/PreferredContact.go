package model

import (
	"fmt"
	"io"
	"strconv"
)

type PreferredContact string

const (
	PreferredContactPhone  PreferredContact = "PHONE"
	PreferredContactMobile PreferredContact = "MOBILE"
	PreferredContactEmail  PreferredContact = "EMAIL"
)

var AllPreferredContact = []PreferredContact{
	PreferredContactPhone,
	PreferredContactMobile,
	PreferredContactEmail,
}

func (e PreferredContact) IsValid() bool {
	switch e {
	case PreferredContactPhone, PreferredContactMobile, PreferredContactEmail:
		return true
	}
	return false
}

func (e PreferredContact) String() string {
	return string(e)
}

func (e *PreferredContact) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = PreferredContact(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid PreferredContact", str)
	}
	return nil
}

func (e PreferredContact) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
