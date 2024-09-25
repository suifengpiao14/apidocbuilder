package apidocbuilder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/suifengpiao14/funcs"
	"github.com/suifengpiao14/lineschema"
	"github.com/suifengpiao14/lineschemagogenerate"
	"github.com/suifengpiao14/pathtransfer"
	"github.com/tidwall/gjson"
)

const (
	HEADER_NAME_CONTENT_TYPE       = "Content-Type"
	Header_Value_Content_Type_Json = "application/json"
)

type Api struct {
	Group  string `json:"group"`
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
	Service             *Service   `json:"service"`
	RequestContentType  string     `json:"requestContentType"` // 内容格式，该头部比较重要且经常使用，单独增加字段
	RequestHeader       Header     `json:"requestHeader"`
	ResponseContentType string     `json:"responseContentType"` // 内容格式，该头部比较重要且经常使用，单独增加字段
	ResponseHeader      Header     `json:"responseHeader"`
	Query               Query      `json:"query"`
	RequestBody         Parameters `json:"requestBody"`
	ResponseBody        Parameters `json:"responseBody"`
	Examples            Examples   `json:"examples"`
	Links               Links      `json:"links"`
	DocumentRef         string     `json:"documentRef"`
}

// GetFirstExample 获取第一个example 模板中有使用
func (api Api) GetFirstExample() (example *Example) {
	if len(api.Examples) > 0 {
		return api.Examples[0]
	}
	return &Example{}
}

func (api *Api) NewExample(request any, response any) (example *Example) {
	if len(api.Examples) == 0 {
		api.Examples = make(Examples, 0)

	}
	headers := make(map[string]string, 0)
	if api.RequestContentType != "" {
		headers["Content-Type"] = api.RequestContentType
	}
	example = &Example{
		Method:      api.Method,
		Title:       api.TitleOrDescription(),
		Summary:     api.Summary,
		URL:         api.Path,
		Headers:     headers,
		ContentType: api.RequestContentType,
	}
	example.SetRequestBody(request).SetResponseBody(response)
	api.Examples = append(api.Examples, example)
	return example
}

func (api *Api) SetContentTypeIfEmpty(requestContentType, responseContentType string) {
	if api.RequestContentType == "" {
		api.RequestContentType = requestContentType
	}
	if api.ResponseContentType == "" {
		api.ResponseContentType = responseContentType
	}
}
func (api *Api) SetContentType(requestContentType, responseContentType string) {
	api.RequestContentType = requestContentType
	api.ResponseContentType = responseContentType
}
func (api *Api) IsSameMethodAndPath(method, path string) (yes bool) {
	yes = strings.EqualFold(api.Method, method) && strings.EqualFold(api.Path, path)
	return yes
}
func (api *Api) IsSameName(name string) (yes bool) {
	yes = strings.EqualFold(api.Name, name)
	return yes
}

// IsRequestContentTypeJson 判断请求是否为json请求格式
func (api Api) IsRequestContentTypeJson() (yes bool) {
	return api.isJson("request")
}

// IsRequestContentTypeJson 判断返回是否为json请求格式
func (api Api) IsResponseContentTypeJson() (yes bool) {
	return api.isJson("response")
}
func (api *Api) WithDocumentRefDomain(domain string) {
	api.DocumentRef = withDomain(domain, api.DocumentRef)
}

func (api Api) isJson(typ string) (yes bool) {
	value := api.RequestContentType
	if strings.EqualFold(typ, "response") {
		value = api.ResponseContentType
	}
	value = strings.ToLower(value)
	lowTypeJson := strings.ToLower(Header_Value_Content_Type_Json)
	yes = strings.Contains(value, lowTypeJson)
	return yes
}

func (api Api) TitleOrDescription() (titleOrDescription string) {
	if api.Title != "" {
		return api.Title
	}
	return api.Description

}

func (api Api) CURLExample() (curlExample string, err error) {
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

func (api *Api) Json() (apiJson string, err error) {
	b, err := json.Marshal(api)
	if err != nil {
		return "", err
	}
	apiJson = string(b)
	return apiJson, nil
}

func (api *Api) Init() {
	if api.Service == nil {
		api.Service = &Service{}
	}
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
		api.Examples = make(Examples, 0)
	}
	if api.Name == "" { // 设置名称，确保名称一定存在
		path := strings.ReplaceAll(strings.Trim(api.Path, "/"), "/", "_")
		api.Name = strings.ToLower(funcs.ToLowerCamel(path))
	}

}

