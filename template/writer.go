package template

import (
	"bytes"
	//"fmt"
	"io/ioutil"
	"text/template"
	"github.com/mdevilliers/redishappy/configuration"
)

type MasterDetails struct {
	ExternalPort int
	Name string
	Ip string 
	Port int
}

type TemplateData struct{
	Clusters *[]MasterDetails
}

func ExecuteTemplate( haproxy *configuration.HAProxy, updates *[]MasterDetails ) (string,error) {

	templateContent, err := ioutil.ReadFile(haproxy.TemplatePath)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New(haproxy.TemplatePath).Parse(string(templateContent))

	if err != nil {
		return "", err
	}

	data := TemplateData{Clusters : updates}
	strBuffer := new(bytes.Buffer)

	err = tmpl.Execute(strBuffer, data)
	if err != nil {
		return "", err
	}
	return strBuffer.String(), nil
}