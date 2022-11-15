package parser

import (
	"fmt"
	"log"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/zeromicro/go-zero/tools/goctl/pkg/parser/api/ast"
	"github.com/zeromicro/go-zero/tools/goctl/pkg/parser/api/scanner"
	"github.com/zeromicro/go-zero/tools/goctl/pkg/parser/api/token"
)

const IDAPI = "api"

type Mode int

type Parser struct {
	filename string
	s        *scanner.Scanner
	errors   []error

	curTok  token.Token
	peekTok token.Token
	nodeSet *ast.NodeSet
	parsed  bool

	tokenNode map[token.Token]*ast.TokenNode
}

func New(nodeSet *ast.NodeSet, filename string, src interface{}) *Parser {
	abs, err := filepath.Abs(filename)
	if err != nil {
		log.Fatalln(err)
	}

	p := &Parser{
		filename:  abs,
		s:         scanner.MustNewScanner(abs, src),
		nodeSet:   nodeSet,
		tokenNode: make(map[token.Token]*ast.TokenNode),
	}

	return p
}

func (p *Parser) Parse() *ast.AST {
	api := &ast.AST{
		Filename: p.filename,
	}
	if !p.init() {
		return nil
	}

	for p.curTokenIsNotEof() {
		stmt := p.parseStmt()
		if isNil(stmt) {
			return nil
		}

		api.Stmts = append(api.Stmts, stmt)
		if !p.nextToken() {
			return nil
		}
	}

	p.parsed = true
	return api
}

func (p *Parser) parseStmt() ast.Stmt {
	switch p.curTok.Type {
	case token.IDENT:
		switch {
		case p.curTok.Is(token.Syntax):
			return p.parseSyntaxStmt()
		case p.curTok.Is(token.Info):
			return p.parseInfoStmt()
		case p.curTok.Is(token.Service):
			return p.parseService()
		default:
			p.expectIdentError(p.curTok, token.Syntax, token.Info, token.Service, token.TYPE)
			return nil
		}
	case token.IMPORT:
		return p.parseImportStmt()
	case token.TYPE:
		return p.parseTypeStmt()
	case token.AT_SERVER:
		return p.parseService()
	default:
		p.errors = append(p.errors, fmt.Errorf("%s unexpected token '%s'", p.curTok.Position.String(), p.peekTok.Text))
		return nil
	}
}

func (p *Parser) parseService() *ast.ServiceStmt {
	var stmt = &ast.ServiceStmt{}
	if p.curTokenIs(token.AT_SERVER) {
		atServerStmt := p.parseAtServerStmt()
		if atServerStmt == nil {
			return nil
		}

		stmt.AtServerStmt = atServerStmt
		if !p.advanceIfPeekTokenIs(token.Service) {
			return nil
		}
	}
	stmt.Service = p.curTok

	if !p.advanceIfPeekTokenIs(token.IDENT) {
		return nil
	}

	// service name expr
	nameExpr := p.parseServiceNameExpr()
	if nameExpr == nil {
		return nil
	}
	stmt.Name = nameExpr

	// token '{'
	if !p.advanceIfPeekTokenIs(token.LBRACE) {
		return nil
	}
	stmt.LBrace = p.curTok

	// service item statements
	routes := p.parseServiceItemsStmt()
	if routes == nil {
		return nil
	}
	stmt.Routes = routes

	// token '}'
	if !p.advanceIfPeekTokenIs(token.RBRACE) {
		return nil
	}
	stmt.RBrace = p.curTok

	return stmt
}

func (p *Parser) parseServiceItemsStmt() []*ast.ServiceItemStmt {
	var stmt []*ast.ServiceItemStmt
	for p.curTokenIsNotEof() && p.peekTokenIsNot(token.RBRACE) {
		item := p.parseServiceItemStmt()
		if item == nil {
			return nil
		}

		stmt = append(stmt, item)
		if p.peekTokenIs(token.RBRACE) {
			break
		}
		if p.notExpectPeekToken(token.AT_DOC, token.AT_HANDLER, token.RBRACE) {
			return nil
		}
	}

	return stmt
}

