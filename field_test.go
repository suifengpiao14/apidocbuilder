package apidocbuilder_test

import (
	"fmt"
	"testing"

	"github.com/suifengpiao14/apidocbuilder"
	"github.com/suifengpiao14/sqlbuilder"
)

type CUTimeFields struct {
	CreateTime sqlbuilder.FieldFn[string]
}

func (f CUTimeFields) Builder() CUTimeFields {
	return CUTimeFields{
		CreateTime: func(value string) *sqlbuilder.Field {
			return sqlbuilder.NewField("").SetName("createAt")
		},
	}
}

type ProfileFields struct {
	Id       sqlbuilder.FieldFn[int]
	Nickname sqlbuilder.FieldFn[string]
	Gender   sqlbuilder.FieldFn[string]
	Email    sqlbuilder.FieldFn[string]

	CUTimeFields
	Times []CUTimeFields
}

func (ProfileFields) Builder() ProfileFields {
	Times := make([]CUTimeFields, 0)
	Times = append(Times, new(CUTimeFields).Builder())
	pf := ProfileFields{
		Id: func(value int) *sqlbuilder.Field {
			return sqlbuilder.NewField(value).SetName("id")
		},
		Nickname: func(value string) *sqlbuilder.Field {
			return sqlbuilder.NewField(value).SetName("nickname")
		},
		Gender: func(value string) *sqlbuilder.Field {
			return sqlbuilder.NewField(value).SetName("gender")
		},

		CUTimeFields: new(CUTimeFields).Builder(),
		Times:        Times,
	}
	return pf
}

func TestProfileDoc(t *testing.T) {
	profileFields := new(ProfileFields).Builder()
	fields := apidocbuilder.FieldStructToArray(profileFields)
	args := apidocbuilder.Fields2DocParams(fields...)
	doc := args.Makedown()
	fmt.Println(doc)
}
