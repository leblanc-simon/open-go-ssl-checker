package templates

import (
	"embed"
)

//go:embed *.html
var files embed.FS

func GetFiles() embed.FS {
	return files
}
