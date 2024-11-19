package components

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/go-sql-driver/mysql" // 假设使用 MySQL 数据库驱动
)

func Init(db *sql.DB, dbtype string) {
	InitializeServiceFactory(dbtype, db)

	log.Println("All services have been successfully initialized based on the database type.")
	// 建造一个单例的 model 容器，给流程模型存到缓存里 根据定义的名称
	initModelMap()
	// 存缓存的流程定义 只需要按照名称 不管版本 ,因为需求里，流程定义数据可能会随时更新的！
	// 需要通过改缓存中的节点信息 从而让剩下的未执行的节点被影响！
	// 因为只开放节点的部分信息修改 不开放结构修改 所以应该是安全的
}

type ServiceFactory interface {
	InitServiceInstance(db *sql.DB)
	GetRuntimeService() RuntimeService
	GetRepositoryService() RepositoryService
	GetNodeService() NodeService
	GetHistoryService() HistoryService
}

var (
	serviceFactoryInstance ServiceFactory
	serviceFactoryOnce     sync.Once
)

func InitializeServiceFactory(dbtype string, db *sql.DB) {
	serviceFactoryOnce.Do(func() {
		switch dbtype {
		case MYSQL_DBNAME:
			serviceFactoryInstance = &MySQLServiceFactory{}
		// 未来如果添加其他数据库类型
		// case "oracle":
		//     serviceFactoryInstance = &OracleServiceFactory{}
		default:
			panic(fmt.Sprintf("Unsupported database type: %s", dbtype))
		}

		serviceFactoryInstance.InitServiceInstance(db)
	})
}

func GetServiceFactory() ServiceFactory {
	if serviceFactoryInstance == nil {
		panic("ServiceFactory is not initialized. Call InitializeServiceFactory first.")
	}
	return serviceFactoryInstance
}
