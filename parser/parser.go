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

const NONE = uint(0)
const (
	LEFT_RECURSIVE = uint(1<<iota)
)

const (
	RESULT_OK = uint(iota)
	RESULT_FAILED
	RESULT_FAILED_CUT
)

type ParserResult struct{
	Result uint
	Next *scanlist.Element // next token on success; undefined on failure.
	Data interface{} // The syntax-tree on success; the error message on failure.
	Pos scanner.Position
}
func ResultOk(next *scanlist.Element,tree interface{}) ParserResult {
	return ParserResult{RESULT_OK,next,tree,scanner.Position{}}
}
func ResultFail(reason string, pos scanner.Position) ParserResult {
	return ParserResult{RESULT_FAILED,nil,reason,pos}
}
func ResultFailCut(reason string, pos scanner.Position) ParserResult {
	return ParserResult{RESULT_FAILED_CUT,nil,reason,pos}
}

type ParseRule interface{
	// left = the Left-Recursive element, if any, else nil
	Parse(p *Parser,tokens *scanlist.Element, left interface{}) ParserResult
}

type Pfunc func(p *Parser,tokens *scanlist.Element, left interface{}) ParserResult
func (pf Pfunc) Parse(p *Parser,tokens *scanlist.Element, left interface{}) ParserResult {
	return pf(p,tokens,left)
}

type OR []ParseRule
func (o OR) Parse(p *Parser,tokens *scanlist.Element, left interface{}) (opr ParserResult) {
	fail := false
	for _,r := range o {
		npr := r.Parse(p,tokens,left)
		switch npr.Result {
		case RESULT_OK: return npr
		case RESULT_FAILED:
			opr = npr
			fail = true
		case RESULT_FAILED_CUT:
			return npr
		}
	}
	if fail { return }
	return ResultFail("no rules!",tokens.Pos)
}

// (Inner)*
type LStar struct {
	Inner ParseRule
}
func (s LStar) Parse(p *Parser,tokens *scanlist.Element, left interface{}) (opr ParserResult) {
	opr = ResultOk(tokens,left)
	for {
		npr := s.Inner.Parse(p,tokens,left)
		switch npr.Result {
		case RESULT_FAILED:
			return
		case RESULT_FAILED_CUT: return npr
		}
		opr = npr
		left = npr.Data
		tokens = npr.Next
	}
}
// (Inner)+
type LPlus struct {
	Inner ParseRule
}
func (s LPlus) Parse(p *Parser,tokens *scanlist.Element, left interface{}) (opr ParserResult) {
	opr = s.Inner.Parse(p,tokens,left)
	if opr.Result!=RESULT_OK { return }
	tokens = opr.Next
	for {
		npr := s.Inner.Parse(p,tokens,left)
		switch npr.Result {
		case RESULT_FAILED:
			return
		case RESULT_FAILED_CUT: return npr
		}
		opr = npr
		left = npr.Data
		tokens = npr.Next
	}
}

// (Inner)* => ARRAY
type ArrayStar struct {
	Inner ParseRule
}
func (s ArrayStar) Parse(p *Parser,tokens *scanlist.Element, left interface{}) (ParserResult) {
	dok := []interface{}{}
	for {
		npr := s.Inner.Parse(p,tokens,nil)
		switch npr.Result {
		case RESULT_FAILED:
			return ResultOk(tokens,dok)
		case RESULT_FAILED_CUT: return npr
		}
		dok = append(dok,npr.Data)
		tokens = npr.Next
	}
	panic("unreachable")
}

// (Inner)+ => ARRAY
type ArrayPlus struct {
	Inner ParseRule
}
func (s ArrayPlus) Parse(p *Parser,tokens *scanlist.Element, left interface{}) (ParserResult) {
	npr := s.Inner.Parse(p,tokens,nil)
	if npr.Result!=RESULT_OK { return npr }
	tokens = npr.Next
	dok := []interface{}{npr.Data}
	for {
		npr := s.Inner.Parse(p,tokens,nil)
		switch npr.Result {
		case RESULT_FAILED:
			return ResultOk(tokens,dok)
		case RESULT_FAILED_CUT: return npr
		}
		dok = append(dok,npr.Data)
		tokens = npr.Next
	}
}


type LSeq []ParseRule
func (s LSeq) Parse(p *Parser,tokens *scanlist.Element, left interface{}) (opr ParserResult) {
	opr = ResultOk(tokens,left)
	for _,r := range s {
		if opr.Result!=RESULT_OK { break }
		opr = r.Parse(p,tokens,opr.Data)
		tokens = opr.Next
	}
	return
}

type ArraySeq []ParseRule
func (s ArraySeq) Parse(p *Parser,tokens *scanlist.Element, left interface{}) (opr ParserResult) {
	arr := make([]interface{},len(s))
	for i,r := range s {
		opr = r.Parse(p,tokens,left)
		if opr.Result!=RESULT_OK { return }
		tokens = opr.Next
		arr[i] = opr.Data
	}
	opr.Data = arr
	return
}

type ruleParser struct{
	phase1 OR
	phase2 OR
}

type Parser struct{
	rules map[string]*ruleParser
}
func (p *Parser) Construct() *Parser {
	p.rules = make(map[string]*ruleParser)
	return p
}
func (p *Parser) Define(n string,left bool,r ParseRule) {
	rp,ok := p.rules[n]
	if !ok {
		rp = new(ruleParser)
		p.rules[n] = rp
	}
	if left {
		rp.phase2 = append(rp.phase2,r)
	} else {
		rp.phase1 = append(rp.phase1,r)
	}
}

// Like .Define(), but does prepend rather than append!
func (p *Parser) DefineBefore(n string,left bool,r ParseRule) {
	rp,ok := p.rules[n]
	if !ok {
		rp = new(ruleParser)
		p.rules[n] = rp
	}
	if left {
		rp.phase2 = append(OR{r},rp.phase2...)
	} else {
		rp.phase1 = append(OR{r},rp.phase1...)
	}
}
func (p *Parser) TouchRule(n string) {
	_,ok := p.rules[n]
	if !ok {
		p.rules[n] = new(ruleParser)
	}
}
func (p *Parser) matchLowLevel(n string,phaseTwo bool,tokens *scanlist.Element) ParserResult {
	rp,ok := p.rules[n]
	if !ok { panic("rule not defined") }
	r1 := rp.phase1.Parse(p,tokens,nil)
	if r1.Result != RESULT_OK { return r1 }
	if !phaseTwo { return r1 }
	r2 := LStar{rp.phase2}.Parse(p,r1.Next,r1.Data)
	return r2
}
func (p *Parser) Match(n string,tokens *scanlist.Element) ParserResult {
	return p.matchLowLevel(n,true,tokens)
}
func (p *Parser) MatchNoLeftRecursion(n string,tokens *scanlist.Element) ParserResult {
	return p.matchLowLevel(n,false,tokens)
}


