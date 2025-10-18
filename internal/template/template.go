package template

import (
	"embed"
	"html/template"
	"net/http"
	"strings"

	"leblanc.io/open-go-ssl-checker/internal/localizer"
	"leblanc.io/open-go-ssl-checker/templates"
)

var files embed.FS = templates.GetFiles()

var funcs template.FuncMap = template.FuncMap{
	"ToUpper": strings.ToUpper,
	"Defer": func(number *int) int {
		return *number // Dereference pointer to get the actual value of number
	},
}

func getTemplates(file string, acceptLanguage string) *template.Template {
	l := localizer.Get(acceptLanguage)
	funcs["Translate"] = l.Translate
	return template.Must(
		template.New("layout.html").Funcs(funcs).ParseFS(files, "layout.html", file+".html"))
}

func Execute(w http.ResponseWriter, file string, acceptLanguage string, data interface{}) {
	getTemplates(file, acceptLanguage).Execute(w, data)
}
