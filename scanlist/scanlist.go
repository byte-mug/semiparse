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

/*
A Lazy-Evaluating Linked list around the "text/scanner" package.
*/
package scanlist

import "text/scanner"

type TokenDict map[string]rune
func (t TokenDict) Get(s string, r rune) rune {
	if t==nil { return r }
	n,ok := t[s]
	if ok { return n }
	return r
}
func (t TokenDict) Join(o TokenDict) TokenDict {
	if o==nil { return t }
	if t==nil { return o }
	n := make(TokenDict)
	for k,v := range t { n[k]=v }
	for k,v := range o { n[k]=v }
	return n
}

type BaseScanner struct{
	scanner.Scanner
	Dict TokenDict
	Concat *Element
}
func (b *BaseScanner) Next() *Element {
	t := b.Scan()
	if t==scanner.EOF { return b.Concat }
	s := b.TokenText()
	e := new(Element)
	e.Token = b.Dict.Get(s,t)
	e.TokenText = s
	e.Pos = b.Pos()
	e.bs = b
	return e
}

type Element struct {
	Token rune
	TokenText string
	Pos scanner.Position
	bs *BaseScanner
	e  *Element
}
func (e *Element) Next() *Element {
	if e.bs==nil { return e.e }
	e.e = e.bs.Next()
	e.bs = nil
	return e.e
}


