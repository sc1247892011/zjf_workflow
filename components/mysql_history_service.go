package components

import (
	"database/sql"
	"fmt"
	"sync"
)

type MySQLHistoryService struct {
	DB *sql.DB
}

var mysqlHistoryServiceInstance *MySQLHistoryService
var mysqlHistoryServiceOnce sync.Once

// InitializeMySQLHistoryService 初始化单例实例
func InitializeMySQLHistoryService(db *sql.DB) {
	mysqlHistoryServiceOnce.Do(func() {
		mysqlHistoryServiceInstance = &MySQLHistoryService{DB: db}
	})
}

// GetMySQLHistoryService 获取单例实例
func GetMySQLHistoryService() *MySQLHistoryService {
	if mysqlHistoryServiceInstance == nil {
		panic("MySQLHistoryService is not initialized. Call InitializeMySQLHistoryService first.")
	}
	return mysqlHistoryServiceInstance
}

func (service *MySQLHistoryService) CopyNodeInstanceById(nodeId int) error {
	query := `
		INSERT INTO historic_node_instance (
		    id
			process_instance_id,
			node_name,
			execution_id,
			output_data,
			previous_execution_id,
			assignee,
			start_time,
			end_time
		)
		SELECT 
		    id,
			process_instance_id,
			node_name,
			execution_id,
			output_data,
			previous_execution_id,
			assignee,
			start_time,
			end_time
		FROM node_instance
		WHERE id = ?
	`

	_, err := service.DB.Exec(query, nodeId)
	if err != nil {
		return fmt.Errorf("failed to copy node instance to historic_node_instance: %v", err)
	}

	return nil
}
