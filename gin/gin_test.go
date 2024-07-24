package gin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/tealeg/xlsx"
	"net/http"
	"reflect"
	"testing"
)

func TestGin(t *testing.T) {
	engine := gin.Default()
	engine.GET("/download/:id", download)
	engine.GET("/client", clientGet)
	engine.POST("/client", clientPost)
	engine.DELETE("/client", clientDelete)
	engine.PUT("/client", clientPut)
	engine.POST("/reflect", testReflect)
	engine.Run(":8080")
}

func download(c *gin.Context) {
	type params struct {
		Id int `uri:"id"`
	}

	style := xlsx.NewStyle()
	style.Alignment.Horizontal = "center"
	style.Alignment.Vertical = "center"
	style.Font.Bold = true
	style.Font.Color = "FF0000"

	header := []string{
		"111111111",
		"222222222",
		"333333333",
	}

	keywords := []string{
		"123", "456",
	}

	var param params
	err := c.BindUri(&param)

	file := xlsx.NewFile()

	sheet, _ := file.AddSheet("sheet1")
	for _, v := range header {
		row := sheet.AddRow()
		cell := row.AddCell()
		cell.SetStyle(style)
		cell.Value = v
	}

	for _, v := range keywords {
		row := sheet.AddRow()
		cell := row.AddCell()
		cell.Value = v
	}

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+"sheet1.xlsx")
	err = file.Write(c.Writer)
	if err != nil {
		fmt.Println(err)
	}
}

func clientGet(c *gin.Context) {
	type params struct {
		Id   int    `form:"id"`
		Name string `form:"name"`
	}
	var p params

	err := c.BindQuery(&p)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	fmt.Println(p)
}

func clientPost(c *gin.Context) {
	type params struct {
		Id   int    `form:"id"`
		Name string `form:"name"`
	}
	var p params

	err := c.ShouldBind(&p)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	fmt.Println(p)
}

func clientDelete(c *gin.Context) {
	type params struct {
		Id   int    `form:"id"`
		Name string `form:"name"`
	}
	var p params

	err := c.BindJSON(&p)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	fmt.Println(p)
}

func clientPut(c *gin.Context) {
	type params struct {
		Id   int    `form:"id"`
		Name string `form:"name"`
	}
	var p params

	err := c.BindJSON(&p)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	fmt.Println(p)
}

func testReflect(c *gin.Context) {
	p := &P{}

	type param struct {
		Name string
		Age  int
	}

	var ps param
	if err := c.BindJSON(&ps); err != nil {
		return
	}
	// 找对应的方法 找不到 -> 方法未找到状态码
	productLinkageValue := reflect.ValueOf(p)
	methodValue := productLinkageValue.MethodByName("Test")
	in := []reflect.Value{reflect.ValueOf(ps.Name), reflect.ValueOf(ps.Age)}
	// 调用对应方法
	results := methodValue.Call(in)
	fmt.Println(results[0].Int())
}
