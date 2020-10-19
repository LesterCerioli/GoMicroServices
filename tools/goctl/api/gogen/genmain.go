package gogen

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/tal-tech/go-zero/tools/goctl/api/spec"
	"github.com/tal-tech/go-zero/tools/goctl/api/util"
	"github.com/tal-tech/go-zero/tools/goctl/templatex"
	goCtlUtil "github.com/tal-tech/go-zero/tools/goctl/util"
	"github.com/tal-tech/go-zero/tools/goctl/vars"
)

const mainTemplate = `package main

import (
	"flag"
	"fmt"

	{{.importPackages}}
)

var configFile = flag.String("f", "etc/{{.serviceName}}.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
`

func genMain(dir string, api *spec.ApiSpec) error {
	name := strings.ToLower(api.Service.Name)
	if strings.HasSuffix(name, "-api") {
		name = strings.ReplaceAll(name, "-api", "")
	}
	goFile := name + ".go"
	fp, created, err := util.MaybeCreateFile(dir, "", goFile)
	if err != nil {
		return err
	}
	if !created {
		return nil
	}
	
	defer func() {
		if err := fp.Close(); err != nil {
			fmt.Printf("Internal error when closing gernerated main file, filename is: %v err is: %v .",fp.Name(), err)
		}
	}()

	parentPkg, err := getParentPackage(dir)
	if err != nil {
		return err
	}

	text, err := templatex.LoadTemplate(category, mainTemplateFile, mainTemplate)
	if err != nil {
		return err
	}

	t := template.Must(template.New("mainTemplate").Parse(text))
	buffer := new(bytes.Buffer)
	err = t.Execute(buffer, map[string]string{
		"importPackages": genMainImports(parentPkg),
		"serviceName":    api.Service.Name,
	})
	if err != nil {
		return nil
	}
	formatCode := formatCode(buffer.String())
	_, err = fp.WriteString(formatCode)
	return err
}

func genMainImports(parentPkg string) string {
	var imports []string
	imports = append(imports, fmt.Sprintf("\"%s\"", goCtlUtil.JoinPackages(parentPkg, configDir)))
	imports = append(imports, fmt.Sprintf("\"%s\"", goCtlUtil.JoinPackages(parentPkg, handlerDir)))
	imports = append(imports, fmt.Sprintf("\"%s\"\n", goCtlUtil.JoinPackages(parentPkg, contextDir)))
	imports = append(imports, fmt.Sprintf("\"%s/core/conf\"", vars.ProjectOpenSourceUrl))
	imports = append(imports, fmt.Sprintf("\"%s/rest\"", vars.ProjectOpenSourceUrl))
	return strings.Join(imports, "\n\t")
}
