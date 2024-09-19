package apidocbuilder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/suifengpiao14/lineschema"
	"github.com/suifengpiao14/lineschemagogenerate"
	"github.com/suifengpiao14/pathtransfer"
	"github.com/tidwall/gjson"
)

const (
	HEADER_NAME_CONTENT_TYPE       = "Content-Type"
	Header_Value_Content_Type_Json = "application/json"
)

type API struct {
	Domain string `json:"domain"`
	Scene  string `json:"scene"`
	Name   string `json:"name"`
	// 标题
	Title string `json:"title"`
	// 路径
	Path   string `json:"path"`
	Method string `json:"method"`
	// 摘要
	Summary string `json:"summary"`
	// 介绍
	Description string `json:"description"`
	// 服务
	Service             Service    `json:"service"`
	RequestContentType  string     `json:"requestContentType"` // 内容格式，该头部比较重要且经常使用，单独增加字段
	RequestHeader       Header     `json:"requestHeader"`
	ResponseContentType string     `json:"responseContentType"` // 内容格式，该头部比较重要且经常使用，单独增加字段
	ResponseHeader      Header     `json:"responseHeader"`
	Query               Query      `json:"query"`
	RequestBody         Parameters `json:"requestBody"`
	ResponseBody        Parameters `json:"responseBody"`
	Examples            []Example  `json:"examples"`
	Links               Links      `json:"links"`
}

// IsRequestContentTypeJson 判断请求是否为json请求格式
func (api API) IsRequestContentTypeJson() (yes bool) {
	return api.isJson("request")
}

// IsRequestContentTypeJson 判断返回是否为json请求格式
func (api API) IsResponseContentTypeJson() (yes bool) {
	return api.isJson("response")
}

func (api API) isJson(typ string) (yes bool) {
	value := api.RequestContentType
	if strings.EqualFold(typ, "response") {
		value = api.ResponseContentType
	}
	value = strings.ToLower(value)
	lowTypeJson := strings.ToLower(Header_Value_Content_Type_Json)
	yes = strings.Contains(value, lowTypeJson)
	return yes
}

func (api API) CURLExample() (curlExample string, err error) {
	var w bytes.Buffer
	w.WriteString("curl ")
	u := url.URL{
		Scheme:   "",
		Host:     "",
		Path:     api.Path,
		RawPath:  api.Path,
		RawQuery: api.Query.Encode(),
	}
	firstU := api.Service.Servers.GetFirst()
	if firstU.URL != "" {
		tu, err := url.Parse(firstU.URL)
		if err != nil {
			return "", err
		}
		u.Scheme = tu.Scheme
		u.Host = tu.Host
		if firstU.Proxy != "" {
			w.WriteString(fmt.Sprintf(` -x%s `, firstU.Proxy))
		}
	}

	w.WriteString(fmt.Sprintf("-X%s", api.Method))

	for _, h := range api.RequestHeader {
		value := h.Example
		if value == "" {
			value = h.Default
		}
		w.WriteString(fmt.Sprintf(` -H '%s: %s'`, h.Name, value))
	}
	contentType := api.RequestHeader.ContentType()
	if contentType == "" {
		contentType = Header_Value_Content_Type_Json
	}
	switch contentType {
	case Header_Value_Content_Type_Json:
		lineschema := api.RequestBody.Lineschema("", false)
		jsonExample, err := lineschema.JsonExample()
		if err != nil {
			return "", err
		}
		if jsonExample != "" {
			w.WriteString(fmt.Sprintf(` -d'%s' `, jsonExample))
		}
	}

	w.WriteString(fmt.Sprintf(` '%s'`, u.String()))
	curlExample = w.String()
	return curlExample, nil
}

func (api *API) Json() (apiJson string, err error) {
	b, err := json.Marshal(api)
	if err != nil {
		return "", err
	}
	apiJson = string(b)
	return apiJson, nil
}

