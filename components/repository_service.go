package components

import (
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// ProcessDefinition 定义了流程定义的数据结构
type ProcessDefinition struct {
	Id                    int
	ProcessDefinitionName string
	Version               int
	XMLContent            []byte
	CreatedAt             time.Time
	CreatedBy             string
	Status                string
	Description           string
}

// RepositoryService 提供了操作流程定义表的接口
type RepositoryService interface {
	SaveProcessDefinition(pd *ProcessDefinition) (int, error)
	GetProcessDefinitionById(id int) (*ProcessDefinition, error)
	GetProcessDefinitionByNameAndVersion(name string, version int) (*ProcessDefinition, error)
	GetLatestProcessDefinitionByName(name string) (*ProcessDefinition, error)
	UpdateProcessDefinition(pd *ProcessDefinition) error
	DeleteProcessDefinition(id int) error
}
