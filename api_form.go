package apidocbuilder

import (
	"fmt"
	"strings"

	"github.com/julvo/htmlgo"
	"github.com/julvo/htmlgo/attributes"
	"github.com/suifengpiao14/funcs"
)

func NewHtmxForm(api Api) HtmxForm {
	return HtmxForm{
		ApiForm: ApiForm{
			api:    api,
			Title:  api.TitleOrDescription(),
			Action: api.Path,
			Method: api.Method,
		},
		HxTarget: "#response-data",
		HxExt:    "jsonpretty",
	}
}

type HtmxForm struct {
	ApiForm
	HxExt    string `json:"hx-ext"`
	HxTarget string `json:"hx-target"`
}

type ApiForm struct {
	api    Api
	Action string `json:"action"`
	Method string `json:"method"`
	Title  string `json:"title"`
}

func (htmxForm HtmxForm) String() (html string) {
	return string(htmxForm.Html())
}

func (htmxForm HtmxForm) Html() (html htmlgo.HTML) {
	attrs := make([]attributes.Attribute, 0)
	attrs = append(attrs, AttrHxTarget(htmxForm.HxTarget))
	attrs = append(attrs, AttrHxExt(htmxForm.HxExt))
	attrs = append(attrs, AttrHxPost(htmxForm.Action))
	// attrs = append(attrs, hxExtAttr)
	// attrs = append(attrs, hxPostAttr)
	attrs = append(attrs, attributes.Method(strings.ToUpper(htmxForm.Method)))
	htmls := make([]htmlgo.HTML, 0)
	for _, p := range htmxForm.api.RequestBody {
		html := Parameter2FormChidren(p)
		htmls = append(htmls, html)
	}
	if len(htmls) == 0 {
		div := htmlgo.Div_(htmlgo.Text("无需入参数"))
		htmls = append(htmls, div)
	}
	submit := TagButton{
		Type:    "submit",
		Text:    "请求",
		WrapDiv: true,
	}
	htmls = append(htmls, submit.Html())
	form := htmlgo.Form(attrs, htmls...)
	return form
}

func AttrHxTarget(data interface{}, templs ...string) attributes.Attribute {
	return Attr("hx-target", data, templs...)
}
func AttrHxPost(data interface{}, templs ...string) attributes.Attribute {
	return Attr("hx-post", data, templs...)
}
func AttrHxExt(data interface{}, templs ...string) attributes.Attribute {
	return Attr("hx-ext", data, templs...)
}

func Attr(name string, data interface{}, templs ...string) attributes.Attribute {
	tplName := funcs.ToCamel(name)
	attr := attributes.Attribute{Data: data, Name: tplName}
	value := "{{.}}"
	if len(templs) > 0 {
		value = strings.Join(templs, " ")
	}
	attr.Templ = fmt.Sprintf(`{{define "%s"}}%s="%s"{{end}}`, tplName, name, value)
	return attr
}

type TagInput struct {
	Label       TagLabel
	Name        string `json:"name"`
	Type        string `json:"type"`
	Value       string `json:"value"`
	Required    bool   `json:"required"`
	Placeholder string `json:"placeholder"`
}

type TagLabel struct {
	Label string `json:"label"`
}

func (t TagLabel) Html() (html htmlgo.HTML) {
	html = htmlgo.Label_(htmlgo.Text(t.Label))
	return html
}

func (tag TagInput) Html() (html htmlgo.HTML) {
	inputAttrs := make([]attributes.Attribute, 0)
	inputAttrs = append(inputAttrs, attributes.Type_(tag.Type))
	inputAttrs = append(inputAttrs, attributes.Name(tag.Name))
	inputAttrs = append(inputAttrs, attributes.Value(tag.Value))
	inputAttrs = append(inputAttrs, attributes.Placeholder_(tag.Placeholder))
	if tag.Required {
		inputAttrs = append(inputAttrs, attributes.Required(true))
	}
	tagInput := htmlgo.Input(inputAttrs)
	div := htmlgo.Div_(tag.Label.Html(), tagInput)
	return div
}

