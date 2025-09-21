package transport

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	xhkdf "golang.org/x/crypto/hkdf"
	"io"
	"strings"
)

func NewNonce() ([]byte, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	return b, err
}

func orderBytes(a, b []byte) (first, second []byte) {
	if len(a) == 0 && len(b) == 0 {
		return a, b
	}
	if string(a) <= string(b) {
		return a, b
	}
	return b, a
}

func DerivePGPWordsSimple(
	serverPub ed25519.PublicKey,
	clientNonce, serverNonce []byte,
	n int,
) ([]string, error) {
	if n < 3 || n > 6 {
		n = 3
	}

	na, nb := orderBytes(clientNonce, serverNonce)
	salt := make([]byte, 0, len(na)+len(nb))
	salt = append(salt, na...)
	salt = append(salt, nb...)

	r := xhkdf.New(sha256.New, serverPub, salt, []byte("SAS/PGP-words v1"))
	buf := make([]byte, n*2)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}

	words := make([]string, n)
	for i := 0; i < n; i++ {
		v := binary.BigEndian.Uint16(buf[i*2:])
		idx := int(v >> 8)
		if i%2 == 0 {
			words[i] = pgpEven[idx]
		} else {
			words[i] = pgpOdd[idx]
		}
	}
	return words, nil
}

func FormatSAS(words []string) string {
	return strings.Join(words, "-")
}

var pgpEven = []string{
	"aardvark", "absurd", "accrue", "acme", "adrift", "adult", "afflict", "ahead", "aimless", "Algol", "allow", "alone", "ammo", "ancient", "apple", "artist",
	"assume", "Athens", "atlas", "Aztec", "baboon", "backfield", "backward", "banjo", "beaming", "Beijing", "belief", "beneath", "bikini", "binocular", "biosphere", "blade",
	"boardwalk", "bodyguard", "bookseller", "borderline", "bottomless", "Bradbury", "bravado", "Brazilian", "breakaway", "Burlington", "businessman", "butterfat", "Camelot", "candidate", "cannonball", "Capricorn",
	"caravan", "caretaker", "celebrate", "cellulose", "certify", "chambermaid", "Cherokee", "Chicago", "clergyman", "coherence", "combustion", "commando", "company", "component", "concurrent", "confidence",
	"conformist", "congregate", "consensus", "consulting", "corporate", "councilman", "crossover", "crucifix", "cumbersome", "customer", "Dakota", "decadence", "December", "decimal", "designing", "detector",
	"detergent", "determine", "dictator", "dinosaur", "direction", "disable", "disbelief", "disruptive", "distortion", "document", "embezzle", "enchanting", "enrollment", "enterprise", "equation", "equipment",
	"escapade", "Eskimo", "everyday", "examine", "existence", "exodus", "fascinate", "filament", "finicky", "forever", "fortitude", "frequency", "gadgetry", "Galveston", "getaway", "glossary",
	"gossamer", "graduate", "gravity", "guitarist", "hamburger", "Hamilton", "handiwork", "hazardous", "headwaters", "hemisphere", "hesitate", "hideaway", "holiness", "hurricane", "hydraulic", "impartial",
	"impetus", "inception", "indigo", "inertia", "infancy", "inferno", "informant", "insincere", "insurgent", "integrate", "intention", "inventive", "Istanbul", "Jamaica", "Jupiter", "leprosy",
	"letterhead", "liberty", "maritime", "matchmaker", "maverick", "Medusa", "megaton", "microscope", "microwave", "midsummer", "millionaire", "miracle", "misnomer", "molasses", "molecule", "Montana",
	"monument", "mosquito", "narrative", "nebula", "newsletter", "Norwegian", "October", "Ohio", "onlooker", "opulent", "orchestra", "oscillator", "outside", "pandemic", "Pandora", "paperweight",
	"paragon", "paragraph", "paramount", "passenger", "pedigree", "Pegasus", "penetrate", "perceptive", "performance", "pharmacy", "phonetic", "photograph", "pioneer", "pocketful", "politeness", "positive",
	"potato", "processor", "provincial", "proximate", "puberty", "publisher", "pyramid", "quantity", "racketeer", "rebellion", "recipe", "recover", "repellent", "replica", "reproduce", "resistor",
	"responsive", "retraction", "retrieval", "retrospect", "revenue", "revival", "revolver", "sandalwood", "sardonic", "Saturday", "savagery", "scavenger", "sensation", "sociable", "souvenir", "specialist",
	"speculate", "stethoscope", "stupendous", "supportive", "surrender", "suspicious", "sympathy", "tambourine", "telephone", "therapist", "tobacco", "tolerance", "tomorrow", "torpedo", "tradition",
	"travesty", "trombonist", "truncated", "typewriter", "ultimate", "undaunted", "underfoot", "unicorn", "unify", "universe", "unravel", "upcoming", "vacancy", "vagabond", "vertigo", "Virginia",
	"visitor", "vocalist", "voyager", "warranty", "Waterloo", "whimsical", "Wichita", "Wilmington", "Wyoming", "yesteryear", "Yucatan", "Yule", "zealous", "zebra", "Zeus", "zodiac",
}

