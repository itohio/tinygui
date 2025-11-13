package widget

import (
	"image/color"

	ui "github.com/itohio/tinygui"
	"tinygo.org/x/tinyfont"
)

// MultilineOrder controls how the most recent line is positioned when drawing.
type MultilineOrder uint8

const (
	// MultilineNewestOnBottom renders older lines first so the newest ones appear last.
	MultilineNewestOnBottom MultilineOrder = iota
	// MultilineNewestOnTop renders newest lines first so recent entries appear at the top.
	MultilineNewestOnTop
)

// MultilineOption customises the behaviour of a multiline widget.
type MultilineOption func(*MultilineBase)

// WithMultilineFont sets the font used for drawing text.
func WithMultilineFont(font tinyfont.Fonter) MultilineOption {
	return func(base *MultilineBase) {
		base.font = font
	}
}

// WithMultilineColor sets the text colour.
func WithMultilineColor(col color.RGBA) MultilineOption {
	return func(base *MultilineBase) {
		base.color = col
	}
}

// WithMultilineOrder defines whether the newest line is drawn at the bottom or top.
func WithMultilineOrder(order MultilineOrder) MultilineOption {
	return func(base *MultilineBase) {
		base.order = order
	}
}

// MultilineBase stores shared properties for multiline widgets.
type MultilineBase struct {
	ui.WidgetBase
	font     tinyfont.Fonter
	color    color.RGBA
	maxLines int
	order    MultilineOrder
	lines    []string
}

// NewMultilineBase constructs the shared base. lineHeight represents the height of a single row.
func NewMultilineBase(width, lineHeight uint16, maxLines int, opts ...MultilineOption) *MultilineBase {
	base := &MultilineBase{
		WidgetBase: ui.NewWidgetBase(width, lineHeight*uint16(maxLines)),
		font:       &tinyfont.TomThumb,
		color:      color.RGBA{255, 255, 255, 255},
		maxLines:   maxLines,
		order:      MultilineNewestOnBottom,
		lines:      nil,
	}
	for _, opt := range opts {
		opt(base)
	}
	return base
}

// SetLines replaces the base content with the provided lines.
func (m *MultilineBase) SetLines(lines []string) {
	m.lines = append(m.lines[:0], lines...)
}

// Lines returns the stored lines.
func (m *MultilineBase) Lines() []string {
	return m.lines
}

// MaxLines returns the maximum number of lines that can be rendered simultaneously.
func (m *MultilineBase) MaxLines() int {
	return m.maxLines
}

// Order exposes the configured drawing order.
func (m *MultilineBase) Order() MultilineOrder {
	return m.order
}

// Draw renders the current lines using the default starting position.
func (m *MultilineBase) Draw(ctx ui.Context) {
	m.DrawAt(ctx, m.defaultStart())
}

// DrawAt renders the view starting at the provided offset.
func (m *MultilineBase) DrawAt(ctx ui.Context, start int) {
	total := len(m.lines)
	if total == 0 {
		return
	}
	start = clampInt(start, 0, m.maxStart())
	visible := m.maxLines
	if visible > total {
		visible = total
	}

	x, y := ctx.DisplayPos()
	lineHeight := int16(m.Height) / int16(m.maxLines)
	if lineHeight <= 0 {
		lineHeight = 1
	}

	switch m.order {
	case MultilineNewestOnTop:
		for i := 0; i < visible; i++ {
			idx := total - 1 - (start + i)
			if idx < 0 {
				break
			}
			tinyfont.WriteLine(ctx.D(), m.font, x, y+lineHeight, m.lines[idx], m.color)
			y += lineHeight
		}
	default: // MultilineNewestOnBottom
		for i := 0; i < visible; i++ {
			idx := start + i
			if idx >= total {
				break
			}
			tinyfont.WriteLine(ctx.D(), m.font, x, y+lineHeight, m.lines[idx], m.color)
			y += lineHeight
		}
	}
}

func (m *MultilineBase) defaultStart() int {
	if m.order == MultilineNewestOnBottom {
		return m.maxStart()
	}
	return 0
}

func (m *MultilineBase) maxStart() int {
	surplus := len(m.lines) - m.maxLines
	if surplus < 0 {
		surplus = 0
	}
	return surplus
}

func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// MultilineLabel renders the current lines using the multiline base.
type MultilineLabel struct {
	*MultilineBase
}

// NewMultilineLabel constructs a multiline label that can be updated by SetLines.
func NewMultilineLabel(width, lineHeight uint16, maxLines int, opts ...MultilineOption) *MultilineLabel {
	return &MultilineLabel{MultilineBase: NewMultilineBase(width, lineHeight, maxLines, opts...)}
}

// Log stores an append-only history capped at capacity.
type Log struct {
	*MultilineBase
	capacity int
}

// SetCapacity adjusts the number of entries retained in history (minimum maxLines).
func (l *Log) SetCapacity(cap int) {
	if cap < l.maxLines {
		cap = l.maxLines
	}
	l.capacity = cap
	if len(l.lines) > l.capacity {
		excess := len(l.lines) - l.capacity
		l.lines = append([]string(nil), l.lines[excess:]...)
	}
}

// NewLog constructs a log view using the legacy constructor signature.
func NewLog(width, lineHeight uint16, maxLines int, font tinyfont.Fonter, col color.RGBA) *Log {
	return NewLogWithOptions(width, lineHeight, maxLines, WithMultilineFont(font), WithMultilineColor(col))
}

