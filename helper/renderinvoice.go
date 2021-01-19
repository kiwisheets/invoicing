package helper

import (
	"fmt"

	"github.com/aymerick/raymond"
	"github.com/kiwisheets/invoicing/model"
	"github.com/sirupsen/logrus"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
)

func RenderInvoice(invoice *model.InvoiceTemplateData) (string, error) {
	if invoice.Company == nil {
		return "", fmt.Errorf("unable to retrieve company")
	}

	if invoice.Client == nil {
		return "", fmt.Errorf("unable to retrieve client")
	}

	tpl, err := raymond.ParseFile("templates/invoice1.handlebars")
	if err != nil {
		logrus.Errorf("error parsing template %s", err)
		return "", err
	}

	result, err := tpl.Exec(invoice)
	if err != nil {
		logrus.Errorf("error rendering template %s", err)
		return "", err
	}

	m := minify.New()
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("text/css", css.Minify)

	s, err := m.String("text/html", result)
	return s, err
}