// GetRequestResponseLineschema 获取输入输出lineschema
func (api *Api) GetRequestResponseLineschema() (reqLineschema *lineschema.Lineschema, respLineschema *lineschema.Lineschema, err error) {
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
func (api *Api) GetRequestResponseSchemaStructs(namespace string) (reqSchemaStructs lineschemagogenerate.Structs, rspSchemaStructs lineschemagogenerate.Structs, err error) {
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

type Apis []Api

var ERROR_NOT_FOUND_API = errors.New("not found api")

func (apis Apis) GetApi(method string, path string) (api *Api, err error) {
	for i := 0; i < len(apis); i++ {
		if apis[i].IsSameMethodAndPath(method, path) {
			return &apis[i], nil
		}
	}
	err = errors.WithMessagef(ERROR_NOT_FOUND_API, "method:%s,path:%s", method, path)
	return nil, err
}
func (apis Apis) SetContentTypeIfEmpty(requestContentType, responseContentType string) {
	for i := 0; i < len(apis); i++ {
		apis[i].SetContentTypeIfEmpty(requestContentType, responseContentType)
	}
}
func (apis Apis) GetApiByName(apiName string) (api *Api, err error) {
	for i := 0; i < len(apis); i++ {
		if apis[i].IsSameName(apiName) {
			return &apis[i], nil
		}
	}
	err = errors.WithMessagef(ERROR_NOT_FOUND_API, "api name:%s", apiName)
	return nil, err
}

func (apis Apis) Json() (apisJson string, err error) {
	b, err := json.Marshal(apis)
	if err != nil {
		return "", err
	}
	apisJson = string(b)
	return apisJson, nil
}

func (a Apis) Len() int           { return len(a) }
func (a Apis) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Apis) Less(i, j int) bool { return a[i].Name < a[j].Name }

func (a Apis) GetGroups() (groups []string) {
	m := make(map[string]bool)
	for _, api := range a {
		if !m[api.Group] {
			m[api.Group] = true
			groups = append(groups, api.Group)
		}
	}
	return groups
}

func (a Apis) GetByGroups(groups ...string) (apis Apis) {
	apis = make(Apis, 0)
	for _, group := range groups {
		for _, api := range a {
			if strings.EqualFold(api.Group, group) {
				apis = append(apis, api)
			}
		}
	}
	return apis
}
func (a *Apis) Append(apis ...Api) *Apis {
	if *a == nil {
		*a = Apis{}
	}
	Apis(apis).Init()
	*a = append(*a, apis...)
	return a
}

func (a Apis) Init() {
	for i := 0; i < len(a); i++ {
		(a)[i].Init()
	}
}
func (a *Apis) WithDocumentRefDomain(domain string) *Apis {
	for i := 0; i < len(*a); i++ {
		api := &(*a)[i]
		api.WithDocumentRefDomain(domain)
	}

	return a
}
func (a *Apis) WithService(service *Service) *Apis {
	for i := 0; i < len(*a); i++ {
		api := &(*a)[i]
		api.Service = service
	}

	return a
}

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

func (api *Api) Example() (example Example, err error) {

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

type Examples []*Example

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

func (example *Example) SetRequestBody(request any) *Example {
	example.RequestBody = MakeBody(request)
	return example
}

func (example *Example) SetResponseBody(response any) *Example {
	example.Response = MakeBody(response)
	return example
}

func InitNilFields(data any) {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Ptr {
		err := errors.New("InitNilFields data must be a pointer")
		panic(err)
	}
	initNilFields(v)
}

func initNilFields(v reflect.Value) {
	// Handle input based on its kind: struct, pointer, slice, array, or interface
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() && v.CanSet() { //为空,并且能初始化（首字母大写属性或者方法）
			sub := reflect.New(v.Type().Elem())
			v.Set(sub)
		}
		initNilFields(v.Elem())
	case reflect.Map:
		// Initialize map if nil
		if v.IsNil() {
			v.Set(reflect.MakeMap(v.Type()))
		}
	case reflect.Interface:
		if v.IsNil() {
			return
		}
		sub := v.Elem()
		if !sub.CanSet() {
			sub = reflect.New(sub.Type()).Elem()
		}
		initNilFields(sub)
		v.Set(sub) // InitNilFields填充值后再赋值
	case reflect.Struct:
		// Handle struct types by initializing fields
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			initNilFields(field)
		}
	case reflect.Slice, reflect.Array:
		if v.Kind() == reflect.Slice && v.CanSet() && v.IsNil() {
			v.Set(reflect.MakeSlice(v.Type(), 1, 1)) // Initialize with a single element slice
		}
		for i := 0; i < v.Len(); i++ {
			initNilFields(v.Index(i))
		}
	}
}

// getRefVariable 获取指针类型数据
func getRefVariable(data any) (ref any) {
	rv := reflect.ValueOf(data)
	if rv.Kind() != reflect.Ptr { // 非指针类型，转为指针类型
		ref := reflect.New(rv.Type())
		ref.Elem().Set(rv)
		data = ref.Interface()
	}
	return data
}

func MakeBody(data any) (s string) {
	//格式化data 对 data 内部的 any 地址 数组等结构进行初始化，确保能正确输出所有结构体结构数据
	data = getRefVariable(data)
	InitNilFields(data)
	switch v := data.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		var out bytes.Buffer
		enc := json.NewEncoder(&out)
		enc.SetIndent("", "    ")
		if err := enc.Encode(data); err != nil {
			err = errors.WithMessagef(err, "makeBody: json.Marshal,data:%v", data)
			panic(err)
		}
		return out.String()
	}
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

func ApiJson2Schema(api Api) (out ApiJson2SchemaOut, err error) {
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
