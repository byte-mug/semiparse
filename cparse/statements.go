/*
MIT License

Copyright (c) 2017 Simon Schmidt

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/


package cparse


import "github.com/byte-mug/semiparse/scanlist"
import "github.com/byte-mug/semiparse/parser"
import "text/scanner"
import "fmt"

const (
	s_none = uint(iota)
	S_EXPR
	S_BLOCK
	S_VARDEC
	S_FOR
	S_IF
	S_IF_ELSE
	S_DO_WHILE
	S_WHILE
)

type VarDecl struct{
	Name string
	Init interface{} // Expr
}
func (v VarDecl) String() string {
	if v.Init!=nil { return fmt.Sprint(v.Name," = ",v.Init) }
	return v.Name
}

type Statement struct{
	Type uint
	Text string
	Data []interface{}
	Pos scanner.Position
}
func (e *Statement) String() string {
	if e==nil { return "NIL" }
	switch e.Type {
	case S_BLOCK:
		return fmt.Sprint("{",e.Data,"}")
	case S_EXPR:
		return fmt.Sprint(e.Data[0],";")
	case S_VARDEC:
		return fmt.Sprint("var ",e.Data[0]," ",e.Data[1:],";")
	case S_FOR:
		return fmt.Sprint("for(",e.Data[0],";",e.Data[1],";",e.Data[2],")",e.Data[3])
	case S_IF:
		return fmt.Sprint("if(",e.Data[0],")",e.Data[1])
	case S_IF_ELSE:
		return fmt.Sprint("if(",e.Data[0],")",e.Data[1]," else ",e.Data[2])
	case S_DO_WHILE:
		return fmt.Sprint("do ",e.Data[0]," while(",e.Data[1],");")
	case S_WHILE:
		return fmt.Sprint("while(",e.Data[0],")",e.Data[1])
	}
	return fmt.Sprint("(",e.Type,"#",e.Text,e.Data,")")
}


func c_statement_prim(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	expr := p.Match("Expr",tokens)
	if expr.Result == parser.RESULT_OK {
		expr.Data = &Statement{S_EXPR,"",aR(expr.Data),tokens.SafePos()}
	}
	return expr
}

func c_statement_vardecl(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	ty := p.Match("Type",tokens)
	switch ty.Result {
	case parser.RESULT_OK:
		err,t := parser.Match(parser.Textify,ty.Next,scanner.Ident); if err!=nil { return parser.ResultFail(fmt.Sprint(err),tokens.SafePos()) }
		v1 := VarDecl{ty.Next.TokenText,nil}
		if ok,t2 := parser.FastMatch(ty.Next,scanner.Ident,'='); ok {
			x1 := p.Match("Expr",t2)
			if x1.Result != parser.RESULT_OK { return x1 }
			v1.Init = x1.Data
			
			t = x1.Next
		}
		vec := aR(ty.Data,v1)
		
		for {
			if ok,t2 := parser.FastMatch(t,',',scanner.Ident,'='); ok {
				t3 := t.Next()
				xn := p.Match("Expr",t2)
				if xn.Result != parser.RESULT_OK { return xn }
				vec = append(vec,VarDecl{t3.TokenText,xn.Data})
				t = xn.Next
				continue
			}
			if ok,t2 := parser.FastMatch(t,',',scanner.Ident); ok {
				t3 := t.Next()
				vec = append(vec,VarDecl{t3.TokenText,nil})
				t = t2
				continue
			}
			err,t = parser.Match(parser.Textify,t,';');
			if err!=nil { return parser.ResultFail(fmt.Sprint(err),tokens.SafePos()) }
			break
		}
		ty.Data = &Statement{S_VARDEC,"var",vec,tokens.Pos}
		ty.Next = t
		return ty
	case parser.RESULT_FAILED_CUT:
		return ty
	}
	
	return parser.ResultFail("Invalid Variable Declaration!",tokens.SafePos())
}

func c_statement(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	if tokens==nil { return parser.ResultFail("EOF!",scanner.Position{}) }
	switch tokens.Token {
	case '{' /*}*/:
		res := parser.ArrayStar{parser.Delegate("Statement")}.Parse(p,tokens.Next(),left)
		if res.Result==parser.RESULT_OK {
			err,t := parser.Match(parser.Textify,res.Next,/*{*/'}')
			if err!=nil { return parser.ResultFail(fmt.Sprint(err),tokens.SafePos()) }
			res.Next = t
			res.Data = &Statement{S_BLOCK,"{}",res.Data.([]interface{}),tokens.Pos}
		}
		return res
	case C_FOR:
		ars := parser.ArraySeq{
			parser.LSeq{parser.Required{'('/*)*/,parser.Textify},
			parser.TokenFinishedOptional{parser.Delegate("Expr"),';'}},
			parser.TokenFinishedOptional{parser.Delegate("Expr"),';'},
			parser.TokenFinishedOptional{parser.Delegate("Expr"),/*(*/')'},
			parser.Delegate("Statement"),
		}.Parse(p,tokens.Next(),left)
		if !ars.Ok() { return ars }
		return parser.ResultOk(ars.Next,&Statement{S_FOR,"for",ars.Data.([]interface{}),tokens.Pos})

	case C_IF:
		ars := parser.ArraySeq{
			parser.LSeq{parser.Required{'('/*)*/,parser.Textify},
			parser.Delegate("Expr")},
			parser.LSeq{parser.Required{/*(*/')',parser.Textify},
			parser.Delegate("Statement")},
		}.Parse(p,tokens.Next(),left)
		if !ars.Ok() { return ars }
		itf := ars.Data.([]interface{})
		if ars.Next.SafeToken() == C_ELSE {
			el := p.Match("Statement",ars.Next.Next())
			if !el.Ok() { return el }
			return parser.ResultOk(ars.Next,&Statement{S_IF_ELSE,"if-else",append(itf,el.Data),tokens.Pos})
		}
		return parser.ResultOk(ars.Next,&Statement{S_IF,"if",itf ,tokens.Pos})
	case C_DO:
		ars := parser.ArraySeq{
			parser.LSeq{parser.Required{C_DO,parser.Textify},
			parser.Delegate("Statement")},
			parser.LSeq{parser.Required{C_WHILE,parser.Textify},parser.Required{'('/*)*/,parser.Textify},
			parser.Delegate("Expr")},
			parser.LSeq{parser.Required{/*(*/')',parser.Textify},parser.Required{';',parser.Textify}},
		}.Parse(p,tokens,left)
		if !ars.Ok() { return ars }
		itf := ars.Data.([]interface{})[:2]
		return parser.ResultOk(ars.Next,&Statement{S_DO_WHILE,"do-while",itf ,tokens.Pos})
	case C_WHILE:
		ars := parser.ArraySeq{
			parser.LSeq{parser.Required{C_WHILE,parser.Textify},parser.Required{'('/*)*/,parser.Textify},
			parser.Delegate("Expr")},
			parser.LSeq{parser.Required{/*(*/')',parser.Textify},parser.Delegate("Statement")},
		}.Parse(p,tokens,left)
		if !ars.Ok() { return ars }
		return parser.ResultOk(ars.Next,&Statement{S_WHILE,"while",ars.Data.([]interface{}),tokens.Pos})
	}
	
	if vd := c_statement_vardecl(p,tokens,left); !vd.TryNextRule() { return vd }
	
	prim := p.Match("StatementPrim",tokens)
	switch prim.Result {
	case parser.RESULT_OK:
		err,t := parser.Match(parser.Textify,prim.Next,';'); if err!=nil { return parser.ResultFail(fmt.Sprint(err),tokens.SafePos()) }
		prim.Next = t
		fallthrough
	case parser.RESULT_FAILED_CUT:
		return prim
	}
	
	return parser.ResultFail("Invalid Statement!",tokens.SafePos())
}
func RegisterStatememt(p *parser.Parser) {
	p.Define("StatementPrim",false,parser.Pfunc(c_statement_prim))
	p.Define("Statement",false,parser.Pfunc(c_statement))
}

