package main

import (
	"runtime"
	"sync"
)

func jaccard(a, b ShingleSet) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 0
	}

	small, large := a, b
	if len(a) > len(b) {
		small, large = b, a
	}

	intersection := 0
	for k := range small {
		if _, ok := large[k]; ok {
			intersection++
		}
	}

	union := len(a) + len(b) - intersection
	return float64(intersection) / float64(union)
}

func SequentialSimilarity(shingleSets []ShingleSet) []PairResult {
	n := len(shingleSets)
	results := make([]PairResult, 0, n*(n-1)/2)

	for i := 0; i < n-1; i++ {
		for j := i + 1; j < n; j++ {
			results = append(results, PairResult{
				I:          i,
				J:          j,
				Similarity: jaccard(shingleSets[i], shingleSets[j]),
			})
		}
	}
	return results
}

func ParallelSimilarity(shingleSets []ShingleSet) []PairResult {
	n := len(shingleSets)
	numPairs := n * (n - 1) / 2
	numWorkers := runtime.NumCPU()

	jobs := make(chan Job, numWorkers*8)
	results := make(chan PairResult, numWorkers*8)

	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				results <- PairResult{
					I:          job.I,
					J:          job.J,
					Similarity: jaccard(job.SetA, job.SetB),
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	go func() {
		for i := 0; i < n-1; i++ {
			for j := i + 1; j < n; j++ {
				jobs <- Job{I: i, J: j, SetA: shingleSets[i], SetB: shingleSets[j]}
			}
		}
		close(jobs)
	}()

	all := make([]PairResult, 0, numPairs)
	for r := range results {
		all = append(all, r)
	}
	return all
}
