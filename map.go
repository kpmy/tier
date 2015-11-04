package tier

import (
	"fmt"
)

type MappedSymbol struct {
	Symbol
	Pos         int
	Row, Column int
}

func (m MappedSymbol) String() string {
	return fmt.Sprint(m.Symbol.Code, " ", m.Pos, "[", m.Row, ":", m.Column, "]")
}

type Mapper interface {
	Scanner
	Map() []MappedSymbol
}

type mp struct {
	sc   Scanner
	data []MappedSymbol
}

func (m *mp) Count() int      { return m.sc.Count() }
func (m *mp) Pos() (int, int) { return m.sc.Pos() }
func (m *mp) Error() error    { return m.sc.Error() }

func (m *mp) Get() Symbol {
	sym := m.sc.Get()
	if m.sc.Error() == nil {
		msym := MappedSymbol{}
		msym.Symbol = sym
		msym.Pos = m.Count()
		msym.Row, msym.Column = m.Pos()
		m.data = append(m.data, msym)
	}
	return sym
}

func (m *mp) Map() []MappedSymbol {
	return m.data
}

func NewMapper(s Scanner) Mapper {
	ret := &mp{}
	ret.sc = s
	return ret
}
