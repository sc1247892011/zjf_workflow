package main

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/Knetic/govaluate"
	"github.com/sc1247892011/zjf_workflow/components"
)

func main() {

	//testReflect()
	//测试解析xml
	testParse()

	//测试流程定义部署 (根据流程名称 和 版本做了 唯一性约束)
	//testRepositoryDeploy()
	//测试启动流程
	//testRuntimeServiceStart()
	//testExpression()
}

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

// 测试流程定义部署
func testRepositoryDeploy() {
	//模拟用户操作

	// 数据库连接信息
	dsn := "root:root@tcp(localhost:3306)/zjf_workflow?charset=utf8mb4&parseTime=True&loc=Local"

	// 打开数据库连接
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Println("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 测试数据库连接
	err = db.Ping()
	if err != nil {
		log.Println("Failed to ping database: %v", err)
	}

	fmt.Println("Successfully connected to the database!")

	//mysqlService := &service.MySQLRepositoryService{DB: db}
	components.Init(db, "mysql")
	mysqlService := components.GetMySQLRepositoryService()
	tx, _ := mysqlService.GetTransaction()
	xmlbyte, _ := components.ReadXMLFile(`D:/workspace/vscodeworkspace/go/moduletest/module/zjf_workflow/xml/leave.xml`)
	// 定义测试数据
	pd := &components.ProcessDefinition{
		ProcessDefinitionName: "Leave Request Process",
		XMLContent:            xmlbyte,
		CreatedAt:             time.Now(),
		CreatedBy:             "SC",
		Status:                "active",
		Description:           "Test process definition",
	}

	//根据名称和版本找流程定义
	//res, _ := mysqlService.GetProcessDefinitionByNameAndVersion(pd.Name, pd.Version)
	//找最新的流程定义
	//res, _ := mysqlService.GetLatestProcessDefinitionByName(pd.Name)
	//删除流程定义
	//mysqlService.DeleteProcessDefinition(res.Id)

	//传入一个xml文件 自动根据当前最新版本 +1 插入
	_, err2 := mysqlService.SaveProcessDefinition(tx, pd)

	if err2 != nil {
		fmt.Printf("Error: %v\n", err2)
	}
}

// 测试xml解析
func testParse() {
	// 文件路径
	//filename := filepath.FromSlash(`D:/workspace/vscodeworkspace/go/moduletest/module/zjf_workflow/xml/leave.xml`)
	filename := filepath.FromSlash(`D:/workspace/vscodeworkspace/go/moduletest/module/zjf_workflow/xml/test.xml`)
	// 调用 ParseXML 函数解析 XML 文件
	model, err := components.ParseXML(filename)
	if err != nil {
		log.Println("Error parsing XML file: %v", err)
	}

	// 输出解析结果
	fmt.Printf("model Name: %s\n", model.ProcessDefinitionName)

	// 打印 StartEvents
	log.Println("\nStart Events:")
	for id, element := range model.StartEvents {
		fmt.Printf("Id: %s, Start Events: %+v\n", id, element)
	}

	// 打印 Tasks
	log.Println("\nTasks:")
	for id, task := range model.Tasks {
		fmt.Printf("Id: %s, Task: %+v\n", id, task)
	}

	// 打印 ParallelGateways
	log.Println("\nParallel Gateways:")
	for id, gateway := range model.ParallelGateways {
		fmt.Printf("Id: %s, Parallel Gateway: %+v\n", id, gateway)
	}

	// 打印 ExclusiveGateways
	log.Println("\nExclusive Gateways:")
	for id, gateway := range model.ExclusiveGateways {
		fmt.Printf("Id: %s, Exclusive Gateway: %+v\n", id, gateway)
	}

	// 打印 EndEvents
	log.Println("\nEnd Events:")
	for id, endEvent := range model.EndEvents {
		fmt.Printf("Id: %s, End Event: %+v\n", id, endEvent)
	}

	// 打印 SequenceFlows
	log.Println("\nSequence Flows:")
	for id, flow := range model.SequenceFlows {
		fmt.Printf("Id: %s, Flow: %+v\n", id, flow)
	}
}

func testRuntimeServiceStart() {
	// 数据库连接信息
	dsn := "root:root@tcp(localhost:3306)/zjf_workflow?charset=utf8mb4&parseTime=True&loc=Local"

	// 打开数据库连接
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Println("Failed to connect to database: %v", err)
	}
	defer db.Close()
	components.Init(db, "mysql")
	// 传递一个processDefinitionKey 找到最新的版本
	mysqlService := components.GetMySQLRuntimeService()
	tx, _ := mysqlService.GetTransaction()
	// 前端 进页面 新建一个工作流 从model里拿startEv的formData
	// 填写完毕 把数据 放到 ctx.Data里 传递给下一个节点
	// 如果是序列流 网关 拿到了ctx.Data 就继续传递
	// 如果是节点 拿自己的输出去替换 ctx.Data ，并且每个节点的输出都要持久化到数据库
	mysqlService.StartProcessInstance(tx, "Leave Request Process", "sc test001", "sc",
		`{
    "startEvent": {
            "employeeName": "John Doe",
            "leaveType": "Sick Leave",
            "startDate": "2024-09-01",
            "endDate": "2024-09-05",
            "reason": "生病了"
        }
}
`)
}

func testExpression() {

	expression := "  approveTask1.approvalStatus == 'Approve' || approveTask2.approvalStatus == 'Reject'"
	parameters := make(map[string]interface{})
	parameters["approveTask1.approvalStatus"] = "Approve"
	parameters["approveTask2.approvalStatus"] = "Approve"
	// 遍历参数，替换表达式中的键为对应的值
	for key, value := range parameters {
		// 将参数中的值替换到表达式中
		strValue := fmt.Sprintf("'%v'", value)
		expression = strings.ReplaceAll(expression, key, strValue)
	}

	expr, err := govaluate.NewEvaluableExpression(expression)
	if err != nil {
		fmt.Errorf("failed to parse expression: %v", err)
	}

	// 评估表达式
	result, err := expr.Evaluate(nil)
	if err != nil {
		fmt.Errorf("failed to evaluate expression: %v", err)
	}
	result2, _ := result.(string)
	fmt.Printf(result2)
}
