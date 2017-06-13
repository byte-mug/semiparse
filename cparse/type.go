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
	t_none = uint(iota)
	T_NAME // Identifiers like "int", "float" or "string" etc...
	T_CONST // Constant value
	T_PTR // Pointer
)

type DType struct{
	Type uint
	Text string
	Data []interface{}
	Pos scanner.Position
}
func (d *DType) String() string {
	if d==nil { return "NIL" }
	switch d.Type {
	case T_NAME:
		return d.Text
	case T_CONST:
		return fmt.Sprint(d.Data[0]," const")
	case T_PTR:
		return fmt.Sprint(d.Data[0],"*")
	}
	return fmt.Sprint("(",d.Type,"#",d.Text,d.Data,")")
}
/*
func c_t_isspec(v interface{}, spec uint) bool {
	d,ok := v.(*DType)
	if !ok { return false }
	return d.Type == spec
}
*/

func c_type(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	switch tokens.SafeToken() {
	case scanner.Ident: return parser.ResultOk(tokens.Next(),&DType{T_NAME,tokens.TokenText,nil,tokens.Pos})
	case C_CONST:
		sub := p.MatchNoLeftRecursion("Type",tokens.Next())
		if sub.Result==parser.RESULT_OK {
			sub.Data = &DType{T_CONST,tokens.TokenText,aR(sub.Data),tokens.Pos}
		}
		return sub
	}
	return parser.ResultFail("Invalid Type!",tokens.Pos)
}

func c_type_trailer(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	switch tokens.SafeToken() {
	case C_CONST:
		return parser.ResultOk(tokens.Next(),&DType{T_CONST,tokens.TokenText,aR(left),tokens.Pos})
	case '*':
		return parser.ResultOk(tokens.Next(),&DType{T_PTR,tokens.TokenText,aR(left),tokens.Pos})
	}
	return parser.ResultFail("Invalid Type!",tokens.Pos)
}

func RegisterType(p *parser.Parser) {
	p.Define("Type",false,parser.Pfunc(c_type))
	p.Define("Type",true,parser.Pfunc(c_type_trailer))
}

