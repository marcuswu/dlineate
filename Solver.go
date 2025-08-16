package dlineate

import (
	"errors"
	"io"
	"math"
	"math/big"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers/svg"

	core "github.com/marcuswu/dlineate/internal"
	"github.com/marcuswu/dlineate/internal/constraint"
	"github.com/marcuswu/dlineate/internal/solver"
	"github.com/marcuswu/dlineate/utils"
)

type Sketch struct {
	plane       *Workplane
	sketch      *core.SketchGraph
	Elements    []*Element
	constraints []*Constraint
	eToC        map[uint][]*Constraint
	passes      int
	Origin      *Element
	XAxis       *Element
	YAxis       *Element
}

func UseLogger(logger zerolog.Logger) {
	utils.Logger = logger
}

// NewSketch creates a new sketch at [0, 0] with standard axis orientation and elements with constraints for origin and X/Y axes
// It returns the new sketch
func NewSketch() *Sketch {
	UseLogger(log.Logger)
	s := new(Sketch)
	s.sketch = core.NewSketch()
	s.passes = 0
	s.Elements = make([]*Element, 0)
	s.constraints = make([]*Constraint, 0)
	s.eToC = make(map[uint][]*Constraint)
	// TODO: These need to be in a special cluster that isn't counted towards solving
	s.Origin = s.addOrigin()
	s.XAxis = s.addAxis(0, -1, 0)
	s.YAxis = s.addAxis(1, 0, 0)
	s.AddAngleConstraint(s.XAxis, s.YAxis, 90, false)
	s.AddCoincidentConstraint(s.Origin, s.XAxis)
	s.AddCoincidentConstraint(s.Origin, s.YAxis)

	// Init logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	return s
}

// SetWorkplane sets the origin and axis orientation of the sketch
func (s *Sketch) SetWorkplane(plane *Workplane) {
	s.plane = plane
}

func (s *Sketch) findConstraints(e *Element) []*Constraint {
	return s.eToC[e.id]
}

func (s *Sketch) findConstraint(ctype ConstraintType, e *Element) (*Constraint, error) {
	for _, c := range s.eToC[e.id] {
		if c.constraintType != ctype || (c.state != Resolved && c.state != Solved) {
			continue
		}
		return c, nil
	}

	return nil, errors.New("no such constraint")
}

func (s *Sketch) nextElementID() uint {
	return uint(len(s.Elements))
}

// AddPoint adds a point to the sketch at [x, y].
// It returns the point element created.
func (s *Sketch) AddPoint(x float64, y float64) *Element {
	p := emptyElement()
	p.id = s.nextElementID()
	p.elementType = Point
	p.values = append(p.values, x)
	p.values = append(p.values, y)
	p.element = s.sketch.AddPoint(big.NewFloat(p.values[0]), big.NewFloat(p.values[1]))
	s.Elements = append(s.Elements, p)
	s.eToC[p.id] = make([]*Constraint, 0)
	return p
}

func (s *Sketch) addOrigin() *Element {
	o := emptyElement()
	o.id = s.nextElementID()
	o.elementType = Point
	o.values = append(o.values, 0)
	o.values = append(o.values, 0)
	s.Elements = append(s.Elements, o)

	var zero big.Float
	zero.SetFloat64(0)
	o.element = s.sketch.AddOrigin(&zero, &zero) // AddLine normalizes a, b, c
	return o
}

func (s *Sketch) addAxis(a float64, b float64, c float64) *Element {
	ax := emptyElement()
	ax.id = s.nextElementID()
	ax.elementType = Axis
	ax.values = append(ax.values, a)
	ax.values = append(ax.values, b)
	ax.values = append(ax.values, c)
	s.Elements = append(s.Elements, ax)

	ax.element = s.sketch.AddAxis(big.NewFloat(a), big.NewFloat(b), big.NewFloat(c)) // AddLine normalizes a, b, c
	return ax
}

