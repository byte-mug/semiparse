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
	e_none = uint(iota)
	E_VAR
	E_INT
	E_FLOAT
	E_CHAR
	E_STRING
	
	E_INCR
	E_DECR
	E_FIELD_DOT // Expr.Field
	E_FIELD_PTR // Expr->Field
	E_UNARY_OP
	E_BINARY_OP
	E_BINARY_OP_ASSIGN
	E_COMPARE
	E_ASSIGN
	E_FUNCTION_CALL // Expr( [Expr [,Expr]* ] )
	E_CONDITIONAL
	E_CAST
	E_INDEX
)

func aR(i ...interface{}) []interface{} { return i }

/* When in doubt, use this! */
var pOS = scanner.Position{}

type Expr struct{
	Type uint
	Text string
	Data []interface{}
	Pos scanner.Position
}
func (e *Expr) String() string {
	if e==nil { return "NIL" }
	switch e.Type {
	case E_BINARY_OP:
		return fmt.Sprint("(",e.Data[0],e.Text,e.Data[1],")")
	case E_BINARY_OP_ASSIGN:
		return fmt.Sprint("(",e.Data[0],e.Text,"=",e.Data[1],")")
	case E_INCR:
		return fmt.Sprint("(",e.Data[0],"++)")
	case E_DECR:
		return fmt.Sprint("(",e.Data[0],"--)")
	case E_FIELD_DOT:
		return fmt.Sprint("(",e.Data,".",e.Text,")")
	case E_FIELD_PTR:
		return fmt.Sprint("(",e.Data,"->",e.Text,")")
	case E_UNARY_OP:
		return fmt.Sprint("(",e.Text,e.Data[0],")")
	case E_VAR,E_INT,E_FLOAT,E_CHAR,E_STRING:
		return fmt.Sprint(e.Text)
	case E_ASSIGN:
		return fmt.Sprint("(",e.Data[0],"=",e.Data[1],")")
	case E_FUNCTION_CALL:
		return fmt.Sprint("(",e.Data[0]," (",e.Data[1:],") )")
	case E_CONDITIONAL:
		return fmt.Sprint("(",e.Data[0],"?",e.Data[1],":",e.Data[2],")")
	case E_CAST:
		return fmt.Sprint("((",e.Data[0],")",e.Data[1],")")
	case E_INDEX:
		return fmt.Sprint(e.Data[0],"[",e.Data[1],"]")
	}
	return fmt.Sprint("(",e.Type,"#",e.Text,e.Data,")")
}

func c_expr_list(p *parser.Parser,tokens *scanlist.Element, sep rune, r []interface{}) parser.ParserResult {
	sub := p.Match("Expr",tokens)
	if sub.Result!=parser.RESULT_OK { return sub }
	r = append(r,sub.Data)
	for sub.Next.SafeToken() == sep {
		sub = p.Match("Expr",sub.Next.Next())
		if sub.Result!=parser.RESULT_OK { return sub }
		r = append(r,sub.Data)
	}
	sub.Data = r
	return sub
}
func c_expr_cast(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	if tokens.SafeToken() != '(' /*)*/ { return parser.ResultFail("No match, next rule!",tokens.SafePos()) }
	tp := p.Match("Type",tokens.Next())
	if tp.Result!=parser.RESULT_OK { return tp }
	
	e,t := parser.Match(parser.Textify,tp.Next,/*(*/')')
	if e!=nil { return parser.ResultFail(fmt.Sprint(e),tp.Next.SafePos()) }
	sub := p.MatchNoLeftRecursion("Expr2",t)
	if sub.Result==parser.RESULT_OK {
		sub.Data = &Expr{E_CAST,"cast",aR(tp.Data,sub.Data),tokens.Pos}
	}
	return sub
}

