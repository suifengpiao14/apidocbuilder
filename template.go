package apidocbuilder

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
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

//go:embed  render
var HtmlTemplateFS embed.FS

func RenderHtml(filename string, data any) (content []byte, err error) {
	fullname := fmt.Sprintf("render/%s", filename)
	fs, err := HtmlTemplateFS.Open(fullname)
	if err != nil {
		return nil, err
	}
	tplContent, err := io.ReadAll(fs)
	if err != nil {
		return nil, err
	}

	tpl, err := template.New("").Funcs(sprig.FuncMap()).Parse(string(tplContent))
	if err != nil {
		return nil, err
	}
	var w bytes.Buffer
	err = tpl.Execute(&w, data)
	if err != nil {
		return nil, err
	}
	content = w.Bytes()
	return content, nil
}

const (
	TPL_NAME_MARKDOWN_DOC  = "markdownDoc"
	TPL_NAME_MARKDOWN_LIST = "markdownList"
	TPL_NAME_HTML_DOC      = "htmlDoc"
)

func Markdown(api API) (out []byte, err error) {
	out, err = ExecTpl(TPL_NAME_MARKDOWN_DOC, api)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func Markdown2HTML(markdownContent []byte) (out []byte, err error) {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)
	var buf bytes.Buffer
	if err := md.Convert(markdownContent, &buf); err != nil {
		return nil, err
	}
	filename := "html_doc.html"
	out, err = RenderHtml(filename, buf.String())
	if err != nil {
		return nil, err
	}

	return out, nil
}
