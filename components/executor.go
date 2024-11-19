package components

import (
	"fmt"
	"reflect"
	"strings"
)

type Executor interface {
	Execute(ctx *WorkflowContext) // 修改接口以使用 WorkflowContext
}

// 执行后续的监听逻辑
func RunListener(Listener string, ctx *WorkflowContext) string {
	listenerArr := strings.Split(Listener, ",")
	var temp string = ""
	for _, value := range listenerArr {
		temp += value
	}
	testReflect()
	return temp
}

// 反射测试
type MyStruct struct{}

func (m MyStruct) Hello() {
	fmt.Println("Hello, World!")
}

func (m MyStruct) Greet(name string) {
	fmt.Println("Hello,", name)
}

func testReflect() {

	// 创建结构体实例
	myStruct := MyStruct{}

	// 获取反射对象
	value := reflect.ValueOf(myStruct)

	// 动态调用方法
	methodName := "Hello"
	method := value.MethodByName(methodName)
	if method.IsValid() {
		method.Call(nil) // 无参数
	}

	// 调用带参数的方法
	methodName = "Greet"
	method = value.MethodByName(methodName)
	if method.IsValid() {
		method.Call([]reflect.Value{reflect.ValueOf("Alice")}) // 带参数
	} else {
		fmt.Println("方法未找到")
	}
}