func (p *Parser) parseServiceItemStmt() *ast.ServiceItemStmt {
	var stmt = &ast.ServiceItemStmt{}
	// statement @doc
	if p.peekTokenIs(token.AT_DOC) {
		if !p.nextToken() {
			return nil
		}
		atDocStmt := p.parseAtDocStmt()
		if atDocStmt == nil {
			return nil
		}
		stmt.AtDoc = atDocStmt
	}

	// statement @handler
	if !p.advanceIfPeekTokenIs(token.AT_HANDLER, token.RBRACE) {
		return nil
	}
	if p.peekTokenIs(token.RBRACE) {
		return stmt
	}
	atHandlerStmt := p.parseAtHandlerStmt()
	if atHandlerStmt == nil {
		return nil
	}
	stmt.AtHandler = atHandlerStmt

	// statement route
	route := p.parseRouteStmt()
	if route == nil {
		return nil
	}
	stmt.Route = route

	p.nodeSet.Append(stmt)
	return stmt
}

func (p *Parser) parseRouteStmt() *ast.RouteStmt {
	var stmt = &ast.RouteStmt{}
	// token http method
	if !p.advanceIfPeekTokenIs(token.HttpMethods...) {
		return nil
	}
	stmt.Method = p.curTok

	// path expression
	pathExpr := p.parsePathExpr()
	if pathExpr == nil {
		return nil
	}
	stmt.Path = pathExpr

	if p.peekTokenIs(token.AT_DOC, token.AT_HANDLER, token.RBRACE) {
		return stmt
	}

	if p.notExpectPeekToken(token.Returns, token.LPAREN) {
		return nil
	}

	if p.peekTokenIs(token.LPAREN) {
		// request expression
		requestBodyStmt := p.parseBodyStmt()
		if requestBodyStmt == nil {
			return nil
		}
		stmt.Request = requestBodyStmt
	}

	if p.notExpectPeekToken(token.Returns, token.AT_DOC, token.AT_HANDLER, token.RBRACE) {
		return nil
	}

	// token 'returns'
	if p.peekTokenIs(token.Returns) {
		if !p.nextToken() {
			return nil
		}
		stmt.Returns = p.curTok

		responseBodyStmt := p.parseBodyStmt()
		if responseBodyStmt == nil {
			return nil
		}

		stmt.Response = responseBodyStmt
	}

	p.nodeSet.Append(stmt)
	return stmt
}

func (p *Parser) parseBodyStmt() *ast.BodyStmt {
	var stmt = &ast.BodyStmt{}
	// token '('
	if !p.advanceIfPeekTokenIs(token.LPAREN) {
		return nil
	}
	stmt.LParen = p.curTok

	expr := p.parseBodyExpr()
	if expr == nil {
		return nil
	}
	stmt.Body = expr

	// token ')'
	if !p.advanceIfPeekTokenIs(token.RPAREN) {
		return nil
	}
	stmt.RParen = p.curTok

	p.nodeSet.Append(stmt)
	return stmt
}

func (p *Parser) parseBodyExpr() *ast.BodyExpr {
	var expr = &ast.BodyExpr{}
	switch {
	case p.peekTokenIs(token.LBRACK): // token '['
		if !p.nextToken() {
			return nil
		}
		expr.LBrack = p.curTok

		// token ']'
		if !p.advanceIfPeekTokenIs(token.RBRACK) {
			return nil
		}
		expr.RBrack = p.curTok

		switch {
		case p.peekTokenIs(token.MUL):
			if !p.nextToken() {
				return nil
			}
			expr.Star = p.curTok
			if !p.advanceIfPeekTokenIs(token.IDENT) {
				return nil
			}
			expr.Value = p.curTok
			return expr
		case p.peekTokenIs(token.IDENT):
			if !p.nextToken() {
				return nil
			}
			expr.Value = p.curTok
			return expr
		default:
			p.expectPeekToken(token.MUL, token.IDENT)
			return nil
		}
	case p.peekTokenIs(token.MUL):
		if !p.nextToken() {
			return nil
		}
		expr.Star = p.curTok
		if !p.advanceIfPeekTokenIs(token.IDENT) {
			return nil
		}
		expr.Value = p.curTok
		return expr
	case p.peekTokenIs(token.IDENT):
		if !p.nextToken() {
			return nil
		}
		expr.Value = p.curTok
		return expr
	default:
		p.expectPeekToken(token.LBRACK, token.MUL, token.IDENT)
		return nil
	}
}

