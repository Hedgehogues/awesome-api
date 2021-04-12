package main

import (
	"fmt"

	cl "github.com/hedgehogues/xxx"
	"github.com/hedgehogues/xxx/example/apis"
)

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func Example() {
	var m cl.Method
	m, err := apis.NewCatAPI(&apis.RequestCatAPI{})
	checkError(err)
	catAPI := cl.NewClient("https", "thatcopy.pw", nil, nil)
	resp, err := catAPI.Request(m, nil)
	checkError(err)
	fmt.Println("Text:", resp.Text())
	fmt.Println("Bytes:", resp.Bytes)
	map_, err := resp.Map()
	checkError(err)
	fmt.Println("Map:", map_)
	respAPI := apis.ResponseCatAPI{}
	err = resp.Struct(&respAPI)
	checkError(err)
	fmt.Println("Struct:", respAPI)
}

func main() {
	Example()
}
