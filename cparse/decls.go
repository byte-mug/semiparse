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

import "strconv"

type ParamDecl struct{
	Type interface{}
	Name string
}

type DeclProtoFunc struct{
	Type interface{}
	Name string
	Arguments []ParamDecl
	Pos scanner.Position
}

type DeclImplFunc struct{
	Type interface{}
	Name string
	Arguments []ParamDecl
	Body interface{}
	Pos scanner.Position
}
type DeclInclude struct{
	HdrName string
}
func (d *DeclInclude) String() string {
	return fmt.Sprint("#include <",d.HdrName,">")
}

type DeclCType struct{
	Inner string
	CType string
}
func (d *DeclCType) String() string {
	return fmt.Sprint(d.Inner," : ",d.CType)
}

type DeclNone struct{}



func C_DeclFragment_Func(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	ars := parser.ArraySeq{
		parser.Delegate("Type"),
		parser.Required{scanner.Ident,parser.Textify},
		parser.Required{'('/*)*/,parser.Textify},
	}.Parse(p,tokens,left)
	if !ars.Ok() { return ars }
	t := ars.Next
	itr := ars.Data.([]interface{})
	
	args := []ParamDecl{}
	if t.SafeToken()!=/*(*/')' {
		for{
			vd := parser.ArraySeq{
				parser.Delegate("Type"),
				parser.Required{scanner.Ident,parser.Textify},
			}.Parse(p,t,left)
			if !vd.Ok() { return ars }
			itr := vd.Data.([]interface{})
			args = append(args,ParamDecl{itr[0],itr[1].(string)})
			t = vd.Next.SafeNext()
			if vd.Next.SafeToken()==',' { continue }
			if vd.Next.SafeToken()!=/*(*/')' {
				return parser.ResultFail(fmt.Sprintf(/*(*/ "Unexpected %s, expected ',' or ')'",
					parser.Textify(vd.Next.SafeToken())),vd.Next.SafePos())
			}
			break
		}
	}
	return parser.ResultOk(t,&DeclProtoFunc{itr[0],itr[1].(string),args,tokens.Pos})
}

func c_declaration_func(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	f := C_DeclFragment_Func(p,tokens,left)
	if !f.Ok() { return f }
	
	if f.Next.SafeToken() == ';' {
		f.Next=f.Next.Next()
		return f
	}
	if f.Next.SafeToken() != '{'/*}*/ {
		return parser.ResultFail(fmt.Sprintf("Unexpected %s, expected '{'" /*}*/,
					parser.Textify(f.Next.SafeToken())),f.Next.SafePos())
	}
	
	el := p.Match("Statement",f.Next)
	if !el.Ok() { return el }
	
	x := f.Data.(*DeclProtoFunc)
	
	return parser.ResultOk(el.Next,&DeclImplFunc{x.Type,x.Name,x.Arguments,el.Data,x.Pos})
}

func c_declaration_libprep(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	vd := parser.ArraySeq{
		parser.Required{'#',parser.Textify},
		parser.RequireText{"cinclude"},
		parser.Required{scanner.String,parser.Textify},
	}.Parse(p,tokens,left)
	switch {
	case vd.Ok():
		s := vd.Data.([]interface{})[2].(string)
		vd.Data = &DeclInclude{s[1:len(s)-1]}
		return vd
	case vd.Cut(): return vd
	}
	
	vd = parser.ArraySeq{
		parser.Required{'#',parser.Textify},
		parser.RequireText{"ctype"},
		parser.Required{scanner.Ident,parser.Textify},
		parser.OR{
			parser.Required{scanner.String,parser.Textify},
			parser.Required{scanner.RawString,parser.Textify},
		},
	}.Parse(p,tokens,left)
	if vd.Ok() {
		i := vd.Data.([]interface{})
		itype := i[2].(string)
		ctype := i[3].(string)
		
		rct,err := strconv.Unquote(ctype)
		if err!=nil { parser.ResultFail(fmt.Sprint(err),tokens.Next().Next().SafePos()) }
		
		vd.Data = &DeclCType{itype,rct}
		return vd
	}
	
	return parser.ResultFail("Next Rule!",tokens.SafePos())
}

func RegisterDeclaration(p *parser.Parser) {
	p.Define("Declaration",false,parser.Pfunc(c_declaration_func))
	p.Define("Declaration",false,parser.Pfunc(c_declaration_libprep))
}

