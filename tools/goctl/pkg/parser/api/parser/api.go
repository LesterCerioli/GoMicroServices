package parser

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/zeromicro/go-zero/core/collection"
	"github.com/zeromicro/go-zero/tools/goctl/pkg/parser/api/ast"
	"github.com/zeromicro/go-zero/tools/goctl/pkg/parser/api/placeholder"
	"github.com/zeromicro/go-zero/tools/goctl/pkg/parser/api/token"
)

type API struct {
	Filename     string
	Syntax       *ast.SyntaxStmt
	info         *ast.InfoStmt    // Info block does not participate in code generation.
	importStmt   []ast.ImportStmt // ImportStmt block does not participate in code generation.
	TypeStmt     []ast.TypeStmt
	ServiceStmts []*ast.ServiceStmt
	importStack  importStack
}

func convert2API(a *ast.AST) (*API, error) {
	if len(a.Stmts) == 0 {
		return nil, fmt.Errorf("syntax error: missing statements")
	}

	var api = new(API)
	api.Filename = a.Filename
	one := a.Stmts[0]
	syntax, ok := one.(*ast.SyntaxStmt)
	if !ok {
		return nil, ast.SyntaxError(one.Pos(), "expected syntax statement, got %T", one)
	}
	api.Syntax = syntax

	for i := 1; i < len(a.Stmts); i++ {
		one := a.Stmts[i]
		switch val := one.(type) {
		case *ast.SyntaxStmt:
			return nil, ast.DuplicateStmtError(val.Pos(), "duplicate syntax statement")
		case *ast.InfoStmt:
			if api.info != nil {
				return nil, ast.DuplicateStmtError(val.Pos(), "duplicate info statement")
			}
			api.info = val
		case ast.ImportStmt:
			api.importStmt = append(api.importStmt, val)
		case ast.TypeStmt:
			api.TypeStmt = append(api.TypeStmt, val)
		case *ast.ServiceStmt:
			api.ServiceStmts = append(api.ServiceStmts, val)
		}
	}

	if err := api.SelfCheck(); err != nil {
		return nil, err
	}
	return api, nil
}

func (api *API) checkImportStmt() error {
	f := newFilter()
	b := f.addCheckItem("import value expression")
	for _, v := range api.importStmt {
		switch val := v.(type) {
		case *ast.ImportLiteralStmt:
			b.check(val.Value)
		case *ast.ImportGroupStmt:
			b.check(val.Values...)
		}
	}
	return f.error()
}

func (api *API) checkInfoStmt() error {
	if api.info == nil {
		return nil
	}
	f := newFilter()
	b := f.addCheckItem("info key expression")
	for _, v := range api.info.Values {
		b.check(v.Key)
	}
	return f.error()
}

func (api *API) checkServiceStmt() error {
	f := newFilter()
	serviceNameChecker := f.addCheckItem("service name expression")
	handlerChecker := f.addCheckItem("handler expression")
	pathChecker := f.addCheckItem("path expression")
	for _, v := range api.ServiceStmts {
		serviceNameChecker.check(token.Token{
			Text:     v.Name.Format(""),
			Position: v.Pos(),
		})
		var group = api.getAtServerValue(v.AtServerStmt, "group")
		for _, item := range v.Routes {
			handlerChecker.check(item.AtHandler.Name)
			path := fmt.Sprintf("[%s]:%s", group, item.Route.Format(""))
			pathChecker.check(token.Token{
				Text:     path,
				Position: item.Route.Pos(),
			})
		}
	}
	return f.error()
}

func (api *API) checkTypeStmt() error {
	f := newFilter()
	b := f.addCheckItem("type expression")
	for _, v := range api.TypeStmt {
		switch val := v.(type) {
		case *ast.TypeLiteralStmt:
			b.check(val.Expr.Name)
		case *ast.TypeGroupStmt:
			for _, expr := range val.ExprList {
				b.check(expr.Name)
			}
		}
	}
	return f.error()
}

func (api *API) checkTypeDeclareContext() error {
	var typeMap = map[string]placeholder.Type{}
	for _, v := range api.TypeStmt {
		switch tp := v.(type) {
		case *ast.TypeLiteralStmt:
			typeMap[tp.Expr.Name.Text] = placeholder.PlaceHolder
		case *ast.TypeGroupStmt:
			for _, v := range tp.ExprList {
				typeMap[v.Name.Text] = placeholder.PlaceHolder
			}
		}
	}

	return api.checkTypeContext(typeMap)
}

