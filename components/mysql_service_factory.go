package components

import (
	"database/sql"
)

type MySQLServiceFactory struct{}

func (f *MySQLServiceFactory) InitServiceInstance(db *sql.DB) {
	InitializeMySQLRuntimeService(db)
	InitializeMySQLRepositoryService(db)
	InitializeMySQLNodeService(db)
	InitializeMySQLHistoryService(db)
}

func (f *MySQLServiceFactory) GetRuntimeService() RuntimeService {
	return GetMySQLRuntimeService()
}

func (f *MySQLServiceFactory) GetRepositoryService() RepositoryService {
	return GetMySQLRepositoryService()
}

func (f *MySQLServiceFactory) GetNodeService() NodeService {
	return GetMySQLNodeService()
}

func (f *MySQLServiceFactory) GetHistoryService() HistoryService {
	return GetMySQLHistoryService()
}
