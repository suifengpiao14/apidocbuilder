package apidocbuilder_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/suifengpiao14/apidocbuilder"
	"github.com/suifengpiao14/sqlbuilder"
)

type makeBodyStruct struct {
	Name               string               `json:"name"`
	Age                int                  `json:"age"`
	Address            string               `json:"address"`
	Items              []MakeBodyStructItem `json:"items"`
	MakeBodyStructItem `json:"makeBodyStructItem"`
	Ids                []int          `json:"ids"`
	Map                map[string]int `json:"map"`
}
type MakeBodyStructItem struct {
	Name     string    `json:"name"`
	Title    string    `json:"title"`
	UserRef  *userBody `json:"userRef"`
	user     *userBody
	UserBody userBody `json:"user"`
	Data     any      `json:"data"`
}
type userBody struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Items []Item
}

type Item struct {
	Name string `json:"name"`
}

type DataJson struct {
	Data json.RawMessage `json:"data"`
}

func TestInitNilFields(t *testing.T) {
	t.Run("simple struct", func(t *testing.T) {
		body := userBody{
			//Name: "a",
		}
		apidocbuilder.InitNilFields(body)
		b, _ := json.Marshal(body)
		s := string(b)
		fmt.Println(s)
	})
	t.Run("more struct", func(t *testing.T) {
		var ub userBody

		body := MakeBodyStructItem{
			UserBody: userBody{Name: "a"},
			Data:     ub.Items,
		}

		apidocbuilder.InitNilFields(&body)
		b, _ := json.Marshal(body)
		s := string(b)
		fmt.Println(s)
	})
	t.Run("complex struct", func(t *testing.T) {
		body := makeBodyStruct{}
		apidocbuilder.InitNilFields(&body)
		b, _ := json.Marshal(&body)
		s := string(b)
		fmt.Println(s)
	})

	t.Run("json.RawMessage", func(t *testing.T) {
		var d DataJson
		apidocbuilder.InitNilFields(&d)
		b, _ := json.Marshal(&d)
		s := string(b)
		fmt.Println(s)
	})

}

func TestMakeBody(t *testing.T) {
	var ub userBody
	body := MakeBodyStructItem{
		UserBody: userBody{Name: "a"},
		Data:     ub.Items,
	}
	s := apidocbuilder.MakeBody(body)
	fmt.Println(s)
}

func TestInitNilFields2(m *testing.T) {
	var api = Pagination{}
	out := ErrorOut{
		Data: api.Out,
	}
	apidocbuilder.InitNilFields(&out)
	b, _ := json.Marshal(&out)
	fmt.Println(string(b))
}

func TestNewExample(t *testing.T) {
	api := Pagination{}
	apiDoc := api.ApiDoc()
	s := apiDoc.Examples[0].Response
	fmt.Println(s)
}

type Pagination struct {
	Params QueryPagination
	Out    PaginationOut
}
type ErrorOut struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func (e ErrorOut) Fields() sqlbuilder.Fields {
	return sqlbuilder.Fields{
		NewCode(e.Code),
		NewMessage(e.Message),
		NewData(e.Data),
	}
}
func NewCode(code string) *sqlbuilder.Field {
	return sqlbuilder.NewStringField(code, "code", "错误编码", 20).SetDescription("0-正常,其它-异常")
}
func NewMessage(message string) *sqlbuilder.Field {
	return sqlbuilder.NewStringField(message, "message", "错误信息", 1024)
}
func NewData(data any) *sqlbuilder.Field {
	return sqlbuilder.NewField(func(_ any) (any, error) {
		if data == nil {
			return nil, nil
		}
		b, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}

		inputValue := string(b)
		return inputValue, nil
	}).SetName("data").SetTitle("返回数据").SetType(sqlbuilder.Schema_doc_Type_null)
}

func (api Pagination) ApiDoc() apidocbuilder.Api {
	reqFields := apidocbuilder.StructToFields(api.Params)
	requestParams := apidocbuilder.Fields2DocParams(reqFields...)
	out := ErrorOut{
		Data: api.Out,
	}
	respFields := apidocbuilder.StructToFields(out)
	responseFields := apidocbuilder.Fields2DocParams(respFields...)
	apiDoc := apidocbuilder.Api{
		Group:        "matecriteria",
		Title:        "列表",
		Description:  "列表",
		RequestBody:  requestParams,
		ResponseBody: responseFields,
	}
	apiDoc.NewExample(api.Params, out)
	return apiDoc
}

type QueryPagination struct {
	BirthYearMin string `json:"birthYearMin"`
	HeightMin    int    `json:"heightMin,string"`
	HeightMax    int    `json:"heightMax,string"`

	WeightMin int `json:"weightMin,string"`
	WeightMax int `json:"weightMax,string"`

	AnnualSalaryMin int `json:"annualSalaryMin,string"`

	Habit         string `json:"habit"`
	RegisterPlace string `json:"registerPlace"`
	LivingPlace   string `json:"livingPlace"`
	NativePlace   string `json:"nativePlace"`
	ZodiacAnimal  string `json:"zodiacAnimal"`
	Constellation string `json:"constellation"`
	Gender        string `json:"gender"`

	PageSize  string `json:"pageSize"`
	PageIndex string `json:"pageIndex"`
}

type PaginationOut struct {
	Pagination PaginationObj  `json:"pagination"`
	Users      []MateCriteria `json:"users"`
}

type PaginationObj struct {
	PageIndex string `json:"pageIndex"`
	PageSize  string `json:"pageSize"`
	Total     string `json:"total"`
}

type MateCriteria struct {
	Id           int    `gorm:"column:Fid" json:"id,string"`
	UserId       string `gorm:"column:FuserId" json:"userId"`
	BirthYearMax string `gorm:"column:Fbirth_year_max" json:"birthYearMax"`
	BirthYearMin string `gorm:"column:Fbirth_year_min" json:"birthYearMin"`
	HeightMin    int    `gorm:"column:Fheight_min" json:"heightMin,string"`
	HeightMax    int    `gorm:"column:Fheight_max" json:"heightMax,string"`

	WeightMin int `gorm:"column:Fweight_min" json:"weightMin,string"`
	WeightMax int `gorm:"column:Fweight_max" json:"weightMax,string"`

	AnnualSalaryMin int `gorm:"column:Fannual_salary_min" json:"annualSalaryMin,string"`

	Habit         string `gorm:"column:Fhabit" json:"habit"`
	RegisterPlace string `gorm:"column:Fregister_place" json:"registerPlace"`
	LivingPlace   string `gorm:"column:Fliving_place" json:"livingPlace"`
	NativePlace   string `gorm:"column:Fnative_place" json:"nativePlace"`
	ZodiacAnimal  string `gorm:"column:Fzodiac_animal" json:"zodiacAnimal"`
	Constellation string `gorm:"column:Fconstellation" json:"constellation"`

	Interest string `gorm:"column:Finterest" json:"interest"`
	Gender   string `gorm:"column:Fgender" json:"gender"`

	CreatedAt string `gorm:"column:Fcreated_at" json:"createdAt"`
	UpdateAt  string `gorm:"column:Fupdated_at" json:"updatedAt"`
}

func TestOutputArr(t *testing.T) {
	var stringArr []string
	out := ErrorOut{
		Data: stringArr,
	}
	respFields := apidocbuilder.StructToFields(out)
	responseFields := apidocbuilder.Fields2DocParams(respFields...)
	fmt.Println(responseFields)
}
