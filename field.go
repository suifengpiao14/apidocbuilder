package apidocbuilder

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/cast"
	"github.com/suifengpiao14/funcs"
	"github.com/suifengpiao14/sqlbuilder"
)

func Fields2DocParams(fs ...*sqlbuilder.Field) (params Parameters) {
	params = make(Parameters, 0)
	fs2 := sqlbuilder.Fields{}
	for _, f := range fs {
		if f == nil {
			continue
		}
		skipNull := false
		schema := f.Schema
		if schema == nil {
			schema = new(sqlbuilder.Schema)
		}

		switch schema.Type {
		case sqlbuilder.Schema_doc_Type_null, sqlbuilder.Schema_doc_Type_object, sqlbuilder.Schema_doc_Type_array:
			docName := f.GetDocName()
			objectChildrenPrefix := fmt.Sprintf("%s.", docName)
			arrayChildrenPrefix := fmt.Sprintf("%s[]", docName)
			for _, f1 := range fs {
				if strings.HasPrefix(f1.GetDocName(), objectChildrenPrefix) || strings.HasPrefix(f1.GetDocName(), arrayChildrenPrefix) {
					skipNull = true
					break
				}
			}
		}
		if !skipNull {
			fs2 = append(fs2, f)
		}

	}
	for _, f := range fs2 {
		dbSchema := f.Schema
		if dbSchema == nil {
			dbSchema = new(sqlbuilder.Schema)
		}
		enum := make([]string, 0)
		enumNames := make([]string, 0)
		for _, v := range dbSchema.Enums {
			enum = append(enum, cast.ToString(v.Key))
			enumNames = append(enumNames, v.Title)
		}
		typ := dbSchema.Type.String()
		if typ == "" {
			typ = "string"
		}
		param := Parameter{
			Fullname:        f.GetDocName(),
			Required:        dbSchema.Required,
			AllowEmptyValue: dbSchema.AllowEmpty(),
			Title:           dbSchema.Title,
			Type:            typ,
			Default:         cast.ToString(dbSchema.Default),
			Description:     dbSchema.Comment,
			Enum:            strings.Join(enum, ", "),
			EnumNames:       strings.Join(enumNames, ", "),
			RegExp:          dbSchema.RegExp,
		}
		params = append(params, param)
	}
	return params

}

func StructFieldCustom(val reflect.Value, structField reflect.StructField, fs sqlbuilder.Fields) sqlbuilder.Fields {
	for _, f := range fs {
		f.SetFieldName(funcs.ToLowerCamel(structField.Name)) //设置列名称
	}
	switch structField.Type.Kind() {
	case reflect.Array, reflect.Slice, reflect.Struct, reflect.Interface:
		if !structField.Anonymous { // 嵌入结构体,文档名称不增加前缀
			for i := 0; i < len(fs); i++ {
				f := fs[i]
				docName := f.GetDocName()
				if docName != "" && !strings.HasPrefix(docName, "[]") {
					docName = fmt.Sprintf(".%s", docName)
				}
				getJsonTag := getJsonTag(structField)
				fName := fmt.Sprintf("%s%s", getJsonTag, docName)
				fName = strings.TrimSuffix(fName, ".")
				f.SetDocName(fName)
			}
		}
	}
	return fs
}

func ArrayFieldCustom(fs sqlbuilder.Fields) sqlbuilder.Fields {
	if len(fs) == 0 { // 数组为空时，自动生成一个字段(对于 data []string 格式 生成 data[]  string)
		f := &sqlbuilder.Field{}
		fs = append(fs, f)
	}
	for _, f := range fs {
		fName := fmt.Sprintf("[].%s", f.GetDocName())
		fName = strings.TrimSuffix(fName, ".")
		f.SetDocName(fName) //设置列名称,f 本身为指针，直接修改f.Name
	}
	return fs
}

func StructToFields(stru any) sqlbuilder.Fields {
	return sqlbuilder.StructToFields(stru, StructFieldCustom, ArrayFieldCustom)
}

func getJsonTag(val reflect.StructField) (jsonTag string) {
	tag := val.Tag.Get("json")
	if tag == "" {
		tag = val.Name
	}
	if tag == "-" {
		tag = ""
	}
	return tag
}

// Struct2ParametersWithCompletment 根据结构体生成参数，并补全文档信息
func Struct2ParametersWithCompletment(stru any) (parameters Parameters) {
	return Struct2Parameters(stru).Complement(Fields2DocParams(StructToFields(stru)...))
}

func Struct2Parameters(stru any) (parameters Parameters) {
	stru = getRefVariable(stru)
	InitNilFields(stru) // 初始化所有字段
	val := reflect.Indirect(reflect.ValueOf(stru))
	parameters = struct2Parameters(val)
	return parameters
}

func struct2Parameters(val reflect.Value) (parameters Parameters) {
	val = reflect.Indirect(val)
	parameters = make(Parameters, 0)
	if !val.IsValid() {
		return parameters
	}
	typ := val.Type()
	switch typ.Kind() {
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			subVal := val.Field(i)
			attr := typ.Field(i)
			jsonTag := getJsonTag(attr)
			if jsonTag == "" { // 没有json tag 直接返回
				continue
			}
			subParameters := struct2Parameters(subVal)
			if len(subParameters) == 0 { // 没有子字段 说明当前字段为基础类型，直接添加本身
				parameter := Parameter{Fullname: jsonTag, Type: attr.Type.Kind().String()}
				parameters.Add(parameter)
				continue
			}

			for i := 0; i < len(subParameters); i++ {
				subParameters[i].Fullname = fmt.Sprintf("%s.%s", jsonTag, subParameters[i].Fullname)
			}
			parameters.Add(subParameters...)
		}

	case reflect.Array, reflect.Slice:
		childTyp := typ.Elem()
		if childTyp.Kind() == reflect.Ptr {
			childTyp = childTyp.Elem()
		}
		childVal := reflect.New(childTyp)
		subParameters := struct2Parameters(childVal)
		for i := 0; i < len(subParameters); i++ {
			subParameters[i].Fullname = fmt.Sprintf("[].%s", subParameters[i].Fullname) // 增加数组前缀
		}
		parameters.Add(subParameters...)
	case reflect.Interface:
		childVal := val.Elem()
		subParameters := struct2Parameters(childVal)
		parameters.Add(subParameters...)
	}

	// 格式化参数
	parameters.FormatField()
	return parameters
}
