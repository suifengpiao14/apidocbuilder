package apidocbuilder

import "encoding/json"

type ApiForm struct {
	api    Api    `json:"api"`
	Action string `json:"action"`
	Method string `json:"method"`
	Title  string `json:"title"`
}

type ApiFormItem struct {
	Label     string `json:"label"`
	FormItem  string `json:"formItem"`
	Name      string `json:"name"`
	InputType string `json:"inputType"`
	Default   string `json:"default"`
	Required  bool   `json:"required"`
}

type ApiFormItems []ApiFormItem

func (fs ApiFormItems) GetState() (state string) {
	m := map[string]any{}
	for _, f := range fs {
		m[f.Name] = f.Default
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		panic(err)
	}
	state = string(b)
	return state
}

func NewApiForm(api Api) ApiForm {
	return ApiForm{
		api:    api,
		Title:  api.TitleOrDescription(),
		Action: api.Path,
		Method: api.Method,
	}
}

func (f ApiForm) Items() (fields ApiFormItems) {
	for _, p := range f.api.RequestBody {
		formItem := "input" //根据需要选择合适表达元素
		inputType := "text" // 根据长度类型，选择合适input 类型

		fields = append(fields, ApiFormItem{
			Label:     p.TitleOrDescription(),
			FormItem:  formItem,
			Name:      p.Name,
			InputType: inputType,
			Default:   p.Default,
			Required:  p.Required,
		})
	}

	return
}

func (f ApiForm) State() (state string) {
	return f.Items().GetState()
}
