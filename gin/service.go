package gin

import "fmt"

type P struct {
}

type PS struct {
	Name string
	Age  int
}

func (p *P) Test(data any) (int, error) {
	d, ok := data.(PS)
	fmt.Println(d, ok)
	return 0, nil
}