func (api *API) Init() {
	if api.Service.Servers == nil {
		api.Service.Servers = make(Servers, 0)
	}
	if api.Service.Contacts == nil {
		api.Service.Contacts = make([]Contact, 0)
	}
	if api.RequestHeader == nil {
		api.RequestHeader = Header{}
	}
	if api.ResponseHeader == nil {
		api.ResponseHeader = Header{}
	}
	if api.Query == nil {
		api.Query = Query{}
	}
	if api.RequestBody == nil {
		api.RequestBody = Parameters{}
	}
	if api.ResponseBody == nil {
		api.ResponseBody = Parameters{}
	}
	if api.Examples == nil {
		api.Examples = make([]Example, 0)
	}

}

// GetRequestResponseLineschema 获取输入输出lineschema
func (api *API) GetRequestResponseLineschema() (reqLineschema *lineschema.Lineschema, respLineschema *lineschema.Lineschema, err error) {
	schemas, err := ApiJson2Schema(*api)
	if err != nil {
		return nil, nil, err
	}
	reqLineschema, err = lineschema.ParseLineschema(schemas.RequestLineSchema)
	if err != nil {
		return nil, nil, err
	}
	respLineschema, err = lineschema.ParseLineschema(schemas.ResponseLineSchema)
	if err != nil {
		return nil, nil, err
	}
	return reqLineschema, respLineschema, nil
}

// GetRequestResponseSchemaStructs 获取接口输入输出schema 结构体
func (api *API) GetRequestResponseSchemaStructs(namespace string) (reqSchemaStructs lineschemagogenerate.Structs, rspSchemaStructs lineschemagogenerate.Structs, err error) {
	reqLineschema, respLineschema, err := api.GetRequestResponseLineschema()
	if err != nil {
		return nil, nil, err
	}

	reqSchemaStructs = lineschemagogenerate.NewSturct(*reqLineschema)
	rspSchemaStructs = lineschemagogenerate.NewSturct(*respLineschema)
	reqSchemaStructs.AddNameprefix(namespace)
	rspSchemaStructs.AddNameprefix(namespace)

	return reqSchemaStructs, rspSchemaStructs, nil
}

type APIs []API

func (apis APIs) Json() (apisJson string, err error) {
	b, err := json.Marshal(apis)
	if err != nil {
		return "", err
	}
	apisJson = string(b)
	return apisJson, nil
}

func (a APIs) Len() int           { return len(a) }
func (a APIs) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a APIs) Less(i, j int) bool { return a[i].Name < a[j].Name }

type Header Parameters

func (h Header) ContentType() (contentType string) {
	for _, parameter := range h {
		if strings.EqualFold(parameter.Name, HEADER_NAME_CONTENT_TYPE) {
			return parameter.Value()
		}
	}
	return ""
}
func (h Header) ToMap() (headerMap map[string]string) {
	headerMap = make(map[string]string)
	for _, header := range h {
		key := header.Name
		value := header.Value()
		headerMap[key] = value
	}
	return headerMap
}
func (h *Header) Add(parameters ...Parameter) {
	tmp := Parameters(*h)
	tmp.Add(parameters...)
	*h = Header(tmp)
}

type Query Parameters

func (q Query) Encode() (query string) {
	ps := Parameters(q)
	ps.FormatField() // 填充名称字段
	urlValues := url.Values{}
	for _, query := range ps {
		key := query.Name
		value := query.Value()
		urlValues.Add(key, value)
	}
	return urlValues.Encode()
}

func (q *Query) Add(parameters ...Parameter) {
	p := Parameters(*q)
	p.Add(parameters...)
	*q = Query(p)
}

