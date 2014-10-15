package types

type HAProxy struct {
	TemplatePath  string `json:"templatePath,omitempty"`
	OutputPath    string `json:"outputPath,omitempty"`
	ReloadCommand string `json:"reloadCommand,omitempty"`
}
