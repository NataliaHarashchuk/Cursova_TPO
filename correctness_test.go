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
	got := cleanText("текст, зі; знаками! пунктуації.")
	if got != "текст знак пунктуації" {
		t.Errorf("got %q, want %q", got, "текст знак пунктуації")
	}
}

func TestCleanText_MultipleSpaces(t *testing.T) {
	got := cleanText("слово    інше   слово")
	if got != "слов інше слов" {
		t.Errorf("got %q, want %q", got, "слов інше слов")
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
		"горутина канал мютекс синхронізація паралельність",
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

func TestPipeline_Identical(t *testing.T) {
	text := "алгоритм дані структура програмування мова функція змінна цикл умова"
	res := PipelineSimilarity([]string{text, text})
	if len(res) != 1 {
		t.Fatalf("expected 1 pair, got %d", len(res))
	}
	if math.Abs(res[0].Similarity-1.0) > eps {
		t.Errorf("identical texts: got %.9f, want 1.0", res[0].Similarity)
	}
}

func TestPipeline_Disjoint(t *testing.T) {
	textA := "алгоритм дані структура програмування мова функція"
	textB := "хмара річка гора вітер сонце місяць зоря"
	res := PipelineSimilarity([]string{textA, textB})
	if len(res) != 1 {
		t.Fatalf("expected 1 pair, got %d", len(res))
	}
	if math.Abs(res[0].Similarity) > eps {
		t.Errorf("disjoint texts: got %.9f, want 0.0", res[0].Similarity)
	}
}

func TestPipeline_PartialOverlap(t *testing.T) {
	textA := "алгоритм дані структура програмування мова функція змінна цикл умова"
	textB := "програмування мова функція змінна цикл умова масив рядок число логіка"
	res := PipelineSimilarity([]string{textA, textB})
	if len(res) != 1 {
		t.Fatalf("expected 1 pair, got %d", len(res))
	}
	sim := res[0].Similarity
	if sim <= 0 || sim >= 1 {
		t.Errorf("partial overlap: got %.9f, expected (0, 1)", sim)
	}
	t.Logf("partial overlap similarity = %.4f", sim)
}

func TestPipeline_PairCount_Small(t *testing.T) {
	texts := []string{
		"алгоритм дані структура програмування мова функція",
		"горутина канал мютекс синхронізація паралельність конкурентність",
		"компілятор інтерпретатор синтаксис семантика модуль бібліотека",
		"хмара річка гора вітер сонце місяць зоря поле ліс степ",
	}
	res := PipelineSimilarity(texts)
	want := 4 * 3 / 2 // = 6
	if len(res) != want {
		t.Errorf("pair count: got %d, want %d", len(res), want)
	}
}

func TestPipeline_PairCount_Large(t *testing.T) {
	corpus, _ := GenerateCorpus(50)
	res := PipelineSimilarity(corpus)
	want := 50 * 49 / 2 // = 1225
	if len(res) != want {
		t.Errorf("pair count: got %d, want %d", len(res), want)
	}
}

func TestPipeline_MatchesSequential_Small(t *testing.T) {
	corpus := []string{
		"алгоритм дані структура програмування мова функція змінна цикл умова масив",
		"горутина канал мютекс синхронізація паралельність конкурентність інтерфейс",
		"компілятор інтерпретатор синтаксис семантика модуль бібліотека рядок число",
		"хмара річка гора вітер сонце місяць зоря поле ліс степ",
		"алгоритм дані структура мова функція змінна цикл умова масив рядок",
		"горутина канал мютекс паралельність конкурентність клас обєкт метод поле",
		"компілятор інтерпретатор синтаксис модуль бібліотека число логіка схожість",
		"вітер сонце місяць зоря поле ліс степ хмара річка гора",
		"програмування мова функція змінна цикл умова масив рядок число логіка",
		"синхронізація паралельність конкурентність інтерфейс клас обєкт метод горутина",
		"семантика модуль бібліотека рядок число логіка компілятор інтерпретатор синтаксис",
		"поле ліс степ хмара річка гора вітер сонце місяць зоря",
		"алгоритм функція змінна цикл умова масив рядок число логіка схожість",
		"канал мютекс синхронізація паралельність конкурентність інтерфейс клас обєкт метод",
		"інтерпретатор синтаксис семантика модуль бібліотека рядок число логіка компілятор",
		"річка гора вітер сонце місяць зоря поле ліс степ хмара",
		"дані структура програмування мова функція змінна цикл умова масив рядок",
		"мютекс синхронізація паралельність конкурентність інтерфейс клас обєкт метод поле",
		"синтаксис семантика модуль бібліотека рядок число логіка компілятор інтерпретатор",
		"гора вітер сонце місяць зоря поле ліс степ хмара річка",
	}

	seqSets := BuildShingleSetsSequential(corpus)
	seqRes := sortResults(SequentialSimilarity(seqSets))

	pipeRes := sortResults(PipelineSimilarity(corpus))

	if len(seqRes) != len(pipeRes) {
		t.Fatalf("result count mismatch: seq=%d pipe=%d", len(seqRes), len(pipeRes))
	}

	for i := range seqRes {
		s, p := seqRes[i], pipeRes[i]
		if s.I != p.I || s.J != p.J {
			t.Errorf("pair[%d]: seq=(%d,%d) pipe=(%d,%d)", i, s.I, s.J, p.I, p.J)
			continue
		}
		if math.Abs(s.Similarity-p.Similarity) > eps {
			t.Errorf("pair(%d,%d): seq=%.9f pipe=%.9f diff=%.2e",
				s.I, s.J, s.Similarity, p.Similarity,
				math.Abs(s.Similarity-p.Similarity))
		}
	}
	t.Logf(" всі %d пар: pipeline == sequential", len(seqRes))
}

func TestPipeline_MatchesSequential_Large(t *testing.T) {
	corpus, _ := GenerateCorpus(100)

	seqSets := BuildShingleSetsSequential(corpus)
	seqRes := sortResults(SequentialSimilarity(seqSets))
	pipeRes := sortResults(PipelineSimilarity(corpus))

	if len(seqRes) != len(pipeRes) {
		t.Fatalf("result count mismatch: seq=%d pipe=%d", len(seqRes), len(pipeRes))
	}

	errCount := 0
	for i := range seqRes {
		s, p := seqRes[i], pipeRes[i]
		if math.Abs(s.Similarity-p.Similarity) > eps {
			errCount++
			if errCount <= 5 {
				t.Errorf("pair(%d,%d): seq=%.9f pipe=%.9f",
					s.I, s.J, s.Similarity, p.Similarity)
			}
		}
	}
	if errCount > 0 {
		t.Errorf("total mismatches: %d / %d", errCount, len(seqRes))
	} else {
		t.Logf(" всі %d пар: pipeline == sequential", len(seqRes))
	}
}

func TestPipeline_NoDuplicatePairs(t *testing.T) {
	corpus, _ := GenerateCorpus(30)
	res := PipelineSimilarity(corpus)

	seen := make(map[[2]int]int, len(res))
	for _, r := range res {
		key := [2]int{r.I, r.J}
		seen[key]++
	}

	duplicates := 0
	for key, cnt := range seen {
		if cnt > 1 {
			duplicates++
			t.Errorf("duplicate pair (%d,%d): appeared %d times", key[0], key[1], cnt)
		}
	}
	want := 30 * 29 / 2 // = 435
	if len(seen) != want {
		t.Errorf("unique pair count: got %d, want %d", len(seen), want)
	}
	if duplicates == 0 {
		t.Logf(" жодних дублікатів, %d унікальних пар", len(seen))
	}
}

func TestPipeline_EmptyCorpus(t *testing.T) {
	res := PipelineSimilarity([]string{})
	if res != nil && len(res) != 0 {
		t.Errorf("empty corpus: expected nil/empty, got %d results", len(res))
	}
}

func TestPipeline_SingleText(t *testing.T) {
	res := PipelineSimilarity([]string{"алгоритм дані структура програмування мова"})
	if len(res) != 0 {
		t.Errorf("single text: expected 0 pairs, got %d", len(res))
	}
}
