//go:build ignore

package main

import (
	"flag"
	"fmt"
	"math"
	"runtime"
	"strings"
	"time"
)

const RUNS = 20

var byNumTexts = []struct{ NumTexts, WordsPerText int }{
	{50, 500},
	{100, 500},
	{200, 500},
	{300, 500},
	{500, 500},
	{700, 500},
	{1000, 500},
	{1500, 500},
	{2000, 500},
	{3000, 500},
	{5000, 500},
	{5000, 500},
}

var byWordsPerText = []struct{ NumTexts, WordsPerText int }{
	{200, 50},
	{200, 100},
	{200, 200},
	{200, 500},
	{200, 700},
	{200, 1000},
	{200, 1500},
	{200, 2000},
	{200, 3000},
	{200, 5000},
	{200, 10000},
}

func calcStats(samples []float64) (mean, stddev, min, max float64) {
	min, max = samples[0], samples[0]
	for _, v := range samples {
		mean += v
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	mean /= float64(len(samples))
	for _, v := range samples {
		d := v - mean
		stddev += d * d
	}
	stddev = math.Sqrt(stddev / float64(len(samples)))
	return
}

func speedup(seq, par float64) float64 {
	if par == 0 {
		return 1
	}
	return seq / par
}

func measureSeq(corpus []string) (meanMs, stdMs float64) {
	samples := make([]float64, RUNS)
	for i := range samples {
		t0 := time.Now()
		sets := BuildShingleSetsSequential(corpus)
		_ = SequentialSimilarity(sets)
		samples[i] = float64(time.Since(t0).Microseconds()) / 1000.0
	}
	mean, std, _, _ := calcStats(samples)
	return mean, std
}

func measurePipeline(corpus []string) (meanMs, stdMs float64) {
	samples := make([]float64, RUNS)
	for i := range samples {
		t0 := time.Now()
		_ = PipelineSimilarity(corpus)
		samples[i] = float64(time.Since(t0).Microseconds()) / 1000.0
	}
	mean, std, _, _ := calcStats(samples)
	return mean, std
}

type rowResult struct {
	key          int
	seqMean      float64
	seqStd       float64
	pipeMean     float64
	pipeStd      float64
	totalSpeedup float64
}

func runExperiment(configs []struct{ NumTexts, WordsPerText int }, varyWords bool) []rowResult {
	rows := make([]rowResult, 0, len(configs))

	for _, cfg := range configs {
		key := cfg.NumTexts
		if varyWords {
			key = cfg.WordsPerText
		}
		fmt.Printf("  [%d текстів × %d слів] генерація...", cfg.NumTexts, cfg.WordsPerText)

		corpus := make([]string, cfg.NumTexts)
		for i := range corpus {
			corpus[i] = generateText(cfg.WordsPerText)
		}
		fmt.Printf(" вимірювання (%d прогонів)...", RUNS)

		seqMean, seqStd := measureSeq(corpus)
		pipeMean, pipeStd := measurePipeline(corpus)
		sp := speedup(seqMean, pipeMean)

		fmt.Printf(" seq=%.1fмс pipe=%.1fмс sp=%.2fx\n", seqMean, pipeMean, sp)

		rows = append(rows, rowResult{
			key:          key,
			seqMean:      seqMean,
			seqStd:       seqStd,
			pipeMean:     pipeMean,
			pipeStd:      pipeStd,
			totalSpeedup: sp,
		})
	}
	return rows
}

func printTable(title string, rows []rowResult, varyWords bool) {
	line110 := strings.Repeat("═", 80)
	dash80 := strings.Repeat("─", 80)

	fmt.Println("\n" + line110)
	fmt.Println("  " + title)
	fmt.Println(line110)

	colLabel := "Текстів"
	if varyWords {
		colLabel = "Слів/текст"
	}
	fmt.Printf("  %-12s  %20s  %20s  %10s\n",
		colLabel, "Seq total (мс)", "Pipeline total (мс)", "Speedup")
	fmt.Println(dash80)

	for _, r := range rows {
		fmt.Printf("  %-12d  %14.2f ±%-5.1f  %14.2f ±%-5.1f  %8.2fx\n",
			r.key,
			r.seqMean, r.seqStd,
			r.pipeMean, r.pipeStd,
			r.totalSpeedup,
		)
	}
	fmt.Println(dash80)
}

func main() {
	procs := flag.Int("procs", runtime.NumCPU(), "кількість логічних процесорів")
	flag.Parse()
	runtime.GOMAXPROCS(*procs)
	numWorkers = *procs

	fmt.Printf("Прогонів на конфігурацію: %d\n", RUNS)
	fmt.Printf("jobBatchSize             : %d пар/batch\n", jobBatchSize)
	fmt.Printf("numWorkers               : %d\n\n", numWorkers)

	fmt.Println("ЕКСПЕРИМЕНТ 1: вплив КІЛЬКОСТІ ТЕКСТІВ")
	rows1 := runExperiment(byNumTexts, false)
	printTable(
		"Кількість текстів vs загальний час (500 слів, shingleSize=3)",
		rows1, false,
	)

	fmt.Println("\n\nЕКСПЕРИМЕНТ 2: вплив ДОВЖИНИ ТЕКСТУ")
	rows2 := runExperiment(byWordsPerText, true)
	printTable(
		"Довжина тексту vs загальний час (200 текстів, shingleSize=3)",
		rows2, true,
	)

	fmt.Println("\nЕксперимент завершено.")
}
