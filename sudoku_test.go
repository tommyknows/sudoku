package sudoku

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestRemove(t *testing.T) {
	is := is.New(t)
	x := []index{"A1", "B1", "C1", "D1", "E1"}
	r := index("C1")
	y := remove(x, r)
	is.True(contains(y, r) == false)
	is.True(contains(x, r))
}

func TestSolve(t *testing.T) {
	s := New("003020600900305001001806400008102900700000008006708200002609500800203009005010300")
	t.Log(s)
	err := s.Solve()
	if err != nil {
		t.Errorf("Could not solve sudoku: %v", s)
	}
	t.Log(s)
	s = New("4.....8.5.3..........7......2.....6.....8.4......1.......6.3.7.5..2.....1.4......")
	t.Log(s)
	err = s.Solve()
	if err != nil {
		t.Errorf("Could not solve sudoku: %v", s)
	}
	t.Log(s)
}

func TestSearch(t *testing.T) {
	sudokus := readSudokusFromFile("sudokus.txt")
	for i, s := range sudokus {
		res := testing.Benchmark(func(b *testing.B) {
			//b.Run(fmt.Sprintf("sudoku %v", i), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				su := New(s)
				err := su.Solve()
				if err != nil {
					b.Error(err)
				}
			}
		})
		fmt.Printf("Mean Time for Sudoku %v: %v\n", i, time.Duration(res.NsPerOp()))
	}
}

func readSudokusFromFile(filename string) []string {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return strings.Split(strings.Trim(string(content), "\n"), "\n")
}
