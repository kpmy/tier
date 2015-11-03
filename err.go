package tier

import (
	"encoding/xml"
	"fmt"
	"github.com/kpmy/ypk/halt"
)

type Marker interface {
	Mark(...interface{})
	FutureMark() Marker
	Log(...interface{})
}

type Error struct {
	XMLName xml.Name
	From    string `xml:"from,attr"`
	Pos     int    `xml:"pos,attr"`
	Line    int    `xml:"line,attr"`
	Column  int    `xml:"column,attr"`
	Message string `xml:",chardata"`
}

func (e *Error) String() string {
	data, _ := xml.Marshal(e)
	return string(data)
}

func Err(sender string, pos, line, col int, msg ...interface{}) *Error {
	err := &Error{From: sender, Pos: pos, Line: line, Column: col, Message: fmt.Sprint(msg...)}
	err.XMLName.Local = "error"
	return err
}

type mark struct {
	rd        int
	line, col int
	marker    Marker
}

func (m *mark) Mark(msg ...interface{}) {
	m.marker.(*mk).m = m
	m.marker.Mark(msg...)
}

func (m *mark) FutureMark() Marker { halt.As(100); return nil }

func (m *mark) Log(msg ...interface{}) { m.marker.Log(msg...) }

type mk struct {
	sc  Scanner
	m   *mark
	log func(...interface{})
}

func (m *mk) Mark(msg ...interface{}) {
	rd := m.sc.Count()
	str, pos := m.sc.Pos()
	if len(msg) == 0 {
		m.m = &mark{rd: rd, line: str, col: pos}
	} else if m.m != nil {
		rd, str, pos = m.m.rd, m.m.line, m.m.col
		m.m = nil
	}
	if m.m == nil {
		panic(Err("parser", rd, str, pos, msg...))
	}
}

func (m *mk) FutureMark() Marker {
	rd := m.sc.Count()
	str, pos := m.sc.Pos()
	ret := &mark{marker: m, rd: rd, line: str, col: pos}
	return ret
}

func (m *mk) Log(msg ...interface{}) {
	if m.log != nil {
		m.log(msg...)
	}
}

func NewMarker(s Scanner, log func(...interface{})) Marker {
	ret := &mk{}
	ret.sc = s
	ret.log = log
	return ret
}
