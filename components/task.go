package components

import (
	"encoding/json"
	"log"
)

// Task 代表 BPMN 中的审批节点
type Task struct {
	ExecutionId string `xml:"executionId,attr"`
	// 绑定 id 属性
	AssigneeType string   `xml:"assigneeType,attr"` //负责人指定方式
	AssigneeKey  string   `xml:"assigneeKey,attr"`  //负责人标识
	Name         string   `xml:"name,attr"`
	Incoming     []string `xml:"Incoming"` // 或者 `[]string`
	Outgoing     []string `xml:"Outgoing"` // 或者 `[]string`
	FormData     string   `xml:"FormData"` // 绑定 <FormData> 子元素
	X            string   `xml:"x,attr"`
	Y            string   `xml:"y,attr"`
	H            string   `xml:"h,attr"`
	W            string   `xml:"w,attr"`
	Listener     string   `xml:"Listener"` // 任务监听 执行完毕之后的后续逻辑 方法名称可以用逗号隔开 传递多段逻辑
}

// Execute 是 Task 节点的执行方法
func (task Task) Execute(ctx *WorkflowContext) {
	//初始化数据库状态
	nodeService := GetServiceFactory().GetNodeService()
	//ctx是从上一个节点传递进来的，所以它的CurrentExecutionId就是上级节点的Id,因为task节点可能有多个输入 所以得从ctx拿上级节点,然后上个节点的输出数据也是从ctx拿
	//因为所有的这些task节点 都是缓存里的 最新的实时数据 所以直接用就行了
	assigneePeopleName := GetAssigneePeopleName(task.AssigneeType, task.AssigneeKey)
	tx := ctx.Tx
	if tx == nil {
		log.Println("Failed to get transaction from ctx ")
		return
	}
	_, initerr := nodeService.InitNodeInstance(tx, ctx.ProcessInstanceId, ctx.ProcessDefinitionName, task.Name, task.ExecutionId, ctx.CurrentExecutionId, assigneePeopleName)
	if initerr != nil {
		log.Println("Failed to InitNodeInstance from database: ", initerr)
		tx.Rollback()
		return
	}

	// 此时流程停止 且上级节点不是并行网关 才可以提交事务，否则此时提交会导致并行网关无法将全部子任务分发
	frontNodeIsParallelGateway := false
	target := ctx.CurrentExecutionId

	for _, gateway := range ctx.Model.ParallelGateways {
		if gateway.ExecutionId == target {
			frontNodeIsParallelGateway = true // 找到匹配的字符串
		}
	}

	if !frontNodeIsParallelGateway {
		tx.Commit()
	}
}

// 修改审批节点状态 把当前节点表单提交的数据放ctx.Data再传递下去
// 前端通过页面 调用接口 查询负责人需要审批的节点 去操作这个方法 表单可以直接从缓存拿
func (task Task) Complete(ctx *WorkflowContext) {
	//更新数据库的输出数据 因为可能存在循环 流程实例和结构id不足以判断唯一性 只有用自增id了
	//需要前端传递task的id 前端把该用户在中流程output为空的数据查询出来
	frontData, err := ParseJSON(ctx.Data)

	if err != nil {
		log.Println("Failed to ParseJSON attribute:", ctx.Data)
		return
	}
	idFloat64, _ := frontData["taskid"].(float64)
	id := int(idFloat64)

	dataBytes, _ := json.Marshal(frontData["outputData"])
	data := string(dataBytes)

	nodeService := GetServiceFactory().GetNodeService()
	tx := ctx.Tx
	if tx == nil {
		log.Println("Failed to get transaction from ctx ")
		return
	}

	updateerr := nodeService.UpdateNodeInstanceOutput(tx, id, data)
	if updateerr != nil {
		log.Println("Failed to update task from database: ", updateerr)
		tx.Rollback()
		return
	}

	ctx.CurrentExecutionId = task.ExecutionId
	//迁徙数据到历史库
	historyService := GetServiceFactory().GetHistoryService()
	copyerr := historyService.CopyNodeInstanceById(tx, id)
	if copyerr != nil {
		log.Println("Failed to copynode to history: ", copyerr)
		tx.Rollback()
		return
	}

	//执行监听
	RunListener(task.Listener, ctx)

	//执行下一个 或者多个 序列流
	for _, value := range task.Outgoing {
		model := *(ctx.Model)
		model.SequenceFlows[value].Execute(ctx)
	}

}

func GetAssigneePeopleName(AssigneeType string, AssigneeKey string) string {

	if AssigneeType == ASSIGNEETYPE_NAME {
		return AssigneeKey
	} else if AssigneeType == ASSIGNEETYPE_COMPANY {
		return parentCompany(AssigneeKey)
	}
	return "sc"
}

func parentCompany(param1 string) string {
	return param1
}
