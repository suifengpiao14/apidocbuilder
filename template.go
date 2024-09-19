package apidocbuilder

import (
	"bytes"
	"embed"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

//go:embed  template
var TemplateFS embed.FS

func ExecTpl(tplName string, data any) (content []byte, err error) {
	tpl, err := template.New("").Funcs(sprig.FuncMap()).ParseFS(TemplateFS, "template/*.tpl")
	if err != nil {
		return nil, err
	}
	var w bytes.Buffer
	err = tpl.ExecuteTemplate(&w, tplName, data)
	if err != nil {
		return nil, err
	}
	content = w.Bytes()
	return content, nil
}

const (
	TPL_NAME_MARKDOWN_DOC = "markdownDoc"
)

func Markdown(api API) (out []byte, err error) {
	out, err = ExecTpl(TPL_NAME_MARKDOWN_DOC, api)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func Markdown2HTML(md []byte) []byte {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}
