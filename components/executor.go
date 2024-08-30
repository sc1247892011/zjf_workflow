package components

type Executor interface {
	Execute(ctx *WorkflowContext) // 修改接口以使用 WorkflowContext
}