func (p *Parser) parsePathExpr() *ast.PathExpr {
	var expr = &ast.PathExpr{}

	for p.curTokenIsNotEof() &&
		p.peekTokenIsNot(token.LPAREN, token.Returns, token.AT_DOC, token.AT_HANDLER, token.RBRACE) {
		// token '/'
		if !p.advanceIfPeekTokenIs(token.QUO) {
			return nil
		}
		expr.Values = append(expr.Values, p.curTok)

		// token ':' or IDENT
		if p.notExpectPeekToken(token.COLON, token.IDENT) {
			return nil
		}

		// token ':'
		if p.peekTokenIs(token.COLON) {
			if !p.nextToken() {
				return nil
			}
			expr.Values = append(expr.Values, p.curTok)
		}

		// path id tokens
		pathTokens := p.parsePathItem()
		if pathTokens == nil {
			return nil
		}
		expr.Values = append(expr.Values, pathTokens...)

		if p.notExpectPeekToken(token.QUO, token.LPAREN, token.Returns, token.AT_DOC, token.AT_HANDLER, token.RBRACE) {
			return nil
		}
	}

	p.nodeSet.Append(expr)
	return expr
}

func (p *Parser) parsePathItem() []token.Token {
	var list []token.Token
	if !p.advanceIfPeekTokenIs(token.IDENT) {
		return nil
	}
	list = append(list, p.curTok)

	for p.curTokenIsNotEof() &&
		p.peekTokenIsNot(token.QUO, token.LPAREN, token.Returns, token.AT_DOC, token.AT_HANDLER, token.RBRACE, token.EOF) {
		if p.peekTokenIs(token.SUB) {
			if !p.nextToken() {
				return nil
			}
			list = append(list, p.curTok)

			if !p.advanceIfPeekTokenIs(token.IDENT) {
				return nil
			}
			list = append(list, p.curTok)
		} else {
			if p.peekTokenIs(token.LPAREN, token.Returns, token.AT_DOC, token.AT_HANDLER, token.RBRACE) {
				return list
			}
			if !p.advanceIfPeekTokenIs(token.IDENT) {
				return nil
			}
			list = append(list, p.curTok)
		}
	}

	return list
}

func (p *Parser) parseServiceNameExpr() *ast.ServiceNameExpr {
	var expr = &ast.ServiceNameExpr{}
	expr.ID = p.curTok

	if p.peekTokenIs(token.SUB) {
		if !p.nextToken() {
			return nil
		}
		expr.Joiner = p.curTok

		if !p.expectPeekToken(IDAPI) {
			return nil
		}
		if !p.nextToken() {
			return nil
		}
		expr.API = p.curTok
	}

	p.nodeSet.Append(expr)
	return expr
}

func (p *Parser) parseAtDocStmt() ast.AtDocStmt {
	if p.notExpectPeekToken(token.LPAREN, token.STRING) {
		return nil
	}
	if p.peekTokenIs(token.LPAREN) {
		return p.parseAtDocGroupStmt()
	}
	return p.parseAtDocLiteralStmt()
}

func (p *Parser) parseAtDocGroupStmt() ast.AtDocStmt {
	var stmt = &ast.AtDocGroupStmt{}
	stmt.AtDoc = p.curTok

	// token '('
	if !p.advanceIfPeekTokenIs(token.LPAREN) {
		return nil
	}
	stmt.LParen = p.curTok

	for p.curTokenIsNotEof() && p.peekTokenIsNot(token.RPAREN) {
		expr := p.parseKVExpression()
		if expr == nil {
			return nil
		}

		stmt.Values = append(stmt.Values, expr)
		if p.notExpectPeekToken(token.RPAREN, token.KEY) {
			return nil
		}
	}

	// token ')'
	if !p.advanceIfPeekTokenIs(token.RPAREN) {
		return nil
	}
	stmt.RParen = p.curTok

	p.nodeSet.Append(stmt)
	return stmt
}

func (p *Parser) parseAtDocLiteralStmt() ast.AtDocStmt {
	var stmt = &ast.AtDocLiteralStmt{}
	stmt.AtDoc = p.curTok

	if !p.advanceIfPeekTokenIs(token.STRING) {
		return nil
	}
	stmt.Value = p.curTok

	p.nodeSet.Append(stmt)
	return stmt
}

