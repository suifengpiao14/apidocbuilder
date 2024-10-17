package apidocbuilder

import (
	"fmt"
	"sort"
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

	submitInput := TagInput{Type: "submit", Value: "请求"}
	htmls = append(htmls, submitInput.Html())
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

type TagTextArea struct {
	Label       TagLabel
	Name        string `json:"name"`
	Value       string `json:"value"`
	Required    bool   `json:"required"`
	Placeholder string `json:"placeholder"`
	Cols        int    `json:"column"`
	Rows        int    `json:"rows"`
}

func (tag TagTextArea) Html() (html htmlgo.HTML) {
	tag.Label.Required = tag.Required
	inputAttrs := htmlgo.Attr(
		attributes.Name(tag.Name),
		attributes.Value(tag.Value),
		attributes.Placeholder_(tag.Placeholder),
		attributes.Rows(tag.Rows),
		attributes.Cols(tag.Cols),
	)

	if tag.Required {
		inputAttrs = append(inputAttrs, attributes.Required_())
	}
	tagInput := htmlgo.Textarea(inputAttrs)
	div := htmlgo.Div_(tag.Label.Html(), tagInput)
	return div
}

type InputTypeRef struct {
	Type     string `json:"type"`
	Format   Format `json:"format"`
	SortDesc int    `json:"sortDesc"`
}

type InputTypeRefs []InputTypeRef

func (a InputTypeRefs) Len() int           { return len(a) }
func (a InputTypeRefs) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a InputTypeRefs) Less(i, j int) bool { return a[i].SortDesc > a[j].SortDesc }

var (
	// 确保遍历顺序
	InputTypeRefDefault = InputTypeRefs{
		{Type: "checkbox", Format: Format{"checkbox"}, SortDesc: 99},
		{Type: "color", Format: Format{"color"}, SortDesc: 98},
		{Type: "date", Format: Format{"date"}, SortDesc: 97},
		{Type: "datetime-local", Format: Format{"datetime"}, SortDesc: 96},
		{Type: "time", Format: Format{"time"}, SortDesc: 95},
		{Type: "week", Format: Format{"week"}, SortDesc: 94},
		{Type: "month", Format: Format{"month"}, SortDesc: 93},
		{Type: "email", Format: Format{"email"}, SortDesc: 92},
		{Type: "file", Format: Format{"file"}, SortDesc: 91},
		{Type: "number", Format: Format{"int", "integer", "float", "range"}, SortDesc: 90},
		{Type: "password", Format: Format{"password"}, SortDesc: 89},
		{Type: "range", Format: Format{"range"}, SortDesc: 88},
		{Type: "search", Format: Format{"search"}, SortDesc: 87},
		{Type: "tel", Format: Format{"tel", "phone"}, SortDesc: 86},
		{Type: "url", Format: Format{"url"}, SortDesc: 85},
		{Type: "text", Format: Format{"string"}, SortDesc: 0},
	}
)

type TagInput struct {
	Label       TagLabel
	Name        string `json:"name"`
	Type        string `json:"type"`
	Value       string `json:"value"`
	Required    bool   `json:"required"`
	Placeholder string `json:"placeholder"`
	Min         int    `json:"min"`
	Max         int    `json:"max"`
}

func (tag TagInput) Format2Type(formats ...string) string {
	sort.Sort(InputTypeRefDefault)
	for _, v := range InputTypeRefDefault {
		if v.Format.Has(formats...) {
			return v.Type
		}
	}
	return "text"
}

type Class []string

func (c *Class) Add(class string) {
	if *c == nil {
		*c = make([]string, 0)
	}
	*c = append(*c, class)
}

func (c *Class) Remove(classes ...string) {
	m := make(map[string]bool)
	for _, v := range classes {
		m[v] = true
	}

	tmpCls := make([]string, 0)
	for _, v := range *c {
		if !m[v] {
			tmpCls = append(tmpCls, v)
		}
	}
	*c = tmpCls
}

func (c *Class) Clearn() {
	*c = make(Class, 0)
}

func (c Class) Attr() attributes.Attribute {
	return attributes.Class_(c...)
}

type TagLabel struct {
	Label string `json:"label"`
	TagRequired
	Class       Class `json:"class"`
	RemoveClass Class `json:"removeClass"`
}

const (
	class_label = "label"
)

func (t TagLabel) Html() (html htmlgo.HTML) {
	t.Class.Add(class_label)
	t.Class.Remove(t.RemoveClass...)
	attrs := make([]attributes.Attribute, 0)
	attrs = append(attrs, t.Class.Attr())
	html = htmlgo.Label(attrs, htmlgo.Text(t.Label), t.TagRequired.Html())
	return html
}

type TagRequired struct {
	Required bool `json:"required"`
}

func (t TagRequired) Html() (html htmlgo.HTML) {
	if t.Required {
		attrs := htmlgo.Attr(
			attributes.Class_("required"),
		)
		html = htmlgo.Span(attrs, htmlgo.Text("*"))
	}
	return html
}

