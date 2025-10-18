package localizer

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	_ "leblanc.io/open-go-ssl-checker/internal/catalog"
)

type Localizer struct {
	printer *message.Printer
}

var matcher = language.NewMatcher(message.DefaultCatalog.Languages())

func Get(id string) Localizer {
	tag, _ := language.MatchStrings(matcher, "", id)
	return Localizer{message.NewPrinter(tag)}
}

func (l Localizer) Translate(key message.Reference, args ...any) string {
	return l.printer.Sprintf(key, args...)
}
