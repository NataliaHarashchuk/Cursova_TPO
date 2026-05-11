package main

const (
	shingleSize  = 3
	numTexts     = 1000
	wordsInTexts = 1000
)

type ShingleSet = map[string]struct{}

type PairResult struct {
	I, J       int
	Similarity float64
}

type Job struct {
	I, J int
	SetA ShingleSet
	SetB ShingleSet
}
