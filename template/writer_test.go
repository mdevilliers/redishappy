package template

import (
	"fmt"
	"github.com/mdevilliers/redishappy/types"
	"testing"
)

func TestLoadTempate(t *testing.T) {

	path := "../example_haproxy_template.cfg"
	master1 := types.MasterDetails{Name: "one", Ip: "10.0.0.1", Port: 2345, ExternalPort: 5432}
	master2 := types.MasterDetails{Name: "two", Ip: "10.0.1.1", Port: 5432, ExternalPort: 2345}
	arr := []types.MasterDetails{master1, master2}

	renderedTemplate, err := RenderTemplate(path, &arr)

	if err != nil {
		t.Error("Error rendering test file")
	}

	fmt.Printf("%s", renderedTemplate)
}

func TestLoadNonExistingTempate(t *testing.T) {

	path := "does_not_exist_template.cfg"
	master1 := types.MasterDetails{Name: "one", Ip: "10.0.0.1", Port: 2345, ExternalPort: 5432}

	arr := []types.MasterDetails{master1}

	_, err := RenderTemplate(path, &arr)

	if err == nil {
		t.Error("Template doesn't exist - this should error")
	}
}
