package server

import (
	"bytes"
	"html/template"
)

//nolint:lll
var dataTemplate = `
{{- /* delete empty line */ -}}
package data

import (
	"context"
	"errors"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"

	//  TODO: modify project name
	// biz "{{ .Project }}/internal/biz"
)

type {{ .Service }}RepoImpl struct {
	data *Data
	log *log.Helper
}

func New{{ .Service }}RepoImpl(data *Data, logger log.Logger) (*{{ .Service }}RepoImpl, error) {

	return &{{ .Service }}RepoImpl{
		data: data,
		log:	  log.NewHelper(logger),
	}, nil
}

{{ range .Methods }}
// {{ .Name }} is
func (d *{{ .Service }}RepoImpl) {{ .Name }}(ctx context.Context) error {
	// TODO database operations

	return nil
}

{{- end }}
`

func (s *Service) executeData() ([]byte, error) {

	buf := new(bytes.Buffer)
	tmpl, err := template.New("data").Parse(dataTemplate)
	if err != nil {
		return nil, err
	}
	if err := tmpl.Execute(buf, s); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
