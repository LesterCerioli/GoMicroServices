package ast

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/tools/goctl/pkg/parser/api/token"
)

/*******************TypeStmt Begin********************/

type TypeStmt interface {
	Stmt
	typeNode()
}

type TypeLiteralStmt struct {
	Type *TokenNode
	Expr *TypeExpr
}

func (t *TypeLiteralStmt) Format(prefix ...string) string {
	w := NewBufferWriter()
	w.Write(withNode(t.Type, t.Expr), withPrefix(prefix...), expectSameLine())
	return w.String()
}

func (t *TypeLiteralStmt) End() token.Position {
	return t.Expr.End()
}

func (t *TypeLiteralStmt) Pos() token.Position {
	return t.Type.Pos()
}

func (t *TypeLiteralStmt) stmtNode() {}
func (t *TypeLiteralStmt) typeNode() {}

type TypeGroupStmt struct {
	Type     *TokenNode
	LParen   *TokenNode
	ExprList []*TypeExpr
	RParen   *TokenNode
}

func (t *TypeGroupStmt) Format(prefix ...string) string {
	if len(t.ExprList) == 0 {
		return ""
	}
	w := NewBufferWriter()
	typeNode := transferTokenNode(t.Type, withTokenNodePrefix(prefix...))
	w.Write(withNode(typeNode, t.LParen), expectSameLine())
	w.NewLine()
	for _, e := range t.ExprList {
		w.Write(withNode(e), withPrefix(peekOne(prefix)+Indent), expectIndentInfix())
		w.NewLine()
	}
	w.WriteText(t.RParen.Format(prefix...))
	return w.String()
}

func (t *TypeGroupStmt) End() token.Position {
	return t.RParen.End()
}

func (t *TypeGroupStmt) Pos() token.Position {
	return t.Type.Pos()
}

func (t *TypeGroupStmt) stmtNode() {}
func (t *TypeGroupStmt) typeNode() {}

/*******************TypeStmt End********************/

/*******************TypeExpr Begin********************/

type TypeExpr struct {
	Name     *TokenNode
	Assign   *TokenNode
	DataType DataType
}

func (e *TypeExpr) Format(prefix ...string) string {
	w := NewBufferWriter()
	dataTypeNode := transfer2TokenNode(e.DataType, false, withTokenNodePrefix(prefix...))
	if e.Assign != nil {
		w.Write(withNode(e.Name, e.Assign, dataTypeNode), expectSameLine())
	} else {
		w.Write(withNode(e.Name, dataTypeNode), expectSameLine())
	}
	return w.String()
}

func (e *TypeExpr) End() token.Position {
	return e.DataType.End()
}

func (e *TypeExpr) Pos() token.Position {
	return e.Name.Pos()
}

func (e *TypeExpr) exprNode() {}

func (e *TypeExpr) isStruct() bool {
	return e.DataType.ContainsStruct()
}

/*******************TypeExpr Begin********************/

/*******************Elem Begin********************/

type ElemExpr struct {
	Name     []*TokenNode
	DataType DataType
	Tag      *TokenNode
}

func (e *ElemExpr) Format(prefix ...string) string {
	w := NewBufferWriter()
	var nameNodeList []*TokenNode
	for idx, n := range e.Name {
		if idx == 0 {
			nameNodeList = append(nameNodeList,
				transferTokenNode(n, ignoreLeadingComment()))
		} else if idx < len(e.Name)-1 {
			nameNodeList = append(nameNodeList,
				transferTokenNode(n, ignoreLeadingComment(), ignoreHeadComment()))
		} else {
			nameNodeList = append(nameNodeList, transferTokenNode(n, ignoreHeadComment()))
		}
	}

	nameNode := transferNilInfixNode(nameNodeList,
		withTokenNodePrefix(prefix...), withTokenNodeInfix(", "))

	var dataTypeOption []tokenNodeOption
	if e.DataType.ContainsStruct() {
		dataTypeOption = append(dataTypeOption, withTokenNodePrefix(peekOne(prefix)+Indent))
	} else {
		dataTypeOption = append(dataTypeOption, withTokenNodePrefix(prefix...))
	}
	dataTypeNode := transfer2TokenNode(e.DataType, false, dataTypeOption...)
	if e.Tag != nil {
		w.Write(withNode(nameNode, dataTypeNode, e.Tag), expectIndentInfix(), expectSameLine())
	} else {
		w.Write(withNode(nameNode, dataTypeNode), expectIndentInfix(), expectSameLine())
	}
	return w.String()
}