func (p *Parser) parseAtHandlerStmt() *ast.AtHandlerStmt {
	var stmt = &ast.AtHandlerStmt{}
	stmt.AtHandler = p.getNode(p.curTok)

	// token IDENT
	if !p.advanceIfPeekTokenIs(token.IDENT) {
		return nil
	}
	stmt.Name = p.getNode(p.curTok)

	p.nodeSet.InsertAfter(stmt, stmt.Name)
	return stmt
}

func (p *Parser) parseAtServerStmt() *ast.AtServerStmt {
	var stmt = &ast.AtServerStmt{}
	stmt.AtServer = p.curTok

	// token '('
	if !p.advanceIfPeekTokenIs(token.LPAREN) {
		return nil
	}
	stmt.LParen = p.curTok

	for p.curTokenIsNotEof() && p.peekTokenIsNot(token.RPAREN) {
		expr := p.parseAtServerKVExpression()
		if expr == nil {
			return nil
		}

		stmt.Values = append(stmt.Values, expr)
		if p.notExpectPeekToken(token.RPAREN, token.KEY) {
			return nil
		}
	}

	// token ')'
	if !p.advanceIfPeekTokenIs(token.RPAREN) {
		return nil
	}
	stmt.RParen = p.curTok

	p.nodeSet.Append(stmt)
	return stmt
}

func (p *Parser) parseTypeStmt() ast.TypeStmt {
	switch {
	case p.peekTokenIs(token.LPAREN):
		return p.parseTypeGroupStmt()
	case p.peekTokenIs(token.IDENT):
		return p.parseTypeLiteralStmt()
	default:
		p.expectPeekToken(token.LPAREN, token.IDENT)
		return nil
	}
}

func (p *Parser) parseTypeLiteralStmt() ast.TypeStmt {
	var stmt = &ast.TypeLiteralStmt{}
	stmt.Type = p.curTok

	expr := p.parseTypeExpr()
	if expr == nil {
		return nil
	}
	stmt.Expr = expr

	p.nodeSet.Append(stmt)
	return stmt
}

func (p *Parser) parseTypeGroupStmt() ast.TypeStmt {
	var stmt = &ast.TypeGroupStmt{}
	stmt.Type = p.curTok

	// token '('
	if !p.nextToken() {
		return nil
	}
	stmt.LParen = p.curTok

	exprList := p.parseTypeExprList()
	if exprList == nil {
		return nil
	}
	stmt.ExprList = exprList

	// token ')'
	if !p.advanceIfPeekTokenIs(token.RPAREN) {
		return nil
	}
	stmt.RParen = p.curTok

	p.nodeSet.Append(stmt)
	return stmt
}

func (p *Parser) parseTypeExprList() []*ast.TypeExpr {
	if !p.expectPeekToken(token.IDENT, token.RPAREN) {
		return nil
	}
	var exprList = make([]*ast.TypeExpr, 0)
	for p.curTokenIsNotEof() && p.peekTokenIsNot(token.RPAREN, token.EOF) {
		expr := p.parseTypeExpr()
		if expr == nil {
			return nil
		}
		exprList = append(exprList, expr)
		if !p.expectPeekToken(token.IDENT, token.RPAREN) {
			return nil
		}
	}
	return exprList
}

func (p *Parser) parseTypeExpr() *ast.TypeExpr {
	var expr = &ast.TypeExpr{}
	// token IDENT
	if !p.advanceIfPeekTokenIs(token.IDENT) {
		return nil
	}
	expr.Name = p.getNode(p.curTok)

	// token '='
	if p.peekTokenIs(token.ASSIGN) {
		if !p.nextToken() {
			return nil
		}
		expr.Assign = p.getNode(p.curTok)
	}

	dt := p.parseDataType()
	if isNil(dt) {
		return nil
	}
	expr.DataType = dt

	p.nodeSet.InsertAfter(expr, expr.DataType)
	return expr
}

