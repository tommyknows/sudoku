package main

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/tommyknows/sudoku/pkg/sudoku"
)

func BenchmarkSolve() {
	sudokus := readSudokusFromFile("sudokus.txt")
	for i, s := range sudokus {
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
		fmt.Printf("Mean Time for Sudoku %v: %v\n", i, time.Duration(res.NsPerOp()))
	}
}

func main() {
	BenchmarkSolve()
}

func readSudokusFromFile(filename string) []string {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return strings.Split(strings.Trim(string(content), "\n"), "\n")
}
