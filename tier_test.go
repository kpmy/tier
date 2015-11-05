package tier

import (
	"bufio"
	"bytes"
	"testing"
)

func defTestOpts() Opts {
	defaultOpts := Opts{}
	defaultOpts.IdentMap = make(map[string]SymCode)
	defaultOpts.IdentMap["BEGIN"] = 100

	defaultOpts.SpaceMap = make(map[string]SymCode)
	defaultOpts.SpaceMap[" "] = 101
	defaultOpts.SpaceMap["\n"] = 102

	defaultOpts.NumContains = func() string { return "ABCDEF" }
	defaultOpts.NumModifiers = func() string { return "UH" }

	defaultOpts.CombinedMap = make(map[string]SymCode)
	defaultOpts.CombinedMap[":"] = 200
	defaultOpts.CombinedMap[":="] = 201
	defaultOpts.CombinedMap[":=="] = 203
	defaultOpts.CombinedMap[">"] = 204
	defaultOpts.CombinedMap["<"] = 205
	defaultOpts.CombinedMap["-"] = 206
	defaultOpts.CombinedMap[";"] = 207

	defaultOpts.CommentTriplet = [3]rune{'(', '*', ')'}
	return defaultOpts
}

func TestScanner(t *testing.T) {
	const testString = `
		BEGIN

		f asdf asdf xx x23 (* dfa3asd *) 33FH 3FU 234U 3.3  : := :== > < 0.12314 003141 -efef23 asdfd asf "dfsdfa sdf asdf " 'df' df'd' ;;      ;
	`

	sc := NewScanner(bufio.NewReader(bytes.NewBufferString(testString)), defTestOpts())
	for sc.Error() == nil {
		t.Log(sc.Get())
	}
}

func TestRunner(t *testing.T) {
	const testString = `
		BEGIN

		f asdf asdf xx x23 (* dfa3asd *) 33FH 3FU 234U 3.3  : := :== > < 0.12314 003141 -efef23 asdfd asf "dfsdfa sdf asdf " 'df' df'd' ;;      ;
	`

	sc := NewScanner(bufio.NewReader(bytes.NewBufferString(testString)), defTestOpts())
	run := NewRunner(sc, NewMarker(sc, func(msg ...interface{}) {
		t.Log(msg...)
	}))
	Debug(run)
	run.Expect(100, "begin expected", 102)
	run.Run(207)
}

func TestMapper(t *testing.T) {
	const testString = `
		BEGIN

		f asdf asdf xx x23 (* dfa3asd *) 33FH 3FU 234U 3.3  : := :== > < 0.12314 003141 -efef23 asdfd asf "dfsdfa sdf asdf " 'df' df'd' ;;      ;
	`

	sc := NewScanner(bufio.NewReader(bytes.NewBufferString(testString)), defTestOpts())
	mp := NewMapper(sc)
	run := NewRunner(mp, NewMarker(sc, func(msg ...interface{}) {
		t.Log(msg...)
	}))
	Debug(run)
	run.Expect(100, "begin expected", 102)
	run.Run(207)
	t.Log(mp.Map())
}
