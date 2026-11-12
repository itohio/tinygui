package app

// liquid is a falling-liquid cellular automaton on a w×h grid: water, and
// the fuel the fire burns. Cells keep a sideways direction; a blocked cell
// slides along the surface, so pools level out quickly and follow gravity
// in any of 8 directions.
const (
	liquidBit   = 1
	liquidMoved = 2
	liquidDir   = 4 // the cell's sideways direction: set = +perpendicular
	// cells a blocked cell can flow sideways in one leg of a slide
	liquidDisperse = 16
	// cells a cell can fall in one step
	liquidFall = 3
	// cells of travel a sliding surface cell may cover in one step
	liquidSlide = 32
)

var gravityNames = []string{"down", "dn-left", "left", "up-left", "up", "up-right", "right", "dn-right"}

// gravity vector per direction (screen coordinates, y down)
var gravityVec = [8][2]int{{0, 1}, {-1, 1}, {-1, 0}, {-1, -1}, {0, -1}, {1, -1}, {1, 0}, {1, 1}}

type liquid struct {
	w, h  int
	state []uint8
	dir   int  // gravity, index into gravityVec
	pour  bool // a two-cell stream enters opposite to gravity, at the centre
	scan  uint8
}

func newLiquid(w, h int) liquid { return liquid{w: w, h: h, state: make([]uint8, w*h)} }

// count is the number of liquid cells.
func (l *liquid) count() int {
	n := 0
	for _, s := range l.state {
		if s&liquidBit != 0 {
			n++
		}
	}
	return n
}

// turn rotates gravity by steps of 45 degrees.
func (l *liquid) turn(by int) { l.dir = (l.dir + by) & 7 }

func (l *liquid) at(x, y int) bool {
	return x >= 0 && y >= 0 && x < l.w && y < l.h && l.state[y*l.w+x]&liquidBit == 0
}

// move puts the cell at x, y to nx, ny with bits (its sideways direction),
// marked moved for this step.
func (l *liquid) move(x, y, nx, ny int, bits uint8) {
	l.state[y*l.w+x] = 0
	l.state[ny*l.w+nx] = liquidBit | liquidMoved | bits
}

// newCell is a liquid cell with a random sideways direction.
func newCell() uint8 { return liquidBit | uint8(rand32()&1)*liquidDir }

// step moves every cell once, as a liquid: along gravity if it can;
// else diagonally down, its own sideways direction first; else up to
// liquidDisperse cells sideways in its direction (stopping where it can fall
// next), turning round when that side is blocked. Spreading several cells a
// step levels the surface quickly (one cell a step piled water up like sand),
// and the direction a cell keeps stops it jittering back and forth.
func (l *liquid) step() {
	g := gravityVec[l.dir]
	gx, gy := g[0], g[1]
	px, py := -gy, gx // perpendicular
	if l.pour {
		// a two-cell stream entering opposite to gravity, at the centre
		cx, cy := l.w/2-gx*(l.w/2-1), l.h/2-gy*(l.h/2-1)
		for _, o := range [2]int{0, 1} {
			x, y := cx+px*o, cy+py*o
			if l.at(x, y) {
				l.state[y*l.w+x] = newCell()
			}
		}
	}
	// Alternate the scan direction on both axes every step (the moved bit
	// keeps any cell from moving twice either way). From the side gravity
	// points to, falling streams move as a block; from the other side a
	// whole face of a heap slides at once (from one side only, each cell fell
	// into the gap its neighbour left and only one per row moved: water piled
	// up like sand). Along the other axis a fixed order favoured one side.
	l.scan++ // cycles through the four combinations
	y0, y1, dy := 0, l.h, 1
	if l.scan&1 != 0 {
		y0, y1, dy = l.h-1, -1, -1
	}
	x0, x1, dx := 0, l.w, 1
	if l.scan&2 != 0 {
		x0, x1, dx = l.w-1, -1, -1
	}
	for y := y0; y != y1; y += dy {
		for x := x0; x != x1; x += dx {
			s := l.state[y*l.w+x]
			if s&liquidBit == 0 || s&liquidMoved != 0 {
				continue
			}
			dbit := s & liquidDir
			if k := l.fall(x, y, gx, gy); k > 0 {
				l.move(x, y, x+gx*k, y+gy*k, dbit)
				continue
			}
			d := -1
			if dbit != 0 {
				d = 1
			}
			switch {
			case l.at(x+gx+px*d, y+gy+py*d):
				l.move(x, y, x+gx+px*d, y+gy+py*d, dbit)
			case l.at(x+gx-px*d, y+gy-py*d):
				l.move(x, y, x+gx-px*d, y+gy-py*d, dbit^liquidDir)
			default:
				if nx, ny := l.slide(x, y, px*d, py*d, gx, gy); nx != x || ny != y {
					l.move(x, y, nx, ny, dbit)
				} else if nx, ny := l.slide(x, y, -px*d, -py*d, gx, gy); nx != x || ny != y {
					l.move(x, y, nx, ny, dbit^liquidDir)
				}
			}
		}
	}
	for i := range l.state {
		l.state[i] &^= liquidMoved
	}
}

// fall is how many free cells (up to liquidFall) a cell can drop along
// gravity this step: falling water speeds up instead of creeping a cell a step.
func (l *liquid) fall(x, y, gx, gy int) int {
	k := 0
	for i := 1; i <= liquidFall; i++ {
		if !l.at(x+gx*i, y+gy*i) {
			break
		}
		k = i
	}
	return k
}

// slide is where a blocked cell ends up flowing sideways (sx, sy) along
// the surface: sideways to the first place it can drop, down, sideways
// again, … within liquidSlide cells of travel. One cell a step left a heap's
// slope to level row by row, slowly (8 s for a tilted basin); sliding the
// whole slope in a step levels it in time proportional to its length.
func (l *liquid) slide(x, y, sx, sy, gx, gy int) (int, int) {
	budget := liquidSlide
	for budget > 0 {
		k := l.reach(x, y, sx, sy, gx, gy)
		if k == 0 {
			break
		}
		x, y = x+sx*k, y+sy*k
		budget -= k
		f := l.fall(x, y, gx, gy)
		if f == 0 {
			break // flat: it spread as far as it may this step
		}
		x, y = x+gx*f, y+gy*f
		budget -= f
	}
	return x, y
}

// reach is how many free cells (up to liquidDisperse) a cell at x, y can flow
// in direction sx, sy, stopping at the first one it could fall from.
func (l *liquid) reach(x, y, sx, sy, gx, gy int) int {
	k := 0
	for i := 1; i <= liquidDisperse; i++ {
		tx, ty := x+sx*i, y+sy*i
		if !l.at(tx, ty) {
			break
		}
		k = i
		if l.at(tx+gx, ty+gy) {
			break
		}
	}
	return k
}