type TagSelect struct {
	Label    TagLabel
	Name     string        `json:"name"`
	Options  SelectOptions `json:"options"`
	Required bool          `json:"required"`
}

func (tag TagSelect) Html() (html htmlgo.HTML) {
	selectAttrs := make([]attributes.Attribute, 0)
	selectAttrs = append(selectAttrs, attributes.Name(tag.Name))
	tagSelect := htmlgo.Select(selectAttrs, tag.Options.Html()...)
	div := htmlgo.Div_(tag.Label.Html(), tagSelect)
	return div
}

type SelectOptions []SelectOption

func (opts SelectOptions) Html() (options []htmlgo.HTML) {
	options = make([]htmlgo.HTML, 0)
	for _, o := range opts {
		options = append(options, o.Html())
	}
	return options
}

type SelectOption struct {
	Label   string `json:"label"`
	Value   any    `json:"value"`
	Checked bool   `json:"checked"`
}

func (o SelectOption) Html() (html htmlgo.HTML) {
	attrs := make([]attributes.Attribute, 0)
	attrs = append(attrs, attributes.Value(o.Value))
	if o.Checked {
		attrs = append(attrs, attributes.Checked(true))

	}
	option := htmlgo.Option(attrs, htmlgo.Text(o.Label))
	return option
}

type TagButton struct {
	Type    string `json:"type"`
	Text    string `json:"label"`
	WrapDiv bool   `json:"wrapDiv"`
}

func (tag TagButton) Html() (html htmlgo.HTML) {
	attrs := make([]attributes.Attribute, 0)
	attrs = append(attrs, attributes.Type_(tag.Type))
	button := htmlgo.Button(attrs, htmlgo.Text(tag.Text))
	html = button
	if tag.WrapDiv {
		html = htmlgo.Div_(button)
	}
	return html
}

func Parameter2FormChidren(p Parameter) (html htmlgo.HTML) {
	if p.Name == "" {
		return
	}
	if p.Enum != "" {
		return Parameter2TagSelect(p).Html()
	}
	html = Parameter2TagInput(p).Html()
	return html
}

func Parameter2TagInput(p Parameter) (tag TagInput) {
	if p.Name == "" {
		return
	}
	realName, _ := isArrayName(p.Name)
	schema := p.Schema
	if schema == nil {
		schema = &Schema{}
	}
	tagInput := TagInput{
		Label:       TagLabel{Label: p.TitleOrDescription()},
		Type:        "text",
		Name:        realName,
		Value:       p.Default,
		Required:    p.Required,
		Placeholder: p.TitleOrDescription(),
	}
	switch schema.Type {
	case "number", "integer", "int":
		tagInput.Type = "number"
	}

	return tagInput
}

func Parameter2TagSelect(p Parameter) (tag TagSelect) {
	if p.Name == "" {
		return
	}
	realName, _ := isArrayName(p.Name)
	schema := p.Schema
	if schema == nil {
		schema = &Schema{}
	}
	tag = TagSelect{Name: realName}
	if p.Enum != "" {
		selectOptions := make([]SelectOption, 0)
		enums := strings.Split(p.Enum, ",")
		enumsName := strings.Split(p.EnumNames, ",")
		nameLen := len(enumsName)
		for i, v := range enums {
			name := v
			if i < nameLen {
				name = enumsName[i]
			}
			checked := false
			if v == p.Default {
				checked = true
			}

			selectOptions = append(selectOptions, SelectOption{
				Label:   name,
				Value:   v,
				Checked: checked,
			})
		}
		tag = TagSelect{
			Label:    TagLabel{Label: p.TitleOrDescription()},
			Name:     realName,
			Required: p.Required,
			Options:  selectOptions,
		}
	}
	return tag

}

func isArrayName(name string) (realName string, isArray bool) {
	realName = name
	isArray = strings.HasSuffix(name, "[]")
	if isArray {
		realName = name[:len(name)-2]
	}
	return realName, isArray

}
