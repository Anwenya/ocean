package reflect

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

type Address struct {
	Local string
}

type User struct {
	Name string
	Age  int
	Addr *Address
}

func T(_sliceMapValue reflect.Value) {
	for i := 0; i < _sliceMapValue.Len(); i++ {
		record := _sliceMapValue.Index(i)
		fmt.Println(record.Kind())
		if record.Elem().Kind() == reflect.Map {
			s := record.Elem().MapIndex(reflect.ValueOf("Name")).Elem().String()
			fmt.Println(s)
			a := record.Elem().MapIndex(reflect.ValueOf("Name")).Elem().String()
			fmt.Println(a)
			record.Elem().SetMapIndex(reflect.ValueOf("Age"), reflect.ValueOf(50))
		}
		fmt.Println(record.Elem().MapIndex(reflect.ValueOf("Age")).Elem().Int())
		reflect.Indirect(reflect.ValueOf(record.Addr().Interface())).Elem()
	}
}

func TestStruct(t *testing.T) {
	var a interface{} = User{}
	str := `{"name":"xiaoming","age":20}`
	userValue := reflect.New(reflect.TypeOf(a)).Elem()
	err := json.Unmarshal([]byte(str), userValue.Addr().Interface())
	if err != nil {
		fmt.Println(err)
	}
	//user := userValue.Interface().(User)
	//fmt.Println(user)
	name := userValue.FieldByName("Addr")
	fmt.Println(name.IsNil())

	_map := map[string]interface{}{
		"Name": "xiaoming",
		"Age":  20,
	}

	_sliceMap := []interface{}{_map}

	_sliceMapValue := reflect.ValueOf(_sliceMap)
	T(_sliceMapValue)

	//s := reflect.ValueOf(_map).MapIndex(reflect.ValueOf("Name")).Elem().String()
	//fmt.Println(s)
}
