package sidecar

import (
	"embed"
	"os"
	"strings"
	"text/template"
)

//go:embed nginx_conf.tpl
var tplfile embed.FS

func confirmslash(path string) string {
	if string(path[len(path)-1]) != "/" {
		return path + "/"
	}
	return path
}

func RenderTemplateByConfig(config *Config) {
	pwd, _ := os.Getwd()
	funcmap := template.FuncMap{
		"ToLower":      strings.ToLower,
		"ConfirmSlash": confirmslash,
	}
	tpl := template.Must(template.New("nginx_conf.tpl").Funcs(funcmap).ParseFS(tplfile, "nginx_conf.tpl"))
	fd := CreateFileIfNotExist(pwd + "/nginx.conf")
	if fd == nil {
		fd = OpenExistFile(pwd + "/nginx.conf")
		fd.Truncate(0)
	}
	err := tpl.Execute(fd, *config)
	if err != nil {
		panic(err)
	}
}
