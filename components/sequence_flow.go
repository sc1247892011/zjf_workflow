package components

import (
	"fmt"
	"log"
	"strings"
)

// SequenceFlow 代表 BPMN 中的序列流。
type SequenceFlow struct {
	ExecutionId string `xml:"executionId,attr"` // 绑定 id 属性
	SourceRef   string `xml:"sourceRef,attr"`   // 绑定 sourceRef 属性
	TargetRef   string `xml:"targetRef,attr"`   // 绑定 targetRef 属性
	Expression  string `xml:"ConditionExpression,omitempty"`
	X           string `xml:"x,attr"`
	Y           string `xml:"y,attr"`
	H           string `xml:"h,attr"`
	W           string `xml:"w,attr"`
	Listener    string `xml:"Listener"` // 任务监听 执行完毕之后的后续逻辑 方法名称可以用逗号隔开 传递多段逻辑
}

func (sequenceFlow SequenceFlow) Execute(ctx *WorkflowContext) {
	//没表达式 直接过
	if strings.Trim(sequenceFlow.Expression, " ") == "" {
		exec := ctx.Model.AllData[sequenceFlow.TargetRef]
		exec.Execute(ctx)
		return
	}

	//根据表达式判断 这个序列流是否往下走
	//startEvent.leaveType== 'Sick Leave'
	attributes := ExtractAttributes(sequenceFlow.Expression)
	log.Println("Extracted attributes:", attributes) // 输出 ["data.value", "data.status"]
	// 为表达式中的变量赋值
	nodeService := GetServiceFactory().GetNodeService()
	parameters, err1 := nodeService.GetAttributeByExpression(ctx.Tx, sequenceFlow.Expression, ctx.ProcessInstanceId)

	if err1 != nil {
		log.Println("Failed to get attribute:", err1)
		return
	}

	result, err := EvaluateExpression(sequenceFlow.Expression, parameters)
	if err != nil {
		log.Println("Failed to parse expression:", err)
		return
	}

	// 判断 result 是否为布尔类型
	if boolResult, ok := result.(bool); ok {
		log.Println("Result is a boolean:", boolResult)
		if boolResult {
			//条件满足 继续走
			//执行监听
			RunListener(sequenceFlow.Listener, ctx)
			//下一步
			exec := ctx.Model.AllData[sequenceFlow.TargetRef]
			exec.Execute(ctx)
		} else {
			return
		}
	} else {
		fmt.Printf("Result is not a boolean, it is of type %T with value %v\n", result, result)
	}
}
