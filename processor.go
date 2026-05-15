package main

import (
	"strings"
	"sync"
	"unicode"
)

var stopWords = map[string]struct{}{

	"в": {}, "у": {}, "на": {}, "з": {}, "із": {}, "зі": {}, "до": {},
	"від": {}, "по": {}, "за": {}, "під": {}, "над": {}, "між": {},
	"через": {}, "без": {}, "про": {}, "при": {}, "для": {}, "після": {},
	"перед": {}, "біля": {}, "крім": {}, "замість": {}, "щодо": {},

	"і": {}, "й": {}, "та": {}, "але": {}, "проте": {}, "однак": {},
	"або": {}, "що": {}, "як": {}, "якщо": {}, "коли": {},
	"хоча": {}, "тому": {}, "бо": {}, "адже": {}, "щоб": {}, "аби": {},
	"поки": {}, "доки": {},

	"не": {}, "ні": {}, "так": {}, "лише": {}, "тільки": {}, "навіть": {},
	"вже": {}, "ще": {}, "ж": {}, "же": {}, "б": {}, "би": {},
	"нехай": {}, "саме": {},

	"я": {}, "ти": {}, "він": {}, "вона": {}, "воно": {}, "ми": {}, "ви": {},
	"вони": {}, "мене": {}, "тебе": {}, "його": {}, "її": {}, "нас": {},
	"вас": {}, "їх": {}, "мені": {}, "тобі": {}, "йому": {}, "їй": {},
	"нам": {}, "вам": {}, "їм": {}, "себе": {}, "цей": {}, "ця": {},
	"це": {}, "ці": {}, "той": {}, "те": {}, "ті": {},
	"який": {}, "яка": {}, "яке": {}, "які": {}, "хто": {},
	"чий": {}, "чия": {}, "чиє": {}, "чиї": {},

	"є": {}, "був": {}, "була": {}, "було": {}, "були": {}, "буде": {},
	"будуть": {}, "бути": {}, "мати": {}, "має": {}, "мав": {}, "мала": {},

	"тут": {}, "там": {}, "де": {}, "куди": {}, "звідки": {},
	"дуже": {}, "більш": {}, "менш": {}, "досить": {}, "майже": {},
	"також": {}, "теж": {}, "ось": {}, "от": {}, "тоді": {},
}

func isStopWord(token string) bool {
	_, ok := stopWords[token]
	return ok
}

func stemming(token string) string {
	runes := []rune(token)
	n := len(runes)
	if n < 5 {
		return token
	}

	suffixes := []struct {
		suf     string
		minRoot int
	}{
		{"ується", 4}, {"уються", 4}, {"ватися", 4},
		{"увати", 4}, {"ювати", 4}, {"овати", 4},
		{"ують", 4}, {"юють", 4}, {"ував", 4}, {"ював", 4},
		{"ання", 4}, {"ення", 4}, {"іння", 4},
		{"ати", 4}, {"яти", 4}, {"іти", 4}, {"ити", 4}, {"ути", 4},
		{"ають", 4}, {"яють", 4}, {"ить", 4}, {"іть", 4},

		{"ість", 4}, {"істю", 4},
		{"ського", 4}, {"зького", 4}, {"цького", 4},
		{"ський", 4}, {"зький", 4}, {"цький", 4},
		{"ових", 4}, {"евих", 4}, {"євих", 4},
		{"овий", 4}, {"евий", 4}, {"євий", 4},
		{"ами", 4}, {"ями", 4},
		{"ого", 4}, {"ього", 4},
		{"ому", 4}, {"ьому", 4},
		{"ові", 4}, {"еві", 4}, {"єві", 4},
		{"ова", 4}, {"ева", 4}, {"єва", 4},
		{"ах", 4}, {"ях", 4},
		{"ою", 4}, {"ею", 4}, {"єю", 4},
		{"ів", 4}, {"ей", 4},
		{"ам", 4}, {"ям", 4},
		{"ий", 4}, {"ій", 4},
		{"и", 4}, {"і", 4}, {"а", 4}, {"я", 4},
		{"у", 4}, {"ю", 4}, {"о", 4},
	}

	for _, rule := range suffixes {
		sufRunes := []rune(rule.suf)
		sufLen := len(sufRunes)
		if n <= sufLen {
			continue
		}
		rootLen := n - sufLen
		if rootLen < rule.minRoot {
			continue
		}
		match := true
		for i, sr := range sufRunes {
			if runes[rootLen+i] != sr {
				match = false
				break
			}
		}
		if match {
			return string(runes[:rootLen])
		}
	}
	return token
}

func cleanText(text string) string {
	text = strings.ToLower(text)

	var sb strings.Builder
	sb.Grow(len(text))
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			sb.WriteRune(r)
		} else {
			sb.WriteRune(' ')
		}
	}

	tokens := strings.Fields(sb.String())
	filtered := make([]string, 0, len(tokens))
	for _, tok := range tokens {
		if !isStopWord(tok) {
			filtered = append(filtered, stemming(tok))
		}
	}
	return strings.Join(filtered, " ")
}

func buildShingles(text string, n int) ShingleSet {
	words := strings.Fields(text)
	shingles := make(ShingleSet)

	if len(words) < n {
		return shingles
	}

	var sb strings.Builder
	for i := 0; i <= len(words)-n; i++ {
		sb.Reset()
		for j := 0; j < n; j++ {
			if j > 0 {
				sb.WriteByte(' ')
			}
			sb.WriteString(words[i+j])
		}
		shingles[sb.String()] = struct{}{}
	}
	return shingles
}

func BuildShingleSetsSequential(corpus []string) []ShingleSet {
	sets := make([]ShingleSet, len(corpus))
	for i, text := range corpus {
		sets[i] = buildShingles(cleanText(text), shingleSize)
	}
	return sets
}

func BuildShingleSetsParallel(corpus []string) []ShingleSet {
	n := len(corpus)
	sets := make([]ShingleSet, n)
	if n == 0 {
		return sets
	}

	chunks := numWorkers
	if chunks > n {
		chunks = n
	}
	chunkSize := (n + chunks - 1) / chunks

	var wg sync.WaitGroup
	for start := 0; start < n; start += chunkSize {
		end := start + chunkSize
		if end > n {
			end = n
		}
		wg.Add(1)
		go func(s, e int) {
			defer wg.Done()
			for i := s; i < e; i++ {
				sets[i] = buildShingles(cleanText(corpus[i]), shingleSize)
			}
		}(start, end)
	}
	wg.Wait()
	return sets
}
