package main

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

func main() {

	fmt.Println(" Аналіз схожості текстів — Алгоритм Шинглів")
	fmt.Printf("\nКількість текстів у корпусі : %d\n", numTexts)
	fmt.Printf("Розмір шингля (N-грама)     : %d слова\n", shingleSize)
	fmt.Printf("Кількість CPU-ядер          : %d\n", runtime.NumCPU())
	fmt.Printf("Розмір batch (пар/Job)      : %d\n", jobBatchSize)

	sep("─")
	fmt.Println("ЕТАП 1 · Генерація корпусу")
	sep("─")

	t0 := time.Now()
	corpus, nGenerated := GenerateCorpus(numTexts)
	tCorpus := time.Since(t0)
	fmt.Printf(" %d текстів згенеровано паралельно за %v\n\n", nGenerated, tCorpus)

	sep("─")
	fmt.Println("ЕТАП 2 · Послідовний алгоритм (baseline)")
	sep("─")
	fmt.Print("  [Послідовно] prep + порівняння всіх пар... ")

	t0 = time.Now()
	seqSets := BuildShingleSetsSequential(corpus)
	seqResults := SequentialSimilarity(seqSets)
	tSeqTotal := time.Since(t0)

	numPairs := numTexts * (numTexts - 1) / 2
	fmt.Printf("готово за %v\n", tSeqTotal)
	fmt.Printf("  Пар оброблено : %d\n\n", numPairs)

	sep("─")
	fmt.Println(" ЕТАП 3 · Паралельний конвеєр (PipelineSimilarity)")
	sep("─")
	fmt.Printf("  [Конвеєр, %d ядер] prep + порівняння... ", runtime.NumCPU())

	t0 = time.Now()
	pipeResults := PipelineSimilarity(corpus)
	tPipeTotal := time.Since(t0)

	fmt.Printf("готово за %v\n", tPipeTotal)
	fmt.Printf("  Пар оброблено : %d\n\n", len(pipeResults))

	sep("═")
	fmt.Println("  ЗВЕДЕНІ РЕЗУЛЬТАТИ")
	sep("═")
	fmt.Printf("  %-38s %14s %14s %10s\n", "Варіант", "Час (seq)", "Час (par)", "Speedup")
	sep("─")
	fmt.Printf("  %-38s %14v %14v %9.2fx\n",
		"Повний цикл (prep + порівняння)",
		tSeqTotal.Round(time.Millisecond),
		tPipeTotal.Round(time.Millisecond),
		ratio(tSeqTotal, tPipeTotal),
	)
	sep("═")

	fmt.Println("\n  Топ-5 найбільш схожих пар текстів (за послідовним результатом):")
	sep("─")
	for rank, r := range topSimilar(seqResults, 5) {
		fmt.Printf("  %d. Текст #%-4d ↔ Текст #%-4d  J = %.4f\n",
			rank+1, r.I+1, r.J+1, r.Similarity)
	}

	const threshold = 0.3
	var sum float64
	above := 0
	for _, r := range seqResults {
		sum += r.Similarity
		if r.Similarity >= threshold {
			above++
		}
	}
	total := float64(len(seqResults))
	sep("─")
	fmt.Printf("  Середній коефіцієнт Жаккара  : %.4f\n", sum/total)
	fmt.Printf("  Пар зі схожістю ≥ %.1f        : %d (%.2f%%)\n",
		threshold, above, 100*float64(above)/total)
	sep("═")
}

func sep(char string) { fmt.Println(strings.Repeat(char, 58)) }

func ratio(a, b time.Duration) float64 {
	if b == 0 {
		return 1
	}
	return float64(a) / float64(b)
}

func topSimilar(results []PairResult, k int) []PairResult {
	cp := make([]PairResult, len(results))
	copy(cp, results)
	if k > len(cp) {
		k = len(cp)
	}
	for i := 0; i < k; i++ {
		maxIdx := i
		for j := i + 1; j < len(cp); j++ {
			if cp[j].Similarity > cp[maxIdx].Similarity {
				maxIdx = j
			}
		}
		cp[i], cp[maxIdx] = cp[maxIdx], cp[i]
	}
	return cp[:k]
}
