package main

type CProperty int
type CMannerOfArticulation int
type CPlaceOfArticulation int
type CPhonation int
type CVoiceOnsetTime int
type CAirstreamMechanism int
type CLength int

type Consonant struct {
	Manner    CMannerOfArticulation
	Place     CPlaceOfArticulation
	Phonation CPhonation
	OnsetTime CVoiceOnsetTime
	Mechanism CAirstreamMechanism
	Length    CLength
	IpaSymbol rune
}

const (
	CONS_MANNER CProperty = iota
	CONS_PLACE
	CONS_PHONATION
	CONS_VOT
	CONS_MECHANISM
	CONS_LENGTH
)

const (
	MANNER_NASAL CMannerOfArticulation = iota
	MANNER_STOP
	MANNER_AFFRICATE
	MANNER_SIBILANT_FRICATIVE
	MANNER_NONSIBILANT_FRICATIVE
	MANNER_APPROXIMANT
	MANNER_FLAP_TAP
	MANNER_TRILL
	MANNER_LATERAL_FRICATIVE
	MANNER_LATERAL_APPROXIMANT
	MANNER_LATERAL_FLAP
)

const (
	PLACE_BILABIAL CPlaceOfArticulation = iota
	PLACE_LABIODENTAL
	PLACE_DENTAL
	PLACE_ALVEOLAR
	PLACE_POSTALVEOLAR
	PLACE_RETROFLEX
	PLACE_ALVEOPALATAL
	PLACE_PALATAL
	PLACE_VELAR
	PLACE_UVULAR
	PLACE_PHARYNGEAL
	PLACE_EPIGLOTTAL
	PLACE_GLOTTAL
)

const (
	PHONATION_VOICELESS CPhonation = iota
	PHONATION_BREATHY
	PHONATION_SLACK
	PHONATION_MODAL
	PHONATION_STIFF
	PHONATION_CREAKY
	PHONATION_CLOSURE
)

const (
	VOT_FORTIS CVoiceOnsetTime = iota
	VOT_TENUIS
	VOT_LENIS
)

const (
	MECHANISM_PULMONIC CAirstreamMechanism = iota
	MECHANISM_EJECTIVE
	MECHANISM_IMPLOSIVE
	MECHANISM_CLICK
)

const (
	LENGTH_SHORT CLength = iota
	LENGTH_GEMINATE
	LENGTH_LONG_GEMINATE
)

var consonants map[rune]Consonant

func init() {
	consonants['m̥'] = Consonant{MANNER_NASAL, PLACE_BILABIAL, PHONATION_VOICELESS}
}
