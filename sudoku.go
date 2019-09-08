package sudoku

import (
	"fmt"
	"strings"
)

// something like "A1"
type index string

// all possible values, as a single string
type value string

// some simple helper functions that handle the conversions
func (v value) remove(val value) value  { return value(strings.ReplaceAll(string(v), string(val), "")) }
func (v value) contains(val value) bool { return strings.Contains(string(v), string(val)) }
func (v value) isZero() bool            { return v == "." || v == "0" }

// this is the actual sudoku
type grid map[index]value

type Sudoku struct {
	// the actual sudoku
	Grid grid
	// map of all fields and their units (row, column & block)
	Units map[index][3][]index
	// the "neighbours" of a field
	Peers map[index][]index
	// an ordered list of fields in the sudoku
	fields []index
}

const (
	digits = value("123456789")
)

// New initialises a new sudoku
func New(fields string) *Sudoku {
	s := Sudoku{
		fields: cross(index("ABCDEFGHI"), index("123456789")),
	}
	s.populateHelpers()
	s.Grid = s.parse(fields)
	return &s
}

// String pretty-prints the sudoku
func (s *Sudoku) String() string {
	// get the maximum width for our values.
	// usually, this should be 1, but if we could not find
	// a solution, it will show all possibilities
	width := 0
	for _, field := range s.fields {
		if width < len(s.Grid[field]) {
			width = len(s.Grid[field])
		}
	}
	line := strings.Repeat("-", ((width+1)*3)+1)
	line = fmt.Sprintf("%v+%v+%v", line, line, line)
	var print string
	for idx, field := range s.fields {
		switch {
		case idx == 0:
			print += "\n "
		case (idx % 27) == 0:
			print += "\n" + line + "\n "
		case (idx % 9) == 0:
			print += "\n "
		case (idx % 3) == 0:
			print += "| "
		}
		value := string(s.Grid[field])
		// print points instead of zeroes for readability
		if value == "0" {
			value = "."
		}
		// center the value if needed
		if len(value) < width {
			value = fmt.Sprintf("%[1]*s", -width, fmt.Sprintf("%[1]*s", (width+len(value))/2, value))
		}
		print += value + " "

	}
	return print
}

// solve a sudoku through constraint propagation. it is possible that
// this does not lead to a completely solved sudoku.
func (s *Sudoku) solve() error {
	g := make(grid)
	for _, field := range s.fields {
		g[field] = value(digits)
	}

	var err error
	for idx, value := range s.Grid {
		// if the value is zero / unknown, we don't want to assign it
		if !digits.contains(value) {
			continue
		}
		g, err = s.assign(idx, value, g)
		if err != nil {
			return err
			//return nil, err
		}
	}
	s.Grid = g
	return nil
	//return g, nil
}

// Solve a sudoku by search and constraint propagation
func (s *Sudoku) Solve() error {
	//var g grid
	//defer func() {
	//s.Grid = g
	//}()
	////g, err := s.solve()
	err := s.solve()
	if err != nil {
		return err
	}
	if s.Grid.solved() {
		return nil
	}
	//g, err := s.search(g)
	s.Grid, err = s.search(s.Grid)
	return err
}

func (s Sudoku) copy() *Sudoku {
	s.Grid = s.Grid.copy()
	return &s
}

func (s *Sudoku) search(g grid) (grid, error) {
	//Chose the unfilled field with the fewest possibilities
	var minField index
	minPoss := 10 // will never have more than 9 possibilities
	for _, field := range s.fields {
		numPoss := len(g[field])
		// we already know what's in there
		if numPoss == 1 {
			continue
		}
		if numPoss < minPoss {
			minField = field
			minPoss = numPoss
		}
	}

	for _, val := range g[minField] {
		newG, err := s.assign(minField, value(val), g.copy())
		if err != nil {
			continue
		}
		if newG.solved() {
			return newG, nil
		}
		newG, err = s.search(newG)
		if err != nil {
			continue
		}
		return newG, nil
	}

	return nil, fmt.Errorf("all possibilties lead nowhere")
}

// copy a grid
func (g grid) copy() grid {
	newGrid := make(grid)
	for k, v := range g {
		newGrid[k] = v
	}
	return newGrid
}

// returns true if the grid is solved
func (g grid) solved() bool {
	for _, field := range g {
		if len(field) > 1 {
			return false
		}
	}
	return true
}

// assign a value to an index and fix the constraints. returns the
// updated grid
func (s *Sudoku) assign(idx index, val value, g grid) (grid, error) {
	var err error
	toRemove := g[idx].remove(val)
	for _, rm := range toRemove {
		// remove the value from the grid
		g, err = s.eliminate(g, idx, value(rm))
		if err != nil {
			return nil, err
		}
	}
	return g, nil
}

// eliminate a possibility from the grid. returns the updated grid
func (s *Sudoku) eliminate(g grid, idx index, val value) (grid, error) {
	var err error
	// check if we already removed the value
	if !g[idx].contains(val) {
		return g, nil
	}

	// remove the value from that field
	g[idx] = g[idx].remove(val)
	// check and update constraints
	switch len(g[idx]) {
	case 0:
		return nil, fmt.Errorf("Removed last value from field %v", idx)
	case 1:
		// if a field is reduced to one value, eliminate that value from the peers
		for _, peer := range s.Peers[idx] {
			g, err = s.eliminate(g, peer, g[idx])
			if err != nil {
				return nil, err
			}
		}
	}
	// If a unit only has one place left for a value, put it there
	for _, unit := range s.Units[idx] {
		var place index
		var possibilities int
		for _, field := range unit {
			if g[field].contains(val) {
				possibilities++
				if possibilities > 1 {
					break
				}
				place = field
			}
		}
		if possibilities == 0 {
			return nil, fmt.Errorf("no place for value %v is left", val)
		}
		if possibilities > 1 {
			continue
		}
		g, err = s.assign(place, val, g)
		if err != nil {
			return nil, err
		}
	}
	return g, nil
}

// parse the sudoku from a string into a grid. the string
// should have either 0s or '.' for empty fields, everything
// else gets ignored
func (s *Sudoku) parse(fields string) grid {
	g := make(map[index]value)
	for i, field := range s.fields {
		val := value(fields[i])
		if !digits.contains(val) && !val.isZero() {
			continue
		}
		g[field] = val
	}
	return g
}

// populateHelpers populates the Peers and Units field
// of the sudoku
func (s *Sudoku) populateHelpers() {
	var (
		unitlist [][]index
		units    [3][]index
		i        int
	)
	const (
		columns = "123456789"
		rows    = "ABCDEFGHI"
	)

	// build the unitlist
	for _, br := range []index{"ABC", "DEF", "GHI"} {
		for _, bc := range []index{"123", "456", "789"} {
			unitlist = append(
				unitlist,
				cross(rows, index(columns[i])),
				cross(index(rows[i]), columns),
				cross(br, bc),
			)
			i++
		}
	}

	// populate units and peers from unitlist
	s.Units = make(map[index][3][]index)
	s.Peers = make(map[index][]index)
	for _, field := range s.fields {
		unitType := 0
		for _, unit := range unitlist {
			if contains(unit, field) {
				units[unitType] = unit
				s.Peers[field] = append(s.Peers[field], remove(unit, field)...)
				unitType++
			}
		}
		s.Units[field] = units
		s.Peers[field] = deduplicate(s.Peers[field])
	}
}
