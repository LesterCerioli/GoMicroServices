package gogen

import (
	"bytes"
	"fmt"
	"strconv"
	"text/template"

	"github.com/tal-tech/go-zero/tools/goctl/api/spec"
	"github.com/tal-tech/go-zero/tools/goctl/api/util"
	"github.com/tal-tech/go-zero/tools/goctl/templatex"
)

const (
	defaultPort = 8888
	etcDir      = "etc"
	etcTemplate = `Name: {{.serviceName}}
Host: {{.host}}
Port: {{.port}}
`
)

func genEtc(dir string, api *spec.ApiSpec) error {
	fp, created, err := util.MaybeCreateFile(dir, etcDir, fmt.Sprintf("%s.yaml", api.Service.Name))
	if err != nil {
		return err
	}
	if !created {
		return nil
	}
	
	defer func() {
		if err := fp.Close(); err != nil {
			fmt.Printf("Internal error when closing gernerated etc file, filename is: %v err is: %v .",fp.Name(), err)
		}
	}()

	service := api.Service
	host, ok := util.GetAnnotationValue(service.Annotations, "server", "host")
	if !ok {
		host = "0.0.0.0"
	}
	port, ok := util.GetAnnotationValue(service.Annotations, "server", "port")
	if !ok {
		port = strconv.Itoa(defaultPort)
	}

	text, err := templatex.LoadTemplate(category, etcTemplateFile, etcTemplate)
	if err != nil {
		return err
	}

	t := template.Must(template.New("etcTemplate").Parse(text))
	buffer := new(bytes.Buffer)
	err = t.Execute(buffer, map[string]string{
		"serviceName": service.Name,
		"host":        host,
		"port":        port,
	})
	if err != nil {
		return err
	}
	formatCode := formatCode(buffer.String())
	_, err = fp.WriteString(formatCode)
	return err
}
