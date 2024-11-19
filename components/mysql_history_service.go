package components

import (
	"database/sql"
	"fmt"
	"sync"
	"time"
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

func (service *MySQLHistoryService) GetTransaction() (*sql.Tx, error) {
	return service.DB.Begin()
}

func (service *MySQLHistoryService) CopyNodeInstanceById(tx *sql.Tx, nodeId int) error {
	query := `
		INSERT INTO historic_node_instance (
		    id,
			process_instance_id,
			process_definition_name,
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
			process_definition_name,
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

	_, err := tx.Exec(query, nodeId)
	if err != nil {
		return fmt.Errorf("failed to copy node instance to historic_node_instance: %v", err)
	}

	return nil
}

func (service *MySQLHistoryService) CopyNodeInstance(tx *sql.Tx, nodeId int, processInstanceId int, processDefinitionName string, nodeName string, executionId string,
	previousExecutionId string, assignee string) (int, error) {
	query := `
        INSERT INTO historic_node_instance (id, process_instance_id, process_definition_name, node_name, execution_id, previous_execution_id, assignee, start_time)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	startTime := time.Now()
	result, err := tx.Exec(query, nodeId, processInstanceId, processDefinitionName, nodeName, executionId, previousExecutionId, assignee, startTime)
	if err != nil {
		return 0, fmt.Errorf("failed to copy node instance to historic: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve last insert id: %v", err)
	}

	return int(id), nil
}

func (service *MySQLHistoryService) GetProcessCompleteTask(ProcessInstanceId int) ([]map[string]interface{}, error) {
	// 创建一个空的map数组用于存储结果
	var results []map[string]interface{}

	// 构建查询语句
	query := `
			SELECT id, process_instance_id,process_definition_name, node_name, execution_id, output_data, previous_execution_id, assignee, start_time, end_time
			FROM node_instance
			WHERE process_instance_id = ?
		`

	// 执行查询
	rows, err := service.DB.Query(query, ProcessInstanceId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 遍历查询结果
	for rows.Next() {
		var (
			id                    int
			processInstanceID     int
			processDefinitionName string
			nodeName              string
			executionID           string
			outputData            sql.NullString
			previousExecutionID   sql.NullString
			assignee              sql.NullString
			startTime             sql.NullString
			endTime               sql.NullString
		)

		// 扫描每一行数据
		err := rows.Scan(&id, &processInstanceID, &processDefinitionName, &nodeName, &executionID, &outputData, &previousExecutionID, &assignee, &startTime, &endTime)
		if err != nil {
			return nil, err
		}

		// 将数据放入map
		result := map[string]interface{}{
			"id":                      id,
			"process_instance_id":     processInstanceID,
			"process_definition_name": processDefinitionName,
			"node_name":               nodeName,
			"execution_id":            executionID,
			"output_data":             outputData,
			"previous_execution_id":   previousExecutionID,
			"assignee":                assignee,
			"start_time":              startTime,
			"end_time":                endTime,
		}

		// 将map放入结果数组
		results = append(results, result)
	}

	// 检查是否有错误
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