func (api *API) checkTypeContext(declareContext map[string]placeholder.Type) error {
	var em = newErrorManager()
	for _, v := range api.TypeStmt {
		switch tp := v.(type) {
		case *ast.TypeLiteralStmt:
			em.add(api.checkTypeExprContext(declareContext, tp.Expr.DataType))
		case *ast.TypeGroupStmt:
			for _, v := range tp.ExprList {
				em.add(api.checkTypeExprContext(declareContext, v.DataType))
			}
		}
	}
	return em.error()
}

func (api *API) checkTypeExprContext(declareContext map[string]placeholder.Type, tp ast.DataType) error {
	switch val := tp.(type) {
	case *ast.ArrayDataType:
		return api.checkTypeExprContext(declareContext, val.DataType)
	case *ast.BaseDataType:
		if IsBaseType(val.Base.Text) {
			return nil
		}
		_, ok := declareContext[val.Base.Text]
		if !ok {
			return ast.SyntaxError(val.Base.Position, "unresolved type <%s>", val.Base.Text)
		}
		return nil
	case *ast.MapDataType:
		var manager = newErrorManager()
		manager.add(api.checkTypeExprContext(declareContext, val.Key))
		manager.add(api.checkTypeExprContext(declareContext, val.Value))
		return manager.error()
	case *ast.PointerDataType:
		return api.checkTypeExprContext(declareContext, val.DataType)
	case *ast.SliceDataType:
		return api.checkTypeExprContext(declareContext, val.DataType)
	case *ast.StructDataType:
		var manager = newErrorManager()
		for _, e := range val.Elements {
			manager.add(api.checkTypeExprContext(declareContext, e.DataType))
		}
		return manager.error()
	}
	return nil
}

func (api *API) getAtServerValue(atServer *ast.AtServerStmt, key string) string {
	if atServer == nil {
		return ""
	}

	for _, val := range atServer.Values {
		if val.Key.Text == key {
			return val.Value.Text
		}
	}

	return ""
}

func (api *API) mergeAPI(in *API) error {
	api.importStack = append(api.importStack, in.importStack...)
	if api.Syntax.Value.Text != in.Syntax.Value.Text {
		return ast.SyntaxError(in.Syntax.Value.Position,
			"multiple syntax value expression, expected <%s>, got <%s>",
			api.Syntax.Value.Text,
			in.Syntax.Value.Text,
		)
	}
	api.TypeStmt = append(api.TypeStmt, in.TypeStmt...)
	api.ServiceStmts = append(api.ServiceStmts, in.ServiceStmts...)
	return nil
}

func (api *API) parseImportedAPI(imports []ast.ImportStmt) ([]*API, error) {
	var list []*API
	if len(imports) == 0 {
		return list, nil
	}

	var importValueSet = collection.NewSet()
	for _, imp := range imports {
		switch val := imp.(type) {
		case *ast.ImportLiteralStmt:
			importValueSet.AddStr(strings.ReplaceAll(val.Value.Text, `"`, ""))
		case *ast.ImportGroupStmt:
			for _, v := range val.Values {
				importValueSet.AddStr(strings.ReplaceAll(v.Text, `"`, ""))
			}
		}
	}

	dir := filepath.Dir(api.Filename)
	for _, impPath := range importValueSet.KeysStr() {
		if !filepath.IsAbs(impPath) {
			impPath = filepath.Join(dir, impPath)
		}
		// import cycle check
		if err := api.importStack.push(impPath); err != nil {
			return nil, err
		}

		p := New(impPath, "", SkipComment)
		ast := p.Parse()
		if err := p.CheckErrors(); err != nil {
			return nil, err
		}

		nestedApi, err := convert2API(ast)
		if err != nil {
			return nil, err
		}

		if err = nestedApi.parseReverse(); err != nil {
			return nil, err
		}

		list = append(list, nestedApi)
		api.importStack.pop()

		if err != nil {
			return nil, err
		}
	}

	return list, nil
}

func (api *API) parseReverse() error {
	list, err := api.parseImportedAPI(api.importStmt)
	if err != nil {
		return err
	}
	for _, e := range list {
		if err = api.mergeAPI(e); err != nil {
			return err
		}
	}
	return nil
}

func (api *API) SelfCheck() error {
	if err := api.parseReverse(); err != nil {
		return err
	}
	if err := api.checkImportStmt(); err != nil {
		return err
	}
	if err := api.checkInfoStmt(); err != nil {
		return err
	}
	if err := api.checkTypeStmt(); err != nil {
		return err
	}
	if err := api.checkServiceStmt(); err != nil {
		return err
	}
	return api.checkTypeDeclareContext()
}