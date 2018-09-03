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

func Color(name string) string {
	switch name {
	case "blue":
		return "#0000FF"
	case "white":
		return "#FFFFFF"
	case "black":
		return "#000000"
	case "grey":
		return "#888888"
	default:
		return "#000000"
	}
}