func (api *API) Example() (example Example, err error) {

	summary := api.Summary
	if summary == "" {
		summary = api.Description
	}

	server := api.Service.Servers.GetFirst()
	urlObj, _ := url.Parse(server.URL)
	urlValues := urlObj.Query()
	for _, query := range api.Query {
		key := query.Name
		value := query.Default
		if query.Example != "" {
			value = query.Example
		}
		urlValues.Add(key, value)
	}
	urlObj.RawQuery = urlValues.Encode()

	urlObj.Path = fmt.Sprintf("%s/%s", strings.TrimRight(urlObj.Path, "/"), strings.TrimLeft(api.Path, "/"))
	requestBody, err := api.RequestBody.Json(false)
	if err != nil {
		return
	}
	response, err := api.ResponseBody.Json(false)
	if err != nil {
		return example, err
	}
	example = Example{
		Method:            api.Method,
		Title:             api.Title,
		Summary:           summary,
		Proxy:             server.Proxy,
		URL:               urlObj.String(),
		Headers:           api.RequestHeader.ToMap(),
		ContentType:       api.RequestHeader.ContentType(),
		RequestBody:       requestBody,
		Response:          response,
		RequestPreScript:  api.Service.RequestPreScript,
		RequestPostScript: api.Service.RequestPostScript,
	}
	return
}

type Link struct {
	Title string `json:"title"`
	Doc   string `json:"doc"`
}

type Links []Link

type Variable struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

type Variables []Variable

func (v *Variables) Json() (jsonStr string, err error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	jsonStr = string(b)
	return jsonStr, err
}

type Navigate struct {
	Title string `json:"title"`
	Route string `json:"route"`
	Sort  string `json:"sort"`
	Name  string `json:"name"`
	Doc   string `json:"doc"`
}

type Navigates []Navigate

var ERROR_NOT_FOUND_NAVIGATE = errors.New("not found navigate")

func (vs Navigates) GetByRoute(route string) (navigate *Navigate, err error) {
	for _, n := range vs {
		if n.Route == route {
			return &n, nil
		}
	}
	err = errors.WithMessage(ERROR_NOT_FOUND_NAVIGATE, fmt.Sprintf("by route:%s", route))
	return nil, err
}

// GetByDoc 通过文档找到对应的导航
func (vs Navigates) GetByDoc(doc string) (navigate *Navigate, err error) {
	for _, n := range vs {
		if strings.HasSuffix(n.Doc, doc) || strings.HasSuffix(doc, n.Doc) {
			return &n, nil
		}
	}
	err = errors.WithMessage(ERROR_NOT_FOUND_NAVIGATE, fmt.Sprintf("by doc:%s", doc))
	return nil, err
}

func (v *Navigates) Json() (jsonStr string, err error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	jsonStr = string(b)
	return jsonStr, err
}

type Server struct {
	// 服务器名称
	Name string `json:"name"`

	Title string `json:"title"`
	// url地址
	URL string `json:"url"`
	// 服务器IP
	IP string `json:"ip"`
	// 介绍
	Description string `json:"description"`
	// 代理地址
	Proxy string `json:"proxy"`
	// 扩展字段
	ExtensionIds string `json:"extensionIds"`
}

type Servers []Server

func (s Servers) GetByName(name string) (server Server, exists bool) {
	for _, server := range s {
		if server.Name == name {
			return server, true
		}
	}
	return server, false
}

func (s Servers) GetFirst() (server Server) {
	if len(s) > 0 {
		return s[0]
	}
	return server
}
func (s Servers) Json() (str string, err error) {
	b, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	str = string(b)
	return str, nil
}

const (
	LANGUAGE_BASH = "bash"
)

type LanguageAlias [][]string

// 顺序很重要，每个数组第第一个为标准名称
var LanguageAliasDefault = LanguageAlias{
	{"bash"},
	{"sh"},
	{"javascript", "js"},
	{"go", "golang"},
	{"php"},
	{"python"},
	{"lua"},
	{"tengo"},
}

