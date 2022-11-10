package ast

import "github.com/zeromicro/go-zero/tools/goctl/pkg/parser/api/token"

type SyntaxStmt struct {
	Syntax token.Token
	Assign token.Token
	Value  token.Token
}

func (s *SyntaxStmt) Pos() token.Position {
	return s.Syntax.Position
}

func (s *SyntaxStmt) Format(prefix string) string {
	w := NewWriter()
	w.WriteWithWhiteSpaceInfix(prefix, s.Syntax, s.Assign, s.Value)
	return w.String()
}

func (s *SyntaxStmt) stmtNode() {}