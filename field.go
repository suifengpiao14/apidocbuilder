package apidocbuilder

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/cast"
	"github.com/suifengpiao14/funcs"
	"github.com/suifengpiao14/sqlbuilder"
)

func Fields2DocParams(fs ...*sqlbuilder.Field) (params DocParams) {
	params = make(DocParams, 0)
	for _, f := range fs {
		dbSchema := f.Schema
		if dbSchema == nil {
			dbSchema = new(sqlbuilder.Schema)
		}
		param := DocParam{
			Name:        f.GetDocName(),
			Required:    dbSchema.Required,
			AllowEmpty:  dbSchema.AllowEmpty(),
			Title:       dbSchema.Title,
			Type:        "string",
			Format:      dbSchema.Type.String(),
			Default:     cast.ToString(dbSchema.Default),
			Description: dbSchema.Comment,
			Enums:       dbSchema.Enums,
			RegExp:      dbSchema.RegExp,
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
	case reflect.Array, reflect.Slice, reflect.Struct:
		if !structField.Anonymous { // 嵌入结构体,文档名称不增加前缀
			for _, f := range fs {
				docName := f.GetDocName()
				if docName != "" && !strings.HasPrefix(docName, "[]") {
					docName = fmt.Sprintf(".%s", docName)
				}
				fName := fmt.Sprintf("%s%s", funcs.ToLowerCamel(structField.Name), docName)
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
		f.SetDocName(fName)
		fs = append(fs, f)
	}
	return fs
}

func FieldStructToArray(stru any) sqlbuilder.Fields {
	return sqlbuilder.FieldStructToArray(stru, StructFieldCustom, ArrayFieldCustom)
}
