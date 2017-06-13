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

/*Extensions to the C-like parser.*/
package ecparse


import "github.com/byte-mug/semiparse/scanlist"
import "github.com/byte-mug/semiparse/parser"
import "text/scanner"
import . "github.com/byte-mug/semiparse/cparse"
import "fmt"


const (
	e_none = uint(iota+100)
	
	// Objekt C eXtensions.
	E_TA_FIELD_DOT // Expr.Field:Type
)

func aR(i ...interface{}) []interface{} { return i }

// Objekt C eXtensions.
func c_expr_trailer_ocx(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	if ok,t := parser.FastMatch(tokens,'.',scanner.Ident,':'); ok {
		tp := p.Match("Type",t)
		if tp.Result!=parser.RESULT_OK { return tp }
		return parser.ResultOk(tp.Next,&Expr{E_TA_FIELD_DOT,tokens.Next().TokenText,aR(left),tokens.Pos})
	}
	
	// parses "Expr.(Type)" as "(Type)Expr".
	if ok,t := parser.FastMatch(tokens,'.','(' /*)*/); ok {
		tp := p.Match("Type",t)
		if tp.Result!=parser.RESULT_OK { return tp }
		e,t := parser.Match(parser.Textify,tp.Next,/*(*/')')
		if e!=nil { return parser.ResultFail(fmt.Sprint(e),tp.Next.SafePos()) }
		return parser.ResultOk(t,&Expr{E_CAST,"cast",aR(tp.Data,left),tokens.Pos})
	}
	{
		/*
		 * Parses "Expr.*" as "*Expr", "Expr.-" as "-Expr", "Expr.!" as "!Expr", etc...
		 */
		ok,t := parser.FastMatch(tokens,'.','*')
		if !ok { ok,t = parser.FastMatch(tokens,'.','+') }
		if !ok { ok,t = parser.FastMatch(tokens,'.','-') }
		if !ok { ok,t = parser.FastMatch(tokens,'.','!') }
		if !ok { ok,t = parser.FastMatch(tokens,'.','~') }
		if !ok { ok,t = parser.FastMatch(tokens,'.','&') }
		if ok {
			return parser.ResultOk(t,&Expr{E_UNARY_OP,tokens.Next().TokenText,aR(left),tokens.Pos})
		}
	}
	
	// XXX: Is this crap really needed in OCX?
	if ok,t := parser.FastMatch(tokens,'[',']','.'); ok {
		sub := p.MatchNoLeftRecursion("Expr",t)
		if sub.Result!=parser.RESULT_OK { return sub }
		sub.Data = &Expr{E_INDEX,"[]",aR(sub.Data,left),tokens.Pos}
		return sub
	}
	
	return parser.ResultFail("No match, next rule!",tokens.SafePos())
}

func RegisterExprOCX(p *parser.Parser) {
	p.DefineBefore("Expr",true,parser.Pfunc(c_expr_trailer_ocx))
}

