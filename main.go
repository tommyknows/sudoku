package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/tommyknows/sudoku/pkg/sudoku"
)

func main() {
	fromInput := flag.String("i", "", "input for the sudoku-solver")
	runBenchmark := flag.Bool("bench", true, "run benchmarks on a sudoku (or the default sudoku if i is not specified)")
	flag.Parse()

	var sudokus []string
	if *fromInput != "" {
		sudokus = append(sudokus, *fromInput)
	} else {
		sudokus = readSudokusFromFile("sudokus.txt")
	}

	f := solve
	if *runBenchmark {
		f = benchmark
	}
	f(sudokus)
}

func benchmark(sudokus []string) {
	for _, s := range sudokus {
		var su *sudoku.Sudoku
		res := testing.Benchmark(func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				su = sudoku.New(s)
				err := su.Solve()
				if err != nil {
					b.Error(err)
				}
			}
		})
		fmt.Println(su)
		fmt.Printf("Mean Time for Sudoku: %v\n", time.Duration(res.NsPerOp()))
	}
}

func solve(sudokus []string) {
	for _, s := range sudokus {
		su := sudoku.New(s)
		if err := su.Solve(); err != nil {
			fmt.Printf("ERROR: %v\n", err)
			continue
		}
		fmt.Println(su)
	}
}

func readSudokusFromFile(filename string) []string {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return strings.Split(strings.Trim(string(content), "\n"), "\n")
}
