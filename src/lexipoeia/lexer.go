/*
 * Based on (and mostly stolen) from Rob Pike's presentation on building
 * a lexical analyzer in Go.
 */

package main

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

/*************************************
 * Lexeme stuff
 ************************************/
type LexemeType int

const (
	LEX_ERROR LexemeType = iota // error occurred, value is error message
	LEX_EOF                     // end of file
	LEX_PHONEME_VARIABLE
	LEX_PHONEME
	LEX_SYLLABLE_PART
	LEX_SYLLABLE_VARIABLE
	LEX_DISALLOWED
	LEX_CONFIG_VARIABLE
	LEX_NUMBER
	LEX_END_DECLARATION
)

type Lexeme struct {
	lexType LexemeType
	value   string
}

func (l Lexeme) String() string {
	switch l.lexType {
	case LEX_EOF:
		return "EOF"
	case LEX_ERROR:
		return l.value
	default:
		return fmt.Sprintf("%q", l.value)
	}
}

/*************************************
 * Lexer stuff
 ************************************/
type Lexer struct {
	input   string      // input being lexed
	start   int         // start position of current Lexeme
	current int         // current position in the input
	width   int         // width of the last rune read (for UTF-8)
	lexemes chan Lexeme // lexeme output channel
}

func NewLexer(input string) *Lexer {
	l := &Lexer{
		input:   strings.TrimSpace(input),
		lexemes: make(chan Lexeme),
	}
	go l.Run()
	return l
}

func (l *Lexer) NextLexeme() Lexeme {
	return <-l.lexemes
}

func (l *Lexer) Run() {
	for state := lexDeclaration; state != nil; {
		state = state(l)
	}
	close(l.lexemes)
}

func (l *Lexer) Emit(t LexemeType) {
	lexeme := Lexeme{t, l.input[l.start:l.current]}
	l.lexemes <- lexeme
	l.start = l.current
}

func (l *Lexer) Next() rune {
	if l.current >= len(l.input) {
		l.width = 0
		return eof
	}
	char, w := utf8.DecodeRuneInString(l.input[l.current:])
	l.width = w
	l.current += l.width
	return char
}

func (l *Lexer) Ignore() {
	l.start = l.current
}

// Can only be called once per call of next
func (l *Lexer) Backup() {
	l.current -= l.width
}

func (l *Lexer) Peek() rune {
	char := l.Next()
	l.Backup()
	return char
}

func (l *Lexer) Accept(valid string) bool {
	if strings.IndexRune(valid, l.Next()) >= 0 {
		return true
	}
	l.Backup()
	return false
}

func (l *Lexer) AcceptRun(valid string) {
	for strings.IndexRune(valid, l.Next()) >= 0 {
	}
	l.Backup()
}

func (l *Lexer) AcceptPred(valid func(rune) bool) bool {
	if valid(l.Next()) {
		return true
	}
	l.Backup()
	return false
}

func (l *Lexer) AcceptPredRun(valid func(rune) bool) {
	for valid(l.Next()) {
	}
	l.Backup()
}

func (l *Lexer) SkipWhitespaceAndComment() {
	l.AcceptPredRun(unicode.IsSpace)
	if l.Accept("(") {
		l.AcceptPredRun(func(char rune) bool {
			return char != ')'
		})
		l.Accept(")")
		l.AcceptPredRun(unicode.IsSpace)
	}
	l.Ignore()
}

func (l *Lexer) Error(format string, args ...interface{}) State {
	l.lexemes <- Lexeme{
		LEX_ERROR,
		fmt.Sprintf(format, args...),
	}
	return nil
}

/*************************************
 * State functions
 ************************************/
const eof = -1

type State func(*Lexer) State

func lexDeclaration(l *Lexer) State {
	l.SkipWhitespaceAndComment()
	switch {
	case l.AcceptPred(unicode.IsLetter):
		l.Backup()
		return lexPhonemeVariable
	case l.Accept("%"):
		l.Ignore()
		return lexSyllableVariable
	case l.Accept("!"):
		l.Emit(LEX_DISALLOWED)
		return lexDisallowed
	case l.Accept("#"):
		l.Ignore()
		return lexConfigVariable
	default:
		return l.Error(fmt.Sprintf("Bad beginning of line: %s", l.Next()))
	}
}

func lexPhonemeVariable(l *Lexer) State {
	if err := lexVariable(l, LEX_PHONEME_VARIABLE); err != nil {
		return err
	}
	return lexPhoneme
}

func lexSyllable(l *Lexer) State {
	if l.Accept("?") {
		l.Ignore()
		l.AcceptRun("0123456789")
		l.Emit(LEX_NUMBER)
	}
	if l.AcceptPred(unicode.IsLetter) {
		l.AcceptPredRun(unicode.IsLetter)
		l.Emit(LEX_PHONEME_VARIABLE)
		l.SkipWhitespaceAndComment()
		if l.Accept("-") {
			l.Ignore()
		}
		l.SkipWhitespaceAndComment()
		return lexSyllable
	} else if l.Accept(";") {
		l.Emit(LEX_END_DECLARATION)
		l.SkipWhitespaceAndComment()
		return lexDeclaration
	}
	return l.Error("Improper syllable variable declaration.")
}

func lexPhoneme(l *Lexer) State {
	if l.AcceptPred(unicode.IsLetter) {
		l.AcceptPredRun(unicode.IsLetter)
		l.Emit(LEX_PHONEME)
		l.SkipWhitespaceAndComment()
		if l.Accept(",") {
			l.Ignore()
		}
		l.SkipWhitespaceAndComment()
		return lexPhoneme
	} else if l.Accept(";") {
		l.Emit(LEX_END_DECLARATION)
		l.SkipWhitespaceAndComment()
		return lexDeclaration
	}
	return l.Error("Expected either letters or ';' as symbol in phoneme group declaration.")
}

func lexSyllableVariable(l *Lexer) State {
	if err := lexVariable(l, LEX_SYLLABLE_VARIABLE); err != nil {
		return err
	}
	return lexSyllable
}

func lexDisallowed(l *Lexer) State {
	if l.AcceptPred(unicode.IsLetter) {
		l.AcceptPredRun(unicode.IsLetter)
		l.Emit(LEX_SYLLABLE_VARIABLE)
		l.SkipWhitespaceAndComment()
		if l.Accept("-") {
			l.Ignore()
		}
		l.SkipWhitespaceAndComment()
		return lexDisallowed
	} else if l.Accept(";") {
		l.Emit(LEX_END_DECLARATION)
		l.SkipWhitespaceAndComment()
		return lexDeclaration
	}
	return l.Error("Improper disallowed statement.")
}

func lexConfigVariable(l *Lexer) State {
	if err := lexVariable(l, LEX_CONFIG_VARIABLE); err != nil {
		return err
	}

	if l.Accept("0123456789") {
		l.AcceptRun("0123456789")
		l.Emit(LEX_NUMBER)
		l.SkipWhitespaceAndComment()
		if l.Accept(";") {
			l.Emit(LEX_END_DECLARATION)
			l.SkipWhitespaceAndComment()
			return lexDeclaration
		}
		return l.Error("Expected semicolon to terminate config variable declaration.")
	}
	return l.Error("Config variables must take only integers as values.")
}

func lexVariable(l *Lexer, lexType LexemeType) State {
	l.AcceptPredRun(unicode.IsLetter)
	l.Emit(lexType)
	l.SkipWhitespaceAndComment()
	if !l.Accept("=:") {
		return l.Error("Expected '=' or ':' after variable name.")
	}
	l.Ignore()
	l.SkipWhitespaceAndComment()
	return nil
}
