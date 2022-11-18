package ast

import (
	"strings"

	"github.com/zeromicro/go-zero/tools/goctl/pkg/parser/api/token"
)

type AtServerStmt struct {
	AtServer *TokenNode
	LParen   *TokenNode
	Values   []*KVExpr
	RParen   *TokenNode
}

func (a *AtServerStmt) Format(prefix ...string) string {
	if len(a.Values) == 0 {
		return ""
	}
	var textList []string
	for _, v := range a.Values {
		if v.Value.IsZeroString() {
			continue
		}
		textList = append(textList, v.Format())
	}
	if len(textList) == 0 {
		return ""
	}

	w := NewBufferWriter()
	w.Write(WithNode(a.AtServer, a.LParen), WithPrefix(prefix...),
		WithMode(ModeExpectInSameLine))
	w.NewLine()
	for _, v := range a.Values {
		w.Write(WithNode(v.Key, v.Value), WithPrefix(peekOne(prefix)+Indent),
			WithMode(ModeExpectInSameLine))
		w.NewLine()
	}
	w.WriteText(a.RParen.Format(prefix...))
	return w.String()
}

func (a *AtServerStmt) End() token.Position {
	return a.RParen.End()
}

func (a *AtServerStmt) Pos() token.Position {
	return a.AtServer.Pos()
}

func (a *AtServerStmt) stmtNode() {}

type AtDocStmt interface {
	Stmt
	atDocNode()
}

type AtDocLiteralStmt struct {
	AtDoc *TokenNode
	Value *TokenNode
}

func (a *AtDocLiteralStmt) Format(prefix ...string) string {
	if a.Value.IsZeroString() {
		return ""
	}
	w := NewBufferWriter()
	w.WriteText(a.AtDoc.Format(prefix...) + WhiteSpace + a.Value.Format())
	return w.String()
}

func (a *AtDocLiteralStmt) End() token.Position {
	return a.Value.End()
}

func (a *AtDocLiteralStmt) atDocNode() {}

func (a *AtDocLiteralStmt) Pos() token.Position {
	return a.AtDoc.Pos()
}

func (a *AtDocLiteralStmt) stmtNode() {}

type AtDocGroupStmt struct {
	AtDoc  *TokenNode
	LParen *TokenNode
	Values []*KVExpr
	RParen *TokenNode
}

func (a *AtDocGroupStmt) Format(prefix ...string) string {
	if len(a.Values) == 0 {
		return ""
	}
	var textList []string
	for _, v := range a.Values {
		if v.Value.IsZeroString() {
			continue
		}
		textList = append(textList, v.Format(peekOne(prefix)+Indent))
	}
	if len(textList) == 0 {
		return ""
	}

	w := NewBufferWriter()
	w.WriteText(a.AtDoc.Format(prefix...) + WhiteSpace + a.LParen.Format())
	w.NewLine()
	w.WriteText(strings.Join(textList, NewLine))
	w.NewLine()
	w.WriteText(a.RParen.Format(prefix...))
	return w.String()
}

func (a *AtDocGroupStmt) End() token.Position {
	return a.RParen.End()
}

func (a *AtDocGroupStmt) atDocNode() {}

func (a *AtDocGroupStmt) Pos() token.Position {
	return a.AtDoc.Pos()
}

func (a *AtDocGroupStmt) stmtNode() {}

type ServiceStmt struct {
	AtServerStmt *AtServerStmt
	Service      *TokenNode
	Name         *ServiceNameExpr
	LBrace       *TokenNode
	Routes       []*ServiceItemStmt
	RBrace       *TokenNode
}

func (s *ServiceStmt) Format(prefix ...string) string {
	w := NewBufferWriter()
	if s.AtServerStmt != nil {
		w.WriteText(s.AtServerStmt.Format(prefix...))
		w.NewLine()
	}
	w.Write(WithNode(s.Service, s.Name, s.LBrace),
		WithPrefix(prefix...), WithMode(ModeExpectInSameLine))
	w.NewLine()
	var textList []string
	for _, v := range s.Routes {
		textList = append(textList, v.Format(peekOne(prefix)+Indent))
	}
	w.WriteText(strings.Join(textList, NewLine))
	w.NewLine()
	w.WriteText(s.RBrace.Format(prefix...))
	return w.String()
}

func (s *ServiceStmt) End() token.Position {
	return s.RBrace.End()
}

func (s *ServiceStmt) Pos() token.Position {
	if s.AtServerStmt != nil {
		return s.AtServerStmt.Pos()
	}
	return s.Service.Pos()
}

func (s *ServiceStmt) stmtNode() {}

type ServiceNameExpr struct {
	Name *TokenNode
}

func (s *ServiceNameExpr) Format(prefix ...string) string {
	w := NewBufferWriter()
	w.WriteText(s.Name.Format(prefix...))
	return w.String()
}