func (lngAlias LanguageAlias) GetByLanguage(language string) (alias []string) {
	language = strings.ToLower(language)
	alias = make([]string, 0)
	for _, row := range lngAlias {
		for _, lang := range row {
			if strings.ToLower(lang) == language {
				return row
			}
		}
	}
	alias = append(alias, language) // 默认保留本身，即无别名
	return alias
}

// 将传入的语言名称转换为标准语言名称(取别名的第一个)
func (lngAlias LanguageAlias) GetStandardLanguage(alias string) (language string, err error) {
	for _, row := range lngAlias {
		for _, lang := range row {
			if strings.EqualFold(alias, lang) {
				return row[0], nil
			}
		}
	}
	err = errors.Errorf("not found StandardLanguage name by alias:%s", alias)
	return "", err
}

const (
	SCRIPT_KEY_LANGUAGE = "language"
	SCRIPT_KEY_CONTENT  = "script"
)

type Script struct {
	Language string `json:"language"`
	Text     string `json:"text"`
}

func (s Script) IsEqual(script Script) (ok bool) {
	ok = strings.EqualFold(s.Language, script.Language) && strings.EqualFold(s.Text, script.Text)
	return ok
}

type Scripts []Script

func (s *Scripts) Add(scripts ...Script) {
	if s == nil {
		s = &Scripts{}
	}
	*s = append(*s, scripts...)
}

func (s *Scripts) FilterByLanguage(language string) (scripts Scripts) {
	scripts = make(Scripts, 0)
	if s == nil {
		return scripts
	}
	alias := LanguageAliasDefault.GetByLanguage(language)
	for _, tmpScript := range *s {
		for _, alia := range alias {
			if strings.EqualFold(tmpScript.Language, alia) {
				scripts.Add(tmpScript)
			}
		}

	}
	return scripts
}

func (s Scripts) First() (script *Script, ok bool) {
	for _, tmpScript := range s {
		return &tmpScript, true

	}
	return nil, false
}

func (scripts Scripts) String() (s string) {
	var w strings.Builder
	for _, script := range scripts {
		w.WriteString(script.Text)
		w.WriteString("\n")
	}
	return w.String()
}

func (s *Scripts) Languages() (languages []string, err error) {
	m := make(map[string]struct{})
	if s == nil {
		return languages, nil
	}

	for _, tmpScript := range *s {
		standardLanguage, err := LanguageAliasDefault.GetStandardLanguage(tmpScript.Language)
		if err != nil {
			return nil, err
		}
		m[standardLanguage] = struct{}{}
	}

	languages = make([]string, 0)
	for lang := range m {
		languages = append(languages, lang)
	}
	sort.Strings(languages)
	return languages, nil
}

type Service struct {
	Name string `json:"name"`
	// 服务标识
	Servers Servers `json:"servers"`
	// 标题
	Title string `json:"title"`
	// 介绍
	Description string `json:"description"`
	// 版本
	Version string `json:"version"`
	// 联系人
	Contacts []Contact `json:"contacts"`
	// 协议
	License string `json:"license"`
	// 鉴权
	Security string `json:"security"`
	// 前置请求脚本
	RequestPreScript Scripts `json:"requestPreScript"`
	// 后置请求脚本
	RequestPostScript Scripts `json:"requestPostScript"`
	// json字符串
	Variables Variables `json:"variables"`
	Navigates Navigates `json:"navigates"`
	Document  string    `json:"document"`
}

func (s *Service) AddServer(servers ...Server) {
	if s.Servers == nil {
		s.Servers = make([]Server, 0)
	}

	for _, server := range servers {
		if server.Title == "" {
			server.Title = makeTitle(server.Description)
		}
		s.Servers = append(s.Servers, server)
	}

	// 去重
	m := make(map[string]Server)
	for _, server := range s.Servers {
		m[server.Name] = server
	}
	newServers := make([]Server, 0)
	for _, server := range m {
		newServers = append(newServers, server)
	}
	s.Servers = newServers

}

