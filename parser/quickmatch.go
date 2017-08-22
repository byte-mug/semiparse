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
A Parser library and framework. Left-recursive rules are not supported without
special care.
*/
package parser

import "text/scanner"
import "github.com/byte-mug/semiparse/scanlist"
import "fmt"

// Textifies a Token-ID
func Textify(r rune) string {
	switch r {
	case scanner.EOF: return "<<EOF>>"
	case scanner.Ident: return "<<Ident>>"
	case scanner.Int: return "<<Int>>"
	case scanner.Float: return "<<Float>>"
	case scanner.Char: return "<<Char>>"
	case scanner.String: return "<<String>>"
	case scanner.RawString: return "<<RawString>>"
	case scanner.Comment: return "<<Comment>>"
	}
	if r>0 { return fmt.Sprintf("'%c'",r) }
	return fmt.Sprintf("#%d",r)
}

var unexpected_eof = fmt.Errorf("Unexpected End-Of-File (EOF)")
var unexpected_syntax = fmt.Errorf("Unexpected Syntax")
func Match(f func(rune) string,t *scanlist.Element,rs ...rune) (error,*scanlist.Element) {
	for _,r := range rs {
		if t==nil {
			if f==nil { return unexpected_eof,t }
			return fmt.Errorf("Unexpected End-Of-File (EOF), expected %s",f(r)),t
		}
		if t.Token!=r {
			if f==nil { return unexpected_syntax,t }
			return fmt.Errorf("Unexpected %s, expected %s",f(t.Token),f(r)),t
		}
		t = t.Next()
	}
	return nil,t
}

func FastMatch(t *scanlist.Element,rs ...rune) (bool,*scanlist.Element) {
	for _,r := range rs {
		if t==nil {
			return false,t
		}
		if t.Token!=r {
			return false,t
		}
		t = t.Next()
	}
	return true,t
}
