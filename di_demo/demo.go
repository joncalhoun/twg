package di_demo

import "errors"

// func Demo() {
// 	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
// 	err := doThing()
// 	if err != nil {
// 		logger.Println("failed to do the thing:", err)
// 		return
// 	}
// 	// pretend to keep going
// }

// func Demo(logger *log.Logger) {
// 	err := doThing()
// 	if err != nil {
// 		logger.Println("failed to do the thing:", err)
// 		return
// 	}
// 	// pretend to keep going
// }

// func Demo(logFn func(...interface{})) {
// 	err := doThing()
// 	if err != nil {
// 		logFn("failed to do the thing:", err)
// 		return
// 	}
// 	// pretend to keep going
// }

func Demo(logger interface {
	Println(...interface{})
}) {
	err := doThing()
	if err != nil {
		logger.Println("failed to do the thing:", err)
		return
	}
	// pretend to keep going
}

func doThing() error {
	// return nil
	return errors.New("fake error")
}

type Logger interface {
	Println(...interface{})
}

type Thing struct {
	Logger Logger
}

func (t Thing) Demo() {
	err := doThing()
	if err != nil {
		t.Logger.Println("failed to do the thing:", err)
		return
	}
	// pretend to keep going
}
