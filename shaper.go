package shaping

type Shaper interface {
	// Shape takes an Input and shapes it into the Output.
	Shape(Input) Output
}