func (tag TagInput) Html() (html htmlgo.HTML) {
	tag.Label.Required = tag.Required
	inputAttrs := make([]attributes.Attribute, 0)
	inputAttrs = append(inputAttrs, attributes.Type_(tag.Type))
	inputAttrs = append(inputAttrs, attributes.Name(tag.Name))
	inputAttrs = append(inputAttrs, attributes.Value(tag.Value))
	inputAttrs = append(inputAttrs, attributes.Placeholder_(tag.Placeholder))
	if tag.Min > 0 {
		inputAttrs = append(inputAttrs, attributes.Min(tag.Min))
	}
	if tag.Max > 0 {
		inputAttrs = append(inputAttrs, attributes.Max(tag.Max))
	}

	if tag.Required {
		inputAttrs = append(inputAttrs, attributes.Required_())
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
	tag.Label.Required = tag.Required
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
		attrs = append(attrs, attributes.Checked_())

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
		if strings.Count(p.Enum, ",") < 3 { // 3个枚举值以内，使用单选框
			return Parameter2Radios(p).Html()
		}
		return Parameter2TagSelect(p).Html()
	}
	schema := p.Schema
	format := p.GetFormat()
	if p.Type == "string" && (format.IsNil() || format.Has("string")) && (schema.MaxLength == 0 || schema.MaxLength >= Schema_MaxLength_textArea) { // 长度不限制，或者过长，使用textarea
		return Parameter2TextArea(p).Html()
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

	tagInput := TagInput{
		Label:       TagLabel{Label: p.TitleOrDescription()},
		Name:        realName,
		Value:       p.Default,
		Required:    p.Required,
		Placeholder: p.TitleOrDescription(),
		Min:         schema.Minimum,
		Max:         schema.Maximum,
	}
	format := p.GetFormat()
	format.Add(tag.Type)
	tagInput.Type = tagInput.Format2Type(format...)

	return tagInput
}

const (
	Schema_textArea_cols = 50 // 50个字符一行
)

func Parameter2TextArea(p Parameter) (tag TagTextArea) {
	if p.Name == "" {
		return
	}
	realName, _ := isArrayName(p.Name)
	schema := p.Schema
	rows := schema.MaxLength / Schema_textArea_cols
	tagInput := TagTextArea{
		Label:       TagLabel{Label: p.TitleOrDescription()},
		Name:        realName,
		Value:       p.Default,
		Required:    p.Required,
		Placeholder: p.TitleOrDescription(),
		Cols:        Schema_textArea_cols,
		Rows:        rows,
	}
	return tagInput
}

type TagRadio struct {
	Label    TagLabel
	Name     string `json:"name"`
	Value    any    `json:"value"`
	Required bool   `json:"required"`
	Checked  bool   `json:"checked"`
}

func (tag TagRadio) Html() (html htmlgo.HTML) {
	//type="radio" name="gender" value="male" checked
	attrs := htmlgo.Attr(
		attributes.Type_("radio"),
		attributes.Name(tag.Name),
		attributes.Value(tag.Value),
	)
	if tag.Required {
		attrs = append(attrs, attributes.Required_())
	}
	if tag.Checked {
		attrs = append(attrs, attributes.Checked_())
	}
	input := htmlgo.Input(attrs)
	label := htmlgo.Label_(input)
	return label
}

type TagRadios struct {
	Label    TagLabel
	Required bool `json:"required"`
	Radios   []TagRadio
}

func (tag TagRadios) Html() (html htmlgo.HTML) {
	tag.Label.Required = tag.Required
	children := make([]htmlgo.HTML, 0)
	children = append(children, tag.Label.Html())
	for i, v := range tag.Radios {
		if tag.Required && i == 0 { // 给第一个标记改组必填
			v.Required = tag.Required
		}
		children = append(children, v.Html())
	}
	html = htmlgo.Div_(children...)
	return html
}

func Parameter2Radios(p Parameter) (tag TagRadios) {
	if p.Name == "" {
		return
	}
	realName, _ := isArrayName(p.Name)

	tag = TagRadios{
		Label:    TagLabel{Label: p.TitleOrDescription()},
		Required: p.Required,
		Radios:   make([]TagRadio, 0),
	}
	if p.Enum != "" {
		enums := strings.Split(p.Enum, ",")
		enumsName := strings.Split(p.EnumNames, ",")
		nameLen := len(enumsName)
		for i, v := range enums {
			label := v
			if i < nameLen {
				label = enumsName[i]
			}
			checked := false
			if v == p.Default {
				checked = true
			}

			radio := TagRadio{
				Label:    TagLabel{Label: label},
				Name:     realName,
				Value:    v,
				Required: p.Required,
				Checked:  checked,
			}
			tag.Radios = append(tag.Radios, radio)
		}

	}
	return tag
}

func Parameter2TagSelect(p Parameter) (tag TagSelect) {
	if p.Name == "" {
		return
	}
	realName, _ := isArrayName(p.Name)
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