// AddLine adds a line to the sketch from [x1, y1] to [x2, y2].
// It returns the line element created.
func (s *Sketch) AddLine(x1 float64, y1 float64, x2 float64, y2 float64) *Element {
	l := emptyElement()
	l.id = s.nextElementID()
	l.elementType = Line
	var a, b, c, t big.Float

	a.SetPrec(utils.FloatPrecision).SetFloat64(y2 - y1) // y' - y
	b.SetPrec(utils.FloatPrecision).SetFloat64(x1 - x2) // x - x'
	c.SetPrec(utils.FloatPrecision).Neg(&a)
	t.SetPrec(utils.FloatPrecision).SetFloat64(x1)
	// c = -ax - by from ax + by + c = 0
	c.Mul(&c, &t)
	t.SetFloat64(y1)
	t.Mul(&t, &b)
	c.Sub(&c, &t)
	l.values = append(l.values, x1)
	l.values = append(l.values, y1)
	l.values = append(l.values, x2)
	l.values = append(l.values, y2)

	l.element = s.sketch.AddLine(&a, &b, &c) // AddLine normalizes a, b, c
	s.Elements = append(s.Elements, l)

	start := s.AddPoint(l.values[0], l.values[1])
	start.isChild = true
	end := s.AddPoint(l.values[2], l.values[3])
	end.isChild = true
	l.children = append(l.children, start)
	s.eToC[start.id] = make([]*Constraint, 0)
	l.children = append(l.children, end)
	s.eToC[end.id] = make([]*Constraint, 0)
	s.eToC[l.id] = make([]*Constraint, 0)
	s.AddDistanceConstraint(l, start, 0.0)
	s.AddDistanceConstraint(l, end, 0.0)
	utils.Logger.Info().
		Uint("line", l.element.GetID()).
		Uint("start", l.children[0].element.GetID()).
		Uint("end", l.children[1].element.GetID()).
		Msg("Added Line")
	return l
}

// AddCircle adds a circle to the sketch at the center point [x, y] with the radius r.
// It returns the circle element created.
func (s *Sketch) AddCircle(x float64, y float64, r float64) *Element {
	c := emptyElement()
	c.id = s.nextElementID()
	c.elementType = Circle
	c.values = append(c.values, x)
	c.values = append(c.values, y)
	c.values = append(c.values, r)

	s.Elements = append(s.Elements, c)

	center := s.AddPoint(c.values[0], c.values[1])
	center.isChild = true
	c.element = center.element

	c.children = append(c.children, center)
	s.eToC[center.id] = make([]*Constraint, 0)
	s.eToC[c.id] = make([]*Constraint, 0)
	utils.Logger.Info().
		Uint("center", c.element.GetID()).
		Msg("Added Circle")
	return c
}

// AddArc adds an arc to the sketch with the center [x1, y1], start point [x2, y2], and end point [x3, y3].
// The arc is created clockwise from start to end point. If the reverse arc is needed, swap start and end.
// It returns the arc element created.
func (s *Sketch) AddArc(x1 float64, y1 float64, x2 float64, y2 float64, x3 float64, y3 float64) *Element {
	a := emptyElement()
	a.id = s.nextElementID()
	a.elementType = Arc
	a.values = append(a.values, x1)
	a.values = append(a.values, y1)
	a.values = append(a.values, x2)
	a.values = append(a.values, y2)
	a.values = append(a.values, x3)
	a.values = append(a.values, y3)

	s.Elements = append(s.Elements, a)

	center := s.AddPoint(a.values[0], a.values[1])
	center.isChild = true
	a.element = center.element
	a.children = append(a.children, center)
	s.eToC[center.id] = make([]*Constraint, 0)

	start := s.AddPoint(a.values[2], a.values[3])
	start.isChild = true
	s.eToC[start.id] = make([]*Constraint, 0)
	end := s.AddPoint(a.values[4], a.values[5])
	end.isChild = true
	s.eToC[end.id] = make([]*Constraint, 0)
	s.eToC[a.id] = make([]*Constraint, 0)
	a.children = append(a.children, start)
	a.children = append(a.children, end)
	s.AddDistanceConstraint(a, start, 0)
	s.AddDistanceConstraint(a, end, 0)
	utils.Logger.Info().
		Uint("arc", a.element.GetID()).
		Uint("start", a.children[1].element.GetID()).
		Uint("end", a.children[2].element.GetID()).
		Msg("Added Arc")
	return a
}

