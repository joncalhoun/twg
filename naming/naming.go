package naming

import "fmt"

type Dog struct {
	Name string
	Age  int
}

func (d Dog) Bark(muzzled bool) {
	if muzzled {
		fmt.Println("woof")
	} else {
		fmt.Println("WOOF!!")
	}
}

func Speak(lang string) {
	switch lang {
	case "spanish":
		fmt.Println("Hola")
	default:
		fmt.Println("Hello")
	}
}
