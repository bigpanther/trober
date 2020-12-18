package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
	"k8s.io/gengo/namer"
	"k8s.io/gengo/types"
)

func main() {
	var openAPIFile string
	flag.StringVar(&openAPIFile, "f", "openapi.yaml", "Specify the filename containing the spec")
	flag.Parse()
	loader := openapi3.NewSwaggerLoader()
	loader.IsExternalRefsAllowed = true

	u, err := url.Parse(openAPIFile)
	var oa *openapi3.Swagger
	if err == nil && u.Scheme != "" && u.Host != "" {
		oa, err = loader.LoadSwaggerFromURI(u)
	} else {
		oa, err = loader.LoadSwaggerFromFile(openAPIFile)
	}
	if err != nil {
		panic(err)
	}
	funcMap := template.FuncMap{
		"ToLower": strings.ToLower,
	}
	// funcMap := template.FuncMap{
	// 	"ToValidate": func(s []string) string {
	// 		var val = ""
	// 		for i, v := range s {
	// 			s[i] = fmt.Sprintf("s == %s", v)
	// 			join
	// 		}
	// 	},
	//}
	t := template.New("").Funcs(funcMap)
	t, err = t.ParseFiles("model_object.go.tmpl", "model_array.go.tmpl", "model_string.go.tmpl")
	if err != nil {
		panic(err)
	}
	n := namer.NewPublicPluralNamer(nil)
	for key, c := range oa.Components.Schemas {
		plural := n.Name(&types.Type{Name: types.Name{Name: key}})
		if c.Ref == "" {
			if c.Value.Type == "string" {
				fmt.Println(key, c.Value.Type)
				f, err := os.OpenFile(fmt.Sprintf("../models/%s.go", ToSnakeCase(key)), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
				if err != nil {
					log.Fatal(err)
				}
				err = f.Truncate(0)
				if err != nil {
					log.Fatal(err)
				}
				err = t.ExecuteTemplate(f, "model_string.go.tmpl", struct {
					Key       string
					PluralKey string
					Schema    *openapi3.Schema
				}{Key: key, PluralKey: plural, Schema: c.Value})
				if err != nil {
					panic(err)
				}
			}
			continue
			err = t.Execute(os.Stdout, struct {
				Key       string
				PluralKey string
				Schema    *openapi3.Schema
			}{Key: key, PluralKey: plural, Schema: c.Value})
			if err != nil {
				panic(err)
			}
		} else {
			fmt.Println(key)
		}
	}

}

var matchFirstCap = regexp.MustCompile("([A-Z])([A-Z][a-z])")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

// ToSnakeCase converts the provided string to snake_case.
// Based on https://gist.github.com/stoewer/fbe273b711e6a06315d19552dd4d33e6
func ToSnakeCase(input string) string {
	output := matchFirstCap.ReplaceAllString(input, "${1}_${2}")
	output = matchAllCap.ReplaceAllString(output, "${1}_${2}")
	output = strings.ReplaceAll(output, "-", "_")
	return strings.ToLower(output)
}