func (s *Sketch) MakeFixed(e *Element) {
	s.sketch.MakeFixed(e.element)
	for _, el := range e.children {
		s.sketch.MakeFixed(el.element)
	}
}

func (s *Sketch) resolveConstraint(c *Constraint) bool {
	if c.state == Resolved {
		return true
	}

	switch c.constraintType {
	case Coincident:
		fallthrough
	case Distance:
		return s.resolveDistanceConstraint(c)
	case Angle:
		fallthrough
	case Perpendicular:
		fallthrough
	case Parallel:
		c.state = Resolved
		return true
	case Ratio:
		return s.resolveRatioConstraint(c)
	case Midpoint:
		return s.resolveMidpointConstraint(c)
	case Tangent:
		return s.resolveTangentConstraint(c)
	}

	return c.state == Resolved
}

func (s *Sketch) resolveConstraints() (int, int) {
	unresolved := 0
	unsolved := 0

	for _, c := range s.constraints {
		if c.state == Unresolved && !s.resolveConstraint(c) {
			unresolved++
		}
		for _, constraint := range c.constraints {
			current, ok := s.sketch.GetConstraint(constraint.GetID())
			if !ok {
				continue
			}
			constraint.Solved = current.Solved
		}
		c.checkSolved()

		if c.state != Solved {
			unsolved++
		}
	}

	return unresolved, unsolved
}

func (s *Sketch) isElementSolved(e *Element) bool {
	return s.sketch.IsElementSolved(e.element)
}

func (s *Sketch) getDistanceConstraint(e *Element) (*Constraint, bool) {
	if e.elementType != Line {
		dc, err := s.findConstraint(Distance, e)
		if err == nil {
			return dc, true
		}

		// if e.elementType != Line { // Move to above
		return nil, false
	}

	for _, c := range s.eToC[e.id] {
		if c.constraintType != Distance || (c.state != Resolved && c.state != Solved) {
			continue
		}
		if len(c.elements) > 1 && c.elements[1] != nil {
			continue
		}
		return c, true
	}

	// Look for a constraint on a line between the start and end
	constraints := s.findConstraints(e.children[0])
	for _, c := range constraints {
		if c.elements[0] == e.children[1] || c.elements[1] == e.children[1] {
			// if c.elements[0] == e.children[1] || c.elements[1] == e.children[2] {
			return c, true
		}
	}

	return nil, false
}

func (s *Sketch) resolveLineLength(e *Element) (float64, bool) {
	if e.elementType != Line {
		return 0, false
	}

	constraints := s.findConstraints(e.children[0])
	for _, c := range constraints {
		if c.constraintType != Distance {
			continue
		}
		if c.elements[0] == e.children[1] || c.elements[1] == e.children[1] {
			val, _ := c.constraints[0].Value.Float64()
			return val, true
		}
	}

	dc, ok := s.getDistanceConstraint(e)
	if ok {
		v, _ := dc.constraints[0].Value.Float64()
		return v, ok
	}

	startConstrained := s.isElementSolved(e.children[0])
	endConstrained := s.isElementSolved(e.children[1])
	if startConstrained && endConstrained {
		// resolve constraint setting p2's distance to the distance from p1 start to p1 end
		v, _ := e.children[0].element.AsPoint().DistanceTo(e.children[1].element.AsPoint()).Float64()

		return v, true
	}

	return 0, false
}

