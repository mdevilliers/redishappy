package template

import (
	"fmt"
	"testing"
	"github.com/mdevilliers/redishappy/configuration"
)

func TestLoadTempate(t *testing.T) {

	proxy := configuration.HAProxy{TemplatePath : "../example_haproxy_template.cfg"}
	master1 := MasterDetails{Name : "one", Ip : "10.0.0.1", Port : 2345, ExternalPort : 5432}
	master2 := MasterDetails{Name : "two", Ip : "10.0.1.1", Port : 5432, ExternalPort : 2345}
	arr := []MasterDetails{master1, master2}

	renderedTemplate, err := ExecuteTemplate( &proxy, &arr )

	if err != nil{
		t.Error("Error writing test file")
	}

	fmt.Printf("%s", renderedTemplate)
}

