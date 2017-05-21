package main

import (
	"flag"
	"log"
	"os"

	"strings"

	"io/ioutil"

	"github.com/blendlabs/template/template"
)

// Variables are a list of commandline variables.
type Variables []string

// Set sets a variable.
func (v *Variables) Set(value string) error {
	*v = append(*v, value)
	return nil
}

func (v *Variables) String() string {
	return "Variable values to set in the template"
}

// Values returns the map of values.
func (v *Variables) Values() (values map[string]string) {
	values = map[string]string{}

	for _, val := range *v {
		pieces := strings.SplitN(val, "=", 2)
		if len(pieces) > 1 {
			values[pieces[0]] = pieces[1]
		}
	}
	return
}

func main() {
	var variables Variables
	flag.Var(&variables, "var", "Variables in the form --var=foo=bar")

	var file string
	flag.StringVar(&file, "f", "", "The file to process")

	var help bool
	flag.BoolVar(&help, "help", false, "Shows this usage message")

	flag.Parse()

	if help {
		flag.Usage()
		os.Exit(0)
	}

	var temp *template.Template
	var err error
	if len(file) > 0 {
		temp, err = template.NewFromFile(file)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		temp = template.New()

		var contents []byte
		contents, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
		temp = temp.WithBody(string(contents))
	}

	vars := variables.Values()
	if len(vars) > 0 {
		for key, value := range vars {
			temp = temp.WithVar(key, value)
		}
	}

	err = temp.Process(os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
}