// NewLogWithOptions constructs a log view using shared multiline options.
func NewLogWithOptions(width, lineHeight uint16, maxLines int, opts ...MultilineOption) *Log {
	base := NewMultilineBase(width, lineHeight, maxLines, opts...)
	log := &Log{MultilineBase: base, capacity: maxLines}
	return log
}

// Append adds a new log entry, keeping only the configured capacity.
func (l *Log) Append(line string) {
	l.lines = append(l.lines, line)
	if l.capacity > 0 && len(l.lines) > l.capacity {
		excess := len(l.lines) - l.capacity
		if excess > 0 {
			l.lines = append([]string(nil), l.lines[excess:]...)
		}
	}
}

// InteractiveMultiline allows scrolling through multiline content.
type InteractiveMultiline struct {
	*MultilineBase
	viewStart int
}

// NewInteractiveMultiline constructs a scrollable multiline widget.
func NewInteractiveMultiline(width, lineHeight uint16, maxLines int, opts ...MultilineOption) *InteractiveMultiline {
	base := NewMultilineBase(width, lineHeight, maxLines, opts...)
	return &InteractiveMultiline{MultilineBase: base, viewStart: base.defaultStart()}
}

// Draw renders using the current view start offset.
func (m *InteractiveMultiline) Draw(ctx ui.Context) {
	m.MultilineBase.DrawAt(ctx, m.viewStart)
}

// Interact scrolls through the history.
func (m *InteractiveMultiline) Interact(cmd ui.UserCommand) bool {
	if len(m.lines) == 0 {
		return false
	}
	switch cmd {
	case ui.UP, ui.NEXT:
		return m.scrollOlder(1)
	case ui.DOWN, ui.PREV:
		return m.scrollNewer(1)
	case ui.LONG_UP:
		return m.scrollOlder(m.MaxLines())
	case ui.LONG_DOWN:
		return m.scrollNewer(m.MaxLines())
	default:
		return false
	}
}

// SetLines updates the content and re-clamps the view region.
func (m *InteractiveMultiline) SetLines(lines []string) {
	m.MultilineBase.SetLines(lines)
	m.viewStart = clampInt(m.viewStart, 0, m.maxStart())
	if m.order == MultilineNewestOnBottom && m.viewStart == 0 {
		m.viewStart = m.defaultStart()
	}
}

func (m *InteractiveMultiline) scrollOlder(step int) bool {
	maxStart := m.maxStart()
	if m.order == MultilineNewestOnBottom {
		newStart := clampInt(m.viewStart-step, 0, maxStart)
		if newStart != m.viewStart {
			m.viewStart = newStart
			return true
		}
		return false
	}
	newStart := clampInt(m.viewStart+step, 0, maxStart)
	if newStart != m.viewStart {
		m.viewStart = newStart
		return true
	}
	return false
}

func (m *InteractiveMultiline) scrollNewer(step int) bool {
	maxStart := m.maxStart()
	if m.order == MultilineNewestOnBottom {
		newStart := clampInt(m.viewStart+step, 0, maxStart)
		if newStart != m.viewStart {
			m.viewStart = newStart
			return true
		}
		return false
	}
	newStart := clampInt(m.viewStart-step, 0, maxStart)
	if newStart != m.viewStart {
		m.viewStart = newStart
		return true
	}
	return false
}

// InteractiveLog augments Log with scrolling controls.
type InteractiveLog struct {
	*Log
	viewStart int
}

// NewInteractiveLog constructs a scrollable log widget.
func NewInteractiveLog(width, lineHeight uint16, maxLines int, opts ...MultilineOption) *InteractiveLog {
	log := NewLogWithOptions(width, lineHeight, maxLines, opts...)
	return &InteractiveLog{Log: log, viewStart: log.defaultStart()}
}

// Draw renders the log at the current view offset.
func (l *InteractiveLog) Draw(ctx ui.Context) {
	l.MultilineBase.DrawAt(ctx, l.viewStart)
}

// Interact scrolls through log history.
func (l *InteractiveLog) Interact(cmd ui.UserCommand) bool {
	switch cmd {
	case ui.UP, ui.NEXT:
		return l.scrollOlder(1)
	case ui.DOWN, ui.PREV:
		return l.scrollNewer(1)
	case ui.LONG_UP:
		return l.scrollOlder(l.MaxLines())
	case ui.LONG_DOWN:
		return l.scrollNewer(l.MaxLines())
	default:
		return false
	}
}

// Append adds a new log entry and repositions the view if necessary.
func (l *InteractiveLog) Append(line string) {
	l.Log.Append(line)
	l.viewStart = clampInt(l.viewStart, 0, l.maxStart())
	if l.order == MultilineNewestOnBottom {
		l.viewStart = l.defaultStart()
	}
}

func (l *InteractiveLog) scrollOlder(step int) bool {
	maxStart := l.maxStart()
	if l.order == MultilineNewestOnBottom {
		newStart := clampInt(l.viewStart-step, 0, maxStart)
		if newStart != l.viewStart {
			l.viewStart = newStart
			return true
		}
		return false
	}
	newStart := clampInt(l.viewStart+step, 0, maxStart)
	if newStart != l.viewStart {
		l.viewStart = newStart
		return true
	}
	return false
}

func (l *InteractiveLog) scrollNewer(step int) bool {
	maxStart := l.maxStart()
	if l.order == MultilineNewestOnBottom {
		newStart := clampInt(l.viewStart+step, 0, maxStart)
		if newStart != l.viewStart {
			l.viewStart = newStart
			return true
		}
		return false
	}
	newStart := clampInt(l.viewStart-step, 0, maxStart)
	if newStart != l.viewStart {
		l.viewStart = newStart
		return true
	}
	return false
}