func c_expr0(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	if tokens==nil { return parser.ResultFail("EOF!",scanner.Position{}) }
	switch tokens.Token {
	case scanner.Ident: return parser.ResultOk(tokens.Next(),&Expr{E_VAR,tokens.TokenText,nil,tokens.Pos})
	case scanner.Int: return parser.ResultOk(tokens.Next(),&Expr{E_INT,tokens.TokenText,nil,tokens.Pos})
	case scanner.Float: return parser.ResultOk(tokens.Next(),&Expr{E_FLOAT,tokens.TokenText,nil,tokens.Pos})
	case scanner.Char: return parser.ResultOk(tokens.Next(),&Expr{E_CHAR,tokens.TokenText,nil,tokens.Pos})
	case scanner.String,scanner.RawString: return parser.ResultOk(tokens.Next(),&Expr{E_STRING,tokens.TokenText,nil,tokens.Pos})
	case '*','+','-','!','~','&':{
		sub := p.MatchNoLeftRecursion("Expr0",tokens.Next())
		if sub.Result==parser.RESULT_OK {
			sub.Data = &Expr{E_UNARY_OP,tokens.TokenText,aR(sub.Data),tokens.Pos}
		}
		return sub
	    }
	case '(': /*)*/{
		sub := p.Match("Expr",tokens.Next())
		if sub.Result==parser.RESULT_OK {/*(*/
			e,t := parser.Match(parser.Textify,sub.Next,')')
			if e!=nil { return parser.ResultFail(fmt.Sprint(e),sub.Next.SafePos()) }
			sub.Next = t
		}
		return sub
	    }
	}
	return parser.ResultFail("Invalid Expression!",tokens.Pos)
}
func c_expr_trailer0(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	if ok,t := parser.FastMatch(tokens,'+','+'); ok {
		return parser.ResultOk(t,&Expr{E_INCR,"++",aR(left),tokens.Pos})
	}
	if ok,t := parser.FastMatch(tokens,'-','-'); ok {
		return parser.ResultOk(t,&Expr{E_DECR,"--",aR(left),tokens.Pos})
	}
	if ok,t := parser.FastMatch(tokens,'-','>',scanner.Ident); ok {
		return parser.ResultOk(t,&Expr{E_FIELD_PTR,tokens.Next().Next().TokenText,aR(left),tokens.Pos})
	}
	if ok,t := parser.FastMatch(tokens,'.',scanner.Ident); ok {
		return parser.ResultOk(t,&Expr{E_FIELD_DOT,tokens.Next().TokenText,aR(left),tokens.Pos})
	}
	if tokens.SafeToken()=='(' /*)*/ {
		if tokens.Next().SafeToken() == /*(*/')' { return parser.ResultOk(tokens.Next().Next(),&Expr{E_FUNCTION_CALL,"()",aR(left),tokens.Pos}) }
		sub := c_expr_list(p,tokens.Next(),',',aR(left))
		if sub.Result==parser.RESULT_OK {/*(*/
			e,t := parser.Match(parser.Textify,sub.Next,')')
			if e!=nil { return parser.ResultFail(fmt.Sprint(e),sub.Next.SafePos()) }
			sub.Next = t
			sub.Data = &Expr{E_FUNCTION_CALL,"()",sub.Data.([]interface{}),tokens.Pos}
		}
		return sub
	}
	if tokens.SafeToken()=='[' /*]*/ {
		sub := p.Match("Expr",tokens.Next())
		if sub.Result==parser.RESULT_OK {
			e,t := parser.Match(parser.Textify,sub.Next,/*[*/']')
			if e!=nil { return parser.ResultFail(fmt.Sprint(e),sub.Next.SafePos()) }
			sub.Next = t
			sub.Data = &Expr{E_INDEX,"[]",aR(left,sub.Data),tokens.Pos}
		}
		return sub
	}
	return parser.ResultFail("No trailer.",tokens.SafePos())
}

