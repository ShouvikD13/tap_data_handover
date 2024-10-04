package database

import (
	"DATA_FWD_TAP/common"
	"DATA_FWD_TAP/util"
	"strconv"

	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

/**********************************************************************************/
/*                                                                                 */
/*  Description       : This package defines a configuration structure for         */
/*                      connecting to a PostgreSQL database. The `PostgreSQLConfig`*/
/*                      struct holds essential details required for establishing   */
/*                      a connection, including username, password, database name, */
/*                      host, port, and SSL mode.                                  */
/*                                                                                 */
/*                      The `GetDatabaseConnection` function formats these values  */
/*                      into a Data Source Name (DSN) string, which is used to     */
/*                      open a connection to the PostgreSQL database using GORM.   */
/*                                                                                 */
/*  Functions         :                                                            */
/*                        - LoadPostgreSQLConfig: Reads the configuration from     */
/*                          an INI file and returns a `PostgreSQLConfig` instance. */
/*                                                                                 */
/*                        - GetDatabaseConnection: Constructs a DSN string from    */
/*                          the provided `PostgreSQLConfig` struct, establishes a  */
/*                          connection to the PostgreSQL database using GORM, and  */
/*                          returns an int status code.                            */
/*                                                                                 */
/*                        - GetDB: Returns the currently established database      */
/*                          connection as a `*gorm.DB` instance.                   */
/*                                                                                 */
/**********************************************************************************/

// ConfigManager implements the DBConfigLoader, DBConnector, and DBAccessor interfaces.
type ConfigManager struct {
	Database struct {
		Db *gorm.DB // GORM database connection instance
	}
	environmentManager *util.EnvironmentManager
	resultTmp          int
	cfg                *common.PostgreSQLConfig
	serviceName        string
	LoggerManager      *util.LoggerManager
}

func NewConfigManager(serviceName string, envManager *util.EnvironmentManager, LM *util.LoggerManager) *ConfigManager {
	return &ConfigManager{
		serviceName:        serviceName,
		environmentManager: envManager,
		LoggerManager:      LM,
	}
}

// LoadPostgreSQLConfig reads the configuration from an INI file and returns a PostgreSQLConfig instance.
// It takes the file name of the INI file as an argument.
func (cm *ConfigManager) LoadPostgreSQLConfig() int {
	portStr := cm.environmentManager.GetProcessSpaceValue("database", "port")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		cm.LoggerManager.LogError(cm.serviceName, "[LoadPostgreSQLConfig] Failed to convert port value '%s' to integer: %v", portStr, err)
		return -1
	}

	cm.cfg = &common.PostgreSQLConfig{
		Host:     cm.environmentManager.GetProcessSpaceValue("database", "host"),
		Port:     port,
		Username: cm.environmentManager.GetProcessSpaceValue("database", "username"),
		Password: cm.environmentManager.GetProcessSpaceValue("database", "password"),
		DBName:   cm.environmentManager.GetProcessSpaceValue("database", "dbname"),
		SSLMode:  cm.environmentManager.GetProcessSpaceValue("database", "sslmode"),
	}

	cm.LoggerManager.LogInfo(cm.serviceName, "[LoadPostgreSQLConfig] PostgreSQL configuration loaded successfully")

	return 0
}

// GetDatabaseConnection constructs a DSN string from the provided PostgreSQLConfig struct,
// establishes a connection to the PostgreSQL database using GORM, and returns an int status code.
// It returns 0 on success and -1 on failure.
func (cm *ConfigManager) GetDatabaseConnection() int {

	cm.LoggerManager.LogInfo(cm.serviceName, "[GetDatabaseConnection] Setting up the database connection")
	dsn := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=%s",
		cm.cfg.Username, cm.cfg.Password, cm.cfg.DBName, cm.cfg.Host, cm.cfg.Port, cm.cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		cm.LoggerManager.LogError(cm.serviceName, "[GetDatabaseConnection] Failed to connect to database: %v", err)
		return -1
	}
	cm.Database.Db = db
	cm.LoggerManager.LogInfo(cm.serviceName, "[GetDatabaseConnection] Database connection established successfully")
	return 0
}

// GetDB returns the currently established database connection as a *gorm.DB instance.

func (cm *ConfigManager) GetDB() *gorm.DB {
	cm.LoggerManager.LogInfo(cm.serviceName, "[GetDB] Returning the Database instance")
	return cm.Database.Db // Return the GORM database connection instance
}
