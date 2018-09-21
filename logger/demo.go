package logger

type Printlner interface {
	Println(v ...interface{})
}

func Demo(logger Printlner) {
	logger.Println("something went wrong!")
}