func c_expr1(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	if tokens==nil { return parser.ResultFail("EOF!",scanner.Position{}) }
	switch tokens.Token {
	case '*','+','-','!','~','&':{
		sub := p.MatchNoLeftRecursion("Expr1",tokens.Next())
		if sub.Result==parser.RESULT_OK {
			sub.Data = &Expr{E_UNARY_OP,tokens.TokenText,aR(sub.Data),tokens.Pos}
		}
		return sub
	    }
	}
	return p.Match("Expr0",tokens)
}
func c_expr2(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	return p.Match("Expr1",tokens)
}
func c_expr3(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	return p.Match("Expr2",tokens)
}
func c_expr_trailer3(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	ok,t := parser.FastMatch(tokens,'*')
	if !ok { ok,t = parser.FastMatch(tokens,'/') }
	if !ok { ok,t = parser.FastMatch(tokens,'%') }
	
	if ok {
		s := tokens.TokenText
		op := E_BINARY_OP
		if t.SafeToken()=='=' {
			op = E_BINARY_OP_ASSIGN
			t = t.SafeNext()
		}
		sub := p.MatchNoLeftRecursion("Expr3",t)
		if sub.Result==parser.RESULT_OK {
			sub.Data = &Expr{op,s,aR(left,sub.Data),tokens.Pos}
		}
		return sub
	}
	return parser.ResultFail("No trailer.",tokens.SafePos())
}

func c_expr4(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	return p.Match("Expr3",tokens)
}
func c_expr_trailer4(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	ok,t := parser.FastMatch(tokens,'+')
	if !ok { ok,t = parser.FastMatch(tokens,'-') }
	
	if ok {
		s := tokens.TokenText
		op := E_BINARY_OP
		if t.SafeToken()=='=' {
			op = E_BINARY_OP_ASSIGN
			t = t.SafeNext()
		}
		sub := p.MatchNoLeftRecursion("Expr4",t)
		if sub.Result==parser.RESULT_OK {
			sub.Data = &Expr{op,s,aR(left,sub.Data),tokens.Pos}
		}
		return sub
	}
	return parser.ResultFail("No trailer.",tokens.SafePos())
}

func c_expr5(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	return p.Match("Expr4",tokens)
}
func c_expr_trailer5(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	var ok bool
	var t  *scanlist.Element
	ok = false
	s := ""
	if !ok { ok,t = parser.FastMatch(tokens,'>','>'); if ok { s=">>" } }
	if !ok { ok,t = parser.FastMatch(tokens,'<','<'); if ok { s="<<" } }
	if !ok { ok,t = parser.FastMatch(tokens,'^') }
	if !ok { ok,t = parser.FastMatch(tokens,'|') }
	if !ok { ok,t = parser.FastMatch(tokens,'&') }
	
	if ok {
		if s=="" { s = tokens.TokenText }
		op := E_BINARY_OP
		if t.SafeToken()=='=' {
			op = E_BINARY_OP_ASSIGN
			t = t.SafeNext()
		}
		sub := p.MatchNoLeftRecursion("Expr5",t)
		if sub.Result==parser.RESULT_OK {
			sub.Data = &Expr{op,s,aR(left,sub.Data),tokens.Pos}
		}
		return sub
	}
	return parser.ResultFail("No trailer.",tokens.SafePos())
}
func c_expr6(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	return p.Match("Expr5",tokens)
}
func c_expr_trailer6(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	s := ""
	ok,t := parser.FastMatch(tokens,'=','='); if ok { s="==" }
	if !ok { ok,t = parser.FastMatch(tokens,'!','='); if ok { s="!=" } }
	if !ok { ok,t = parser.FastMatch(tokens,'<','='); if ok { s="<=" } }
	if !ok { ok,t = parser.FastMatch(tokens,'<'); if ok { s="<" } }
	if !ok { ok,t = parser.FastMatch(tokens,'>','='); if ok { s=">=" } }
	if !ok { ok,t = parser.FastMatch(tokens,'>'); if ok { s=">" } }
	if ok {
		sub := p.MatchNoLeftRecursion("Expr6",t)
		if sub.Result==parser.RESULT_OK {
			sub.Data = &Expr{E_COMPARE,s,aR(left,sub.Data),tokens.Pos}
		}
		return sub
	}
	return parser.ResultFail("No trailer.",tokens.SafePos())
}