func (p *Parser) parseDataType() ast.DataType {
	switch {
	case p.peekTokenIs(token.Any):
		return p.parseAnyDataType()
	case p.peekTokenIs(token.LBRACE):
		return p.parseStructDataType()
	case p.peekTokenIs(token.IDENT):
		if !p.nextToken() {
			return nil
		}
		return &ast.BaseDataType{Base: p.curTok}
	case p.peekTokenIs(token.MAP):
		return p.parseMapDataType()
	case p.peekTokenIs(token.LBRACK):
		if !p.nextToken() {
			return nil
		}
		switch {
		case p.peekTokenIs(token.RBRACK):
			return p.parseSliceDataType()
		case p.peekTokenIs(token.INT, token.ELLIPSIS):
			return p.parseArrayDataType()
		default:
			p.expectPeekToken(token.RBRACK, token.INT, token.ELLIPSIS)
			return nil
		}
	case p.peekTokenIs(token.ANY):
		return p.parseInterfaceDataType()
	case p.peekTokenIs(token.MUL):
		return p.parsePointerDataType()
	default:
		p.expectPeekToken(token.IDENT, token.LBRACK, token.MAP, token.ANY, token.MUL, token.LBRACE)
		return nil
	}
}
func (p *Parser) parseStructDataType() *ast.StructDataType {
	var tp = &ast.StructDataType{}
	if !p.nextToken() {
		return nil
	}
	tp.LBrace = p.curTok

	if p.notExpectPeekToken(token.IDENT, token.RBRACE) {
		return nil
	}
	// ElemExprList
	elems := p.parseElemExprList()
	if elems == nil {
		return nil
	}
	tp.Elements = elems

	if !p.advanceIfPeekTokenIs(token.RBRACE) {
		return nil
	}
	tp.RBrace = p.curTok

	p.nodeSet.Append(tp)
	return tp
}

func (p *Parser) parseElemExprList() ast.ElemExprList {
	var list = make(ast.ElemExprList, 0)
	for p.curTokenIsNotEof() && p.peekTokenIsNot(token.RBRACE, token.EOF) {
		if p.notExpectPeekToken(token.IDENT, token.RBRACE) {
			return nil
		}
		expr := p.parseElemExpr()
		if expr == nil {
			return nil
		}
		list = append(list, expr)
		if p.notExpectPeekToken(token.IDENT, token.RBRACE) {
			return nil
		}
	}

	return list
}

func (p *Parser) parseElemExpr() *ast.ElemExpr {
	var expr = &ast.ElemExpr{}
	if !p.advanceIfPeekTokenIs(token.IDENT) {
		return nil
	}
	expr.Name = append(expr.Name, p.curTok)

	if p.notExpectPeekToken(token.COMMA, token.IDENT, token.LBRACK, token.MAP, token.ANY, token.MUL, token.LBRACE) {
		return nil
	}

	for p.peekTokenIs(token.COMMA) {
		if !p.nextToken() {
			return nil
		}
		if !p.advanceIfPeekTokenIs(token.IDENT) {
			return nil
		}
		expr.Name = append(expr.Name, p.curTok)
	}

	dt := p.parseDataType()
	if isNil(dt) {
		return nil
	}
	expr.DataType = dt

	if p.notExpectPeekToken(token.RAW_STRING, token.IDENT, token.RBRACE) {
		return nil
	}

	if p.peekTokenIs(token.RAW_STRING) {
		if !p.nextToken() {
			return nil
		}
		expr.Tag = p.curTok
	}

	p.nodeSet.Append(expr)
	return expr
}

func (p *Parser) parseAnyDataType() *ast.AnyDataType {
	var tp = &ast.AnyDataType{}
	if !p.nextToken() {
		return nil
	}
	tp.Any = p.curTok
	p.nodeSet.Append(tp)
	return tp
}

func (p *Parser) parsePointerDataType() *ast.PointerDataType {
	var tp = &ast.PointerDataType{}
	if !p.nextToken() {
		return nil
	}
	tp.Star = p.curTok

	if p.notExpectPeekToken(token.IDENT, token.LBRACK, token.MAP, token.ANY, token.MUL) {
		return nil
	}
	// DataType
	dt := p.parseDataType()
	if isNil(dt) {
		return nil
	}
	tp.DataType = dt

	p.nodeSet.Append(tp)
	return tp
}

func (p *Parser) parseInterfaceDataType() *ast.InterfaceDataType {
	var tp = &ast.InterfaceDataType{}
	if !p.nextToken() {
		return nil
	}
	tp.Interface = p.curTok
	p.nodeSet.Append(tp)
	return tp
}

