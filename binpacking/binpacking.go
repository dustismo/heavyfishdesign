package binpacking

import (
	"math"
)

type BinBoundary interface {
	GetWidth() float64
	GetHeight() float64
}

type Bin struct {
	Width  float64
	Height float64
	X      float64
	Y      float64
	Object interface{}
	// Packable    BinPackable
	Rotated     bool // if true this should be rotated by 90deg
	HasObject   bool
	HasChildren bool
	LeftChild   *Bin
	RightChild  *Bin
}

type Container struct {
	Root   *Bin
	Width  float64
	Height float64
	X      float64
	Y      float64
}

type paddedPackable struct {
	object   interface{}
	boundary BinBoundary
	padding  float64
}

func (p *paddedPackable) GetWidth() float64 {
	return (p.padding * 2) + p.boundary.GetWidth()
}

func (p *paddedPackable) GetHeight() float64 {
	return (p.padding * 2) + p.boundary.GetHeight()
}

func NewContainer(x float64, y float64, width float64, height float64) *Container {
	return &Container{
		Root:   NewBin(x, y, width, height, false),
		Width:  width,
		Height: height,
		X:      x,
		Y:      y,
	}
}

// creates a new container with a single element.  It will not be possible to add
// to this container.
func NewSingleObjectContainer(obj interface{}, x float64, y float64, width float64, height float64) *Container {
	container := NewContainer(x, y, width, height)
	container.Root.HasObject = true
	container.Root.Object = obj
	return container
}

// finds all the empty bins, useful for finding how
// much space is left over
func (c *Container) GetEmptyBins() []*Bin {
	return c.Root.getEmpties([]*Bin{})
}

// returns the total area in unit^2 of emptyness
func (c *Container) GetEmptyArea() float64 {
	area := 0.0
	for _, b := range c.GetEmptyBins() {
		area += b.Width * b.Height
	}
	return area
}

func (c *Container) IsEmpty() bool {
	return !c.Root.HasChildren && !c.Root.HasObject
}

// Inserts the packable into the container and returns the bin it was placed in
// if possible
func (c *Container) Insert(object interface{}, boundary BinBoundary) (bool, Bin) {
	inserted, bin := c.Root.Insert(object, boundary)
	if !inserted {
		bin = &Bin{}
	}

	return inserted, *bin
}

// inserts into the container with the specified padding on all sides
func (c *Container) InsertWithPadding(object interface{}, boundary BinBoundary, padding float64) (bool, Bin) {

	packable := &paddedPackable{object, boundary, padding}
	inserted, bin := c.Insert(packable, packable)
	if !inserted {
		return inserted, bin
	}
	// adjust the x and y of the bin
	newBin := NewBin(
		bin.X+padding,
		bin.Y+padding,
		boundary.GetWidth(),
		boundary.GetHeight(),
		bin.Rotated,
	)
	newBin.HasObject = true
	newBin.Object = object
	return true, *newBin
}

func NewBin(x float64, y float64, width float64, height float64, rotated bool) *Bin {
	return &Bin{
		X:           x,
		Y:           y,
		Width:       width,
		Height:      height,
		HasObject:   false,
		HasChildren: false,
		Rotated:     rotated,
	}
}

func (b *Bin) getEmpties(collect []*Bin) []*Bin {
	if b.HasObject {
		return collect
	}
	if b.HasChildren {
		collect = b.LeftChild.getEmpties(collect)
		collect = b.RightChild.getEmpties(collect)
	} else {
		// no children and no packable
		collect = append(collect, b)
	}
	return collect
}

// is this split on the horizontal axis: i.e. left means top, right means bottom
func (b *Bin) IsHorizontalSplit() bool {
	if !b.HasChildren {
		return false
	}
	return b.LeftChild.X == b.RightChild.X
}

// based on algorith described here:
// http://blackpawn.com/texts/lightmaps/

// Recursively attempts to add the packable in the bin
// returns (true, bin) or (false, nil)
func (b *Bin) Insert(object interface{}, boundary BinBoundary) (bool, *Bin) {
	if b.HasChildren {
		// attempt to insert into one of the children
		inserted, bin := b.LeftChild.Insert(object, boundary)
		if inserted {
			return inserted, bin
		}
		return b.RightChild.Insert(object, boundary)
	}

	width := boundary.GetWidth()
	height := boundary.GetHeight()

	if width == 0 || height == 0 ||
		math.IsNaN(width) || math.IsNaN(height) {
		return false, nil
	}
	rotated := false

	// this node doesnt have any children

	if b.Height < height || b.Width < width {
		// too small, try to rotate
		rotated = true
		width = boundary.GetHeight()
		height = boundary.GetWidth()
		if b.Height < height || b.Width < width {
			// still too small, return
			return false, nil
		}
	}

	if b.HasObject {
		// already a packable here
		return false, nil
	}

	if b.Height == height && b.Width == width {
		// fits perfectly!
		b.Object = object
		b.HasObject = true
		return true, b
	}

	// split and add children
	b.HasChildren = true

	// split horizontal or vertical?
	// we attempt to maximize left over space

	// TODO: what if we allow rotate 90deg?

	// should rotate?

	if (b.Width - width) > (b.Height - height) {
		// split on vertical axis
		b.LeftChild = NewBin(
			b.X,
			b.Y,
			width,
			b.Height,
			rotated,
		)

		b.RightChild = NewBin(
			b.X+width,
			b.Y,
			b.Width-width,
			b.Height,
			rotated,
		)
	} else {
		// split on horizontal axis
		b.LeftChild = NewBin(
			b.X,
			b.Y,
			b.Width,
			height,
			rotated,
		)

		b.RightChild = NewBin(
			b.X,
			b.Y+height,
			b.Width,
			b.Height-height,
			rotated,
		)
	}
	// now insert into left child
	return b.LeftChild.Insert(object, boundary)
}
