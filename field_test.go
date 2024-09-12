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

type Account struct {
	Identify string `json:"identify"`
	Password string `json:"password"`
}

func (a Account) Fields() sqlbuilder.Fields {
	return sqlbuilder.Fields{
		sqlbuilder.NewField(a.Identify).SetName("identify").SetTitle("账号"),
		sqlbuilder.NewField(a.Password).SetName("password").SetTitle("密码"),
		sqlbuilder.NewField("").SetName("createAt").SetTitle("创建时间"),
	}
}

type Book struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
}

func (b Book) Fields() sqlbuilder.Fields {
	return sqlbuilder.Fields{
		sqlbuilder.NewField(b.Id).SetName("id").SetTitle("用户ID"),
		sqlbuilder.NewField(b.Title).SetName("title").SetTitle("书名"),
	}
}

type User struct {
	Id       int     `json:"id"`
	Name     string  `json:"name"`
	Nickname string  `json:"nickname"`
	Account  Account `json:"account"`
	Books    []*Book `json:"books"`
}

func (u User) Fields() sqlbuilder.Fields {
	return sqlbuilder.Fields{
		sqlbuilder.NewField(u.Id).SetName("id").SetTitle("用户ID").SetType(sqlbuilder.Schema_Type_int),
		sqlbuilder.NewField(u.Name).SetName("name").SetTitle("用户名"),
		sqlbuilder.NewField(u.Nickname).SetName("nickname").SetTitle("昵称"),
	}
}

func TestUser(t *testing.T) {
	user := User{}
	fields := apidocbuilder.FieldStructToArray(user)
	args := apidocbuilder.Fields2DocParams(fields...)
	doc := args.Makedown()
	fmt.Println(doc)
}

func TestStruct2DocName(t *testing.T) {
	user := User{}
	fields := user.Fields()
	fields = append(fields, user.Account.Fields()...)
	book := Book{}
	fields = append(fields, book.Fields()...)
	docFields := apidocbuilder.Struct2DocName(user, fields)
	args := apidocbuilder.Fields2DocParams(docFields...)
	doc := args.Makedown()
	fmt.Println(doc)
}
