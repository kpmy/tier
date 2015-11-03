package tier

import (
	"fmt"
	"github.com/kpmy/ypk/assert"
)

type Runner interface {
	Next() Symbol
	//expect is the most powerful step forward runner, breaks the compilation if unexpected sym found
	Expect(SymCode, interface{}, ...SymCode)
	//await runs for the sym through skip list, but may not find the sym
	Await(SymCode, ...SymCode) bool
	//pass runs through skip list
	Pass(...SymCode)
	//run runs to the first sym through any other sym
	Run(SymCode)
	//Is current symbol?
	Is(SymCode) bool
}

type rn struct {
	sc    Scanner
	done  bool
	debug bool
	sym   Symbol

	marker Marker
}

func (r *rn) Next() Symbol {
	r.done = true
	r.sym = r.sc.Get()
	if r.debug {
		r.marker.Log("`" + fmt.Sprint(r.sym) + "`")
	}
	return r.sym
}

//expect is the most powerful step forward runner, breaks the compilation if unexpected sym found
func (r *rn) Expect(sym SymCode, msg interface{}, skip ...SymCode) {
	assert.For(r.done, 20)
	if !r.Await(sym, skip...) {
		r.marker.Mark(msg)
	}
	r.done = false
}

//await runs for the sym through skip list, but may not find the sym
func (r *rn) Await(sym SymCode, skip ...SymCode) bool {
	assert.For(r.done, 20)
	skipped := func() (ret bool) {
		for _, v := range skip {
			if v == r.sym.Code {
				ret = true
			}
		}
		return
	}

	for sym != r.sym.Code && skipped() && r.sc.Error() == nil {
		r.Next()
	}
	r.done = r.sym.Code != sym
	return r.sym.Code == sym
}

//pass runs through skip list
func (r *rn) Pass(skip ...SymCode) {
	skipped := func() (ret bool) {
		for _, v := range skip {
			if v == r.sym.Code {
				ret = true
			}
		}
		return
	}
	for skipped() && r.sc.Error() == nil {
		r.Next()
	}
}

//run runs to the first sym through any other sym
func (r *rn) Run(sym SymCode) {
	if r.sym.Code != sym {
		for r.sc.Error() == nil && r.Next().Code != sym {
			if r.sc.Error() != nil {
				r.marker.Mark(sym, " not found")
				break
			}
		}
	}
}

func (r *rn) Is(sym SymCode) bool {
	return r.sym.Code == sym
}

func NewRunner(s Scanner, m Marker) Runner {
	ret := &rn{}
	ret.sc = s
	ret.marker = m
	ret.Next()
	return ret
}

func Debug(r Runner) Runner {
	if rr, ok := r.(*rn); ok {
		rr.debug = !rr.debug
	}
	return r
}
