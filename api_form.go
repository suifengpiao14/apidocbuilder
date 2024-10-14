package apidocbuilder

import (
	"encoding/json"
	"strings"
)

type ApiForm struct {
	api    Api
	Action string `json:"action"`
	Method string `json:"method"`
	Title  string `json:"title"`
}

type ApiFormItem struct {
	Label          string         `json:"label"`
	FormItem       string         `json:"formItem"`
	Name           string         `json:"name"`
	InputType      string         `json:"inputType"`
	Default        string         `json:"default"`
	Required       bool           `json:"required"`
	SelectMultiple bool           `json:"selectMultiple"`
	SelectOptions  []SelectOption `json:"SelectOptions"`
}

type SelectOption struct {
	Label string `json:"label"`
	Value any    `json:"value"`
}

type ApiFormItems []ApiFormItem

func (fs ApiFormItems) GetState() (state string) {
	m := map[string]any{}
	for _, f := range fs {
		var value any
		value = f.Default
		if f.SelectMultiple {
			value = []any{f.Default}
		}
		m[f.Name] = value

	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		panic(err)
	}
	state = string(b)
	return state
}

func isArrayName(name string) (realName string, isArray bool) {
	realName = name
	isArray = strings.HasSuffix(name, "[]")
	if isArray {
		realName = name[:len(name)-2]
	}
	return realName, isArray

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
		if p.Name == "" {
			continue
		}
		formItem := "input" //根据需要选择合适表达元素
		inputType := "text" // 根据长度类型，选择合适input 类型
		realName, isArray := isArrayName(p.Name)
		schema := p.Schema
		if schema == nil {
			schema = &Schema{}
		}
		selectOptions := make([]SelectOption, 0)
		if schema.Enum != "" {
			formItem = "select"
			enums := strings.Split(schema.Enum, ",")
			enumsName := strings.Split(schema.EnumNames, ",")
			nameLen := len(enumsName)
			for i, v := range enums {
				name := v
				if i < nameLen {
					name = enumsName[i]
				}
				selectOptions = append(selectOptions, SelectOption{
					Label: name,
					Value: v,
				})

			}
		}

		switch schema.Type {
		case "number", "integer":
			inputType = "number"
		}
		field := ApiFormItem{
			Label:          p.TitleOrDescription(),
			FormItem:       formItem,
			Name:           realName,
			InputType:      inputType,
			Default:        p.Default,
			Required:       p.Required,
			SelectMultiple: isArray,
			SelectOptions:  selectOptions,
		}

		fields = append(fields, field)

	}

	return
}

func (f ApiForm) State() (state string) {
	return f.Items().GetState()
}
