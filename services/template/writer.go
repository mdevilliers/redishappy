package template

import (
	"bytes"
	"github.com/mdevilliers/redishappy/types"
	"io/ioutil"
	"sort"
	"text/template"
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