var pgpOdd = []string{
	"adroitness", "adviser", "aftermath", "aggregate", "alkaline", "almighty", "amulet", "amusement", "antenna", "applicant", "Apollo", "armistice", "article", "asteroid", "Atlantic", "atmosphere",
	"autopsy", "Babylon", "backwater", "barbecue", "belowground", "bifocals", "bodyguard", "bookseller", "borderline", "bottomless", "Bradbury", "bravado", "Brazilian", "breakaway", "Burlington", "businessman",
	"butterfat", "Camelot", "candidate", "cannonball", "Capricorn", "caravan", "caretaker", "celebrate", "cellulose", "certify", "chambermaid", "Cherokee", "Chicago", "clergyman", "coherence", "combustion",
	"commando", "company", "component", "concurrent", "confidence", "conformist", "congregate", "consensus", "consulting", "corporate", "councilman", "crossover", "crucifix", "cumbersome", "customer", "Dakota",
	"decadence", "December", "decimal", "designing", "detector", "detergent", "determine", "dictator", "dinosaur", "direction", "disable", "disbelief", "disruptive", "distortion", "document",
	"embezzle", "enchanting", "enrollment", "enterprise", "equation", "equipment", "escapade", "Eskimo", "everyday", "examine", "existence", "exodus", "fascinate", "filament", "finicky",
	"forever", "fortitude", "frequency", "gadgetry", "Galveston", "getaway", "glossary", "gossamer", "graduate", "gravity", "guitarist", "hamburger", "Hamilton", "handiwork", "hazardous", "headwaters",
	"hemisphere", "hesitate", "hideaway", "holiness", "hurricane", "hydraulic", "impartial", "impetus", "inception", "indigo", "inertia", "infancy", "inferno", "informant", "insincere",
	"insurgent", "integrate", "intention", "inventive", "Istanbul", "Jamaica", "Jupiter", "leprosy", "letterhead", "liberty", "maritime", "matchmaker", "maverick", "Medusa", "megaton", "microscope",
	"microwave", "midsummer", "millionaire", "miracle", "misnomer", "molasses", "molecule", "Montana", "monument", "mosquito", "narrative", "nebula", "newsletter", "Norwegian", "October", "Ohio",
	"onlooker", "opulent", "orchestra", "oscillator", "outside", "pandemic", "Pandora", "paperweight", "paragon", "paragraph", "paramount", "passenger", "pedigree", "Pegasus", "penetrate", "perceptive",
	"performance", "pharmacy", "phonetic", "photograph", "pioneer", "pocketful", "politeness", "positive", "potato", "processor", "provincial", "proximate", "puberty", "publisher", "pyramid", "quantity",
	"racketeer", "rebellion", "recipe", "recover", "repellent", "replica", "reproduce", "resistor", "responsive", "retraction", "retrieval", "retrospect", "revenue", "revival", "revolver", "sandalwood",
	"sardonic", "Saturday", "savagery", "scavenger", "sensation", "sociable", "souvenir", "specialist", "speculate", "stethoscope", "stupendous", "supportive", "surrender", "suspicious", "sympathy", "tambourine",
	"telephone", "therapist", "tobacco", "tolerance", "tomorrow", "torpedo", "tradition", "travesty", "trombonist", "truncated", "typewriter", "ultimate", "undaunted", "underfoot", "unicorn", "unify",
	"universe", "unravel", "upcoming", "vacancy", "vagabond", "vertigo", "Virginia", "visitor", "vocalist", "voyager", "warranty", "Waterloo", "whimsical", "Wichita", "Wilmington", "Wyoming",
	"yesteryear", "Yucatan", "Yule", "zealous", "zebra", "Zeus", "zodiac", "aardvark", "absurd", "accrue", "acme", "adrift", "adult", "afflict", "ahead", "aimless", "Algol",
}
