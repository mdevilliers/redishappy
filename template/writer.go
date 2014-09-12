package template

import (
	"bytes"
	"io/ioutil"
	"text/template"
)

type MasterDetails struct {
	ExternalPort int
	Name         string
	Ip           string
	Port         int
}

type TemplateData struct {
	Clusters *[]MasterDetails
}

func RenderTemplate(templatePath string, updates *[]MasterDetails) (string, error) {

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
