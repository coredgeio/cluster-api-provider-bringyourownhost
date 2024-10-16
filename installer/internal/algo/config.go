package algo

import (
	_ "embed"
	"os"
)

var (
	//go:embed config.toml.tpl
	config string
)

func GetConfigTomlEchoString() string {
	data := MakeRenderData()
	sandboxImg, found := os.LookupEnv("CKP_K8S_SANDBOX_IMAGE")
	if found {
		data.Data["SANDBOX_IMAGE"] = sandboxImg
	} else {
		data.Data["SANDBOX_IMAGE"] = ""
	}
	str, _ := RenderTemplate("config", config, &data)
	return str
}
