package main

import "sync"

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

func PipelineSimilarity(corpus []string) []PairResult {
	n := len(corpus)
	if n == 0 {
		return nil
	}

	numChunks := numWorkers
	if numChunks > n {
		numChunks = n
	}
	chunkSize := (n + numChunks - 1) / numChunks

	type ChunkResult struct {
		Start int
		Sets  []ShingleSet
	}

	processedChunks := make(chan ChunkResult, numChunks)

	var prepWg sync.WaitGroup
	for start := 0; start < n; start += chunkSize {
		end := start + chunkSize
		if end > n {
			end = n
		}
		prepWg.Add(1)
		go func(s, e int) {
			defer prepWg.Done()
			sets := make([]ShingleSet, e-s)
			for i := s; i < e; i++ {
				sets[i-s] = buildShingles(cleanText(corpus[i]), shingleSize)
			}
			processedChunks <- ChunkResult{Start: s, Sets: sets}
		}(start, end)
	}
	go func() { prepWg.Wait(); close(processedChunks) }()

	jobs := make(chan BatchJob, numWorkers*8)
	cmpResults := make(chan []PairResult, numWorkers*8)
	allSets := make([]ShingleSet, n)

	var cmpWg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		cmpWg.Add(1)
		go func() {
			defer cmpWg.Done()
			for batch := range jobs {
				out := make([]PairResult, 0, len(batch.Pairs))
				for _, p := range batch.Pairs {
					out = append(out, PairResult{
						I:          p.I,
						J:          p.J,
						Similarity: jaccard(allSets[p.I], allSets[p.J]),
					})
				}
				cmpResults <- out
			}
		}()
	}
	go func() { cmpWg.Wait(); close(cmpResults) }()

	go func() {
		type chunkMeta struct{ start, end int }
		var ready []chunkMeta
		buf := make([]PairIndex, 0, jobBatchSize)

		flush := func() {
			if len(buf) > 0 {
				jobs <- BatchJob{Pairs: buf}
				buf = make([]PairIndex, 0, jobBatchSize)
			}
		}

		emit := func(i, j int) {
			if i > j {
				i, j = j, i
			}
			buf = append(buf, PairIndex{I: i, J: j})
			if len(buf) == jobBatchSize {
				flush()
			}
		}

		for cr := range processedChunks {
			for i, s := range cr.Sets {
				allSets[cr.Start+i] = s
			}
			newMeta := chunkMeta{start: cr.Start, end: cr.Start + len(cr.Sets)}

			for i := newMeta.start; i < newMeta.end-1; i++ {
				for j := i + 1; j < newMeta.end; j++ {
					emit(i, j)
				}
			}

			for _, prev := range ready {
				for i := prev.start; i < prev.end; i++ {
					for j := newMeta.start; j < newMeta.end; j++ {
						emit(i, j)
					}
				}
			}
			ready = append(ready, newMeta)
		}
		flush()
		close(jobs)
	}()

	numPairs := n * (n - 1) / 2
	all := make([]PairResult, 0, numPairs)
	for batch := range cmpResults {
		all = append(all, batch...)
	}
	return all
}