func c_expr7(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	return p.Match("Expr6",tokens)
}
func c_expr_trailer7(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	var t *scanlist.Element
	ok := false
	s := ""
	if !ok { ok,t = parser.FastMatch(tokens,'&','&'); if ok { s="&&" } }
	if !ok { ok,t = parser.FastMatch(tokens,'|','|'); if ok { s="||" } }
	
	if ok {
		if s=="" { s = tokens.TokenText }
		op := E_BINARY_OP
		if t.SafeToken()=='=' {
			op = E_BINARY_OP_ASSIGN
			t = t.SafeNext()
		}
		sub := p.MatchNoLeftRecursion("Expr7",t)
		if sub.Result==parser.RESULT_OK {
			sub.Data = &Expr{op,s,aR(left,sub.Data),tokens.Pos}
		}
		return sub
	}
	return parser.ResultFail("No trailer.",tokens.SafePos())
}

func c_expr8(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	return p.Match("Expr7",tokens)
}
func c_expr_trailer8(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	if tokens.SafeToken()=='?' {
		sub := p.Match("Expr7",tokens.Next())
		if sub.Result!=parser.RESULT_OK { return sub }
		e,t := parser.Match(parser.Textify,sub.Next,':')
		if e!=nil { return parser.ResultFail(fmt.Sprint(e),sub.Next.SafePos()) }
		sub2 := p.MatchNoLeftRecursion("Expr7",t)
		if sub2.Result!=parser.RESULT_OK { return sub2 }
		return parser.ResultOk(sub2.Next,&Expr{E_CONDITIONAL,"?:",aR(left,sub.Data,sub2.Data),tokens.Pos})
	}
	return parser.ResultFail("No trailer.",tokens.SafePos())
}

func c_expr(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	return p.Match("Expr8",tokens)
}
func c_expr_trailer(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	ok,t := parser.FastMatch(tokens,'=')
	
	if ok {
		sub := p.Match("Expr",t)
		if sub.Result==parser.RESULT_OK {
			sub.Data = &Expr{E_ASSIGN,"=",aR(left,sub.Data),tokens.Pos}
		}
		return sub
	}
	
	return parser.ResultFail("No trailer.",tokens.SafePos())
}

/*
The following expression classes are registered.
	'Expr0' // Primary Expression + Suffix
	'Expr1' // Unary Expression
	'Expr2' // Cast Expression
	'Expr3' // Multiplicative Expression
	'Expr4' // Additive Expression
	'Expr5' // Bitwise expression
	'Expr6' // Relational expression
	'Expr7' // Logical expression
	'Expr8' // Conditional expression
	'Expr'  // Expression
*/
func RegisterExpr(p *parser.Parser) {
	p.Define("Expr0",false,parser.Pfunc(c_expr0))
	p.Define("Expr0",true,parser.Pfunc(c_expr_trailer0))
	
	p.Define("Expr1",false,parser.Pfunc(c_expr1))
	p.Define("Expr2",false,parser.Pfunc(c_expr2))
	p.Define("Expr3",false,parser.Pfunc(c_expr3))
	p.Define("Expr3",true,parser.Pfunc(c_expr_trailer3))
	p.Define("Expr4",false,parser.Pfunc(c_expr4))
	p.Define("Expr4",true,parser.Pfunc(c_expr_trailer4))
	p.Define("Expr5",false,parser.Pfunc(c_expr5))
	p.Define("Expr5",true,parser.Pfunc(c_expr_trailer5))
	p.Define("Expr6",false,parser.Pfunc(c_expr6))
	p.Define("Expr6",true,parser.Pfunc(c_expr_trailer6))
	p.Define("Expr7",false,parser.Pfunc(c_expr7))
	p.Define("Expr7",true,parser.Pfunc(c_expr_trailer7))
	p.Define("Expr8",false,parser.Pfunc(c_expr8))
	p.Define("Expr8",true,parser.Pfunc(c_expr_trailer8))
	
	p.Define("Expr",false,parser.Pfunc(c_expr))
	p.Define("Expr",true,parser.Pfunc(c_expr_trailer))
}

/*
Should be called only after RegisterExpr()!
Adds the casting operation (Type)Expr.
*/
func RegisterExprCast(p *parser.Parser) {
	p.TouchRule("Type")
	p.DefineBefore("Expr2",false,parser.Pfunc(c_expr_cast))
}