func (s *Service) AddConcat(concats ...Contact) {
	if s.Contacts == nil {
		s.Contacts = make([]Contact, 0)
	}
	s.Contacts = append(s.Contacts, concats...)
}

func (s *Service) AddVariable(variables ...Variable) {
	if s.Variables == nil {
		s.Variables = make([]Variable, 0)
	}
	s.Variables = append(s.Variables, variables...)
}

func (s *Service) AddNavigate(navigates ...Navigate) {
	if s.Navigates == nil {
		s.Navigates = make(Navigates, 0)
	}
	for _, nav := range navigates {
		if nav.Route == "" {
			nav.Route = nav.Doc
			lastSlashIndex := strings.LastIndex(nav.Route, "/")
			if lastSlashIndex > -1 {
				nav.Route = nav.Route[lastSlashIndex:]
			}
			firstDotIndex := strings.Index(nav.Route, ".")
			if firstDotIndex > -1 {
				nav.Route = nav.Route[:firstDotIndex]
			}
			nav.Route = fmt.Sprintf("/%s", strings.Trim(nav.Route, "/"))
		}
		s.Navigates = append(s.Navigates, nav)
	}

	// 去重
	m := make(map[string]Navigate)
	for _, navigate := range s.Navigates {
		m[navigate.Name] = navigate
	}
	newNavigates := make([]Navigate, 0)
	for _, server := range m {
		newNavigates = append(newNavigates, server)
	}
	s.Navigates = newNavigates
}

type Example struct {
	// 标签,mock数据时不同接口案例优先返回相同tag案例
	Tag    string `json:"tag,omitempty"`
	Method string `json:"method,omitempty"`
	// 案例名称
	Title string `json:"title,omitempty"`
	// 简介
	Summary string `json:"summary,omitempty"`
	// URL
	URL   string `json:"url,omitempty"`
	Proxy string `json:"proxy,omitempty"`
	// query 鉴权
	Auth string `json:"auth,omitempty"`
	// query 请求头
	Headers map[string]string `json:"headers,omitempty"`
	// 请求格式(application/json-json,plain/text-文本)
	ContentType       string  `json:"contentType,omitempty"`
	RequestPreScript  Scripts `json:"requestPreScript,omitempty"`
	RequestPostScript Scripts `json:"requestPostScript,omitempty"`
	// query 请求体
	RequestBody string `json:"requestBody,omitempty"`
	// query 返回体测试脚本
	TestScript string `json:"testScript,omitempty"`
	// 请求体
	Response string `json:"response,omitempty"`
}

type Parameter struct {
	Title           string  `json:"title"` // 验证规则标识
	Schema          *Schema // 全称
	Fullname        string  `json:"fullname,omitempty"` // 名称(冗余local.en)
	Name            string  `json:"name,omitempty"`     // 参数类型(string-字符,int-整型,number-数字,array-数组,object-对象)
	Type            string  `json:"type,omitempty"`     // 参数所在的位置(body-BODY,head-HEAD,path-PATH,query-QUERY,cookie-COOKIE)
	Position        string  `json:"position,omitempty"`
	Format          string  `json:"format,omitempty"` // 案例
	Example         string  `json:"example,omitempty"`
	Default         string  `json:"default,omitempty"`                // 是否弃用(true-是,false-否)
	Deprecated      string  `json:"deprecated,omitempty"`             // 是否必须(true-是,false-否)
	Required        bool    `json:"required,omitempty,string"`        // 对数组、对象序列化方法,参照openapi parameters.style
	Serialize       string  `json:"serialize,omitempty"`              // 对象的key,是否单独成参数方式,参照openapi parameters.explode(true-是,false-否)
	Explode         string  `json:"explode,omitempty"`                // 是否容许空值(true-是,false-否)
	AllowEmptyValue bool    `json:"allowEmptyValue,omitempty,string"` // 特殊字符是否容许出现在uri参数中(true-是,false-否)
	AllowReserved   string  `json:"allowReserved,omitempty"`          // 简介
	Description     string  `json:"description,omitempty"`
	Enum            string  `json:"enum,omitempty"`
	EnumNames       string  `json:"enumNames,omitempty"`
	RegExp          string  `json:"regExp"`         // 验证规则
	Vocabulary      string  `json:"vocabularyDict"` // 词汇
}

