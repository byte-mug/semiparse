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
	E_FIELD
	E_UNARY_OP
	E_BINARY_OP
	E_BINARY_OP_ASSIGN
	E_COMPARE
	E_ASSIGN
)

func aR(i ...interface{}) []interface{} { return i }

type Expr struct{
	Type uint
	Text string
	Data []interface{}
}
func (e *Expr) String() string {
	if e==nil { return "NIL" }
	switch e.Type {
	case E_BINARY_OP:
		return fmt.Sprint("(",e.Data[0],e.Text,e.Data[1],")")
	case E_BINARY_OP_ASSIGN:
		return fmt.Sprint("(",e.Data[0],e.Text,"=",e.Data[1],")")
	case E_INCR,E_DECR:
		return fmt.Sprint("(",e.Data,e.Text,")")
	case E_FIELD:
		return fmt.Sprint("(",e.Data,"->",e.Text,")")
	case E_UNARY_OP:
		return fmt.Sprintf("(",e.Text,e.Data,")")
	case E_VAR,E_INT,E_FLOAT,E_CHAR,E_STRING:
		return fmt.Sprint(e.Text)
	case E_ASSIGN:
		return fmt.Sprint("(",e.Data[0],"=",e.Data[1],")")
	}
	return fmt.Sprint("(",e.Type,"#",e.Text,e.Data,")")
}


func c_expr(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	if tokens==nil { return parser.ResultFail("EOF!",scanner.Position{}) }
	switch tokens.Token {
	case scanner.Ident: return parser.ResultOk(tokens.Next(),&Expr{E_VAR,tokens.TokenText,nil})
	case scanner.Int: return parser.ResultOk(tokens.Next(),&Expr{E_INT,tokens.TokenText,nil})
	case scanner.Float: return parser.ResultOk(tokens.Next(),&Expr{E_FLOAT,tokens.TokenText,nil})
	case scanner.Char: return parser.ResultOk(tokens.Next(),&Expr{E_CHAR,tokens.TokenText,nil})
	case scanner.String,scanner.RawString: return parser.ResultOk(tokens.Next(),&Expr{E_STRING,tokens.TokenText,nil})
	case '*','+','-','!','~':{
		sub := p.MatchNoLeftRecursion("Expr",tokens.Next())
		if sub.Result==parser.RESULT_OK {
			sub.Data = &Expr{E_UNARY_OP,tokens.TokenText,aR(sub.Data)}
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
func c_expr_trailer(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	if ok,t := parser.FastMatch(tokens,'+','+'); ok {
		return parser.ResultOk(t,&Expr{E_INCR,"++",aR(left)})
	}
	if ok,t := parser.FastMatch(tokens,'-','-'); ok {
		return parser.ResultOk(t,&Expr{E_DECR,"--",aR(left)})
	}
	if ok,t := parser.FastMatch(tokens,'-','>',scanner.Ident); ok {
		return parser.ResultOk(t,&Expr{E_FIELD,tokens.Next().Next().TokenText,aR(left)})
	}
	/*
	TODO: function(call);
	*/
	
	ok,t := parser.FastMatch(tokens,'+')
	s := ""
	if !ok { ok,t = parser.FastMatch(tokens,'-') }
	if !ok { ok,t = parser.FastMatch(tokens,'*') }
	if !ok { ok,t = parser.FastMatch(tokens,'/') }
	if !ok { ok,t = parser.FastMatch(tokens,'%') }
	if !ok { ok,t = parser.FastMatch(tokens,'>','>'); if ok { s=">>" } }
	if !ok { ok,t = parser.FastMatch(tokens,'<','<'); if ok { s="<<" } }
	if !ok { ok,t = parser.FastMatch(tokens,'&','&'); if ok { s="&&" } }
	if !ok { ok,t = parser.FastMatch(tokens,'|','|'); if ok { s="||" } }
	if !ok { ok,t = parser.FastMatch(tokens,'^') }
	
	if ok {
		if s=="" { s = tokens.TokenText }
		op := E_BINARY_OP
		if t.SafeToken()=='=' {
			op = E_BINARY_OP_ASSIGN
			t = t.SafeNext()
		}
		sub := p.MatchNoLeftRecursion("Expr",t)
		if sub.Result==parser.RESULT_OK {
			sub.Data = &Expr{op,s,aR(left,sub.Data)}
		}
		return sub
		/*
		if sub.Result==parser.RESULT_OK {
			sub.Data = &Expr{E_BINARY_OP,tokens.TokenText,aR(left,sub.Data)}
		}
		*/
	}
	
	ok,t = parser.FastMatch(tokens,'=','='); if ok { s="==" }
	if !ok { ok,t = parser.FastMatch(tokens,'!','='); if ok { s="!=" } }
	if !ok { ok,t = parser.FastMatch(tokens,'<','='); if ok { s="<=" } }
	if !ok { ok,t = parser.FastMatch(tokens,'<','<'); if ok { s="<<" } }
	if !ok { ok,t = parser.FastMatch(tokens,'>','='); if ok { s=">=" } }
	if !ok { ok,t = parser.FastMatch(tokens,'>','>'); if ok { s=">>" } }
	if ok {
		sub := p.MatchNoLeftRecursion("Expr",t)
		if sub.Result==parser.RESULT_OK {
			sub.Data = &Expr{E_COMPARE,s,aR(left,sub.Data)}
		}
		return sub
	}
	
	ok,t = parser.FastMatch(tokens,'=')
	
	if ok {
		sub := p.Match("Expr",t)
		if sub.Result==parser.RESULT_OK {
			sub.Data = &Expr{E_ASSIGN,"=",aR(left,sub.Data)}
		}
		return sub
	}
		
	return parser.ResultFail("No trailer.",tokens.SafePos())
}

func RegisterExpr(p *parser.Parser) {
	p.Define("Expr",false,parser.Pfunc(c_expr))
	p.Define("Expr",true,parser.Pfunc(c_expr_trailer))
}

