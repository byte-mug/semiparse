# Parser


```go

package main

import "github.com/byte-mug/semiparse/scanlist"
import "github.com/byte-mug/semiparse/parser"
import "strings"
import "fmt"
import "text/scanner"

const src = `
A B
C
`
func Expr(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	if tokens!=nil {
		return parser.ResultOk(tokens.Next(),tokens.TokenText)
	}
	return parser.ResultFailCut("EOF!",scanner.Position{})
}
func ExprTrail(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	if tokens!=nil {
		return parser.ResultOk(tokens.Next(),fmt.Sprint("<",left,",",tokens.TokenText,">",))
	}
	return parser.ResultFail("EOF!",scanner.Position{})
}
func buildParser() *parser.Parser {
	p := new(parser.Parser).Construct()
	p.Define("Expr",false,parser.Pfunc(Expr))
	p.Define("Expr",true,parser.Pfunc(ExprTrail))
	return p
}

func main() {
	s := new(scanlist.BaseScanner)
	s.Init(strings.NewReader(src))
	l := s.Next()
	p := buildParser()
	//for ;l!=nil; l = l.Next() {
	//	fmt.Println(l.Token,l.TokenText,l.Pos)
	//}
	fmt.Println(p.Match("Expr",l))
}

```