func (p *Parameter) Value() (value string) {
	value = p.Default
	if p.Example != "" {
		value = p.Example
	}
	return value
}

func (p *Parameter) Copy() (copy Parameter) {
	copy = *p
	if p.Schema != nil {
		schema := p.Schema.Copy()
		copy.Schema = &schema
	}
	return copy
}

func (p *Parameter) SetSchema(schemaJson string) {
	schema := &Schema{}
	json.Unmarshal([]byte(schemaJson), schema) // ignore error
	p.Schema = schema
	//schema 部分值重新赋值
	p.completeSchema()
}

func (p *Parameter) completeSchema() {
	if p.Schema == nil {
		p.Schema = &Schema{}
	}
	schema := p.Schema
	//schema 部分值重新赋值
	if p.Format != "" {
		schema.Format = p.Format
	}
	schema.Type = p.Type
	schema.Required = p.Required
	schema.AllowEmptyValue = p.AllowEmptyValue
	schema.Description = p.Description
	if p.Example != "" {
		schema.Example = p.Example
	}
	if p.Default != "" {
		schema.Default = p.Default
	}
	if len(p.Enum) > 0 {
		schema.Enum = p.Enum
	}
	if len(p.EnumNames) > 0 {
		schema.EnumNames = p.EnumNames
	}
	if schema.Comments == "" {
		schema.Comments = schema.Description
	}
	if schema.Title == "" {
		schema.Title = makeTitle(p.Description) // 此处暂时赋值给title，后续优化
	}
	if schema.Description == schema.Title {
		schema.Description = "" // 避免相同提示重复
	}

}

func (p *Parameter) FormatField() {
	if p.Name == "" { // 格式化名称
		name := p.Fullname
		lastDotIndex := strings.LastIndex(name, ".")
		if lastDotIndex > -1 {
			name = name[lastDotIndex+1:]
		}
		(*p).Name = name
	}

}

type Parameters []Parameter

func (a Parameters) Len() int           { return len(a) }
func (a Parameters) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Parameters) Less(i, j int) bool { return a[i].Fullname < a[j].Fullname }

func (p *Parameters) Add(parameters ...Parameter) {
	tmp := Parameters(parameters)
	tmp.FormatField()
	*p = append(*p, parameters...)
}

// Transfers 收集跟随参数的词汇
func (ps *Parameters) Transfers(namespace string) (vs pathtransfer.Transfers) {
	vs = make(pathtransfer.Transfers, 0)
	for _, p := range *ps {
		path := fmt.Sprintf("%s.%s", namespace, p.Fullname)
		path = strings.ReplaceAll(path, "[]", ".#")
		v := pathtransfer.Transfer{
			Src: pathtransfer.TransferUnit{
				Path: pathtransfer.Path(path),
				Type: p.Type,
			},
			Dst: pathtransfer.TransferUnit{
				Path: pathtransfer.Path(p.Vocabulary),
			},
		}
		vs.AddReplace(v)
	}
	return vs
}

const (
	PARAMETER_ATTR_POSITION_ENUM_HEADER = "header"
)

