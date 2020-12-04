package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
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
				err = t.ExecuteTemplate(os.Stdout, "model_string.go.tmpl", struct {
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
