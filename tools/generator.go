package main

import (
	"flag"
	"fmt"
	"net/url"

	"github.com/getkin/kin-openapi/openapi3"
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
	for _, c := range oa.Components.Schemas {
		fmt.Println(c)
	}

}
