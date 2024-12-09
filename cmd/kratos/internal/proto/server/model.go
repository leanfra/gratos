package server

import (
	"bytes"
	"html/template"
)

type Field struct {
	Name     string
	Type     string
	Repeated bool
}

type Model struct {
	Name   string
	Fields []Field
}

type Models struct {
	Models []Model
}

var modelTemplate = `
{{- /* delete empty line */ -}}
package biz

{{ range $m := .Models}}

type {{ $m.Name}} struct {
	{{- range $f := $m.Fields }}
	{{ $f.Name }} {{ if $f.Repeated }} []{{ end }}{{ $f.Type }}
	{{- end }} 
}

{{- end }}

`

func (m *Models) execute() ([]byte, error) {

	buf := new(bytes.Buffer)
	tmpl, err := template.New("models").Parse(modelTemplate)
	if err != nil {
		return nil, err
	}
	if err := tmpl.Execute(buf, m); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
