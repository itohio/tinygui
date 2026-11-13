package layout

// Arrangers place every child at once (measure, then arrange), which the
// cursor strategies above can't: they see one child at a time, so they can't
// centre a row, split space by weights or give a table equal columns.
// Children keep their own size (TinyGUI widgets are fixed-size); an arranger
// decides the cell each child gets and where the child sits inside it.
//
// Containers use them through container.WithArranger.

// Rect is a child's place relative to the top-left of the arranged box.
type Rect struct {
	X, Y int16
	W, H uint16
}

// Sizes lists the children's sizes; containers implement it.
type Sizes interface {
	Len() int
	SizeAt(i int) (w, h uint16)
}

// Arranger fills out[i] with the place of child i inside a w×h box.
// len(out) is items.Len(). Arrange must not allocate.
type Arranger interface {
	Arrange(w, h uint16, items Sizes, out []Rect)
}

// Measurer reports the size an arranger needs for its children, for
// containers created with a zero width or height.
type Measurer interface {
	Measure(items Sizes) (w, h uint16)
}

// Align places children along an axis.
type Align uint8

const (
	Start Align = iota
	Center
	End
	// SpaceBetween spreads the free space evenly between children (main
	// axis only; elsewhere it acts as Start).
	SpaceBetween
)

// offset of an item of size n in a span of size total.
func (a Align) offset(total, n int16) int16 {
	switch a {
	case Center:
		return (total - n) / 2
	case End:
		return total - n
	}
	return 0
}

// Line arranges children in one row (or one column when Vertical).
//
// Without weights the children are packed with Gap between them and the
// group is placed by Main (Start, Center, End, SpaceBetween). With weights,
// a child with weight 0 gets its own size and the others share the rest of
// the main axis in proportion to their weights; each child sits in its cell
// by Main. Across the line every child is placed by Cross.
type Line struct {
	Vertical    bool
	Gap         int16
	Main, Cross Align
	Weights     []uint8 // optional, one per child; missing entries are 0
}

// Row lays children left to right.
func Row(gap int16, main, cross Align, weights ...uint8) *Line {
	return &Line{Gap: gap, Main: main, Cross: cross, Weights: weights}
}

// Column lays children top to bottom.
func Column(gap int16, main, cross Align, weights ...uint8) *Line {
	return &Line{Vertical: true, Gap: gap, Main: main, Cross: cross, Weights: weights}
}

func (l *Line) axes(w, h uint16) (main, cross int16) {
	if l.Vertical {
		return int16(h), int16(w)
	}
	return int16(w), int16(h)
}

func (l *Line) size(items Sizes, i int) (main, cross int16) {
	w, h := items.SizeAt(i)
	return l.axes(w, h)
}

func (l *Line) weight(i int) int16 {
	if i < len(l.Weights) {
		return int16(l.Weights[i])
	}
	return 0
}

func (l *Line) Arrange(w, h uint16, items Sizes, out []Rect) {
	n := items.Len()
	if n == 0 {
		return
	}
	boxMain, boxCross := l.axes(w, h)
	var used, sumW int16
	for i := range n {
		m, _ := l.size(items, i)
		if wt := l.weight(i); wt > 0 {
			sumW += wt
		} else {
			used += m
		}
	}
	gaps := l.Gap * int16(n-1)
	free := boxMain - used - gaps
	pos, gap := int16(0), l.Gap
	if sumW == 0 {
		switch l.Main {
		case SpaceBetween:
			if n > 1 && free > 0 {
				gap += free / int16(n-1)
			}
		default:
			pos = l.Main.offset(free, 0)
		}
	}
	shared, left := free, sumW
	for i := range n {
		m, c := l.size(items, i)
		cell := m
		if wt := l.weight(i); wt > 0 {
			// the last weighted child takes the rounding remainder
			cell = shared
			if left > wt {
				cell = int16(int32(shared) * int32(wt) / int32(left)) // int32: shared×weight overflows int16
			}
			shared -= cell
			left -= wt
		}
		inCell := int16(0)
		if sumW > 0 {
			inCell = l.Main.offset(cell, m)
		}
		cr := l.Cross.offset(boxCross, c)
		wi, hi := items.SizeAt(i)
		if l.Vertical {
			out[i] = Rect{X: cr, Y: pos + inCell, W: wi, H: hi}
		} else {
			out[i] = Rect{X: pos + inCell, Y: cr, W: wi, H: hi}
		}
		pos += cell + gap
	}
}

