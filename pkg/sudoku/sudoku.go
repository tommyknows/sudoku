// package sudoku solves sudoku puzzles through constraint proppagation and
// search. it is based on the article here: https://norvig.com/sudoku.html
package sudoku

import (
	"fmt"
	"strings"
)

const (
	digits = value("123456789")
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

// returns true if the grid is solved
func (g grid) solved() bool {
	for _, field := range g {
		if len(field) > 1 {
			return false
		}
	}
	return true
}

type Sudoku struct {
	// the actual sudoku
	grid
	// map of all fields and their units (row, column & block)
	units map[index][3][]index
	// the "neighbours" of a field
	peers map[index][]index
	// an ordered list of fields in the sudoku
	fields []index
}

// New initialises a new sudoku, parsing the given fields. empty fields
// should be indicated with either '0' or '.'. All other characters
// will be ignored (apart from actual digits).
func New(fields string) *Sudoku {
	s := Sudoku{
		fields: cross(index("ABCDEFGHI"), index("123456789")),
	}
	s.populateHelpers()
	s.parse(fields)
	return &s
}

// String pretty-prints the sudoku
func (s *Sudoku) String() string {
	// get the maximum width for our values.
	// usually, this should be 1, but if we could not find
	// a solution, it will show all possibilities
	width := 0
	for _, field := range s.fields {
		if width < len(s.grid[field]) {
			width = len(s.grid[field])
		}
	}
	line := strings.Repeat("-", ((width+1)*3)+1)
	line = fmt.Sprintf("%v+%v+%v", line, line, line)
	var print string
	for idx, field := range s.fields {
		switch {
		case idx == 0:
			print += " "
		case (idx % 27) == 0:
			print += "\n" + line + "\n "
		case (idx % 9) == 0:
			print += "\n "
		case (idx % 3) == 0:
			print += "| "
		}
		value := string(s.grid[field])
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

// Solve a sudoku by constraint propagation and search, if necessary
func (s *Sudoku) Solve() error {
	if err := s.constraintPropagation(); err != nil {
		return err
	}
	if s.grid.solved() {
		return nil
	}
	return s.search()
}

// search solves a sudoku through guessing
func (s *Sudoku) search() error {
	//Chose the unfilled field with the fewest possibilities
	minField := s.minimumValues()
	// try through the possible values
	for _, val := range s.grid[minField] {
		sc, err := s.try(value(val), minField)
		if err != nil {
			continue
		}
		// if we got here, that means we have a solution.
		*s = *sc
		return nil
	}
	return fmt.Errorf("all possibilties lead nowhere")
}

// constraintPropagation solves a sudoku solely through constraint propagation.
// it is possible that this does not lead to a completely solved sudoku.
func (s *Sudoku) constraintPropagation() error {
	toSolve := s.grid
	s.grid = make(grid)
	for _, field := range s.fields {
		s.grid[field] = digits
	}

	for idx, value := range toSolve {
		// if the value is zero / unknown, we don't want to assign it
		if !digits.contains(value) {
			continue
		}
		if err := s.assign(value, idx); err != nil {
			return err
		}
	}
	return nil
}

// minimumValues gets the field that has the lowest
// number of possible values
func (s *Sudoku) minimumValues() index {
	var minField index
	minPoss := 10 // will never have more than 9 possibilities
	for _, field := range s.fields {
		numPoss := len(s.grid[field])
		// we already know what's in there
		if numPoss == 1 {
			continue
		}
		if numPoss < minPoss {
			minField = field
			minPoss = numPoss
		}
	}
	return minField
}

// try to set a value at the given field, returning a copy of
// the sudoku with the tried value filled in, or an error if
// a contradiction has been detected, making this move invalid
func (s *Sudoku) try(val value, idx index) (*Sudoku, error) {
	// create a copy as we are guessing
	sc := s.copy()
	if err := sc.assign(val, idx); err != nil {
		return nil, err
	}
	if !sc.grid.solved() {
		// take another guess
		if err := sc.search(); err != nil {
			return nil, err
		}
	}
	return sc, nil
}

// assign a value to an index and fix the constraints.
func (s *Sudoku) assign(val value, idx index) error {
	for _, rm := range s.grid[idx].remove(val) {
		// remove the value from the grid
		if err := s.removeAt(value(rm), idx); err != nil {
			return err
		}
	}
	return nil
}

// removeAt removes a possibility from a field and update its peers
func (s *Sudoku) removeAt(val value, idx index) error {
	// check if we already removed the value
	if !s.grid[idx].contains(val) {
		return nil
	}

	// remove the value from that field
	s.grid[idx] = s.grid[idx].remove(val)
	switch len(s.grid[idx]) {
	case 0:
		return fmt.Errorf("removed last value from field %v", idx)
	case 1:
		// if a field is reduced to one value, eliminate that value from the peers
		if err := s.removeFromPeers(idx); err != nil {
			return err
		}
	}
	for _, unit := range s.units[idx] {
		// If a unit only has one place left for a value, put it there
		p, place := s.singlePossibility(val, unit)
		if place == "" {
			return fmt.Errorf("no place for value %v is left", val)
		}
		if !p {
			continue
		}
		if err := s.assign(val, place); err != nil {
			return err
		}
	}
	return nil
}

// removes the given value from all peers of idx
func (s *Sudoku) removeFromPeers(idx index) error {
	val := s.grid[idx]
	for _, peer := range s.peers[idx] {
		if err := s.removeAt(val, peer); err != nil {
			return err
		}
	}
	return nil
}

// singlePossibility returns true when the give value only has one possibility
// in the given units, and returns the index. if there is no possibility left,
// the index will be an empty string.
func (s *Sudoku) singlePossibility(val value, unit []index) (found bool, field index) {
	for _, f := range unit {
		if s.grid[f].contains(val) {
			// second possibility
			if found {
				return false, field
			}
			found = true
			field = f
		}
	}
	return found, field
}

func (v value) validCharacter() bool {
	return digits.contains(v) || v.isZero()
}

// grid sets the grid by parsing the sudoku from a string. the
// string should have either 0s or '.' for empty fields, everything
// else gets ignored
func (s *Sudoku) parse(fields string) {
	s.grid = make(map[index]value)
	i := 0
	for _, v := range fields {
		val := value(v)
		if !val.validCharacter() {
			continue
		}
		s.grid[s.fields[i]] = val
		i++
	}
}

// creates a copy of the Sudoku, including a deep-copy
// of the main sudoku grid. Useful when doing guesswork
// that could be wrong
func (s Sudoku) copy() *Sudoku {
	newGrid := make(grid)
	for k, v := range s.grid {
		newGrid[k] = v
	}
	s.grid = newGrid
	return &s
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
	s.units = make(map[index][3][]index)
	s.peers = make(map[index][]index)
	for _, field := range s.fields {
		unitType := 0
		for _, unit := range unitlist {
			if contains(unit, field) {
				units[unitType] = unit
				s.peers[field] = append(s.peers[field], remove(unit, field)...)
				unitType++
			}
		}
		s.units[field] = units
		s.peers[field] = deduplicate(s.peers[field])
	}
}
