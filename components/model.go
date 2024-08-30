package components

import "sync"

// Model 代表整个流程模型，包含所有元素和序列流
type Model struct {
	ProcessDefinitionName string // 模型的名称，用于标识不同的模型
	Version               int
	StartEvents           map[string]StartEvent       // 存储所有的开始事件，使用唯一Id作为键
	Tasks                 map[string]Task             // 存储所有的任务，使用唯一Id作为键
	ParallelGateways      map[string]ParallelGateway  // 存储所有的并行网关，使用唯一Id作为键
	ExclusiveGateways     map[string]ExclusiveGateway // 存储所有的互斥网关，使用唯一Id作为键
	EndEvents             map[string]EndEvent         // 存储所有的结束事件，使用唯一Id作为键
	SequenceFlows         map[string]SequenceFlow     // 存储所有的序列流，使用唯一Id作为键
	AllData               map[string]Executor         // 冗余数据
}

// NewModel 创建并初始化一个新的模型，并为其设置名称
func NewModel(processDefinitionName string) *Model {
	return &Model{
		ProcessDefinitionName: processDefinitionName,
		StartEvents:           make(map[string]StartEvent),
		Tasks:                 make(map[string]Task),
		ParallelGateways:      make(map[string]ParallelGateway),
		ExclusiveGateways:     make(map[string]ExclusiveGateway),
		EndEvents:             make(map[string]EndEvent),
		SequenceFlows:         make(map[string]SequenceFlow),
		AllData:               make(map[string]Executor),
	}
}

// AddStartEvent 向模型中添加开始事件
func (model *Model) AddStartEvent(ExecutionId string, startEvent StartEvent) {
	model.StartEvents[ExecutionId] = startEvent
	model.AllData[ExecutionId] = startEvent
}

// AddTask 向模型中添加任务
func (model *Model) AddTask(ExecutionId string, task Task) {
	model.Tasks[ExecutionId] = task
	model.AllData[ExecutionId] = task
}

// AddParallelGateway 向模型中添加并行网关
func (model *Model) AddParallelGateway(ExecutionId string, gateway ParallelGateway) {
	model.ParallelGateways[ExecutionId] = gateway
	model.AllData[ExecutionId] = gateway
}

// AddExclusiveGateway 向模型中添加互斥网关
func (model *Model) AddExclusiveGateway(ExecutionId string, gateway ExclusiveGateway) {
	model.ExclusiveGateways[ExecutionId] = gateway
	model.AllData[ExecutionId] = gateway
}

// AddEndEvent 向模型中添加结束事件
func (model *Model) AddEndEvent(ExecutionId string, endEvent EndEvent) {
	model.EndEvents[ExecutionId] = endEvent
	model.AllData[ExecutionId] = endEvent
}

// AddSequenceFlow 向模型中添加序列流
func (model *Model) AddSequenceFlow(ExecutionId string, flow SequenceFlow) {
	model.SequenceFlows[ExecutionId] = flow
	model.AllData[ExecutionId] = flow
}

// Process 代表整个流程定义，用于从XML解析得到的数据结构，用来提供给go做xml解析的数组模型，后续要根据id来找到对应的节点，得转成map
type Process struct {
	Name string `xml:"name,attr"`

	// 存储所有的开始事件
	StartEvents []StartEvent `xml:"StartEvent"`

	// 存储所有的任务
	Tasks []Task `xml:"Task"`

	// 存储所有的并行网关
	ParallelGateways []ParallelGateway `xml:"ParallelGateway"`

	// 存储所有的互斥网关
	ExclusiveGateways []ExclusiveGateway `xml:"ExclusiveGateway"`

	// 存储所有的结束事件
	EndEvents []EndEvent `xml:"EndEvent"`

	// 存储所有的序列流
	SequenceFlows []SequenceFlow `xml:"SequenceFlow"`
}

var (
	modelInstance *map[string]*Model
	modelOnce     sync.Once
)

func initModelMap() *map[string]*Model {
	modelOnce.Do(func() {
		temp := make(map[string]*Model)
		modelInstance = &temp
	})
	return modelInstance
}

func GetModelMap() *map[string]*Model {
	if modelInstance == nil {
		panic("modelInstance is not initialized. Call initModelMap first.")
	}
	return modelInstance
}
