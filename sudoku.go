package sudoku

import (
	"fmt"
	"strings"
)

type index string // something like "A1"

type value string // all possible values, as a single string

type grid map[index]value

type Sudoku struct {
	// the actual sudoku
	Grid grid
	// map of all fields and their units
	// 0 -> rows
	// 1 -> columns
	// 2 -> blocks
	Units map[index][3][]index
	// the "neighbors" of a field
	Peers map[index][]index
	//Peers []index
	// all the digits
	digits string
	// all the rows / columns
	rows    index
	columns index
	squares []index
}

// New creates a new sudoku, parsing the given field into it
func New() *Sudoku {
	s := Sudoku{
		digits:  "123456789",
		rows:    "ABCDEFGHI",
		columns: "123456789",
		squares: cross(index("ABCDEFGHI"), index("123456789")),
	}
	s.populateHelpers()
	return &s
}

func (s *Sudoku) String() string {
	width := 0
	for _, square := range s.squares {
		if width < len(s.Grid[square]) {
			width = len(s.Grid[square])
		}
	}
	line := strings.Repeat("-", ((width+1)*3)+1)
	line = fmt.Sprintf("%v+%v+%v", line, line, line)
	var print string
	for idx, square := range s.squares {
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
		value := string(s.Grid[square])
		// center the value if needed
		if len(value) < width {
			value = fmt.Sprintf("%[1]*s", -width, fmt.Sprintf("%[1]*s", (width+len(value))/2, value))
		}
		print += value + " "

	}
	return print
}

func (s *Sudoku) solve(fields string) (grid, bool) {
	g := make(grid)
	for _, square := range s.squares {
		g[square] = value(s.digits)
	}

	var permissible bool
	for idx, value := range s.parse(fields) {
		// if the value is zero / unknown, we don't want to assign it
		if !strings.Contains(s.digits, string(value)) {
			continue
		}
		g, permissible = s.assign(idx, value, g)
		if !permissible {
			return nil, false
		}
	}
	return g, true
}

func (s *Sudoku) Solve(fields string) bool {
	g, permissible := s.search(s.solve(fields))
	if permissible {
		s.Grid = g
	}
	return permissible
}

func (s *Sudoku) search(g grid, permissible bool) (grid, bool) {
	if !permissible {
		return nil, permissible
	}
	solved := true
	for _, square := range s.squares {
		if len(g[square]) > 1 {
			solved = false
			break
		}
	}
	if solved {
		return g, true
	}

	//Chose the unfilled square s with the fewest possibilities
	var minField index
	minPoss := 10 // will never have more than 9 possibilities
	for _, square := range s.squares {
		numPoss := len(g[square])
		// we already know what's in there
		if numPoss == 1 {
			continue
		}
		if numPoss < minPoss {
			minField = square
			minPoss = numPoss
		}
	}

	for _, val := range g[minField] {
		newG, permissible := s.search(s.assign(minField, value(val), g.copy()))
		if permissible {
			return newG, true
		}
	}

	return nil, false
}

// copy a grid
func (g grid) copy() grid {
	newGrid := make(grid)
	for k, v := range g {
		newGrid[k] = v
	}
	return newGrid
}

// assign a value to an index and fix the constraints. returns the
// updated grid
func (s *Sudoku) assign(idx index, val value, g grid) (grid, bool) {
	toRemove := strings.ReplaceAll(string(g[idx]), string(val), "")
	var permissible bool
	for _, rm := range toRemove {
		// remove the value from the grid
		g, permissible = s.eliminate(idx, value(rm), g)
		if !permissible {
			return nil, false
		}
	}
	return g, true
}

// eliminate a possibility from the grid. returns the updated grid
func (s *Sudoku) eliminate(idx index, val value, g grid) (grid, bool) {
	var permissible bool
	// check if we already removed the value
	if !strings.Contains(string(g[idx]), string(val)) {
		return g, true
	}

	// remove the value from that field
	g[idx] = value(strings.ReplaceAll(string(g[idx]), string(val), ""))
	// check and update constraints
	switch len(g[idx]) {
	case 0:
		return nil, false // we removed the last value
	case 1:
		// if a field is reduced to one value, eliminate that value from the peers
		lastVal := g[idx]
		for _, peer := range s.Peers[idx] {
			g, permissible = s.eliminate(peer, lastVal, g)
			if !permissible {
				return nil, false
			}
		}
	}
	// If a unit is reduced to only one place for a value, then put it there.
	for _, unit := range s.Units[idx] {
		var valPlaces []index
		// get all the places in that unit where the value COULD be
		for _, field := range unit {
			if strings.Contains(string(g[field]), string(val)) {
				valPlaces = append(valPlaces, field)
			}
		}
		switch len(valPlaces) {
		case 0:
			return nil, false // no place for that value is left
		case 1:
			g, permissible = s.assign(valPlaces[0], val, g)
			if !permissible {
				return nil, false
			}
		}
	}
	return g, true
}

func (s *Sudoku) parse(fields string) map[index]value {
	grid := make(map[index]value)
	for i, square := range s.squares {
		if !strings.Contains(s.digits, string(fields[i])) && fields[i] != byte('.') && fields[i] != byte('0') {
			continue
		}
		grid[square] = value(fields[i])
	}
	return grid
}

// populateHelpers populates the Peers and Units field
// of the sudoku
func (s *Sudoku) populateHelpers() {
	var unitlist [][]index
	for _, c := range s.columns {
		unitlist = append(unitlist, cross(s.rows, index(c)))
	}
	for _, r := range s.rows {
		unitlist = append(unitlist, cross(index(r), s.columns))
	}
	for _, br := range []index{"ABC", "DEF", "GHI"} {
		for _, bc := range []index{"123", "456", "789"} {
			unitlist = append(unitlist, cross(br, bc))
		}
	}
	s.Units = make(map[index][3][]index)
	s.Peers = make(map[index][]index)
	var units [3][]index
	for _, square := range s.squares {
		unitType := 0
		for _, unit := range unitlist {
			if contains(unit, square) {
				units[unitType] = unit
				s.Peers[square] = append(s.Peers[square], remove(unit, square)...)
				unitType++
			}
		}
		s.Units[square] = units
	}

	// remove duplicates from the peers
	for i := range s.Peers {
		s.Peers[i] = deduplicate(s.Peers[i])
	}
}

func cross(a, b index) []index {
	i := make([]index, len(a)*len(b))
	idx := 0
	for _, sa := range a {
		for _, sb := range b {
			i[idx] = index(sa) + index(sb)
			idx++
		}
	}
	return i
}
