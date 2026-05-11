package main

import (
	"math"
	"sort"
	"testing"
)

const eps = 1e-9

func makeSet(keys ...string) ShingleSet {
	s := make(ShingleSet, len(keys))
	for _, k := range keys {
		s[k] = struct{}{}
	}
	return s
}

func sortResults(rs []PairResult) []PairResult {
	sort.Slice(rs, func(a, b int) bool {
		if rs[a].I != rs[b].I {
			return rs[a].I < rs[b].I
		}
		return rs[a].J < rs[b].J
	})
	return rs
}

func TestCleanText_LowerCase(t *testing.T) {
	if got := cleanText("АЛГОРИТМ Дані"); got != "алгоритм дані" {
		t.Errorf("got %q, want %q", got, "алгоритм дані")
	}
}

func TestCleanText_Punctuation(t *testing.T) {
	if got := cleanText("текст, зі; знаками! пунктуації."); got != "текст зі знаками пунктуації" {
		t.Errorf("got %q", got)
	}
}

func TestCleanText_MultipleSpaces(t *testing.T) {
	if got := cleanText("слово    інше   слово"); got != "слово інше слово" {
		t.Errorf("got %q", got)
	}
}

func TestCleanText_Empty(t *testing.T) {
	if got := cleanText(""); got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestBuildShingles_KnownCount(t *testing.T) {
	s := buildShingles("а б в г", 3)
	if len(s) != 2 {
		t.Errorf("got %d shingles, want 2", len(s))
	}
	for _, sh := range []string{"а б в", "б в г"} {
		if _, ok := s[sh]; !ok {
			t.Errorf("missing shingle %q", sh)
		}
	}
}

func TestBuildShingles_TooShort(t *testing.T) {
	if s := buildShingles("а б", 3); len(s) != 0 {
		t.Errorf("got %d shingles, want 0", len(s))
	}
}

func TestBuildShingles_Duplicates(t *testing.T) {
	if s := buildShingles("а а а а а", 3); len(s) != 1 {
		t.Errorf("got %d shingles, want 1", len(s))
	}
}

func TestJaccard_Identical(t *testing.T) {
	a := makeSet("а б в", "б в г", "в г д")
	b := makeSet("а б в", "б в г", "в г д")
	if got := jaccard(a, b); math.Abs(got-1.0) > eps {
		t.Errorf("got %.9f, want 1.0", got)
	}
}

func TestJaccard_Disjoint(t *testing.T) {
	a := makeSet("а б в", "б в г")
	b := makeSet("х у з", "у з ж")
	if got := jaccard(a, b); math.Abs(got) > eps {
		t.Errorf("got %.9f, want 0.0", got)
	}
}

func TestJaccard_KnownValue(t *testing.T) {
	a := makeSet("а б в", "б в г", "в г д")
	b := makeSet("б в г", "в г д", "г д е")
	if got := jaccard(a, b); math.Abs(got-0.5) > eps {
		t.Errorf("got %.9f, want 0.5", got)
	}
}

func TestJaccard_OneEmpty(t *testing.T) {
	if got := jaccard(makeSet("а б в"), make(ShingleSet)); math.Abs(got) > eps {
		t.Errorf("got %.9f, want 0.0", got)
	}
}

func TestJaccard_BothEmpty(t *testing.T) {
	if got := jaccard(make(ShingleSet), make(ShingleSet)); math.Abs(got) > eps {
		t.Errorf("got %.9f, want 0.0", got)
	}
}

func TestSeqSimilarity_Identical(t *testing.T) {
	text := "алгоритм дані структура програмування мова функція змінна"
	sets := BuildShingleSetsSequential([]string{text, text})
	res := SequentialSimilarity(sets)
	if len(res) != 1 {
		t.Fatalf("expected 1 pair, got %d", len(res))
	}
	if math.Abs(res[0].Similarity-1.0) > eps {
		t.Errorf("identical texts: got %.9f, want 1.0", res[0].Similarity)
	}
}

func TestSeqSimilarity_Disjoint(t *testing.T) {
	textA := "алгоритм дані структура програмування мова функція"
	textB := "хмара річка гора вітер сонце місяць зоря"
	sets := BuildShingleSetsSequential([]string{textA, textB})
	res := SequentialSimilarity(sets)
	if len(res) != 1 {
		t.Fatalf("expected 1 pair, got %d", len(res))
	}
	if math.Abs(res[0].Similarity) > eps {
		t.Errorf("disjoint texts: got %.9f, want 0.0", res[0].Similarity)
	}
}

func TestSeqSimilarity_PairCount(t *testing.T) {
	texts := []string{
		"алгоритм дані структура програмування мова",
		"горутина канал блокування синхронізація паралельність",
		"компілятор інтерпретатор синтаксис семантика модуль",
		"хмара річка гора вітер сонце місяць зоря",
	}
	sets := BuildShingleSetsSequential(texts)
	res := SequentialSimilarity(sets)
	if len(res) != 6 {
		t.Errorf("pair count: got %d, want 6", len(res))
	}
}

func TestSeqSimilarity_PartialOverlap(t *testing.T) {
	textA := "алгоритм дані структура програмування мова функція змінна цикл умова"
	textB := "програмування мова функція змінна цикл умова масив рядок число логіка"
	sets := BuildShingleSetsSequential([]string{textA, textB})
	res := SequentialSimilarity(sets)
	sim := res[0].Similarity
	if sim <= 0 || sim >= 1 {
		t.Errorf("partial overlap: got %.9f, expected value in (0, 1)", sim)
	}
	t.Logf("partial overlap similarity = %.4f (очікується > 0 та < 1)", sim)
}

func TestParSimilarity_Identical(t *testing.T) {
	text := "алгоритм дані структура програмування мова функція змінна"
	sets := BuildShingleSetsParallel([]string{text, text})
	res := ParallelSimilarity(sets)
	if len(res) != 1 {
		t.Fatalf("expected 1 pair, got %d", len(res))
	}
	if math.Abs(res[0].Similarity-1.0) > eps {
		t.Errorf("identical texts: got %.9f, want 1.0", res[0].Similarity)
	}
}

func TestParSimilarity_Disjoint(t *testing.T) {
	textA := "алгоритм дані структура програмування мова функція"
	textB := "хмара річка гора вітер сонце місяць зоря"
	sets := BuildShingleSetsParallel([]string{textA, textB})
	res := ParallelSimilarity(sets)
	if len(res) != 1 {
		t.Fatalf("expected 1 pair, got %d", len(res))
	}
	if math.Abs(res[0].Similarity) > eps {
		t.Errorf("disjoint texts: got %.9f, want 0.0", res[0].Similarity)
	}
}

func TestParSimilarity_MatchesSequential_Small(t *testing.T) {
	corpus := []string{
		"алгоритм дані структура програмування мова функція змінна цикл умова масив",
		"горутина канал блокування синхронізація паралельність конкурентність інтерфейс",
		"компілятор інтерпретатор синтаксис семантика модуль бібліотека рядок число",
		"хмара річка гора вітер сонце місяць зоря поле ліс степ",
		"алгоритм дані структура мова функція змінна цикл умова масив рядок",
		"горутина канал блокування паралельність конкурентність клас обєкт метод поле",
		"компілятор інтерпретатор синтаксис модуль бібліотека число логіка схожість",
		"вітер сонце місяць зоря поле ліс степ хмара річка гора",
		"програмування мова функція змінна цикл умова масив рядок число логіка",
		"синхронізація паралельність конкурентність інтерфейс клас обєкт метод горутина",
		"семантика модуль бібліотека рядок число логіка компілятор інтерпретатор синтаксис",
		"поле ліс степ хмара річка гора вітер сонце місяць зоря",
		"алгоритм функція змінна цикл умова масив рядок число логіка схожість",
		"канал блокування синхронізація паралельність конкурентність інтерфейс клас обєкт метод",
		"інтерпретатор синтаксис семантика модуль бібліотека рядок число логіка компілятор",
		"річка гора вітер сонце місяць зоря поле ліс степ хмара",
		"дані структура програмування мова функція змінна цикл умова масив рядок",
		"блокування синхронізація паралельність конкурентність інтерфейс клас обєкт метод поле",
		"синтаксис семантика модуль бібліотека рядок число логіка компілятор інтерпретатор",
		"гора вітер сонце місяць зоря поле ліс степ хмара річка",
	}

	seqSets := BuildShingleSetsSequential(corpus)
	parSets := BuildShingleSetsParallel(corpus)

	seqRes := sortResults(SequentialSimilarity(seqSets))
	parRes := sortResults(ParallelSimilarity(parSets))

	if len(seqRes) != len(parRes) {
		t.Fatalf("result count mismatch: seq=%d par=%d", len(seqRes), len(parRes))
	}

	for i := range seqRes {
		s, p := seqRes[i], parRes[i]
		if s.I != p.I || s.J != p.J {
			t.Errorf("pair[%d]: seq=(%d,%d) par=(%d,%d)", i, s.I, s.J, p.I, p.J)
			continue
		}
		if math.Abs(s.Similarity-p.Similarity) > eps {
			t.Errorf("pair(%d,%d): seq=%.9f par=%.9f diff=%.2e",
				s.I, s.J, s.Similarity, p.Similarity, math.Abs(s.Similarity-p.Similarity))
		}
	}
}

func TestParSimilarity_MatchesSequential_Large(t *testing.T) {
	corpus, _ := GenerateCorpus(100)

	seqSets := BuildShingleSetsSequential(corpus)
	parSets := BuildShingleSetsParallel(corpus)

	seqRes := sortResults(SequentialSimilarity(seqSets))
	parRes := sortResults(ParallelSimilarity(parSets))

	if len(seqRes) != len(parRes) {
		t.Fatalf("result count mismatch: seq=%d par=%d", len(seqRes), len(parRes))
	}

	errCount := 0
	for i := range seqRes {
		s, p := seqRes[i], parRes[i]
		if math.Abs(s.Similarity-p.Similarity) > eps {
			errCount++
			if errCount <= 5 { // показуємо не більше 5 помилок
				t.Errorf("pair(%d,%d): seq=%.9f par=%.9f",
					s.I, s.J, s.Similarity, p.Similarity)
			}
		}
	}
	if errCount > 0 {
		t.Errorf("total mismatches: %d / %d", errCount, len(seqRes))
	}
}

func TestParSimilarity_PairCount(t *testing.T) {
	corpus, _ := GenerateCorpus(50)
	sets := BuildShingleSetsParallel(corpus)
	res := ParallelSimilarity(sets)
	want := 50 * 49 / 2 // = 1225
	if len(res) != want {
		t.Errorf("pair count: got %d, want %d", len(res), want)
	}
}