func (e *ElemExpr) End() token.Position {
	if e.Tag != nil {
		return e.Tag.End()
	}
	return e.DataType.End()
}

func (e *ElemExpr) Pos() token.Position {
	if len(e.Name) > 0 {
		return e.Name[0].Pos()
	}
	return token.IllegalPosition
}

func (e *ElemExpr) exprNode() {}

/*******************Elem End********************/

/*******************ElemExprList Begin********************/

type ElemExprList []*ElemExpr

func (e ElemExprList) Pos() token.Position {
	if len(e) > 0 {
		return e[0].Pos()
	}
	return token.IllegalPosition
}

func (e ElemExprList) exprNode() {}

/*******************ElemExprList Begin********************/

/*******************DataType Begin********************/

type DataType interface {
	Expr
	dataTypeNode()
	CanEqual() bool
	ContainsStruct() bool
	RawText() string
}

type AnyDataType struct {
	Any     *TokenNode
	isChild bool
}

func (t *AnyDataType) Format(prefix ...string) string {
	return t.Any.Format(prefix...)
}

func (t *AnyDataType) End() token.Position {
	return t.Any.End()
}

func (t *AnyDataType) RawText() string {
	return t.Any.Token.Text
}

func (t *AnyDataType) ContainsStruct() bool {
	return false
}

func (t *AnyDataType) Pos() token.Position {
	return t.Any.Pos()
}

func (t *AnyDataType) exprNode() {}

func (t *AnyDataType) dataTypeNode() {}

func (t *AnyDataType) CanEqual() bool {
	return true
}

type ArrayDataType struct {
	LBrack   *TokenNode
	Length   *TokenNode
	RBrack   *TokenNode
	DataType DataType
	isChild  bool
}

func (t *ArrayDataType) Format(prefix ...string) string {
	w := NewBufferWriter()
	lbrack := transferTokenNode(t.LBrack, ignoreLeadingComment())
	lengthNode := transferTokenNode(t.Length, ignoreLeadingComment())
	rbrack := transferTokenNode(t.RBrack, ignoreHeadComment())
	var dataType *TokenNode
	var options []tokenNodeOption
	options = append(options, withTokenNodePrefix(prefix...))
	if t.isChild {
		options = append(options, ignoreComment())
	} else {
		options = append(options, ignoreHeadComment())
	}

	dataType = transfer2TokenNode(t.DataType, false, options...)
	node := transferNilInfixNode([]*TokenNode{lbrack, lengthNode, rbrack, dataType})
	w.Write(withNode(node))
	return w.String()
}

func (t *ArrayDataType) End() token.Position {
	return t.DataType.End()
}

func (t *ArrayDataType) RawText() string {
	return ""
}

func (t *ArrayDataType) ContainsStruct() bool {
	return t.DataType.ContainsStruct()
}

func (t *ArrayDataType) CanEqual() bool {
	return t.DataType.CanEqual()
}

func (t *ArrayDataType) Pos() token.Position {
	return t.LBrack.Pos()
}

func (t *ArrayDataType) exprNode()     {}
func (t *ArrayDataType) dataTypeNode() {}

// BaseDataType is a common id type which contains bool, uint8, uint16, uint32,
// uint64, int8, int16, int32, int64, float32, float64, complex64, complex128,
// string, int, uint, uintptr, byte, rune, any.
type BaseDataType struct {
	Base    *TokenNode
	isChild bool
}

func (t *BaseDataType) Format(prefix ...string) string {
	return t.Base.Format(prefix...)
}

func (t *BaseDataType) End() token.Position {
	return t.Base.End()
}

func (t *BaseDataType) RawText() string {
	return t.Base.Token.Text
}

func (t *BaseDataType) ContainsStruct() bool {
	return false
}

func (t *BaseDataType) CanEqual() bool {
	return true
}

func (t *BaseDataType) Pos() token.Position {
	return t.Base.Pos()
}

func (t *BaseDataType) exprNode()     {}
func (t *BaseDataType) dataTypeNode() {}

type InterfaceDataType struct {
	Interface *TokenNode
	isChild   bool
}

func (t *InterfaceDataType) Format(prefix ...string) string {
	return t.Interface.Format(prefix...)
}

