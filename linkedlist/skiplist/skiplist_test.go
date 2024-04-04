package skiplist

import (
	"fmt"
	"testing"
)

func TestRange(t *testing.T) {
	sk := NewSkiplist[int, int]()
	sk.Put(1, 1)
	sk.Put(2, 2)
	sk.Put(3, 3)

	res := sk.Range(2, 5)
	fmt.Println(res)
	res = sk.Range(1, 2)
	fmt.Println(res)
	res = sk.Range(4, 5)
	fmt.Println(res)

	sk.Delete(2)
	res = sk.Range(1, 5)
	fmt.Println(res)

	sk.Put(4)
	res = sk.Range(1, 5)
	fmt.Println(res)

	res = sk.Get(1)
	fmt.Println(res)
}
