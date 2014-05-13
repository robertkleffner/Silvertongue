package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() < 1 {
		usage()
	}

	/*var spec Specification
	spec.GenerateCount = 100
	spec.Seed = 12345
	spec.MeanSyllables = 3
	spec.LowDeviation = 2
	spec.HighDeviation = 4
	spec.GenerateCount = 100

	consonants := PhonemeGroup{"p", "t", "m", "n", "k", "h", "'", "w", "v", "l"}
	vowels := PhonemeGroup{"i", "u", "e", "o", "a"}

	cm := Phoneme{"C", 75}
	c := Phoneme{"C", 100}
	v := Phoneme{"V", 100}
	cv := Syllable{c, v}
	mcv := Syllable{cm, v}

	spec.PhonemeVariables = make(map[string]PhonemeGroup)
	spec.PhonemeVariables["V"] = vowels
	spec.PhonemeVariables["C"] = consonants
	spec.SyllableVariables = make(map[string]Syllable)
	spec.SyllableVariables["cv"] = cv
	spec.SyllableVariables["*cv"] = mcv
	spec.SyllableNames = []string{"cv", "*cv"}
	spec.DisallowedSequences = []SyllableSequence{SyllableSequence{"*cv", "*cv"}}*/

	spec := LoadSpecification(flag.Arg(0))
	output := ""
	if flag.NArg() > 1 {
		output = flag.Arg(1)
	}
	Generate(spec, output)
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: lexipoeia [inputfile]\n")
	flag.PrintDefaults()
	os.Exit(2)
}