func (t *InterfaceDataType) End() token.Position {
	return t.Interface.End()
}

func (t *InterfaceDataType) RawText() string {
	return t.Interface.Token.Text
}

func (t *InterfaceDataType) ContainsStruct() bool {
	return false
}

func (t *InterfaceDataType) CanEqual() bool {
	return true
}

func (t *InterfaceDataType) Pos() token.Position {
	return t.Interface.Pos()
}

func (t *InterfaceDataType) exprNode() {}

func (t *InterfaceDataType) dataTypeNode() {}

type MapDataType struct {
	Map     *TokenNode
	LBrack  *TokenNode
	Key     DataType
	RBrack  *TokenNode
	Value   DataType
	isChild bool
}

func (t *MapDataType) Format(prefix ...string) string {
	w := NewBufferWriter()
	mapNode := transferTokenNode(t.Map, ignoreLeadingComment())
	lbrack := transferTokenNode(t.LBrack, ignoreLeadingComment())
	rbrack := transferTokenNode(t.RBrack, ignoreComment())
	var keyOption, valueOption []tokenNodeOption
	keyOption = append(keyOption, ignoreComment())
	valueOption = append(valueOption, withTokenNodePrefix(prefix...))

	if t.isChild {
		valueOption = append(valueOption, ignoreComment())
	} else {
		valueOption = append(valueOption, ignoreHeadComment())
	}

	keyDataType := transfer2TokenNode(t.Key, true, keyOption...)
	valueDataType := transfer2TokenNode(t.Value, false, valueOption...)
	node := transferNilInfixNode([]*TokenNode{mapNode, lbrack, keyDataType, rbrack, valueDataType})
	w.Write(withNode(node))
	return w.String()
}

func (t *MapDataType) End() token.Position {
	return t.Value.End()
}

func (t *MapDataType) RawText() string {
	return fmt.Sprintf("map[%s]%s", t.Key.RawText(), t.Value.RawText())
}

func (t *MapDataType) ContainsStruct() bool {
	return t.Key.ContainsStruct() || t.Value.ContainsStruct()
}

func (t *MapDataType) CanEqual() bool {
	return false
}

func (t *MapDataType) Pos() token.Position {
	return t.Map.Pos()
}

func (t *MapDataType) exprNode()     {}
func (t *MapDataType) dataTypeNode() {}

type PointerDataType struct {
	Star     *TokenNode
	DataType DataType
	isChild  bool
}

func (t *PointerDataType) Format(prefix ...string) string {
	w := NewBufferWriter()
	star := transferTokenNode(t.Star, ignoreLeadingComment())
	var dataTypeOption []tokenNodeOption
	dataTypeOption = append(dataTypeOption, withTokenNodePrefix(prefix...))
	dataTypeOption = append(dataTypeOption, ignoreHeadComment())
	dataType := transfer2TokenNode(t.DataType, false, dataTypeOption...)
	node := transferNilInfixNode([]*TokenNode{star, dataType})
	w.Write(withNode(node))
	return w.String()
}

func (t *PointerDataType) End() token.Position {
	return t.DataType.End()
}

func (t *PointerDataType) RawText() string {
	return "*" + t.DataType.RawText()
}

func (t *PointerDataType) ContainsStruct() bool {
	return t.DataType.ContainsStruct()
}

func (t *PointerDataType) CanEqual() bool {
	return t.DataType.CanEqual()
}

func (t *PointerDataType) Pos() token.Position {
	return t.Star.Pos()
}

func (t *PointerDataType) exprNode()     {}
func (t *PointerDataType) dataTypeNode() {}

type SliceDataType struct {
	LBrack   *TokenNode
	RBrack   *TokenNode
	DataType DataType
	isChild  bool
}

func (t *SliceDataType) Format(prefix ...string) string {
	w := NewBufferWriter()
	lbrack := transferTokenNode(t.LBrack, ignoreLeadingComment())
	rbrack := transferTokenNode(t.RBrack, ignoreHeadComment())
	dataType := transfer2TokenNode(t.DataType, false, withTokenNodePrefix(prefix...), ignoreHeadComment())
	node := transferNilInfixNode([]*TokenNode{lbrack, rbrack, dataType})
	w.Write(withNode(node))
	return w.String()
}

func (t *SliceDataType) End() token.Position {
	return t.DataType.End()
}

