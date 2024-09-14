package apidocbuilder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"github.com/suifengpiao14/sqlbuilder"
)

var ApiDocument_markdown_template = ``

func ApiDocumentMarkdonwnTplDefault() (tpl *template.Template, err error) {
	tpl, err = template.New("").Parse(ApiDocument_markdown_template)
	if err != nil {
		return nil, errors.WithMessage(err, "ApiDocumentMarkdonwnTplDefault")
	}
	return tpl, nil
}

type ApiDocument struct {
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	Path             string    `json:"path"`
	Method           string    `json:"method"`
	ContentType      string    `json:"contentType"`
	InputParameters  DocParams `json:"inputParameters"`
	OutputParameters DocParams `json:"outputParameters"`
}

func (doc ApiDocument) Markdown(tpl *template.Template) (out string, err error) {
	var w bytes.Buffer
	err = tpl.Execute(&w, doc)
	if err != nil {
		return "", err
	}
	out = w.String()
	return out, nil
}

type DocParam struct {
	Name        string           `json:"name"`
	Required    bool             `json:"required,string"`
	AllowEmpty  bool             `json:"allowEmpty,string"`
	Title       string           `json:"title"`
	Type        string           `json:"type"`
	Format      string           `json:"format"`
	Default     string           `json:"default"`
	Description string           `json:"description"`
	Example     string           `json:"example"`
	Enums       sqlbuilder.Enums `json:"enums"`
	RegExp      string           `json:"regExp"`
}

type DocParams []DocParam

func (args DocParams) Makedown() string {
	var w bytes.Buffer
	w.WriteString(`|名称|类型|必填|格式|标题|默认值|描述|案例|`)
	w.WriteString("\n")
	w.WriteString(`|:--|:--|:--|:--|:--|:--|:--|:--|:--|:--|`)
	w.WriteString("\n")
	for _, arg := range args {
		description := arg.Description
		if len(arg.Enums) > 0 {
			description = fmt.Sprintf("%s(%s)", description, arg.Enums.String())
		}
		if arg.RegExp != "" {
			description = fmt.Sprintf("%s(匹配规则:%s)", description, arg.RegExp)
		}
		row := fmt.Sprintf(`|%s|%s|%s|%s|%s|%s|%s|%s|`,
			arg.Name,
			arg.Type,
			cast.ToString(arg.Required),
			arg.Format,
			arg.Title,
			arg.Default,
			description,
			arg.Example,
		)
		w.WriteString(row)
		w.WriteString("\n")
	}
	return w.String()
}

func (args DocParams) JsonExample(pretty bool) string {
	m := map[string]any{}
	for _, arg := range args {
		m[arg.Name] = arg.Example
		if m[arg.Name] == "" {
			m[arg.Name] = arg.Default
		}
	}
	var w bytes.Buffer
	marshal := json.NewEncoder(&w)
	marshal.SetIndent("", " ")
	marshal.Encode(m)
	return w.String()

}
