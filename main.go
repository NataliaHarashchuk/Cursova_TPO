package main

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

func main() {
	fmt.Println("Аналіз схожості текстів — Алгоритм Шинглів")
	fmt.Printf("\nКількість текстів у корпусі : %d\n", numTexts)
	fmt.Printf("Розмір шингля (N-грама)     : %d слова\n", shingleSize)
	fmt.Printf("Кількість CPU-ядер          : %d\n", runtime.NumCPU())

	fmt.Println("ЕТАП 1 · Генерація корпусу")

	t0 := time.Now()
	corpus, nGenerated := GenerateCorpus(numTexts)
	tCorpus := time.Since(t0)

	fmt.Printf("%d текстів згенеровано паралельно за %v\n\n", nGenerated, tCorpus)

	fmt.Println("ЕТАП 2 · Препроцесинг (cleanText + buildShingles)")

	fmt.Print("Послідовно будуємо шингл-множини... ")
	t0 = time.Now()
	_ = BuildShingleSetsSequential(corpus)
	tSeqPrep := time.Since(t0)
	fmt.Printf("готово за %v\n", tSeqPrep)

	fmt.Print("Паралельно будуємо шингл-множини... ")
	t0 = time.Now()
	shingleSets := BuildShingleSetsParallel(corpus)
	tParPrep := time.Since(t0)
	fmt.Printf("готово за %v\n", tParPrep)

	numPairs := numTexts * (numTexts - 1) / 2
	fmt.Printf("\n Прискорення препроцесингу   : %.2fx\n", ratio(tSeqPrep, tParPrep))
	fmt.Printf("Загальна кількість пар      : %d\n\n", numPairs)

	fmt.Println(" ЕТАП 3 · Порівняння пар (коефіцієнт Жаккара)")

	fmt.Print("Послідовно порівнюємо пари... ")
	t0 = time.Now()
	seqResults := SequentialSimilarity(shingleSets)
	tSeqCmp := time.Since(t0)
	fmt.Printf("готово за %v  (%d пар)\n", tSeqCmp, len(seqResults))

	fmt.Printf("Паралельно порівнюємо пари (%d воркерів)... ", runtime.NumCPU())
	t0 = time.Now()
	parResults := ParallelSimilarity(shingleSets)
	tParCmp := time.Since(t0)
	fmt.Printf("готово за %v  (%d пар)\n", tParCmp, len(parResults))

	fmt.Printf("\n Прискорення порівняння      : %.2fx\n\n", ratio(tSeqCmp, tParCmp))

	fmt.Println("  ЗВЕДЕНІ РЕЗУЛЬТАТИ")
	fmt.Printf("  %-32s %12s %12s %10s\n", "Етап", "Послідовно", "Паралельно", "Speedup")
	sep("─")
	printRow("Препроцесинг (cleanText+shingle)", tSeqPrep, tParPrep)
	printRow("Порівняння пар (Jaccard)", tSeqCmp, tParCmp)
	sep("─")
	printRow("ЗАГАЛОМ", tSeqPrep+tSeqCmp, tParPrep+tParCmp)
	sep("═")

	fmt.Println("\n  Топ-5 найбільш схожих пар текстів:")
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

func sep(char string) {
	fmt.Println(strings.Repeat(char, 52))
}

func ratio(a, b time.Duration) float64 {
	if b == 0 {
		return 1
	}
	return float64(a) / float64(b)
}

func printRow(label string, seq, par time.Duration) {
	fmt.Printf("  %-32s %12v %12v %9.2fx\n",
		label,
		seq.Round(time.Microsecond),
		par.Round(time.Microsecond),
		ratio(seq, par))
}

func topSimilar(results []PairResult, k int) []PairResult {
	cp := make([]PairResult, len(results))
	copy(cp, results)
	if k > len(cp) {
		k = len(cp)
	}
	for i := 0; i < k; i++ {
		max := i
		for j := i + 1; j < len(cp); j++ {
			if cp[j].Similarity > cp[max].Similarity {
				max = j
			}
		}
		cp[i], cp[max] = cp[max], cp[i]
	}
	return cp[:k]
}
