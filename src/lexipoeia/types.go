package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

type PhonemeGroup []string

type Phoneme struct {
	GroupVariable string
	PercentChance int
}

type Syllable []Phoneme

type SyllableSequence []string

func (syls SyllableSequence) IsContainedIn(sequence []string) bool {
	if len(syls) > len(sequence) {
		return false
	}

	index := 0
	for _, str := range sequence {
		if str == syls[index] {
			index++
		} else {
			index = 0
			if str == syls[index] {
				index++
			}
		}

		if index == len(syls) {
			return true
		}
	}

	return false
}

type Specification struct {
	MeanSyllables       int
	LowDeviation        int
	HighDeviation       int
	GenerateCount       int
	Seed                int64
	PhonemeVariables    map[string]PhonemeGroup
	PhonemeNames        []string
	SyllableVariables   map[string]Syllable
	SyllableNames       []string
	DisallowedSequences []SyllableSequence
}

func LoadSpecification(filename string) Specification {
	read, err := os.Open(filename)
	if err != nil {
		fmt.Println("Could not open the given input file; it may not exist.")
		os.Exit(1)
	}

	defer func() {
		if err := read.Close(); err != nil {
			panic(err)
		}
	}()

	input, err := ioutil.ReadAll(read)
	if err != nil {
		panic(err.Error())
	}

	spec := parseSpecification(string(input))
	if !validateSpecification(spec) {
		os.Exit(1)
	}
	return spec
}

func validateSpecification(spec Specification) bool {
	// are all the disallowed sequences valid syllable names?
	// only emit warnings for these
	for _, seq := range spec.DisallowedSequences {
		for _, name := range seq {
			contains := false
			for _, sylName := range spec.SyllableNames {
				if sylName == name {
					contains = true
					break
				}
			}
			if !contains {
				fmt.Printf("Warning: name '%s' in a disallowed sequence is not a syllable variable name.\n", name)
			}
		}
	}

	// now make sure that every phoneme in the syllables corresponds to phoneme group name
	for sylVar, syl := range spec.SyllableVariables {
		for _, phoneme := range syl {
			contains := false
			for _, phoName := range spec.PhonemeNames {
				if phoName == phoneme.GroupVariable {
					contains = true
					break
				}
			}
			if !contains {
				fmt.Printf("Error: name '%s' in syllable variable '%s' is not a defined phoneme group name.\n", phoneme.GroupVariable, sylVar)
				return false
			}
		}
	}

	return true
}

func parseSpecification(input string) Specification {
	var spec Specification
	spec.PhonemeVariables = make(map[string]PhonemeGroup)
	spec.SyllableVariables = make(map[string]Syllable)

	lexer := NewLexer(input)
	empty := false
	for !empty {
		select {
		case lexeme, ok := <-lexer.lexemes:
			if !ok {
				empty = true
			} else {
				switch lexeme.lexType {
				case LEX_PHONEME_VARIABLE:
					spec.PhonemeVariables[lexeme.value] = parsePhonemeVariable(lexer)
					spec.PhonemeNames = append(spec.PhonemeNames, lexeme.value)
				case LEX_SYLLABLE_VARIABLE:
					spec.SyllableVariables[lexeme.value] = parseSyllableVariable(lexer)
					spec.SyllableNames = append(spec.SyllableNames, lexeme.value)
				case LEX_DISALLOWED:
					spec.DisallowedSequences = append(spec.DisallowedSequences, parseDisallowed(lexer))
				case LEX_CONFIG_VARIABLE:
					parseConfigVariable(&spec, lexeme, lexer)
				}
			}
		}
	}

	return spec
}

func parsePhonemeVariable(l *Lexer) PhonemeGroup {
	group := PhonemeGroup{}
	for lexeme := range l.lexemes {
		if lexeme.lexType == LEX_END_DECLARATION {
			break
		}
		group = append(group, lexeme.value)
	}
	return group
}

func parseSyllableVariable(l *Lexer) Syllable {
	phonemes := []Phoneme{}
	for lexeme := range l.lexemes {
		if lexeme.lexType == LEX_END_DECLARATION {
			break
		}
		p := Phoneme{}
		if lexeme.lexType == LEX_NUMBER {
			num, err := strconv.ParseInt(lexeme.value, 10, 32)
			if err != nil {
				fmt.Printf("Bad number format: %s\n", lexeme.value)
				os.Exit(1)
			}
			lexeme = <-l.lexemes
			p.PercentChance = int(num)
			if p.PercentChance > 100 {
				fmt.Println("Chance of phoneme can't be greater than 100%.")
				os.Exit(1)
			}
		} else {
			p.PercentChance = 100
		}
		if lexeme.lexType == LEX_PHONEME_VARIABLE {
			p.GroupVariable = lexeme.value
		} else {
			fmt.Printf("Expected a phoneme variable name, but got %s\n", lexeme.value)
			os.Exit(1)
		}

		phonemes = append(phonemes, p)
	}
	return phonemes
}

func parseDisallowed(l *Lexer) SyllableSequence {
	seq := SyllableSequence{}
	for lexeme := range l.lexemes {
		if lexeme.lexType == LEX_END_DECLARATION {
			break
		}
		seq = append(seq, lexeme.value)
	}
	return seq
}

func parseConfigVariable(spec *Specification, lexeme Lexeme, l *Lexer) {
	num := int64(-1)
	next := <-l.lexemes
	if next.lexType == LEX_NUMBER {
		temp, err := strconv.ParseInt(next.value, 10, 64)
		if err != nil {
			fmt.Printf("Bad number format: %s\n", next.value)
			os.Exit(1)
		}
		num = temp
	} else {
		fmt.Printf(next.value)
		os.Exit(1)
	}
	switch lexeme.value {
	case "mean":
		spec.MeanSyllables = int(num)
	case "lowDeviation":
		spec.LowDeviation = int(num)
	case "highDeviation":
		spec.HighDeviation = int(num)
	case "words":
		spec.GenerateCount = int(num)
	case "seed":
		spec.Seed = num
	default:
		fmt.Printf("Unknown config variable '%s'\n", lexeme.value)
		os.Exit(1)
	}
}
