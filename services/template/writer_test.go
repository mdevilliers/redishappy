package template

import (
	"fmt"
	"testing"

	"github.com/mdevilliers/redishappy/types"
)

func TestLoadTempate(t *testing.T) {

	path := "../../build/configs/redis-haproxy/haproxy_template.cfg"
	collection := types.NewMasterDetailsCollection()
	collection.AddOrReplace(&types.MasterDetails{Name: "one", Ip: "10.0.0.1", Port: 2345, ExternalPort: 5432})
	collection.AddOrReplace(&types.MasterDetails{Name: "two", Ip: "10.0.1.1", Port: 5432, ExternalPort: 2345})

	renderedTemplate, err := RenderTemplate(path, &collection)

	if err != nil {
		t.Error("Error rendering test file")
	}

	fmt.Printf("%s", renderedTemplate)
}

func TestLoadNonExistingTempate(t *testing.T) {

	path := "does_not_exist_template.cfg"
	collection := types.NewMasterDetailsCollection()

	_, err := RenderTemplate(path, &collection)

	if err == nil {
		t.Error("Template doesn't exist - this should error")
	}
}
