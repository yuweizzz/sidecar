package sidecar

import (
	"strings"
	"text/template"
)

func confirmslash(path string) string {
	if string(path[len(path)-1]) != "/" {
		return path + "/"
	}
	return path
}

func RenderTemplateByConfig(path string, config *Config) {
	t, err := template.New("nginx_conf.tpl").Funcs(template.FuncMap{
		"ToLower":      strings.ToLower,
		"ConfirmSlash": confirmslash,
	}).ParseFiles("nginx_conf.tpl")
	if err != nil {
		panic(err)
	}
	fd := CreateFileIfNotExist(path + "/nginx.conf")
	if fd == nil {
		fd = OpenExistFile(path + "/nginx.conf")
		fd.Truncate(0)
	}
	err = t.ExecuteTemplate(fd, "nginx_conf.tpl", *config)
	if err != nil {
		panic(err)
	}
}