func (s *ServiceNameExpr) End() token.Position {
	return s.Name.End()
}

func (s *ServiceNameExpr) Pos() token.Position {
	return s.Name.Pos()
}

func (s *ServiceNameExpr) exprNode() {}

type AtHandlerStmt struct {
	AtHandler *TokenNode
	Name      *TokenNode
}

func (a *AtHandlerStmt) Format(prefix ...string) string {
	w := NewBufferWriter()
	w.WriteText(a.AtHandler.Format(prefix...) + WhiteSpace + a.Name.Format())
	return w.String()
}

func (a *AtHandlerStmt) End() token.Position {
	return a.Name.End()
}

func (a *AtHandlerStmt) Pos() token.Position {
	return a.AtHandler.Pos()
}

func (a *AtHandlerStmt) stmtNode() {}

type ServiceItemStmt struct {
	AtDoc     AtDocStmt
	AtHandler *AtHandlerStmt
	Route     *RouteStmt
}

func (s *ServiceItemStmt) Format(prefix ...string) string {
	w := NewBufferWriter()
	if s.AtDoc != nil {
		w.WriteText(s.AtDoc.Format(prefix...))
		w.NewLine()
	}
	w.WriteText(s.AtHandler.Format(prefix...))
	w.NewLine()
	w.WriteText(s.Route.Format(prefix...))
	return w.String()
}

func (s *ServiceItemStmt) End() token.Position {
	return s.Route.End()
}

func (s *ServiceItemStmt) Pos() token.Position {
	if s.AtDoc != nil {
		return s.AtDoc.Pos()
	}
	return s.AtHandler.Pos()
}

func (s *ServiceItemStmt) stmtNode() {}

type RouteStmt struct {
	Method   *TokenNode
	Path     *PathExpr
	Request  *BodyStmt
	Returns  *TokenNode
	Response *BodyStmt
}

func (r *RouteStmt) Format(prefix ...string) string {
	w := NewBufferWriter()
	if r.Request == nil {
		w.Write(WithNode(r.Method, r.Path), WithPrefix(prefix...),
			WithMode(ModeExpectInSameLine))
	} else if r.Returns == nil {
		w.Write(WithNode(r.Method, r.Path, r.Request),
			WithPrefix(prefix...), WithMode(ModeExpectInSameLine))
	} else {
		w.Write(WithNode(r.Method, r.Path, r.Request,
			r.Returns, r.Response), WithPrefix(prefix...),
			WithMode(ModeExpectInSameLine))
	}
	return w.String()
}

func (r *RouteStmt) End() token.Position {
	if r.Response != nil {
		return r.Response.End()
	}
	if r.Returns != nil {
		return r.Returns.Pos()
	}
	if r.Request != nil {
		return r.Request.End()
	}
	return r.Path.End()
}

func (r *RouteStmt) Pos() token.Position {
	return r.Method.Pos()
}

func (r *RouteStmt) stmtNode() {}

type PathExpr struct {
	Value *TokenNode
}

func (p *PathExpr) Format(prefix ...string) string {
	return p.Value.Format(prefix...)
}

func (p *PathExpr) End() token.Position {
	return p.Value.End()
}

func (p *PathExpr) Pos() token.Position {
	return p.Value.Pos()
}

func (p *PathExpr) exprNode() {}

type BodyStmt struct {
	LParen *TokenNode
	Body   *BodyExpr
	RParen *TokenNode
}

func (b *BodyStmt) Format(prefix ...string) string {
	return "(" + b.Body.Format() + ")"
}

func (b *BodyStmt) End() token.Position {
	return b.RParen.End()
}

func (b *BodyStmt) Pos() token.Position {
	return b.LParen.Pos()
}

func (b *BodyStmt) stmtNode() {}

type BodyExpr struct {
	LBrack *TokenNode
	RBrack *TokenNode
	Star   *TokenNode
	Value  *TokenNode
}

func (e *BodyExpr) End() token.Position {
	return e.Value.End()
}

func (e *BodyExpr) Format(prefix ...string) string {
	w := NewBufferWriter()
	if e.LBrack != nil {
		if e.Star != nil {
			w.WriteText(peekOne(prefix) + "[]*" + e.Value.Format(prefix...))
		} else {
			w.WriteText(peekOne(prefix) + "[]" + e.Value.Format(prefix...))
		}
	} else if e.Star != nil {
		w.WriteText(peekOne(prefix) + "*" + e.Value.Format(prefix...))
	} else {
		w.WriteText(peekOne(prefix) + e.Value.Format(prefix...))
	}
	return w.String()
}

func (e *BodyExpr) Pos() token.Position {
	if e.LBrack != nil {
		return e.LBrack.Pos()
	}
	if e.Star != nil {
		return e.Star.Pos()
	}
	return e.Value.Pos()
}

func (e *BodyExpr) exprNode() {}