func (t *SliceDataType) RawText() string {
	return fmt.Sprintf("[]%s", t.DataType.RawText())
}

func (t *SliceDataType) ContainsStruct() bool {
	return t.DataType.ContainsStruct()
}

func (t *SliceDataType) CanEqual() bool {
	return false
}

func (t *SliceDataType) Pos() token.Position {
	return t.LBrack.Pos()
}

func (t *SliceDataType) exprNode()     {}
func (t *SliceDataType) dataTypeNode() {}

type StructDataType struct {
	LBrace   *TokenNode
	Elements ElemExprList
	RBrace   *TokenNode
	isChild  bool
}

func (t *StructDataType) Format(prefix ...string) string {
	w := NewBufferWriter()
	if len(t.Elements) == 0 {
		lbrace := transferTokenNode(t.LBrace, withTokenNodePrefix(prefix...), ignoreLeadingComment())
		rbrace := transferTokenNode(t.RBrace, ignoreHeadComment())
		brace := transferNilInfixNode([]*TokenNode{lbrace, rbrace})
		w.Write(withNode(brace), expectSameLine())
		return w.String()
	}
	w.WriteText(t.LBrace.Format(NilIndent))
	w.NewLine()
	for _, e := range t.Elements {
		//w.Write(withNode(e), withPrefix(peekOne(prefix)+Indent))
		var nameNode *TokenNode
		nameNode = transferTokenNode(e.Name[0], withTokenNodePrefix(peekOne(prefix)+Indent))
		if len(e.Name) > 1 {
			var nameNodeList []*TokenNode
			for idx, n := range e.Name {
				if idx == 0 {
					nameNodeList = append(nameNodeList,
						transferTokenNode(n, ignoreLeadingComment()))
				} else if idx < len(e.Name)-1 {
					nameNodeList = append(nameNodeList,
						transferTokenNode(n, ignoreLeadingComment(), ignoreHeadComment()))
				} else {
					nameNodeList = append(nameNodeList, transferTokenNode(n, ignoreHeadComment()))
				}
			}

			nameNode = transferNilInfixNode(nameNodeList,
				withTokenNodePrefix(peekOne(prefix)+Indent), withTokenNodeInfix(", "))
		}
		var dataTypeOption []tokenNodeOption
		if e.DataType.ContainsStruct() {
			dataTypeOption = append(dataTypeOption, withTokenNodePrefix(peekOne(prefix)+Indent))
		} else {
			dataTypeOption = append(dataTypeOption, withTokenNodePrefix(prefix...))
		}

		dataTypeNode := transfer2TokenNode(e.DataType, false, dataTypeOption...)
		if e.Tag != nil {
			if e.DataType.ContainsStruct() {
				w.Write(withNode(nameNode, dataTypeNode, e.Tag), expectSameLine())
			} else {
				w.Write(withNode(nameNode, e.DataType, e.Tag), expectIndentInfix(), expectSameLine())
			}
		} else {
			if e.DataType.ContainsStruct() {
				w.Write(withNode(nameNode, dataTypeNode), expectSameLine())
			} else {
				w.Write(withNode(nameNode, e.DataType), expectIndentInfix(), expectSameLine())
			}
		}
		w.NewLine()
	}
	w.WriteText(t.RBrace.Format(prefix...))
	return w.String()
}

func (t *StructDataType) End() token.Position {
	return t.RBrace.End()
}

func (t *StructDataType) RawText() string {
	b := bytes.NewBuffer(nil)
	b.WriteRune('{')
	for _, v := range t.Elements {
		b.WriteRune('\n')
		var nameList []string
		for _, n := range v.Name {
			nameList = append(nameList, n.Token.Text)
		}
		b.WriteString(fmt.Sprintf("%s %s %s", strings.Join(nameList, ", "), v.DataType.RawText(), v.Tag.Token.Text))
	}
	b.WriteRune('\n')
	b.WriteRune('}')
	return b.String()
}

func (t *StructDataType) ContainsStruct() bool {
	return true
}

func (t *StructDataType) CanEqual() bool {
	for _, v := range t.Elements {
		if !v.DataType.CanEqual() {
			return false
		}
	}
	return true
}

func (t *StructDataType) Pos() token.Position {
	return t.LBrace.Pos()
}

func (t *StructDataType) exprNode()     {}
func (t *StructDataType) dataTypeNode() {}

/*******************DataType End********************/
