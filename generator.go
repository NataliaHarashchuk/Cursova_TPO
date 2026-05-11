package main

import (
	"math/rand/v2"
	"strings"
	"sync"
	"sync/atomic"
)

var wordPool = []string{
	"алгоритм", "дані", "структура", "програмування", "мова", "функція",
	"змінна", "цикл", "умова", "масив", "рядок", "число", "логіка",
	"компілятор", "інтерпретатор", "синтаксис", "семантика", "модуль",
	"бібліотека", "інтерфейс", "клас", "обєкт", "метод", "поле",
	"горутина", "канал", "мютекс", "синхронізація", "паралельність",
	"конкурентність", "схожість", "шингл", "жаккар", "текст", "аналіз",
}

func generateText(wordsCount int) string {
	words := make([]string, wordsCount)
	poolLen := len(wordPool)
	for i := range words {
		words[i] = wordPool[rand.IntN(poolLen)]
	}
	return strings.Join(words, " ")
}

func GenerateCorpus(n int) ([]string, int64) {
	corpus := make([]string, n)

	baseA := strings.Repeat("алгоритм дані структура програмування мова функція змінна цикл умова масив ", 3)
	baseB := strings.Repeat("горутина канал мютекс синхронізація паралельність конкурентність інтерфейс клас ", 3)
	baseC := strings.Repeat("компілятор інтерпретатор синтаксис семантика модуль бібліотека рядок число логіка поле ", 3)

	var (
		wg        sync.WaitGroup
		generated atomic.Int64
	)

	for i := range corpus {
		wg.Add(1)
		go func(idx int) {
			defer func() {
				wg.Done()
				generated.Add(1)
			}()
			switch {
			case idx%4 == 0:
				corpus[idx] = baseA + generateText(15)
			case idx%4 == 1:
				corpus[idx] = baseB + generateText(15)
			case idx%6 == 2:
				corpus[idx] = baseC + generateText(15)
			default:
				corpus[idx] = generateText(wordsInTexts)
			}
		}(i)
	}
	wg.Wait()

	return corpus, generated.Load()
}
