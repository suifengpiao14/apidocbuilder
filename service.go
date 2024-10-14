package apidocbuilder

import (
	"encoding/json"
	"fmt"
	"strings"
)

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
	Variables           Variables `json:"variables"`
	Navigates           Navigates `json:"navigates"`
	DocumentRef         string    `json:"documentRef"`
	Apis                Apis      `json:"apis"`
	requestContentType  string    // 批量给apis 设置
	responseContentType string    // 批量给apis 设置
}

func (s *Service) RegisterContentType(requestContentType, responseContentType string) {
	s.requestContentType, s.responseContentType = requestContentType, responseContentType
	s.Apis.SetContentTypeIfEmpty(requestContentType, responseContentType)
}

func (s *Service) AddApi(apis ...Api) {
	// 默认请求和响应内容类型
	Apis(apis).SetContentTypeIfEmpty(s.requestContentType, s.responseContentType)
	s.Apis.Append(apis...).WithService(s)
}

func (s Service) GetApi(method, path string) (api *Api, err error) {
	return s.Apis.GetApi(method, path)
}

func (s Service) GetApiByName(apiName string) (api *Api, err error) {
	return s.Apis.GetApiByName(apiName)
}

func (s *Service) TitleOrDescription() string {
	if s.Title != "" {
		return s.Title
	}
	return s.Description
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

func (s *Service) WithDocumentRefDomain(domain string) *Service {
	s.DocumentRef = withDomain(domain, s.DocumentRef)
	if s.Apis == nil {
		s.Apis = make(Apis, 0)
	}
	s.Apis.WithDocumentRefDomain(domain)

	return s
}

func withDomain(domain string, path string) string {
	if path == "" {
		return path
	}
	if strings.HasPrefix(path, "http") {
		return path
	}
	return fmt.Sprintf("%s%s", domain, path)

}
func (s *Service) GetFormPath() (formUrl string) {
	return getFormPath(s.DocumentRef)
}

func getFormPath(documentRef string) (formUrl string) {
	return fmt.Sprintf("%s/form", documentRef)
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