func (p *Parser) parseMapDataType() *ast.MapDataType {
	var tp = &ast.MapDataType{}
	if !p.nextToken() {
		return nil
	}
	tp.Map = p.curTok

	// token '['
	if !p.advanceIfPeekTokenIs(token.LBRACK) {
		return nil
	}
	tp.LBrack = p.curTok

	// DataType
	dt := p.parseDataType()
	if isNil(dt) {
		return nil
	}
	tp.Key = dt

	// token  ']'
	if !p.advanceIfPeekTokenIs(token.RBRACK) {
		return nil
	}
	tp.RBrack = p.curTok

	// DataType
	dt = p.parseDataType()
	if isNil(dt) {
		return nil
	}
	tp.Value = dt

	p.nodeSet.Append(tp)
	return tp
}

func (p *Parser) parseArrayDataType() *ast.ArrayDataType {
	var tp = &ast.ArrayDataType{}
	tp.LBrack = p.curTok

	// token INT | ELLIPSIS
	if !p.nextToken() {
		return nil
	}
	tp.Length = p.curTok

	// token ']'
	if !p.advanceIfPeekTokenIs(token.RBRACK) {
		return nil
	}
	tp.RBrack = p.curTok

	// DataType
	dt := p.parseDataType()
	if isNil(dt) {
		return nil
	}
	tp.DataType = dt

	p.nodeSet.Append(tp)
	return tp
}

func (p *Parser) parseSliceDataType() *ast.SliceDataType {
	var tp = &ast.SliceDataType{}
	tp.LBrack = p.curTok

	// token ']'
	if !p.advanceIfPeekTokenIs(token.RBRACK) {
		return nil
	}
	tp.RBrack = p.curTok

	// DataType
	dt := p.parseDataType()
	if isNil(dt) {
		return nil
	}
	tp.DataType = dt

	p.nodeSet.Append(tp)
	return tp
}

func (p *Parser) parseImportStmt() ast.ImportStmt {
	if p.notExpectPeekToken(token.LPAREN, token.STRING) {
		return nil
	}

	if p.peekTokenIs(token.LPAREN) {
		return p.parseImportGroupStmt()
	}

	return p.parseImportLiteralStmt()
}

func (p *Parser) parseImportLiteralStmt() ast.ImportStmt {
	var stmt = &ast.ImportLiteralStmt{}
	stmt.Import = p.getNode(p.curTok)

	// token STRING
	if !p.advanceIfPeekTokenIs(token.STRING) {
		return nil
	}
	stmt.Value = p.getNode(p.curTok)

	p.nodeSet.InsertAfter(stmt, stmt.Value)
	return stmt
}

func (p *Parser) parseImportGroupStmt() ast.ImportStmt {
	var stmt = &ast.ImportGroupStmt{}
	stmt.Import = p.getNode(p.curTok)

	// token '('
	if !p.advanceIfPeekTokenIs(token.LPAREN) { // assert: dead code
		return nil
	}
	stmt.LParen = p.getNode(p.curTok)

	// token STRING
	for p.curTokenIsNotEof() && p.peekTokenIsNot(token.RPAREN) {
		if !p.advanceIfPeekTokenIs(token.STRING) {
			return nil
		}
		stmt.Values = append(stmt.Values, p.getNode(p.curTok))

		if p.notExpectPeekToken(token.RPAREN, token.STRING) {
			return nil
		}
	}

	// token ')'
	if !p.advanceIfPeekTokenIs(token.RPAREN) {
		return nil
	}
	stmt.RParen = p.getNode(p.curTok)

	p.nodeSet.InsertAfter(stmt, stmt.RParen)
	return stmt
}

func (p *Parser) parseInfoStmt() *ast.InfoStmt {
	var stmt = &ast.InfoStmt{}
	stmt.Info = p.getNode(p.curTok)

	// token '('
	if !p.advanceIfPeekTokenIs(token.LPAREN) {
		return nil
	}
	stmt.LParen = p.getNode(p.curTok)

	for p.curTokenIsNotEof() && p.peekTokenIsNot(token.RPAREN) {
		expr := p.parseKVExpression()
		if expr == nil {
			return nil
		}

		stmt.Values = append(stmt.Values, expr)
		if p.notExpectPeekToken(token.RPAREN, token.KEY) {
			return nil
		}
	}

	// token ')'
	if !p.advanceIfPeekTokenIs(token.RPAREN) {
		return nil
	}
	stmt.RParen = p.getNode(p.curTok)

	p.nodeSet.InsertAfter(stmt, stmt.RParen)
	return stmt
}

