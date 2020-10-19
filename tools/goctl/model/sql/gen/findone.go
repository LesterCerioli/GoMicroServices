package gen

import (
	"github.com/tal-tech/go-zero/tools/goctl/model/sql/template"
	"github.com/tal-tech/go-zero/tools/goctl/templatex"
	"github.com/tal-tech/go-zero/tools/goctl/util/stringx"
)

func genFindOne(table Table, withCache bool) (string, error) {
	camel := table.Name.ToCamel()
	text, err := templatex.LoadTemplate(category, findOneTemplateFile, template.FindOne)
	if err != nil {
		return "", err
	}
	output, err := templatex.With("findOne").
		Parse(text).
		Execute(map[string]interface{}{
			"withCache":                 withCache,
			"upperStartCamelObject":     camel,
			"lowerStartCamelObject":     stringx.From(camel).UnTitle(),
			"originalPrimaryKey":        table.PrimaryKey.Name.Source(),
			"lowerStartCamelPrimaryKey": stringx.From(table.PrimaryKey.Name.ToCamel()).UnTitle(),
			"dataType":                  table.PrimaryKey.DataType,
			"cacheKey":                  table.CacheKey[table.PrimaryKey.Name.Source()].KeyExpression,
			"cacheKeyVariable":          table.CacheKey[table.PrimaryKey.Name.Source()].Variable,
		})
	if err != nil {
		return "", err
	}
	return output.String(), nil
}
