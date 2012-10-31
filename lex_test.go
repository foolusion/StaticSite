package parse

import (
	"reflect"
	"testing"
)

type lexTest struct {
	name  string
	input string
	items []item
}

var (
	tEOF = item{itemEOF, ""}
)

var lexTests = []lexTest{
	{"empty", "", []item{tEOF}},
	{"spaces", " \t\n", []item{{itemText, " \t\n"}, tEOF}},
	{"text", "now is the time\nfor all good things", []item{{itemText, "now is the time\nfor all good things"}, tEOF}},
	{"2par", "go is fun\n\nrodents are gross", []item{{itemText, "go is fun"}, {itemText, "rodents are gross"}, tEOF}},
	{"h1", "# This is a header\n", []item{{itemH1, "#"}, {itemText, "This is a header"}, tEOF}},
	{"par head par", "par one\n\n# header\n\npar two",
		[]item{{itemText, "par one"}, {itemH1, "#"}, {itemText, "header"}, {itemText, "par two"}, tEOF}},
}

func collect(t *lexTest) (items []item) {
	l := lex(t.name, t.input)
	for {
		item := l.nextItem()
		items = append(items, item)
		if item.typ == itemEOF || item.typ == itemError {
			break
		}
	}
	return
}

func TestLex(t *testing.T) {
	for _, test := range lexTests {
		items := collect(&test)
		if !reflect.DeepEqual(items, test.items) {
			t.Errorf("%s: got\n\t%v\nexpected\n\t%v", test.name, items, test.items)
		}
	}
}