func (p *Parser) parseAtServerKVExpression() *ast.KVExpr {
	var expr = &ast.KVExpr{}

	// token IDENT
	if !p.advanceIfPeekTokenIs(token.KEY, token.RPAREN) {
		return nil
	}

	expr.Key = ast.NewTokenNode(p.curTok)

	var valueTok token.Token
	if p.peekTokenIs(token.QUO) {
		if !p.nextToken() {
			return nil
		}
		slashTok := p.curTok
		if !p.advanceIfPeekTokenIs(token.IDENT) {
			return nil
		}
		idTok := p.curTok
		valueTok = token.Token{
			Text:     slashTok.Text + idTok.Text,
			Position: slashTok.Position,
		}
	} else {
		if !p.advanceIfPeekTokenIs(token.IDENT) {
			return nil
		}
		valueTok = p.curTok
	}

	for {
		if p.peekTokenIs(token.QUO) {
			if !p.nextToken() {
				return nil
			}
			slashTok := p.curTok
			if !p.advanceIfPeekTokenIs(token.IDENT) {
				return nil
			}
			idTok := p.curTok
			valueTok = token.Token{
				Text:     valueTok.Text + slashTok.Text + idTok.Text,
				Position: valueTok.Position,
			}
		} else {
			break
		}
	}

	valueTok.Type = token.PATH
	expr.Value = ast.NewTokenNode(valueTok)

	p.nodeSet.InsertAfter(expr, expr.Value)
	return expr
}

func (p *Parser) parseKVExpression() *ast.KVExpr {
	var expr = &ast.KVExpr{}

	// token IDENT
	if !p.advanceIfPeekTokenIs(token.KEY) {
		return nil
	}
	expr.Key = p.getNode(p.curTok)

	// token STRING
	if !p.advanceIfPeekTokenIs(token.STRING) {
		return nil
	}
	expr.Value = p.getNode(p.curTok)
	p.nodeSet.InsertAfter(expr, expr.Value)
	return expr
}

func (p *Parser) parseSyntaxStmt() *ast.SyntaxStmt {
	var stmt = &ast.SyntaxStmt{}
	stmt.Syntax = p.getNode(p.curTok)

	// token '='
	if !p.advanceIfPeekTokenIs(token.ASSIGN) {
		return nil
	}
	stmt.Assign = p.getNode(p.curTok)

	// token STRING
	if !p.advanceIfPeekTokenIs(token.STRING) {
		return nil
	}
	stmt.Value = p.getNode(p.curTok)

	p.nodeSet.InsertAfter(stmt, stmt.Value)
	return stmt
}

func (p *Parser) curTokenIsNotEof() bool {
	return p.curTokenIsNot(token.EOF)
}

func (p *Parser) curTokenIsNot(expected token.Type) bool {
	return p.curTok.Type != expected
}

func (p *Parser) curTokenIs(expected ...interface{}) bool {
	for _, v := range expected {
		switch val := v.(type) {
		case token.Type:
			if p.curTok.Type == val {
				return true
			}
		case string:
			if p.curTok.Text == val {
				return true
			}
		}
	}
	return false
}

func (p *Parser) advanceIfPeekTokenIs(expected ...interface{}) bool {
	if p.expectPeekToken(expected...) {
		if !p.nextToken() {
			return false
		}
		return true
	}

	return false
}

func (p *Parser) peekTokenIs(expected ...interface{}) bool {
	for _, v := range expected {
		switch val := v.(type) {
		case token.Type:
			if p.peekTok.Type == val {
				return true
			}
		case string:
			if p.peekTok.Text == val {
				return true
			}
		}
	}
	return false
}

func (p *Parser) peekTokenIsNot(expected ...interface{}) bool {
	for _, v := range expected {
		switch val := v.(type) {
		case token.Type:
			if p.peekTok.Type == val {
				return false
			}
		case string:
			if p.peekTok.Text == val {
				return false
			}
		}
	}
	return true
}

