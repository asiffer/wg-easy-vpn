package main

import (
	"fmt"
	"html/template"
	"os"
	"strings"
	"testing"
)

const (
	titleSize = 80
	titlePad  = "-"
	// docFile   = "wg-easy-vpn.ex"
)

var manPageTemplate = `
.ll 80
.TH wg-easy-vpn 8 "April 14 2020" wg-easy-vpn "Wireguard Easy VPN"

.SH NAME
{{.Name}} \- {{.Usage}}
.SH SYNOPSIS
.B {{.Name}}
[ \fICOMMAND\fP ] [ \fIOPTIONS\fP ] {{raw "<"}}\fI{{.ArgsUsage}}\fP{{raw ">"}}
.BR

.SH DESCRIPTION
.ll 80
{{.Description}}

.SH COMMANDS

.nf
.ta 15 {{range .Commands}}
\fB{{.Name}}\fP 	{{.Usage}} {{end}}

.SH OPTIONS
{{range .Commands}}
\fB{{.Name}}\fP
.nf
.ta 30
.in 10 {{range .Flags}} 
--{{.Name}}{{if .Aliases}}{{range .Aliases}},-{{.}}{{end}}{{end}} {{if .DefaultText}}{{raw "<"}}\fI{{.DefaultText}}\fP{{raw ">"}}{{end}}	{{.Usage}}{{end}}
.in
{{end}}

.SH SEE ALSO
.BR wg (1).
.BR wg-quick (1),

.SH AUTHOR
.B {{.Name}}
was written by {{range .Authors}}
.MT {{.Email}}
{{.Name}} 
.ME
{{end}}

.SH COPYRIGHT
{{.Copyright}}
.

`

func title(s string) {
	length := len(s) + 2
	var leftLength, rightLength int
	if (titleSize-length)%2 == 0 {
		leftLength = (titleSize - length) / 2

	} else {
		leftLength = (titleSize - length - 1) / 2
	}
	rightLength = titleSize - leftLength - length
	fmt.Println(strings.Repeat(titlePad, leftLength),
		s,
		strings.Repeat(titlePad, rightLength))
}

func TestGenDoc(t *testing.T) {
	manFile := "/tmp/wg-easy-vpn.ex"
	f, err := os.Create(manFile)
	if err != nil {
		t.Error(err)
	}
	// if md, err := app.ToMan(); err == nil {
	// 	f.WriteString(md)
	// } else {
	// 	t.Error(err)
	// }
	tp := template.New("man")
	tp.Funcs(template.FuncMap{
		"raw": func(html string) template.HTML {
			return template.HTML(html)
		},
	})
	tp, err = tp.Parse(manPageTemplate)
	if err != nil {
		t.Error(err)
	}
	// remove \n and \t
	app.Description = strings.Replace(app.Description, "\n", "", -1)
	app.Description = strings.Replace(app.Description, "\t", "", -1)
	tp.ExecuteTemplate(f, "man", app)
	f.Close()
}
