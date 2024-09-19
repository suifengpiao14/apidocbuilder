package apidocbuilder_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/suifengpiao14/apidocbuilder"
)

func TestMarkdown(t *testing.T) {
	api := apidocbuilder.API{Name: "test", Description: "测试"}
	out, err := apidocbuilder.Markdown(api)
	require.NoError(t, err)
	s := string(out)
	fmt.Println(s)
}

func TestHtml(t *testing.T) {
	api := apidocbuilder.API{Name: "test", Description: "测试"}
	md, err := apidocbuilder.Markdown(api)
	require.NoError(t, err)
	htm, err := apidocbuilder.Markdown2HTML(md)
	require.NoError(t, err)
	s := string(htm)
	fmt.Println(s)
}
