package main

import (
	"flag"
	"io"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/Masterminds/sprig"
)

type TemplateData struct {
}

func main() {
	var writer io.Writer = os.Stdout

	inputFile := flag.String("template", "-", "Template file")
	outputFile := flag.String("output", "", "File to output template to")

	flag.Parse()

	if *outputFile != "" {
		wr, err := os.Create(*outputFile)

		if err != nil {
			panic(err)
		}

		defer func() {
			if err := wr.Close(); err != nil {
				panic(err)
			}
		}()

		writer = wr
	}

	var name string = "main"
	var err error

	template := template.New(name).
		Funcs(sprig.TxtFuncMap()).
		Funcs(getSSMFuncMap())

	if *inputFile == "-" {
		data, err := ioutil.ReadAll(os.Stdin)

		if err != nil {
			panic(err)
		}

		template, err = template.Parse(string(data))

		if err != nil {
			panic(err)
		}
	} else {
		template, err = template.ParseFiles(*inputFile)

		if err != nil {
			panic(err)
		}

		name = *inputFile
	}

	data := TemplateData{}
	err = template.ExecuteTemplate(writer, name, &data)

	if err != nil {
		panic(err)
	}
}
