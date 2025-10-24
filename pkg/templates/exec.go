package templates

import (
	"bytes"
	_ "embed"
	"html/template"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"

	"github.com/SOTBI-LLC/sotbi.lib/pkg/times"
)

//go:embed _header.html
var header string

//go:embed _footer.html
var footer string

// TemplateData struct.
type TemplateData struct {
	Data any
	IP   string
}

func ExecTemplate(data any, templ, ip string) (res string, err error) {
	emailData := TemplateData{data, ip}
	funcs := template.FuncMap{
		"getRequestFiles": GetRequestFiles,
		"formatDate": func(date time.Time) string {
			return date.In(times.GetMoscowLocation()).Format("02.01.2006 15:04")
		},
	}

	var buf bytes.Buffer

	t, err := template.New("title").Funcs(funcs).Parse(templ)
	if err != nil {
		return
	}

	if err = t.Execute(&buf, emailData); err != nil {
		return
	}

	buffer := new(bytes.Buffer)
	md := goldmark.New(goldmark.WithExtensions(extension.GFM))

	if err := md.Convert(buf.Bytes(), buffer); err != nil {
		return "", err
	}

	res = header + buffer.String() + footer

	return res, nil
}

func GetRequestFiles(sType uint) string {
	var b strings.Builder

	first := true

	if sType&1 != 0 {
		b.WriteString("TXT")

		first = false
	}

	if sType&2 != 0 {
		if !first {
			b.WriteString(" + ")
		}

		b.WriteString("PDF")

		first = false
	}

	if sType&4 != 0 {
		if !first {
			b.WriteString(" + ")
		}

		b.WriteString("Excel")
	}

	return b.String()
}
