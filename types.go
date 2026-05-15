package main

import "runtime"

const (
	shingleSize  = 3
	numTexts     = 1000
	wordsInTexts = 1000
	jobBatchSize = 64
)

var numWorkers = runtime.NumCPU()

type ShingleSet = map[string]struct{}

type PairResult struct {
	I, J       int
	Similarity float64
}

type PairIndex struct {
	I, J int
}

type BatchJob struct {
	Pairs []PairIndex
}

type Job struct {
	I, J int
	SetA ShingleSet
	SetB ShingleSet
}
