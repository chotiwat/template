package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"strings"

	"io/ioutil"

	"strconv"

	"runtime"

	"bytes"

	"github.com/blendlabs/template/template"
	"gopkg.in/yaml.v2"
)

var (
	// Version is the app version.
	Version string
)

// Includes are a collection of template files to include as sub templates.
type Includes []string

// Set sets the value.
func (v *Includes) Set(value string) error {
	*v = append(*v, value)
	return nil
}

func (v *Includes) String() string {
	return "Files to include as sub templates"
}

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

// Numbers represent float typed variables.
type Numbers []string

// Set sets a variable.
func (n *Numbers) Set(value string) error {
	*n = append(*n, value)
	return nil
}

func (n *Numbers) String() string {
	return "Number variable values to set in the template"
}

// Values returns the map of values.
func (n *Numbers) Values() (values map[string]interface{}, err error) {
	values = map[string]interface{}{}

	var value float64
	for _, val := range *n {
		pieces := strings.SplitN(val, "=", 2)
		if len(pieces) > 1 {
			value, err = strconv.ParseFloat(pieces[1], 64)
			if err != nil {
				return
			}
			values[pieces[0]] = value
		}
	}
	return
}

func loadVarsFile(path string) (map[string]interface{}, error) {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	output := map[string]interface{}{}
	err = yaml.Unmarshal(contents, &output)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func main() {
	var templateFile string
	flag.StringVar(&templateFile, "f", "", "Template file to process; if \"-\", will read from os.Stdin")

	var varsFile string
	flag.StringVar(&varsFile, "v", "", "Vars file to process")

	var outFile string
	flag.StringVar(&outFile, "o", "", "Output file")

	var variables Variables
	flag.Var(&variables, "var", "Variables in the form --var=foo=bar")

	var numbers Numbers
	flag.Var(&numbers, "num", "Number variables in the form --num=foo=3.14")

	var includes Includes
	flag.Var(&includes, "include", "Files to include as sub templates")

	var help bool
	flag.BoolVar(&help, "help", false, "Shows this usage message")

	var versionFlag bool
	flag.BoolVar(&versionFlag, "version", false, "Shows the app version")

	flag.Usage = func() {
		if len(Version) == 0 {
			Version = "master"
		}
		fmt.Fprintf(os.Stderr, "%s version %s\n\n", os.Args[0], Version)
		fmt.Fprintf(os.Stderr, "Find more information at https://github.com/blendlabs/template\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample Usage:\n")
		fmt.Fprintf(os.Stderr, "Read a template file: \"template -f template.yml\"\n")
		fmt.Fprintf(os.Stderr, "Read a template from stdin: \"echo '{{ .Var \"foo\" }}' | template -f -\"\n")
	}

	flag.Parse()

	if help {
		flag.Usage()
		os.Exit(0)
	}

	if versionFlag {
		if len(Version) == 0 {
			Version = "master"
		}
		fmt.Fprintf(os.Stdout, "%s version %s %s/%s\n", os.Args[0], Version, runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

	var temp *template.Template
	var err error
	if len(templateFile) > 0 && templateFile == "-" {
		temp = template.New()

		var contents []byte
		contents, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
		temp = temp.WithBody(string(contents))
	} else if len(templateFile) > 0 {
		temp, err = template.NewFromFile(templateFile)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		flag.Usage()
		os.Exit(1)
	}

	if len(includes) > 0 {
		for _, include := range includes {
			var contents []byte
			contents, err = ioutil.ReadFile(include)
			if err != nil {
				log.Fatal(err)
			}
			temp = temp.WithInclude(string(contents))
		}
	}

	if len(varsFile) > 0 {
		vars, err := loadVarsFile(varsFile)
		if err != nil {
			log.Fatal(err)
		}
		for key, value := range vars {
			temp = temp.WithVar(key, value)
		}
	}

	vars := variables.Values()
	if len(vars) > 0 {
		for key, value := range vars {
			temp = temp.WithVar(key, value)
		}
	}

	numVars, err := numbers.Values()
	if err != nil {
		log.Fatal(err)
	}
	if len(numVars) > 0 {
		for key, value := range numVars {
			temp = temp.WithVar(key, value)
		}
	}

	buffer := bytes.NewBuffer(nil)
	err = temp.Process(buffer)
	if err != nil {
		log.Fatal(err)
	}

	if len(outFile) > 0 {
		f, err := os.Create(outFile)
		if err != nil {
			log.Fatal(err)
		}
		_, err = buffer.WriteTo(f)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		buffer.WriteTo(os.Stdout)
	}
}
