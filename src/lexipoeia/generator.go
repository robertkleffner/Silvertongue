package main

import (
	"io"
	"math/rand"
	"os"
)

func Generate(spec Specification, filename string) {
	var out io.Writer
	if filename != "" {
		temp, err := os.Create(filename)
		if err != nil {
			panic(err)
		}
		defer func() {
			if err := temp.Close(); err != nil {
				panic(err)
			}
		}()
		out = temp
	} else {
		out = os.Stdout
	}

	rand.Seed(spec.Seed)
	for i := 0; i < spec.GenerateCount; i++ {
		io.WriteString(out, nextWord(spec)+"\n")
	}
}

func nextWord(spec Specification) string {
	syllables := nextSyllableCount(spec)
	sequence := generateSequence(syllables, spec)

	word := ""
	for _, syll := range sequence {
		word += generateSyllable(syll, spec)
	}
	return word
}

func nextSyllableCount(spec Specification) int {
	min := spec.MeanSyllables
	if spec.LowDeviation != 0 {
		min = spec.MeanSyllables - rand.Intn(spec.LowDeviation+1)
	}

	max := spec.MeanSyllables
	if spec.HighDeviation != 0 {
		max = spec.MeanSyllables + rand.Intn(spec.HighDeviation+1)
	}

	syllables := min
	if max-min != 0 {
		syllables = rand.Intn(max-min) + min
	}
	return syllables
}

func nextSyllable(spec Specification) string {
	return spec.SyllableNames[rand.Intn(len(spec.SyllableNames))]
}

func generateSyllable(syllableVariable string, spec Specification) string {
	result := ""
	syllable := spec.SyllableVariables[syllableVariable]
	for _, phoneme := range syllable {
		chance := rand.Intn(100)
		if chance < phoneme.PercentChance {
			result += generatePhoneme(phoneme.GroupVariable, spec)
		}
	}
	return result
}

func generatePhoneme(phonemeVariable string, spec Specification) string {
	group := spec.PhonemeVariables[phonemeVariable]
	return group[rand.Intn(len(group))]
}

func generateSequence(syllables int, spec Specification) []string {
	sequence := make([]string, 0, syllables)

	valid := false
	for !valid {
		for i := 0; i < syllables; i++ {
			sequence = append(sequence, nextSyllable(spec))
		}
		valid = true
		for _, disallowed := range spec.DisallowedSequences {
			if disallowed.IsContainedIn(sequence) {
				valid = false
				sequence = sequence[:0]
				break
			}
		}
	}
	return sequence
}
