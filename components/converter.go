package components

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
)

// ParseXML 解析BPMN XML内容并将其转换为Model对象
func ParseXMLByte(byteValue []byte) (*Model, error) {
	// Unmarshal XML数据
	var process Process // 假设在 elements 包中定义了 Process 结构体
	err := xml.Unmarshal(byteValue, &process)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal XML: %v", err)
	}

	// 使用流程的名称创建一个新的Model
	model := NewModel(process.Name)

	// 遍历所有的XML元素并添加到Model中

	// 添加 StartEvent 到 Model
	for _, startEvent := range process.StartEvents {
		model.AddStartEvent(startEvent.ExecutionId, startEvent)
	}

	// 添加 Task 到 Model
	for _, task := range process.Tasks {
		model.AddTask(task.ExecutionId, task)
	}

	// 添加 ParallelGateway 到 Model
	for _, gateway := range process.ParallelGateways {
		model.AddParallelGateway(gateway.ExecutionId, gateway)
	}

	// 添加 ExclusiveGateway 到 Model
	for _, gateway := range process.ExclusiveGateways {
		model.AddExclusiveGateway(gateway.ExecutionId, gateway)
	}

	// 添加 EndEvent 到 Model
	for _, endEvent := range process.EndEvents {
		model.AddEndEvent(endEvent.ExecutionId, endEvent)
	}

	// 添加 SequenceFlow 到 Model
	for _, flow := range process.SequenceFlows {
		model.AddSequenceFlow(flow.ExecutionId, flow)
	}

	return model, nil
}

// ParseXML 解析BPMN XML文件并将其转换为Model对象
func ParseXML(filename string) (*Model, error) {
	// 读取XML文件
	xmlFile, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open XML file: %v", err)
	}
	defer xmlFile.Close()

	byteValue, _ := ioutil.ReadAll(xmlFile)

	// Unmarshal XML数据
	var process Process
	err = xml.Unmarshal(byteValue, &process)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal XML: %v", err)
	}

	// 使用流程的名称创建一个新的Model
	model := NewModel(process.Name)
	// 遍历所有的XML元素并添加到Model中

	// 添加 StartEvent 到 Model
	for _, startEvent := range process.StartEvents {
		model.AddStartEvent(startEvent.ExecutionId, startEvent)
	}

	// 添加 Task 到 Model
	for _, task := range process.Tasks {
		model.AddTask(task.ExecutionId, task)
	}

	// 添加 ParallelGateway 到 Model
	for _, gateway := range process.ParallelGateways {
		model.AddParallelGateway(gateway.ExecutionId, gateway)
	}

	// 添加 ExclusiveGateway 到 Model
	for _, gateway := range process.ExclusiveGateways {
		model.AddExclusiveGateway(gateway.ExecutionId, gateway)
	}

	// 添加 EndEvent 到 Model
	for _, endEvent := range process.EndEvents {
		model.AddEndEvent(endEvent.ExecutionId, endEvent)
	}

	// 添加 SequenceFlow 到 Model
	for _, flow := range process.SequenceFlows {
		model.AddSequenceFlow(flow.ExecutionId, flow)
	}

	return model, nil
}
