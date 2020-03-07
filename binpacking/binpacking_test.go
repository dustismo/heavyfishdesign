package binpacking

import "testing"

// a bin for testing
type MockBin struct {
	Width  float64
	Height float64
}

func (m MockBin) GetWidth() float64 {
	return m.Width
}

func (m MockBin) GetHeight() float64 {
	return m.Height
}

func TestBinPacking(t *testing.T) {
	
}
