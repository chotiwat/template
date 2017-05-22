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

	"github.com/blendlabs/template/template"
	"gopkg.in/yaml.v2"
)

var (
	// Version is the app version.
	Version string
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
	flag.StringVar(&templateFile, "f", "", "Template file to process; if unset, will read from <stdin>")

	var varsFile string
	flag.StringVar(&varsFile, "v", "", "Vars file to process")

	var variables Variables
	flag.Var(&variables, "var", "Variables in the form --var=foo=bar")

	var numbers Numbers
	flag.Var(&numbers, "num", "Number variables in the form --num=foo=3.14")

	var help bool
	flag.BoolVar(&help, "help", false, "Shows this usage message")

	var versionFlag bool
	flag.BoolVar(&versionFlag, "version", false, "Shows the app version")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s version %s\n", os.Args[0], Version)
		fmt.Fprintln(os.Stderr, "usage:")
		flag.PrintDefaults()
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
	if len(templateFile) > 0 {
		temp, err = template.NewFromFile(templateFile)
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

	err = temp.Process(os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
}
