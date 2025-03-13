package main

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

func main() {
	args := os.Args
	var filename string

	if len(args) < 2 {
		filename = "input"
	} else {
		filename = os.Args[1]
	}

	input, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading file: ", err)
		return
	}

	strInp := string(input)

	fmt.Println("Running Synchronous Version on ", filename)
	start := time.Now()
	cost := calcTotalCost(strInp)
	end := time.Now()
	timeSeq := end.Sub(start).Nanoseconds()

	fmt.Println("Cost: ", cost)
	fmt.Println("executed in ", timeSeq, "ns.")

	fmt.Println("Running Parallel Version (walk to nearest newline) with 16 threads on ", filename)
	start = time.Now()
	cost = calcTotalCostParallelWalkToNearestNewLine(strInp, 16)
	end = time.Now()
	timePar := end.Sub(start).Nanoseconds()

	fmt.Println("Cost: ", cost)
	fmt.Println("executed in ", timePar, "ns.")
	fmt.Println("Speedup: ", (float32(timeSeq) / float32(timePar)))
}

func calcTotalCost(inp string) int {
	res := 0

	// i := 0
	chars := 0
	bytes := 0
	for _, c := range inp {
		if c != '\n' {
			chars += 1
			bytes += utf8.RuneLen(c)
		} else {
			// add := strToCost(chars, bytes)
			// fmt.Printf("cost for {\"%.12s\", chars(%d), bytes(%d)} is %d\n", inp[i:j], chars, bytes, add)
			res += strInfoToCost(chars, bytes)
			// i = j + 1
			chars = 0
			bytes = 0
		}
	}

	if chars != 0 && bytes != 0 {
		// add := strToCost(chars, bytes)
		// fmt.Printf("cost for {\"%.12s\", chars(%d), bytes(%d)} is %d\n", inp[i:j], chars, bytes, add)
		res += strInfoToCost(chars, bytes)
	}
	return res
}

func strInfoToCost(chars, bytes int) int {
	val := 0
	if chars <= 140 && bytes <= 160 {
		val = 13
	} else if bytes <= 160 {
		val = 11
	} else if chars <= 140 {
		val = 7
	}

	return val
}

func calcTotalCostParallelWalkToNearestNewLine(inp string, nThreads int) int {
	var wg sync.WaitGroup
	var res int32
	strLen := len(inp)
	index := 0
	prevIndex := 0

	for i := 0; i < nThreads && index < strLen; i++ {
		index = index + (strLen-index)/(nThreads-i)
		for index+1 < strLen && inp[index] != '\n' {
			index += 1
		}

		wg.Add(1)
		go func(begin, end int) {
			atomic.AddInt32(&res, int32(calcTotalCost(inp[begin:end])))
			defer wg.Done()
		}(prevIndex, index)
		prevIndex = index + 1
	}

	wg.Wait()

	return int(res)
}

type record struct {
	chars, bytes int
}

func calcTotalCostParallelCoalesceToSingleThread(inp string, nThreads int) int {
	var wg sync.WaitGroup
	var res int32
	pre := make([]record, nThreads)
	post := make([]record, nThreads)
	chunkSize := len(inp) / nThreads

	for i := 0; i < nThreads; i++ {
		wg.Add(1)
		j, k := chunkSize*i, chunkSize*(i+1)

		if i == nThreads-1 {
			k = len(inp)
		}

		go func(begin, end, index int) {

			atomic.AddInt32(&res, int32(calcTotalCost(inp[begin:end])))
			defer wg.Done()
		}(j, k, i)
	}
	wg.Wait()

	for i := 1; i < nThreads; i++ {
		res += int32(strInfoToCost(pre[i].chars+post[i-1].chars, pre[i].bytes+post[i-1].bytes))
	}

	return int(res)
}

func calcTotalCostReturnTrailing(inp string) (int, int, int) {
	res := 0

	chars := 0
	bytes := 0
	for _, c := range inp {
		if c != '\n' {
			chars += 1
			bytes += utf8.RuneLen(c)
		} else {
			res += strInfoToCost(chars, bytes)
			chars = 0
			bytes = 0
		}
	}

	return res, chars, bytes
}
