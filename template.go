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

func newTplInstance() *template.Template {
	return template.New("").Funcs(sprig.FuncMap())
}

func RenderHtml(tplInstance *template.Template, filename string, data any) (content []byte, err error) {
	fullname := fmt.Sprintf("render/%s", filename)
	fs, err := HtmlTemplateFS.Open(fullname)
	if err != nil {
		return nil, err
	}
	tplContent, err := io.ReadAll(fs)
	if err != nil {
		return nil, err
	}

	tpl, err := tplInstance.Parse(string(tplContent))
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
	TPL_NAME_MARKDOWN_DOC     = "markdownDoc"
	TPL_NAME_MARKDOWN_SERVICE = "markdownService"
	TPL_NAME_HTML_DEBUGGING   = "debugging"
)

func Api2Markdown(api Api) (out []byte, err error) {
	out, err = ExecTpl(TPL_NAME_MARKDOWN_DOC, api)
	if err != nil {
		return nil, err
	}
	return out, nil
}
func Service2Markdown(service Service) (out []byte, err error) {
	out, err = ExecTpl(TPL_NAME_MARKDOWN_SERVICE, service)
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
	out, err = RenderHtml(newTplInstance(), filename, buf.String())
	if err != nil {
		return nil, err
	}

	return out, nil
}

type ServiceRender struct {
	Service
	activeApi *Api `json:"-"`
}

func (s *ServiceRender) SetActiveApi(api Api) { // 渲染html时使用
	api.DocumentRef = s.DocumentRef
	api.Service = &s.Service
	s.activeApi = &api
}
func (s ServiceRender) GetActiveApi() *Api { // 渲染html时使用
	if s.activeApi == nil && len(s.Apis) > 0 {
		return &s.Apis[0]
	}
	return s.activeApi
}
func (s ServiceRender) IsActiveApi(api Api) bool { // 判断是否是当前api
	return s.activeApi.IsSameMethodAndPath(api.Method, api.Path)
}

func (s ServiceRender) ActiveClass(api Api, activeCalss string) (out string) { // ActiveClass判断是否是当前api 是则返回 activeCalss，否则返回空字符串
	if s.IsActiveApi(api) {
		return activeCalss
	}
	return ""
}

func (s ServiceRender) GetCurrentApiContent() (out string, err error) {
	currentApi := s.GetActiveApi()
	if currentApi == nil {
		return "", nil
	}
	b, err := Api2Markdown(*currentApi)
	if err != nil {
		return "", err
	}
	b, err = Markdown2HTML(b)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func RenderService(serviceRender ServiceRender, currentApiName string) (out []byte, err error) {
	if currentApiName != "" {
		api, err := serviceRender.GetApiByName(currentApiName)
		if err != nil {
			return nil, err
		}
		serviceRender.SetActiveApi(*api)
	}

	filename := "html_service.html"
	out, err = RenderHtml(newTplInstance(), filename, serviceRender)
	if err != nil {
		return nil, err
	}

	return out, nil
}
func RenderForm(serviceRender ServiceRender, currentApiName string) (out []byte, err error) {
	if currentApiName != "" {
		api, err := serviceRender.GetApiByName(currentApiName)
		if err != nil {
			return nil, err
		}
		serviceRender.SetActiveApi(*api)
	}

	filename := "html_form.html"
	apiForm := NewApiForm(*serviceRender.activeApi)
	out, err = RenderHtml(newTplInstance().Delims("[[", "]]"), filename, apiForm)
	if err != nil {
		return nil, err
	}

	return out, nil
}
