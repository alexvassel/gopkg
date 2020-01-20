package types

import "strings"

var nl2brReplacer = strings.NewReplacer("\n", "<br/>")

func NlToBr(txt string) string {
	return nl2brReplacer.Replace(txt)
}