func (p *Parser) notExpectPeekToken(expected ...interface{}) bool {
	if !p.peekTokenIsNot(expected...) {
		return false
	}

	var expectedString []string
	for _, v := range expected {
		expectedString = append(expectedString, fmt.Sprintf("'%s'", v))
	}

	var got string
	if p.peekTok.Type == token.ILLEGAL {
		got = p.peekTok.Text
	} else {
		got = p.peekTok.Type.String()
	}

	var err error
	if p.peekTok.Type == token.EOF {
		position := p.curTok.Position
		position.Column = position.Column + len(p.curTok.Text)
		err = fmt.Errorf(
			"%s syntax error: expected %s, got '%s'",
			position,
			strings.Join(expectedString, " | "),
			got)
	} else {
		err = fmt.Errorf(
			"%s syntax error: expected %s, got '%s'",
			p.peekTok.Position,
			strings.Join(expectedString, " | "),
			got)
	}
	p.errors = append(p.errors, err)

	return true
}

func (p *Parser) expectPeekToken(expected ...interface{}) bool {
	if p.peekTokenIs(expected...) {
		return true
	}

	var expectedString []string
	for _, v := range expected {
		expectedString = append(expectedString, fmt.Sprintf("'%s'", v))
	}

	var got string
	if p.peekTok.Type == token.ILLEGAL {
		got = p.peekTok.Text
	} else {
		got = p.peekTok.Type.String()
	}

	var err error
	if p.peekTok.Type == token.EOF {
		position := p.curTok.Position
		position.Column = position.Column + len(p.curTok.Text)
		err = fmt.Errorf(
			"%s syntax error: expected %s, got '%s'",
			position,
			strings.Join(expectedString, " | "),
			got)
	} else {
		err = fmt.Errorf(
			"%s syntax error: expected %s, got '%s'",
			p.peekTok.Position,
			strings.Join(expectedString, " | "),
			got)
	}
	p.errors = append(p.errors, err)

	return false
}

func (p *Parser) expectIdentError(tok token.Token, expected ...interface{}) {
	var expectedString []string
	for _, v := range expected {
		expectedString = append(expectedString, fmt.Sprintf("'%s'", v))
	}

	p.errors = append(p.errors, fmt.Errorf(
		"%s syntax error: expected %s, got '%s'",
		tok.Position,
		strings.Join(expectedString, " | "),
		tok.Type.String()))
}

func (p *Parser) init() bool {
	if !p.nextToken() {
		return false
	}
	return p.nextToken()
}

func (p *Parser) nextToken() bool {
	var err error
	p.curTok = p.peekTok
	if p.curTok.Valid() {
		tokenNode := ast.NewTokenNode(p.curTok)
		p.tokenNode[p.curTok] = tokenNode
		p.nodeSet.Append(tokenNode)
	}

	p.peekTok, err = p.s.NextToken()
	if err != nil {
		p.errors = append(p.errors, err)
		return false
	}

	for p.peekTok.Type == token.COMMENT || p.peekTok.Type == token.DOCUMENT {
		p.nodeSet.Append(&ast.CommentStmt{Comment: p.peekTok})
		p.peekTok, err = p.s.NextToken()
		if err != nil {
			p.errors = append(p.errors, err)
			return false
		}

	}

	return true
}

func (p *Parser) getNode(tok token.Token) *ast.TokenNode {
	return p.tokenNode[tok]
}

func isNil(v interface{}) bool {
	if v == nil {
		return true
	}

	vo := reflect.ValueOf(v)
	if vo.Kind() == reflect.Ptr {
		return vo.IsNil()
	}
	return false
}

func (p *Parser) CheckErrors() error {
	if len(p.errors) == 0 {
		return nil
	}
	var errors []string
	for _, e := range p.errors {
		errors = append(errors, e.Error())
	}
	return fmt.Errorf(strings.Join(errors, "\n"))
}

/************************以下函数仅用于单元测试************************/
func (p *Parser) hasNoErrors() bool {
	return len(p.errors) == 0
}

func (p *Parser) parseForUintTest() *ast.AST {
	api := &ast.AST{}
	if !p.init() {
		return nil
	}

	for p.curTokenIsNotEof() {
		stmt := p.parseStmtForUniTest()
		if isNil(stmt) {
			return nil
		}

		api.Stmts = append(api.Stmts, stmt)
		if !p.nextToken() {
			return nil
		}
	}

	return api
}

func (p *Parser) parseStmtForUniTest() ast.Stmt {
	switch p.curTok.Type {
	case token.AT_SERVER:
		return p.parseAtServerStmt()
	case token.AT_HANDLER:
		return p.parseAtHandlerStmt()
	case token.AT_DOC:
		return p.parseAtDocStmt()
	}
	return nil
}
