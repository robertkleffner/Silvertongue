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
	LEX_SYLLABLE_VARIABLE
	LEX_DISALLOWED
	LEX_CONFIG_VARIABLE
	LEX_NUMBER
	LEX_VARIABLE
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
	case l.AcceptPred(validVariableName):
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

func lexPhoneme(l *Lexer) State {
	// phonemes are letter strings, optionally followed by a number
	if l.AcceptPred(validVariableName) {
		l.AcceptPredRun(validVariableName)
		l.Emit(LEX_VARIABLE)
		if l.Accept("0123456789") {
			l.AcceptRun("0123456789")
			l.Emit(LEX_NUMBER)
		}
		l.SkipWhitespaceAndComment()
		return lexPhoneme
	}
	if l.Accept(";") {
		l.Emit(LEX_END_DECLARATION)
		l.SkipWhitespaceAndComment()
		return lexDeclaration
	}
	return l.Error(fmt.Sprintf("Possible unterminated phoneme group declaration: %s", l.Next()))
}

func lexSyllableVariable(l *Lexer) State {
	if err := lexVariable(l, LEX_SYLLABLE_VARIABLE); err != nil {
		return err
	}
	return lexSyllable
}

func lexSyllable(l *Lexer) State {
	// syllables can optionally begin with a number
	if l.Accept("0123456789") {
		l.AcceptRun("0123456789")
		l.Emit(LEX_NUMBER)
	}
	if l.AcceptPred(validVariableName) {
		l.AcceptPredRun(validVariableName)
		l.Emit(LEX_VARIABLE)
		l.SkipWhitespaceAndComment()
		return lexSyllable
	}
	if l.Accept(";") {
		l.Emit(LEX_END_DECLARATION)
		l.SkipWhitespaceAndComment()
		return lexDeclaration
	}
	return l.Error(fmt.Sprintf("Invalid symbol '%s' in syllable variable definition.", l.Next()))
}

func lexDisallowed(l *Lexer) State {
	if l.AcceptPred(validVariableName) {
		l.AcceptPredRun(validVariableName)
		l.Emit(LEX_VARIABLE)
		l.SkipWhitespaceAndComment()
		return lexDisallowed
	}
	if l.Accept(";") {
		l.Emit(LEX_END_DECLARATION)
		l.SkipWhitespaceAndComment()
		return lexDeclaration
	}
	return l.Error(fmt.Sprintf("Disallowed statement contains invalid symbol (possibly unterminated): %s.", l.Next()))
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
		return l.Error(fmt.Sprintf("Expected semicolon to terminate config variable declaration, but got '%s' instead.", l.Next()))
	}
	return l.Error(fmt.Sprintf("Config variables must take only integers as values, but got '%s' instead.", l.Next()))
}

func lexVariable(l *Lexer, lexType LexemeType) State {
	l.AcceptPredRun(validVariableName)
	l.Emit(lexType)
	l.SkipWhitespaceAndComment()
	if !l.Accept("=:") {
		return l.Error("Expected '=' or ':' after variable name declaration.")
	}
	l.Ignore()
	l.SkipWhitespaceAndComment()
	return nil
}

func validVariableName(char rune) bool {
	return !unicode.IsSpace(char) && !unicode.IsDigit(char) &&
		char != ';' && char != '(' &&
		char != ')' && char != '#' &&
		char != '!' && char != '%' &&
		char != '=' && char != ':' &&
		char != eof
}
