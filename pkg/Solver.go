package dlineate

import (
	core "github.com/marcuswu/dlineate/internal/core"
	el "github.com/marcuswu/dlineate/internal/element"
)

type Sketch struct {
	sketch *core.SketchGraph
}

func NewSketch() *Sketch {
	s := new(Sketch)
	s.sketch = core.NewSketch()
}

func (s *Sketch) AddElement(e *Element) {
	e.addToSketch(s.sketch)
}