func (p *Parameters) Lineschema(id string, withHeader bool) (lineSchema lineschema.Lineschema) {

	lineSchema = lineschema.Lineschema{
		Meta: &lineschema.Meta{
			Version: "http://json-schema.org/draft-07/schema#",
			ID:      id,
		},
		Items: make(lineschema.LineschemaItems, 0),
	}
	for _, parameter := range *p {
		if !withHeader && parameter.Position == PARAMETER_ATTR_POSITION_ENUM_HEADER {
			continue
		}
		parameter.completeSchema() //完善schema
		if parameter.Schema != nil {
			item := parameter.Schema.ToLineSchemaItem(parameter.Fullname)
			lineSchema.Items.Add(&item)
		}
	}
	return lineSchema
}

func (p *Parameters) Json(pretty bool) (jsonStr string, err error) {
	id := "example"
	lineschema := p.Lineschema(id, false)
	jsonStr, err = lineschema.JsonExample()
	if err != nil {
		return "", err
	}
	jsonStr = gjson.Get(jsonStr, "@this|@pretty").String()
	return jsonStr, nil
}

// FormatField 填充name字段
func (p *Parameters) FormatField() {
	for i, param := range *p {
		param.FormatField()
		(*p)[i] = param
	}
}

type Schema struct {
	// 标题, apiformview 需要使用
	Title string `json:"title"`
	// 描述
	Description string `json:"description,omitempty"`
	Comments    string `json:"comments,omitempty"`
	// 备注
	Remark string `json:"remark,omitempty"`
	// 类型(integer-整数,array-数组,string-字符串,object-对象)
	Type string `json:"type"`
	// 案例
	Example  string `json:"example,omitempty"`
	Examples string `json:"examples,omitempty"`
	// 是否弃用(true-是,false-否)
	Deprecated bool `json:"deprecated,omitempty,string"`
	// 是否必须(true-是,false-否)
	Required bool `json:"required,omitempty,string"`
	// 枚举值
	Enum string `json:"enum,omitempty"`
	// 枚举名称
	EnumNames string `json:"enumNames,omitempty"`
	// 格式
	Format string `json:"format,omitempty"`
	// 默认值
	Default string `json:"default,omitempty"`
	// 是否可以为空(true-是,false-否)
	Nullable string `json:"nullable,omitempty"`
	// 倍数
	MultipleOf int `json:"multipleOf,omitempty,string"`
	// 最大值
	Maximum int `json:"maxnum,omitempty,string"`
	// 是否不包含最大项(true-是,false-否)
	ExclusiveMaximum bool `json:"exclusiveMaximum,omitempty,string"`
	// 最小值
	Minimum int `json:"minimum,omitempty,string"`
	// 是否不包含最小项(true-是,false-否)
	ExclusiveMinimum bool `json:"exclusiveMinimum,omitempty,string"`
	// 最大长度
	MaxLength int `json:"maxLength,omitempty,string"`
	// 最小长度
	MinLength int `json:"minLength,omitempty,string"`
	// 正则表达式
	Pattern string `json:"pattern,omitempty"`
	// 最大项数
	MaxItems int `json:"maxItems,omitempty,string"`
	// 最小项数
	MinItems int `json:"minItems,omitempty,string"`
	// 所有项是否需要唯一(true-是,false-否)
	UniqueItems bool `json:"uniqueItems,omitempty,string"`
	// 最多属性项
	MaxProperties int `json:"maxProperties,omitempty,string"`
	// 最少属性项
	MinProperties int `json:"minProperties,omitempty,string"`
	// 所有
	AllOf []*Schema `json:"allOf,omitempty"`
	// 只满足一个
	OneOf []*Schema `json:"oneOf,omitempty"`
	// 任何一个SchemaID
	AnyOf []*Schema `json:"anyOf,omitempty"`
	// 是否容许空值(true-是,false-否)
	AllowEmptyValue bool `json:"allowEmptyValue,omitempty"`
	// 特殊字符是否容许出现在uri参数中(true-是,false-否)
	AllowReserved string `json:"allowReserved,omitempty"`
	// 不包含的schemaID
	Not string `json:"not,omitempty"`
	// boolean
	AdditionalProperties string `json:"additionalProperties,omitempty"`
	// schema鉴别
	Discriminator string `json:"discriminator,omitempty"`
	// 是否只读(true-是,false-否)
	ReadOnly bool `json:"readOnly,omitempty,string"`
	// 是否只写(true-是,false-否)
	WriteOnly bool `json:"writeOnly,omitempty,string"`
	// xml对象
	XML string `json:"xml,omitempty"`
	// 附加文档
	ExternalDocs string `json:"externalDocs,omitempty"`
	// 附加文档
	ExternalPros string `json:"externalPros,omitempty"`
	// 扩展字段
	Extensions string `json:"extensions,omitempty"`
	// 简介
	Summary   string `json:"summary,omitempty"`
	SchemaRef string `json:"$schema,omitempty"`
	//Schema     *Schema            `json:"schema,omitempty"`
	Items      *Schema            `json:"items,omitempty"`
	Properties map[string]*Schema `json:"properties,omitempty"`
}

