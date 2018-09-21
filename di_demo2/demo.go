package di_demo2

import (
	"log"
	"os"
	"sync"
)

// // Global state ex initial code
// func SomeFunc() {
// 	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
// 	// ...
// 	logger.Println("this is a log")
// }

// var logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)

// func SomeFunc(logger interface {
// 	Println(...interface{})
// }) {
// 	// ...
// 	logger.Println("this is a log")
// }

type Logger interface {
	Println(...interface{})
}

type Thing struct {
	Logger Logger
	once   sync.Once
}

func (t *Thing) logger() Logger {
	t.once.Do(func() {
		if t.Logger == nil {
			t.Logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
		}
	})
	return t.Logger
}

func (t *Thing) SomeFunc() {
	// ...
	t.logger().Println("this is a log")
}
