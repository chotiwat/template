template
========

Template is a thin wrapper on `text/template`, the golang templating engine. The primary usecase for the utility is to dynamically modify config templates.

## Usage

Typical usage is to read a file and apply a couple variables.

Example Template (`ingress.yml.template`):

```yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: {{ .Var "service" }}
  ref: {{ .Env "CURRENT_REF" }}
```

If we then run the following:

```bash
> CURRENT_REF="abcdef" template -f ingress.yml.template --var service="my service"
```

`template` will then print to the screen the updated template:

```yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: my service
  ref: abcdef
```

You can alternately read from `stdin` by omitting the `-f` flag.

## Commandline Flag Reference

### `-f <TEMPLATE PATH>`

The `-f` flag specifies an input file. If it is not present, `template` will read from `stdin`.

### `-var <KEY>=<VALUE>`

The `-var` flag specifies a variable for the template.

## Template Helper Reference

### `.Var`

Var will return a variable as set by the commandline. It takes the variable name as the first parameter. It can take a default value as a second parameter. If no default is specified, and the variable is not present, this will cause an error.

```go
{{ .Var "<var name>" }}
```

With a default:

```go
{{ .Var "<var name>" "<default value>" }}
```

### `.Env`

Env will return an environment variable. It takes the environment variable name as the first parameter. It can take a default value as a second parameter. If no default is specified, and the environment variable is not present, this will cause an error.

```go
{{ .Env "<var name>" }}
```

With a default:

```go
{{ .Env "<var name>" "<default value>" }}
```