func (s *Sketch) resolveCurveRadius(e *Element) (*big.Float, bool) {
	if e.elementType != Arc && e.elementType != Circle {
		return nil, false
	}

	dc, _ := s.getDistanceConstraint(e)
	// Have a distance constraint already marked as resolved before solving begins!
	if dc != nil {
		var v big.Float
		v.SetPrec(utils.FloatPrecision).SetFloat64(dc.dataValue)
		if len(dc.constraints) > 0 {
			v.Copy(&dc.constraints[0].Value)
		}
		return &v, true
	}

	// Circles and Arcs with solved center and solved elements coincident or distance to the circle / arc
	if centerSolved := s.isElementSolved(e.children[0]); centerSolved {
		// Find constraints against the curve itself (not against its center or other child elements)
		eAll := s.findConstraints(e)
		var other *Element = nil
		for _, ec := range eAll {
			if ec.constraintType != Distance && ec.constraintType != Coincident {
				continue
			}
			other = ec.elements[0]
			if other.id == e.id {
				other = ec.elements[1]
			}
			if !s.isElementSolved(other) {
				continue
			}
			var dv, radius big.Float
			dv.SetPrec(utils.FloatPrecision).SetFloat64(ec.dataValue)
			// Other & e have a distance constraint between them. dist(other, e.center) - c.value is radius
			distFromCurve := other.element.AsPoint().DistanceTo(e.children[0].element.AsPoint())
			radius.Sub(distFromCurve, &dv)
			return &radius, true
		}
	}

	return nil, false
}

func (s *Sketch) checkCircleConstraints() {
	for _, e := range s.Elements {
		if e.elementType != Circle {
			continue
		}

		constraints := s.eToC[e.id]
		if len(constraints) == 1 && constraints[0].constraintType == Distance &&
			len(constraints[0].elements) == 1 {
			// Ignore this constraint because there is nothing else to relate the edge of the circle to
			c := constraints[0]
			e.constraints = make([]*constraint.Constraint, 0)
			s.eToC[e.id] = make([]*Constraint, 0)
			index := -1
			for i := range s.constraints {
				if constraints[i] == c {
					index = i
					break
				}
			}
			if index >= 0 {
				s.constraints = append(s.constraints[:index], s.constraints[index+1:]...)
			}
		}
	}
}

// Solve attempts to solve the sketch by translating and rotating elements until they meet all constraints provided.
// After a solve, each Element's ConstraintLevel will be defined.
// It returns an error if one is encountered during the solve.
func (s *Sketch) Solve() error {
	solveState := solver.None
	passes := 0

	unresolved := 0
	unsolved := 0
	for _, c := range s.constraints {
		if c.state == Unresolved {
			unresolved++
		}
		if c.state != Solved {
			unsolved++
		}
	}
	utils.Logger.Info().
		Int("total", len(s.constraints)).
		Int("unresolved", unresolved).
		Int("unsolved", unsolved).
		Msg("Initial constraint state.")

	// This isn't correct -- should run until everything is solved
	// lastUnsolved := 0
	// lastUnresolved := 0
	s.checkCircleConstraints()
	lastUnresolved, lastUnsolved := s.resolveConstraints()
	for numUnresolved, numUnsolved := lastUnresolved+1, lastUnresolved; /*numUnsolved > 0 ||*/ numUnresolved > 0; numUnresolved, numUnsolved = s.resolveConstraints() {
		if lastUnsolved == numUnsolved && lastUnresolved == numUnresolved {
			utils.Logger.Debug().
				Int("last unsolved", lastUnsolved).
				Int("current unsolved", numUnsolved).
				Int("last unresolved", lastUnresolved).
				Int("current unresolved", numUnresolved).
				Msg("Exiting solve loop early")
			solveState = solver.NonConvergent
			break
		}
		utils.Logger.Info().
			Int("unresolved", numUnresolved).
			Int("unsolved", numUnsolved).
			Msgf("State prior to pass %d", passes+1)
		utils.Logger.Info().Msgf("Running solve pass %d", passes+1)
		s.sketch.ResetClusters() // TODO: this probably needs a reset between passes!
		// Rebuild cluster 0
		s.sketch.BuildClusters() // TODO: this probably needs a reset between passes!
		if utils.LogLevel() <= zerolog.DebugLevel {
			utils.Logger.Info().Int("unresolved", numUnresolved).Msgf("Writing clustered.dot")
			s.ExportGraphViz("clustered.dot")
		}
		solveState = s.sketch.Solve()
		lastUnresolved = numUnresolved
		lastUnsolved = numUnsolved
		passes++
	}
	s.passes += passes

	var copyElements func(e *Element, sketch *core.SketchGraph)
	copyElements = func(e *Element, sketch *core.SketchGraph) {
		if el, ok := s.sketch.GetElement(e.element.GetID()); ok {
			e.element = el
		}
		for _, child := range e.children {
			copyElements(child, sketch)
		}
	}

	// Load solved values back into our elements
	for _, e := range s.Elements {
		// copy internal elements back into external elements
		copyElements(e, s.sketch)
		e.valuesFromSketch(s)
	}

	if s.sketch.Conflicting().Count() > 0 {
		solveState = solver.OverConstrained
	}

	switch solveState {
	case solver.Solved:
		return nil
	default:
		return errors.New("failed to solve completely")
	}
}

