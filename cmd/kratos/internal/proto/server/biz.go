package server

import (
	"bytes"
	"html/template"
)

//nolint:lll
var bizTemplate = `
{{- /* delete empty line */ -}}
package biz

import (
	{{- if .UseContext }}
	"context"
	{{- end }}
	{{- if .UseIO }}
	"io"
	{{- end }}
)

type {{ .Service }}Repo interface {
{{- range .Methods }}
    {{ .Name}}(ctx context.Context) error
{{- end }}
}

type {{ .Service }}Usecase struct {
	repo {{ .Service }}Repo	
	log *log.Helper
}

func New{{ .Service }}Usecase(repo {{ .Service }}Repo, logger log.Logger) *{{ .Service }}Usecase {
	return &{{ .Service }}Usecase{
		repo: repo,
		log:	  log.NewHelper(logger),
	}
}

{{ range .Methods }}

// {{ .Name }} is 
func (s *{{ .Service }}Usecase) {{ .Name }}(ctx context.Context)  error {
	return s.repo.{{ .Name }}(ctx)
}

{{- end }}
`

func (s *Service) executeBiz() ([]byte, error) {
	const empty = "google.protobuf.Empty"
	buf := new(bytes.Buffer)
	for _, method := range s.Methods {
		if (method.Type == unaryType && (method.Request == empty || method.Reply == empty)) ||
			(method.Type == returnsStreamsType && method.Request == empty) {
			s.GoogleEmpty = true
		}
		if method.Type == twoWayStreamsType || method.Type == requestStreamsType {
			s.UseIO = true
		}
		if method.Type == unaryType {
			s.UseContext = true
		}
	}
	tmpl, err := template.New("biz").Parse(bizTemplate)
	if err != nil {
		return nil, err
	}
	if err := tmpl.Execute(buf, s); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
