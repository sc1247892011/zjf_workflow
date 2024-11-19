package components

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLRepositoryService struct {
	DB *sql.DB
}

var mysqlRepositoryServiceInstance *MySQLRepositoryService
var mysqlRepositoryServiceOnce sync.Once

// InitializeMySQLRepositoryService 初始化单例实例
func InitializeMySQLRepositoryService(db *sql.DB) {
	mysqlRepositoryServiceOnce.Do(func() {
		mysqlRepositoryServiceInstance = &MySQLRepositoryService{DB: db}
	})
}

// GetMySQLRepositoryService 获取单例实例
func GetMySQLRepositoryService() *MySQLRepositoryService {
	if mysqlRepositoryServiceInstance == nil {
		panic("MySQLRepositoryService is not initialized. Call InitializeMySQLRepositoryService first.")
	}
	return mysqlRepositoryServiceInstance
}

func (service *MySQLRepositoryService) GetTransaction() (*sql.Tx, error) {
	return service.DB.Begin()
}

// SaveProcessDefinition 插入新的流程定义到数据库中
func (service *MySQLRepositoryService) SaveProcessDefinition(tx *sql.Tx, pd *ProcessDefinition) (int, error) {
	query := `
     INSERT INTO process_definition (process_definition_name, version, xml_content, created_at, created_by, status, description)
SELECT 
    ?,
    IFNULL(MAX(version) + 1, 1),
    ?, ?, ?, ?, ?
FROM 
    (SELECT * FROM process_definition WHERE process_definition_name = ?) AS pd;

    `
	result, err := tx.Exec(query, pd.ProcessDefinitionName, pd.XMLContent, pd.CreatedAt, pd.CreatedBy, pd.Status, pd.Description, pd.ProcessDefinitionName)
	if err != nil {
		return 0, fmt.Errorf("failed to save process definition: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve last insert id: %v", err)
	}

	return int(id), nil
}

// GetProcessDefinitionById 根据Id获取流程定义
func (service *MySQLRepositoryService) GetProcessDefinitionById(id int) (*ProcessDefinition, error) {
	query := `SELECT id, process_definition_name, version, xml_content, created_at, created_by, status, description FROM process_definition WHERE id = ?`
	pd := &ProcessDefinition{}
	err := service.DB.QueryRow(query, id).Scan(&pd.Id, &pd.ProcessDefinitionName, &pd.Version, &pd.XMLContent, &pd.CreatedAt, &pd.CreatedBy, &pd.Status, &pd.Description)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get process definition by Id: %v", err)
	}
	return pd, nil
}

// GetProcessDefinitionByNameAndVersion 根据流程名称和版本号获取流程定义
func (service *MySQLRepositoryService) GetProcessDefinitionByNameAndVersion(name string, version int) (*ProcessDefinition, error) {
	query := `SELECT id, process_definition_name, version, xml_content, created_at, created_by, status, description FROM process_definition WHERE name = ? AND version = ?`
	pd := &ProcessDefinition{}
	err := service.DB.QueryRow(query, name, version).Scan(&pd.Id, &pd.ProcessDefinitionName, &pd.Version, &pd.XMLContent, &pd.CreatedAt, &pd.CreatedBy, &pd.Status, &pd.Description)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get process definition by process_definition_name and version: %v", err)
	}
	return pd, nil
}

// GetProcessDefinitionByNameAndVersion 根据流程名称获取最新流程定义
func (service *MySQLRepositoryService) GetLatestProcessDefinitionByName(name string) (*ProcessDefinition, error) {
	query := `SELECT id, process_definition_name, version, xml_content, created_at, created_by, status, description FROM process_definition pd WHERE pd.version = (SELECT MAX(version)
FROM process_definition WHERE process_definition_name = pd.process_definition_name) AND pd.process_definition_name = ? `
	pd := &ProcessDefinition{}
	err := service.DB.QueryRow(query, name).Scan(&pd.Id, &pd.ProcessDefinitionName, &pd.Version, &pd.XMLContent, &pd.CreatedAt, &pd.CreatedBy, &pd.Status, &pd.Description)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get process definition by name and version: %v", err)
	}
	return pd, nil
}

// UpdateProcessDefinition 更新流程定义 只允许改数据 不允许改结构 ，名称和版本都不变，这个限制得在前端做
func (service *MySQLRepositoryService) UpdateProcessDefinition(tx *sql.Tx, pd *ProcessDefinition) error {
	query := `
        UPDATE process_definition
        SET  xml_content = ?, created_by = ?
        WHERE id = ?
    `
	_, err := tx.Exec(query, pd.XMLContent, pd.CreatedBy, pd.Id)
	if err != nil {
		return fmt.Errorf("failed to update process definition: %v", err)
	}
	return nil
}

// DeleteProcessDefinition 根据Id删除流程定义
func (service *MySQLRepositoryService) DeleteProcessDefinition(tx *sql.Tx, id int) error {
	query := `DELETE FROM process_definition WHERE id = ?`
	_, err := tx.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete process definition: %v", err)
	}
	return nil
}
