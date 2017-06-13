# Parser


```go
package main

import "github.com/byte-mug/semiparse/scanlist"
import "github.com/byte-mug/semiparse/parser"
import "github.com/byte-mug/semiparse/cparse"
import "github.com/byte-mug/semiparse/ecparse"
import "strings"
import "fmt"
//import "text/scanner"

const src = `
//object.getX:int(arg);
i[].Array;
//a = b + c;
`


func buildParser() *parser.Parser {
	p := new(parser.Parser).Construct()
	cparse.RegisterExpr(p)
	cparse.RegisterType(p)
	cparse.RegisterExprCast(p)
	ecparse.RegisterExprOCX(p)
	return p
}

func main() {
	s := new(scanlist.BaseScanner)
	s.Init(strings.NewReader(src))
	s.Dict = cparse.CKeywords
	l := s.Next()
	p := buildParser()
	res := p.Match("Expr",l)
	fmt.Println(res.Result)
	fmt.Println(res.Data)
}

```