package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const secret = `
envFroms:
  env:{{- range $key, $value := .}}
    {{ $key }} : {{ $value }}
{{- end -}}
`

// Load resources to struct
type Config struct {
	EnvFroms map[string]interface{} `yaml:"envFroms" mapstructure:"envFroms"`
}

func main() {

	// recive file path
	flag.String("secrets", "secrets.yaml", "Secrets file need to be base64.")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	i := viper.GetString("secrets")

	// load file
	file, err := ioutil.ReadFile(i)
	if err != nil {
		fmt.Print(err)
	}

	// viper
	viper.SetConfigType("yaml")
	err = viper.ReadConfig(bytes.NewBuffer(file))
	if err != nil {
		log.Fatal("Failed to read viper config", err)
	}

	// parse to struct
	cfg := &Config{}
	err = viper.Unmarshal(&cfg)
	if err != nil {
		panic(err)
	}

	// base64 value
	env := cfg.EnvFroms["env"].(map[string]interface{})
	for key, value := range env {
		env[key] = base64.StdEncoding.EncodeToString([]byte(value.(string)))
	}
	// toUpper
	lf := make(map[string]interface{}, len(env))
	for k, v := range env {
		lf[strings.ToUpper(k)] = v
	}

	// declaire template
	t := template.Must(template.New("secrets").Parse(secret))

	f, err := os.OpenFile(i, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Fatal(err)
	}
	// apply the template to the vars map and write the result to file.
	err = t.Execute(f, lf)
	if err != nil {
		panic(err)
	}
	// closefile
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}
