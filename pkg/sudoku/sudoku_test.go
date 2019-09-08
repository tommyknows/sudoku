package sudoku

import (
	"fmt"
	"testing"

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

// This is how godoc parsed ExampleExamples_output() that was shown above.
func ExampleSudoku() {
	s := New("003020600900305001001806400008102900700000008006708200002609500800203009005010300")
	// show the parsed sudoku
	fmt.Print(s)
	err := s.Solve()
	if err != nil {
		panic(err)
	}
	fmt.Print(s)
}
