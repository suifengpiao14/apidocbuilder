package apidocbuilder_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/suifengpiao14/apidocbuilder"
)

var apis = apidocbuilder.Apis{
	{Name: "test", Description: "测试", DocumentRef: "http://doc.com/test"},
	{Name: "test2", Description: "测试2", DocumentRef: "http://doc.com/test2"},
}

var service = apidocbuilder.Service{
	Name:        "test service",
	Description: "测试 服务描述",
	Servers: apidocbuilder.Servers{
		{Name: "local", Description: "开发环境", URL: "http://api.com", IP: "127.0.0.1"},
		{Name: "test", Description: "测试环境", URL: "http://api.com", IP: "10.10.111.12"},
	},
	Apis: apis,
}

func TestApi2Markdown(t *testing.T) {
	api := apis[0]
	out, err := apidocbuilder.Api2Markdown(api)
	require.NoError(t, err)
	s := string(out)
	fmt.Println(s)
}
func TestService2Markdown(t *testing.T) {

	out, err := apidocbuilder.Service2Markdown(service)
	require.NoError(t, err)
	s := string(out)
	fmt.Println(s)
}

func TestApiHtml(t *testing.T) {
	api := apis[0]
	md, err := apidocbuilder.Api2Markdown(api)
	require.NoError(t, err)
	htm, err := apidocbuilder.Markdown2HTML(md)
	require.NoError(t, err)
	s := string(htm)
	fmt.Println(s)
}
func TestServiceHtml(t *testing.T) {
	md, err := apidocbuilder.Service2Markdown(service)
	require.NoError(t, err)
	htm, err := apidocbuilder.Markdown2HTML(md)
	require.NoError(t, err)
	s := string(htm)
	fmt.Println(s)
}
