package tier

import (
	"fmt"
	"github.com/kpmy/ypk/assert"
	"io"
	"strconv"
	"strings"
	"unicode"
)

type SymCode int

const (
	None SymCode = iota
	Ident
	Number
	String

	UserIota = 100
)

type Symbol struct {
	Code  SymCode
	Value string

	StringOpts struct {
		Apos bool
	}

	NumberOpts struct {
		Modifier string
		Period   bool
	}
}

type Opts struct {
	IdentMap      map[string]SymCode
	IdentStarts   func() string
	IdentContains func() string

	NumContains  func() string
	NumModifiers func() string

	SpaceMap map[string]SymCode

	CombinedMap map[string]SymCode

	CommentTriplet [3]rune

	Skip func(r rune) bool

	NoStrings bool
	NoNum     bool
}

type Scanner interface {
	Get() Symbol
	Error() error

	Count() int
	Pos() (int, int)
}

func Token2(r rune) string {
	return string([]rune{r})
}

func Token(r rune) string {
	if unicode.IsSpace(r) || int(r) <= int(' ') {
		return strconv.Itoa(int(r)) + "U"
	} else {
		return string([]rune{r})
	}
}

func (s Symbol) String() string {
	return fmt.Sprint("sym: `", s.Code, "` ", s.Value)
}

func (sym SymCode) String() (s string) {
	switch sym {
	case None:
		s = "none"
	case Ident:
		s = "ident"
	case Number:
		s = "num"
	case String:
		s = "string"
	default:
		s = strconv.Itoa(int(sym))
	}
	return
}

type sc struct {
	rd  io.RuneReader
	err error

	ch  rune
	pos int

	opts Opts

	lines struct {
		count int
		last  int
		crlf  bool
		lens  map[int]func() (int, int)
	}
}

func (s *sc) Count() int { return s.pos }

func (s *sc) Pos() (int, int) { return s.lines.count, s.lines.last }

func (s *sc) Get() (sym Symbol) {
	for stop := s.err != nil; !stop; {
		sym = s.get()
		stop = sym.Code != 0 || s.err != nil
	}
	return
}

func (s *sc) Error() error {
	return s.err
}

func (s *sc) mark(msg ...interface{}) {
	//log.Println("at pos ", s.pos, " ", fmt.Sprintln(msg...))
	l, c := s.Pos()
	panic(Err("scanner", s.Count(), l, c, msg...))
}

func (o Opts) safeIdentContains() string {
	if o.IdentContains != nil {
		return o.IdentContains()
	}
	return ""
}

func (o Opts) safeIdentStarts() string {
	if o.IdentStarts != nil {
		return o.IdentStarts()
	}
	return ""
}

func (o Opts) safeNumContains() string {
	if o.NumContains != nil {
		return o.NumContains()
	}
	return ""
}

func (o Opts) safeNumModifiers() string {
	if o.NumContains != nil {
		return o.NumModifiers()
	}
	return ""
}

func (o Opts) safeSkip(r rune) (skip bool) {
	if o.Skip != nil {
		skip = o.Skip(r)
	}
	return
}

func (o Opts) isIdentLetter(r rune) bool {
	return o.isIdentFirstLetter(r) || unicode.IsDigit(r) || strings.ContainsRune(o.safeIdentContains(), r)
}

func (o Opts) isIdentFirstLetter(r rune) bool {
	return unicode.IsLetter(r) || strings.ContainsRune(o.safeIdentStarts(), r)
}

func (o Opts) validate() {
	assert.For(o.CommentTriplet[0] != o.CommentTriplet[1], 20)
	assert.For(o.CommentTriplet[1] != o.CommentTriplet[2], 20)
}

func (s *sc) ident() (sym Symbol) {
	assert.For(s.opts.isIdentFirstLetter(s.ch), 20, "letter must be first")
	buf := make([]rune, 0)
	for s.err == nil && s.opts.isIdentLetter(s.ch) {
		buf = append(buf, s.ch)
		s.next()
	}
	if s.err == nil || s.err == io.EOF {
		sym.Value = string(buf)
		if code, ok := s.opts.IdentMap[sym.Value]; ok {
			sym.Code = code
		} else {
			sym.Code = Ident
		}
	} else {
		s.mark("error while reading ident ", s.err)
	}
	return
}

const dec = "0123456789"

//first char always 0..9
func (s *sc) num() (sym Symbol) {
	assert.For(unicode.IsDigit(s.ch), 20, "digit expected")
	var buf []rune
	var mbuf []rune
	hasDot := false

	for {
		buf = append(buf, s.ch)
		s.next()
		if s.ch == '.' {
			if !hasDot {
				hasDot = true
			} else if hasDot {
				s.mark("dot unexpected")
			}
		}
		if s.err != nil || !(s.ch == '.' || strings.ContainsRune(dec, s.ch) || strings.ContainsRune(s.opts.safeNumContains(), s.ch)) {
			break
		}
	}
	if strings.ContainsRune(s.opts.safeNumModifiers(), s.ch) {
		mbuf = append(mbuf, s.ch)
		s.next()
	}
	if strings.ContainsAny(string(buf), s.opts.safeNumContains()) && len(mbuf) == 0 {
		s.mark("modifier expected")
	}
	if s.err == nil || s.err == io.EOF {
		sym.Code = Number
		sym.Value = string(buf)
		sym.NumberOpts.Modifier = string(mbuf)
		sym.NumberOpts.Period = hasDot
	} else {
		s.mark("error reading number")
	}
	return
}