func (s *Schema) Copy() (copy Schema) {
	copy = *s
	return copy
}

func (s *Schema) ToLineSchemaItem(fullname string) (lineschemaItem lineschema.LineschemaItem) {
	lineschemaItem = lineschema.LineschemaItem{
		Comments:         s.Comments,
		Type:             s.Type,
		Enum:             s.Enum,
		EnumNames:        s.EnumNames,
		MultipleOf:       s.MultipleOf,
		Maximum:          s.Maximum,
		ExclusiveMaximum: s.ExclusiveMaximum,
		Minimum:          s.Minimum,
		ExclusiveMinimum: s.ExclusiveMinimum,
		MaxLength:        s.MaxLength,
		MinLength:        s.MinLength,
		Pattern:          s.Pattern,
		MaxItems:         s.MaxItems,
		MinItems:         s.MinItems,
		UniqueItems:      s.UniqueItems,
		MaxContains:      uint(s.MaxProperties),
		MinContains:      uint(s.MinProperties),
		MaxProperties:    s.MaxProperties,
		MinProperties:    s.MinProperties,
		Required:         s.Required,
		Format:           s.Format,
		Description:      s.Description,
		Default:          s.Default,
		Deprecated:       s.Deprecated,
		ReadOnly:         s.ReadOnly,
		WriteOnly:        s.WriteOnly,
		Example:          s.Example,
		Examples:         s.Examples,
		Fullname:         fullname,
		Title:            s.Title,
		AllowEmptyValue:  s.AllowEmptyValue,
	}
	return lineschemaItem
}

type Contact struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email"`
}

type ApiJson2SchemaOut struct {
	RequestJsonSchema  string `json:"requestJsonschema"`
	RequestExample     string `json:"requestExample"`
	RequestLineSchema  string `json:"requestLineSchema"`
	ResponseLineSchema string `json:"responseLineSchema"`
}

func ApiJson2Schema(api API) (out ApiJson2SchemaOut, err error) {
	parameters := Parameters(api.RequestBody)
	hp := Parameters(api.RequestHeader)
	qp := Parameters(api.Query)
	parameters.Add(hp...)
	parameters.Add(qp...)

	queryLineSchema := parameters.Lineschema("in", false)
	b, err := queryLineSchema.JsonSchema()
	if err != nil {
		return out, err
	}
	out.RequestJsonSchema = string(b)
	out.RequestExample, err = queryLineSchema.JsonExample()
	if err != nil {
		return out, err
	}
	out.RequestLineSchema = queryLineSchema.String()

	responseParameter := Parameters(api.ResponseBody)
	responseLineSchema := responseParameter.Lineschema("out", false)
	out.ResponseLineSchema = responseLineSchema.String()
	return out, nil
}

func makeTitle(description string) (title string) {
	if description == "" {
		return ""
	}
	reg := regexp.MustCompile("[\u4e00-\u9fa5\\w]+")
	title = reg.FindString(description)
	return title
}
