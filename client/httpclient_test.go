package client

import (
	"fmt"
	"net/url"
	"testing"
)

func TestHttpClient(t *testing.T) {
	type JsonData struct {
		Id   int64
		Name string
	}

	jsonData := JsonData{1, "Jack"}

	urlData := url.Values{}
	urlData.Set("id", "1")
	urlData.Set("name", "李明")

	host := "http://127.0.0.1:8080"
	path := "/client"

	response, err := Get(host, path).Param(urlData).Do()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(response)

	response, err = Post(host, path).Form(urlData).Do()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(response)

	response, err = Delete(host, path).Json(jsonData).Do()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(response)

	response, err = Put(host, path).Json(jsonData).Do()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(response)
}
