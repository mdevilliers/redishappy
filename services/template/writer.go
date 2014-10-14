package template

import (
	"bytes"
	"io/ioutil"
	"sort"
	"text/template"

	"github.com/mdevilliers/redishappy/types"
)

type TemplateData struct {
	Clusters []*types.MasterDetails
}

func RenderTemplate(templatePath string, updates *types.MasterDetailsCollection) (string, error) {

	templateContent, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New(templatePath).Parse(string(templateContent))

	if err != nil {
		return "", err
	}

	sortedUpdates := updates.Items()
	sort.Sort(types.ByName(sortedUpdates))

	data := TemplateData{Clusters: sortedUpdates}
	strBuffer := new(bytes.Buffer)

	err = tmpl.Execute(strBuffer, data)
	if err != nil {
		return "", err
	}
	return strBuffer.String(), nil
}

func WriteFile(outputFilePath string, content string) error {
	return ioutil.WriteFile(outputFilePath, []byte(content), 0666)
}
