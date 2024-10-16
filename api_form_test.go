package apidocbuilder_test

import (
	"fmt"
	"testing"

	"github.com/suifengpiao14/apidocbuilder"
)

func TestA(t *testing.T) {
	api := apidocbuilder.Api{}
	htmxForm := apidocbuilder.NewHtmxForm(api)
	ht := htmxForm.Html()
	fmt.Println(ht)

}
