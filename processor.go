package main

import (
	"strings"
	"sync"
	"unicode"
)

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

	return strings.Join(strings.Fields(sb.String()), " ")
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
	sets := make([]ShingleSet, len(corpus))

	var wg sync.WaitGroup
	for i, text := range corpus {
		wg.Add(1)
		go func(idx int, t string) {
			defer wg.Done()
			sets[idx] = buildShingles(cleanText(t), shingleSize)
		}(i, text)
	}
	wg.Wait()

	return sets
}
