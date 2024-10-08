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
		skipNull := false
		switch f.Schema.Type {
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
		format := typ
		if format == "string" {
			format = ""
		}

		param := Parameter{
			Fullname:        f.GetDocName(),
			Required:        dbSchema.Required,
			AllowEmptyValue: dbSchema.AllowEmpty(),
			Title:           dbSchema.Title,
			Type:            "string", // 类型全部转为string
			Format:          format,   //记录真正的类型
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
	for _, f := range fs {
		fName := fmt.Sprintf("[].%s", f.GetDocName())
		fName = strings.TrimSuffix(fName, ".")
		f.SetDocName(fName) //设置列名称,f 本身为指针，直接修改f.Name
	}
	return fs
}

func FieldStructToArray(stru any) sqlbuilder.Fields {
	return sqlbuilder.FieldStructToArray(stru, StructFieldCustom, ArrayFieldCustom)
}

// Struct2Fields 结构体转文档参数名称，再通过名称 匹配fields 集合，即可生成文档参数
func Struct2Fields(stru any, fs sqlbuilder.Fields) (fields sqlbuilder.Fields) { // todo 测试未通过，计划合并FieldStructToArray 实现 Struct2Fields
	stru = getRefVariable(stru)
	InitNilFields(stru) // 初始化所有字段
	val := reflect.Indirect(reflect.ValueOf(stru))
	names := struct2DocName(val)
	fields = make(sqlbuilder.Fields, 0)
	for i := 0; i < len(names); i++ {
		fullname := names[i]
		index := 0
		name := fullname
		for index > -1 {
			field := nameFindFieldDefaultFn(name, fs)
			if field != nil {
				cp := field.Copy()
				cp.SetDocName(fullname)
				fields = append(fields, cp)
				break
			}
			index = strings.Index(name, ".")
			if index > -1 {
				name = name[index+1:]
			}
		}
	}

	return fields
}
func nameFindFieldDefaultFn(name string, fs sqlbuilder.Fields) *sqlbuilder.Field {
	name = strings.ReplaceAll(name, "[]", "")
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.Trim(name, "_")
	camelName := funcs.ToLowerCamel(name)
	for _, f := range fs {
		if f == nil {
			continue
		}
		if f.GetDocName() == camelName {
			return f
		}
	}
	return nil
}

func struct2DocName(val reflect.Value) (names []string) {
	val = reflect.Indirect(val)
	names = make([]string, 0)
	typ := val.Type()
	switch typ.Kind() {
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			subVal := val.Field(i)
			attr := typ.Field(i)
			subNames := struct2DocName(subVal)
			jsonTag := getJsonTag(attr)
			if jsonTag == "" { // 没有json tag 直接返回
				continue
			}
			if len(subNames) == 0 { // 没有子字段 说明当前字段为基础类型，直接添加本身
				names = append(names, jsonTag)
				continue
			}

			for i := 0; i < len(subNames); i++ {
				subNames[i] = fmt.Sprintf("%s.%s", jsonTag, subNames[i])
			}
			names = append(names, subNames...)
		}

	case reflect.Array, reflect.Slice:
		childTyp := typ.Elem()
		if childTyp.Kind() == reflect.Ptr {
			childTyp = childTyp.Elem()
		}
		childVal := reflect.New(childTyp)
		subNames := struct2DocName(childVal)
		for i := 0; i < len(subNames); i++ {
			subNames[i] = fmt.Sprintf("[].%s", subNames[i]) // 增加数组前缀
		}
		names = append(names, subNames...)
	}

	return names
}

func getJsonTag(val reflect.StructField) (jsonTag string) {
	tag := val.Tag.Get("json")
	if tag == "-" {
		tag = ""
	}
	return tag
}
