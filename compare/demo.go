package compare

func Square(i int) int {
	return i * i
}

// Dog is used a demo type
type Dog struct {
	Name string
	Age  int
}

// DogWithFn is used a demo type
type DogWithFn struct {
	Name string
	Age  int
	Fn   func()
}
