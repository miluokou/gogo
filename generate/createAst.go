package generate

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
)

func CreateAst(c *gin.Context) {
	packageName := "job"       // 自定义 method 变量名
	methodName := "testMethod" // 自定义 method 变量名

	code := fmt.Sprintf(`
package %s

import "fmt"

func %s() {
	fmt.Println("Hello, World!")
}`, packageName, methodName)

	// 将代码字符串写入新文件
	err := ioutil.WriteFile(fmt.Sprintf("controllers/job/%s.go", methodName), []byte(code), 0644)
	if err != nil {
		panic(err)
	}
}