// Measure is the packed size: the sum along the line plus gaps, the largest
// across it.
func (l *Line) Measure(items Sizes) (uint16, uint16) {
	var sum, most int16
	for i := range items.Len() {
		m, c := l.size(items, i)
		sum += m
		most = max(most, c)
	}
	if n := items.Len(); n > 1 {
		sum += l.Gap * int16(n-1)
	}
	if l.Vertical {
		return uint16(most), uint16(sum)
	}
	return uint16(sum), uint16(most)
}

// Wrap fills rows left to right, starting a new row when the next child
// doesn't fit the width. Each row is placed by Main, each child within its
// row by Cross, and the block of rows vertically by Block.
type Wrap struct {
	Gap                int16
	Main, Cross, Block Align
}

func (f *Wrap) Arrange(w, h uint16, items Sizes, out []Rect) {
	n := items.Len()
	// pass 1: total height, to place the block
	var total int16
	for i := 0; i < n; {
		j, _, rh := f.row(w, items, i)
		if total > 0 {
			total += f.Gap
		}
		total += rh
		i = j
	}
	y := f.Block.offset(int16(h), total)
	for i := 0; i < n; {
		j, rw, rh := f.row(w, items, i)
		free := int16(w) - rw
		x, gap := f.Main.offset(free, 0), f.Gap
		if f.Main == SpaceBetween && j-i > 1 && free > 0 {
			gap += free / int16(j-i-1)
		}
		for k := i; k < j; k++ {
			cw, ch := items.SizeAt(k)
			out[k] = Rect{X: x, Y: y + f.Cross.offset(rh, int16(ch)), W: cw, H: ch}
			x += int16(cw) + gap
		}
		y += rh + f.Gap
		i = j
	}
}

// row returns the end (exclusive) of the row starting at i, its packed
// width and its height. A row holds at least one child.
func (f *Wrap) row(w uint16, items Sizes, i int) (end int, rw, rh int16) {
	n := items.Len()
	for end = i; end < n; end++ {
		cw, ch := items.SizeAt(end)
		next := rw + int16(cw)
		if end > i {
			next += f.Gap
		}
		if end > i && next > int16(w) {
			break
		}
		rw = next
		rh = max(rh, int16(ch))
	}
	return end, rw, rh
}

// Table puts children in Cols equal-width columns, row by row; a row is as
// tall as its tallest child, and each child sits in its cell by H and V.
type Table struct {
	Cols       int
	GapX, GapY int16
	H, V       Align
}

func (t *Table) Arrange(w, h uint16, items Sizes, out []Rect) {
	n := items.Len()
	cols := max(t.Cols, 1)
	cellW := (int16(w) - t.GapX*int16(cols-1)) / int16(cols)
	y := int16(0)
	for r := 0; r < n; r += cols {
		end := min(r+cols, n)
		var rh int16
		for k := r; k < end; k++ {
			_, ch := items.SizeAt(k)
			rh = max(rh, int16(ch))
		}
		for k := r; k < end; k++ {
			cw, ch := items.SizeAt(k)
			x := int16(k-r) * (cellW + t.GapX)
			out[k] = Rect{X: x + t.H.offset(cellW, int16(cw)), Y: y + t.V.offset(rh, int16(ch)), W: cw, H: ch}
		}
		y += rh + t.GapY
	}
}

// Measure is the size Arrange fills: every column as wide as the widest
// child, rows as tall as their tallest child, with the gaps.
func (t *Table) Measure(items Sizes) (uint16, uint16) {
	n := items.Len()
	if n == 0 {
		return 0, 0
	}
	cols := max(t.Cols, 1)
	var cellW, h int16
	for r := 0; r < n; r += cols {
		var rh int16
		for k := r; k < min(r+cols, n); k++ {
			cw, ch := items.SizeAt(k)
			cellW = max(cellW, int16(cw))
			rh = max(rh, int16(ch))
		}
		h += rh
		if r > 0 {
			h += t.GapY
		}
	}
	return uint16(int16(cols)*cellW + t.GapX*int16(cols-1)), uint16(h)
}

// Anchor stacks every child in the whole box, placed by H and V and moved by
// DX, DY: centred dialogs, corner badges, layers drawn in child order.
type Anchor struct {
	H, V   Align
	DX, DY int16
}

func (a *Anchor) Arrange(w, h uint16, items Sizes, out []Rect) {
	for i := range items.Len() {
		cw, ch := items.SizeAt(i)
		out[i] = Rect{X: a.H.offset(int16(w), int16(cw)) + a.DX, Y: a.V.offset(int16(h), int16(ch)) + a.DY, W: cw, H: ch}
	}
}

// Measure is the largest child.
func (a *Anchor) Measure(items Sizes) (uint16, uint16) {
	var w, h uint16
	for i := range items.Len() {
		cw, ch := items.SizeAt(i)
		w, h = max(w, cw), max(h, ch)
	}
	return w, h
}
