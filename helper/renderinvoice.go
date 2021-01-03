package helper

import (
	"github.com/aymerick/raymond"
	"github.com/kiwisheets/invoicing/model"
	"github.com/sirupsen/logrus"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
)

func RenderInvoice(invoice *model.InvoiceTemplateData) (string, error) {
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
