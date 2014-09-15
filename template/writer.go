package template

import (
	"bytes"
	"github.com/mdevilliers/redishappy/types"
	"io/ioutil"
	"text/template"
)

type TemplateData struct {
	Clusters *[]types.MasterDetails
}

func RenderTemplate(templatePath string, updates *[]types.MasterDetails) (string, error) {

	templateContent, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New(templatePath).Parse(string(templateContent))

	if err != nil {
		return "", err
	}

	data := TemplateData{Clusters: updates}
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