func (s *Sketch) ConflictingConstraints() []*Constraint {
	conflicting := make([]*Constraint, 0)
	for _, c := range s.constraints {
		for _, ic := range c.constraints {
			if s.sketch.Conflicting().Contains(ic.GetID()) {
				conflicting = append(conflicting, c)
				break
			}
		}
	}

	return conflicting
}

func (s *Sketch) calculateRectangle(scale float64) (float64, float64, float64, float64) {
	minX := math.MaxFloat64
	minY := math.MaxFloat64
	maxX := math.MaxFloat64 * -1
	maxY := math.MaxFloat64 * -1

	for _, e := range s.Elements {
		lX, lY, hX, hY := e.minMaxXY()
		if lX < minX {
			minX = lX
		}
		if lY < minY {
			minY = lY
		}
		if hX > maxX {
			maxX = hX
		}
		if hY > maxY {
			maxY = hY
		}
	}
	return minX * scale, minY * scale, maxX * scale, maxY * scale
}

func (s *Sketch) ExportImage(filename string, args ...float64) error {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	return s.WriteImage(f, args...)
}

// ExportImage exports an image representing the current state of the sketch.
// The origin and axes will be colored gray. Fully constrained solved elements will be colored black.
// Other elements will be colored blue.
// It returns an error if unable to open the output file.
func (s *Sketch) WriteImage(out io.Writer, args ...float64) error {
	width := 150.0
	height := 150.0
	scale := 1.0

	if len(args) > 0 {
		width = args[0]
	}
	if len(args) > 1 {
		height = args[1]
	}

	minx, miny, maxx, maxy := s.calculateRectangle(scale)

	// Calculate viewbox
	vw := float64(maxx - minx)
	vh := float64(maxy - miny)

	scaleX := width / vw
	scaleY := height / vh
	scale = scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	c := canvas.New(width, height)
	ctx := canvas.NewContext(c)
	ctx.SetCoordSystem(canvas.CartesianI)
	ctx.SetCoordRect(canvas.Rect{X0: minx, Y0: miny, X1: vw, Y1: vh}, width, height)

	ctx.SetStrokeColor(canvas.Gray)
	ctx.SetStrokeWidth(0.5)
	ctx.MoveTo(0, 0)
	ctx.LineTo((maxx * scale), 0)
	ctx.Close()
	ctx.Stroke()
	ctx.MoveTo(0, (miny * scale))
	ctx.LineTo(0, (maxy * scale))
	ctx.Close()
	ctx.Stroke()

	for _, e := range s.Elements {
		e.DrawToSVG(s, ctx, scale)
	}

	c.Fit(5.0)

	svg := svg.New(out, c.W, c.H, &svg.Options{})

	c.RenderTo(svg)

	return svg.Close()
}

func (s *Sketch) ExportGraphViz(filename string) error {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(s.sketch.ToGraphViz())

	return err
}
