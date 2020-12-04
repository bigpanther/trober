package main

import (
	"flag"
	"fmt"
	"net/url"

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
	//t, err := template.ParseFiles("model.go.tmpl")
	if err != nil {
		panic(err)
	}
	n := namer.NewPublicPluralNamer(nil)
	for key, _ := range oa.Components.Schemas {
		fmt.Println(key)
		plural := n.Name(&types.Type{Name: types.Name{Name: key}})
		fmt.Println(plural)

		//t.Execute(os.Stdout, key)
	}

}
