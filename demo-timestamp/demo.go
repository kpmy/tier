package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/kpmy/tier"
	"io"
	"log"
)

const (
	Dash tier.SymCode = iota + tier.UserIota
	Z
	Colon
	Plus

	Minus = Dash
)

type Parser struct {
	tier.Marker
	r tier.Runner
}

func (p *Parser) FutureMark() tier.Marker { panic(126) }

func (p *Parser) Mark(msg ...interface{}) {
	panic(fmt.Sprint(msg...))
}

func (p *Parser) Log(msg ...interface{}) {
	log.Println(msg...)
}

func (p *Parser) ConnectTo(rd io.Reader) {
	sc := tier.NewScanner(bufio.NewReader(rd), opts)
	p.r = tier.NewRunner(sc, p)
	//tier.Debug(p.r)
}

var opts tier.Opts

func parse(ts string) {
	p := &Parser{}
	p.ConnectTo(bytes.NewBufferString(ts))
	log.Println(p.r.Assert(tier.Number, "number expected").Value)
	p.r.Assert(Dash, "dash expected")
	log.Println(p.r.Assert(tier.Number, "number expected").Value)
	p.r.Assert(Dash, "dash expected")
	log.Println(p.r.Assert(tier.Number, "number expected").Value)

	offset := ""
	getOffset := func() (offset string) {
		offset = p.r.Assert(tier.Number, "number expected").Value
		p.r.Assert(Colon, "colon expected")
		offset = offset + ":" + p.r.Assert(tier.Number, "number expected").Value
		return
	}

	if p.r.Await(Minus) {
		p.r.Next()
		offset = "-" + getOffset()
	} else if p.r.Is(Plus) {
		p.r.Next()
		offset = "+" + getOffset()
	} else if p.r.Is(Z) {
		offset = "Z"
	}

	if offset != "" {
		log.Println(offset)
	}
	log.Println()
}

func init() {
	log.SetFlags(0)

	opts = tier.DefaultOpts
	opts.CombinedMap["-"] = Dash
	opts.CombinedMap["+"] = Plus
	opts.CombinedMap[":"] = Colon

	opts.IdentMap["Z"] = Z
}

//xml datetime format
func main() {
	parse("2015-11-05")
	parse("2015-11-05Z")
	parse("2015-11-05+04:00")
	parse("2015-11-05-01:00")
}
