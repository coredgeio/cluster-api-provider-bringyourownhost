package algo

import (
	"bytes"
	"text/template"

	"github.com/Masterminds/sprig"
)

type RenderData struct {
	Funcs template.FuncMap
	Data  map[string]interface{}
}

func MakeRenderData() RenderData {
	return RenderData{
		Funcs: template.FuncMap{},
		Data:  map[string]interface{}{},
	}
}

// RenderTemplate reads, renders, and attempts to parse a yaml or
// json file representing one or more k8s api objects
func RenderTemplate(name, raw string, d *RenderData) (string, error) {
	rendered, err := renderTemplateInternal(name, raw, d)
	if err != nil {
		return "", err
	}
	return rendered.String(), nil
}

func renderTemplateInternal(name, raw string, d *RenderData) (*bytes.Buffer, error) {
	tmpl := template.New(name).Option("missingkey=error")
	if d.Funcs != nil {
		tmpl.Funcs(d.Funcs)
	}

	// Add universal functions
	tmpl.Funcs(template.FuncMap{"getOr": getOr, "isSet": isSet})
	tmpl.Funcs(sprig.TxtFuncMap())

	if _, err := tmpl.Parse(raw); err != nil {
		return nil, err
	}

	rendered := &bytes.Buffer{}
	if err := tmpl.Execute(rendered, d.Data); err != nil {
		return nil, err
	}
	return rendered, nil
}

func RenderTemplateToBytes(name, raw string, d *RenderData) ([]byte, error) {
	rendered, err := renderTemplateInternal(name, raw, d)
	if err != nil {
		return nil, err
	}
	return rendered.Bytes(), nil
}
