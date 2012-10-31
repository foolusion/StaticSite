package parse

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type item struct {
	typ ItemType
	val string
}

type ItemType int

const (
	itemError ItemType = iota
	itemEOF
	itemH1
	itemText
)

const eof = -1

func (i item) String() string {
	switch i.typ {
	case itemEOF:
		return "EOF"
	case itemError:
		return i.val
	}
	if len(i.val) > 80 {
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name  string    // used only for error reports.
	input string    // the string being scanned.
	state stateFn   // the next lexing function to enter.
	start int       // start position of this item.
	pos   int       // current position in the input.
	width int       // width of last rune read
	items chan item // channel of scanned items.
}

func lex(name, input string) *lexer {
	l := &lexer{
		name:  name,
		input: input,
		state: lexNewLine,
		items: make(chan item, 2),
	}
	return l
}

// emit passes an item back to the client.
func (l *lexer) emit(t ItemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

const (
	h1     = "#"
	endPar = "\n\n"
)

func lexNewLine(l *lexer) stateFn {
	for strings.HasPrefix(l.input[l.pos:], "\n") {
		l.pos += len("\n")
		l.ignore()
	}
	if strings.HasPrefix(l.input[l.pos:], h1) {
		return lexHeader
	}
	return lexText
}

func lexText(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], endPar) {
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexNewLine // Next State.
		}
		if l.next() == eof {
			break
		}
	}
	// Correctly reached EOF.
	if l.pos > l.start {
		l.emit(itemText)
	}
	l.emit(itemEOF) // Useful to make EOF a token.
	return nil      // Stop the run loop.
}

func lexHeader(l *lexer) stateFn {
	l.pos += len(h1)
	for strings.HasPrefix(l.input[l.pos:], h1) {
		l.pos += len(h1)
	}
	l.emit(itemH1)
	return lexHeaderText
}

func lexHeaderText(l *lexer) stateFn {
	for isSpace(l.next()) {
		l.ignore()
	}
	i := strings.Index(l.input[l.pos:], "\n")
	if i > 0 {
		l.pos += i
	}
	if l.pos > l.start {
		l.emit(itemText)
	}
	return lexNewLine
}

// next returns the next rune in the input.
func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

// error returns an error token and terminates the scan by passing back a nil
// pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, fmt.Sprintf(format, args...)}
	return nil
}

// nextItem returns the next item from the input.
func (l *lexer) nextItem() item {
	for {
		select {
		case item := <-l.items:
			return item
		default:
			l.state = l.state(l)
		}
	}
	panic("not reached")
}

// isSpace reports whether r is a space character
func isSpace(r rune) bool {
	switch r {
	case ' ', '\t':
		return true
	}
	return false
}
