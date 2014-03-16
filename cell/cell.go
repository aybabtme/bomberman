package cell

type GameObject interface {
	String() string
	Draw(x, y int)
	Traversable() bool
}

type Exported struct {
	Name string
}

// Cell is a cell on the board. A cell can have many z layers.
type Cell struct {
	base    GameObject
	zLayers []GameObject
	X       int
	Y       int
}

// NewCell creates a cell with base as z layer 0.
func NewCell(base GameObject, x, y int) *Cell {
	return &Cell{
		base:    base,
		zLayers: make([]GameObject, 0, 0),
		X:       x,
		Y:       y,
	}
}

// Top returns the object on top of the z layers.
func (c *Cell) Top() GameObject {
	if len(c.zLayers) == 0 {
		return c.base
	}
	return c.zLayers[len(c.zLayers)-1]
}

// Push adds an object to the top of the z layers.
func (c *Cell) Push(o GameObject) {
	c.zLayers = append(c.zLayers, o)
}

// Pop returns the top object.  It will remove the object from the layers unless
// it's the base layer.
func (c *Cell) Pop() (GameObject, bool) {
	if len(c.zLayers) == 0 {
		return c.base, false
	}
	var pop GameObject
	pop = c.zLayers[len(c.zLayers)-1]
	c.zLayers = c.zLayers[:len(c.zLayers)-1]
	return pop, true
}

// RemoveLayer removes the object at layer z, unless if z is the base layer. If
// z is out of bound, this will panic.
func (c *Cell) RemoveLayer(z int) GameObject {
	// Case z == 0
	if z == 0 {
		return c.base
	}

	// Case z >= len, z <= 0: panic
	removed := c.zLayers[z-1]

	// Case z == 1: will skip body of loop
	for i := z - 1; i < len(c.zLayers)-1; i++ {
		c.zLayers[i] = c.zLayers[i+1]
	}

	// z has at least 1 element
	c.zLayers = c.zLayers[:len(c.zLayers)-1]
	return removed
}

// Remove finds an GameObject to remove in the Cells z layers. The base
// layer can't be removed.
func (c *Cell) Remove(toRemove GameObject) bool {
	if c.base == toRemove {
		return false
	}
	for z, obj := range c.zLayers {
		if obj == toRemove {
			c.RemoveLayer(z + 1)
			return true
		}
	}
	return false
}

// Layer looks up the object at layer z. If z is out of bound, this will panic.
func (c *Cell) Layer(z int) GameObject {
	if z == 0 {
		return c.base
	}
	// Invalid Z will panic
	return c.zLayers[z-1]
}

// Depth gives the depth of the z layer, including the base layer.
func (c *Cell) Depth() int {
	return 1 + len(c.zLayers)
}

func (c *Cell) Export() *Exported {
	return &Exported{
		Name: c.Top().String(),
	}
}