func (s *sc) str() string {
	assert.For(s.ch == '"' || s.ch == '\'' || s.ch == '`', 20, "quote expected")
	var buf []rune
	ending := s.ch
	s.next()
	for ; s.err == nil && s.ch != ending; s.next() {
		buf = append(buf, s.ch)
	}
	if s.err == nil {
		s.next()
	} else {
		s.mark("string expected")
	}
	return string(buf)
}

func (s *sc) next() rune {
	read := 0
	s.ch, read, s.err = s.rd.ReadRune()
	if s.err == nil {
		s.pos += read
	}
	if s.ch == '\r' || s.ch == '\n' {
		s.line()
	} else {
		s.lines.last++
	}
	//log.Println(Token(s.ch), s.err)
	return s.ch
}

func (s *sc) line() {
	if s.ch == '\r' {
		s.lines.crlf = true
	}
	if (s.lines.crlf && s.ch == '\r') || (!s.lines.crlf && s.ch == '\n') {
		s.lines.lens[s.lines.count] = func() (int, int) {
			return s.lines.count, s.pos
		}
		s.lines.count++
		s.lines.last = 1
	} else if s.lines.crlf && s.ch == '\n' {
		s.lines.last--
	}
}

func (s *sc) comment() {
	assert.For(s.ch == '*', 20, "expected ", s.opts.CommentTriplet[1], "got ", Token(s.ch))
	for {
		for s.err == nil && s.ch != s.opts.CommentTriplet[1] {
			if s.ch == s.opts.CommentTriplet[0] {
				if s.next() == s.opts.CommentTriplet[1] {
					s.comment()
				}
			} else {
				s.next()
			}
		}
		for s.err == nil && s.ch == s.opts.CommentTriplet[1] {
			s.next()
		}
		if s.err != nil || s.ch == s.opts.CommentTriplet[2] {
			break
		}
	}
	if s.err == nil {
		s.next()
	} else {
		s.mark("unclosed comment")
	}
}

func (s *sc) filter(r ...rune) SymCode {
	var run func(keys map[string]SymCode, r ...rune) SymCode

	run = func(keys map[string]SymCode, r ...rune) (ret SymCode) {
		key := string(r)
		ret = keys[key]
		continues := make(map[string]SymCode)
		for k, v := range keys {
			if key != k && strings.HasPrefix(k, key) {
				continues[k] = v
			}
		}
		//log.Println(ret, continues)
		if len(continues) > 0 {
			nr := []rune(key)
			nr = append(nr, s.next())
			if x := run(continues, nr...); x != None {
				ret = x
			}
		}
		return
	}

	return run(s.opts.CombinedMap, r...)
}

func (s *sc) get() (sym Symbol) {
	justRune := func() {
		if symCode := s.filter(s.ch); symCode != None {
			sym.Code = symCode
			s.next()
		} else {
			if !s.opts.safeSkip(s.ch) {
				s.mark("unhandled ", "`", Token(s.ch), "`")
			}
			s.next()
		}
	}

	switch s.ch {
	case s.opts.CommentTriplet[0]:
		if s.next() == s.opts.CommentTriplet[1] {
			s.comment()
		} else if symCode := s.filter(s.opts.CommentTriplet[0], s.ch); symCode != None {
			sym.Code = symCode
		} else {
			sym.Code = s.filter(s.opts.CommentTriplet[0])
		}
	case '"', '\'', '`':
		if !s.opts.NoStrings {
			sym.StringOpts.Apos = (s.ch == '\'' || s.ch == '`')
			sym.Value = s.str()
			sym.Code = String
		} else {
			justRune()
		}
	default:
		switch {
		case s.opts.isIdentFirstLetter(s.ch):
			sym = s.ident()
		case unicode.IsSpace(s.ch):
			sym.Value = Token2(s.ch)
			sym.Code, _ = s.opts.SpaceMap[sym.Value]
			s.next()
		case unicode.IsDigit(s.ch):
			if !s.opts.NoNum {
				sym = s.num()
			} else {
				justRune()
			}
		default:
			justRune()
		}
	}
	return
}

func NewScanner(rd io.RuneReader, opts ...Opts) Scanner {
	ret := &sc{}
	ret.rd = rd
	ret.lines.lens = make(map[int]func() (int, int))
	ret.lines.count++
	if len(opts) > 0 {
		ret.opts = opts[0]
	} else {
		ret.opts = DefaultOpts
	}
	ret.opts.validate()
	ret.next()
	return ret
}

var DefaultOpts Opts

func init() {
	DefaultOpts.IdentMap = make(map[string]SymCode)

	DefaultOpts.SpaceMap = make(map[string]SymCode)

	DefaultOpts.NumContains = func() string { return "ABCDEF" }
	DefaultOpts.NumModifiers = func() string { return "U" }

	DefaultOpts.CombinedMap = make(map[string]SymCode)

	DefaultOpts.CommentTriplet = [3]rune{'(', '*', ')'}
}
