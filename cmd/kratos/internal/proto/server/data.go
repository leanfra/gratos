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
    "gorm.io/gorm"

	//  TODO: modify project name
	// biz "{{ .Project }}/internal/biz"
)

type {{ .Service }}RepoImpl struct {
	db *gorm.DB
}

func New{{ .Service }}RepoImpl(dsn string) (*{{ .Service }}RepoImpl, error) {

	var _db *gorm.DB
	var err error
	if len(dsn) > 0 {
        if dsn[:5] == "mysql" {
            // 连接 MySQL 数据库
            _db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
        } else if dsn[:6] == "sqlite" {
            // 连接 SQLite 数据库
            _db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
        } else {
            return nil, errors.New("Unsupported database DSN format")
        }
    } else {
        return nil, errors.New("DSN is not provided")
    }

	if err != nil {
		return nil, err
	}
	return &{{ .Service }}RepoImpl{
		db: _db,
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
