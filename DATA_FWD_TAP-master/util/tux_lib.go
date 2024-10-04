package util

import (
	"DATA_FWD_TAP/common"
	"fmt"

	"gorm.io/gorm"
)

/**********************************************************************************/
/*                                                                                */
/*  Constants:                                                                    */
/*    - TRAN_TIMEOUT: Placeholder for transaction timeout.                        */
/*    - REMOTE_TRNSCTN: Constant for remote transaction type.                     */
/*    - LOCAL_TRNSCTN: Constant for local transaction type.                       */
/*                                                                                */
/*  Structs:                                                                      */
/*    - TransactionManager: Manages database transactions using GORM.             */
/*                                                                                */
/*  Functions:                                                                    */
/*    - NewTransactionManager: Creates and returns a new TransactionManager      */
/*      instance with the specified service name and database accessor.           */
/*                                                                                */
/*    - FnBeginTran: Begins a new transaction. Determines if the transaction is   */
/*      local or remote based on the database connection status.                   */
/*                                                                                */
/*    - PingDatabase: Checks if the database connection is alive by pinging it.   */
/*                                                                                */
/*    - FnCommitTran: Commits the transaction if it is local, or skips commit     */
/*      for remote transactions.                                                  */
/*                                                                                */
/*    - FnAbortTran: Rolls back the transaction if it is local, or skips rollback */
/*      for remote transactions.                                                  */
/*                                                                                */
/**********************************************************************************/

const (
	TRAN_TIMEOUT   = 300 // placeholder for transaction timeout . i have to use this .
	REMOTE_TRNSCTN = 2
	LOCAL_TRNSCTN  = 1
)

type TransactionManager struct {
	ServiceName   string
	TranType      int
	Tran          *gorm.DB
	DbAccessor    common.DBAccessor
	LoggerManager *LoggerManager
}

func NewTransactionManager(ServiceName string, dbAccessor common.DBAccessor, LM *LoggerManager) *TransactionManager {
	return &TransactionManager{
		ServiceName:   ServiceName,
		DbAccessor:    dbAccessor,
		LoggerManager: LM,
	}
}

func (tm *TransactionManager) FnBeginTran() int {

	// Check if a transaction is already active (GORM does not provide tpgetlev equivalent, manage this externally if needed)

	db := tm.DbAccessor.GetDB() // Get the database instance on this instance we will check if there is any active transaction.
	fmt.Println(tm.DbAccessor)
	fmt.Println(db)
	err := tm.PingDatabase(db) // Use PingDatabase to check if there is any active transaction.

	if err == nil {
		tm.Tran = db.Begin() // Begin the transaction here. This line represents something unique. When we call db.Begin(), it returns a new *gorm.DB instance, and we are supposed to use this instance throughout the transaction. This is because we have to maintain the 'isolation'.
		if tm.Tran.Error != nil {
			errMsg := fmt.Sprintf("[FnBeginTran] [Error: tpbegin error: %v]", tm.Tran.Error)
			tm.LoggerManager.LogError(tm.ServiceName, errMsg)
			return -1
		}
		tm.TranType = LOCAL_TRNSCTN
		tm.LoggerManager.LogInfo(tm.ServiceName, "[FnBeginTran] Transaction started (LOCAL)")

	} else {
		tm.TranType = REMOTE_TRNSCTN
		tm.LoggerManager.LogInfo(tm.ServiceName, "[FnBeginTran] Transaction is already in (REMOTE)")

	}
	return 0
}

func (tm *TransactionManager) PingDatabase(db *gorm.DB) error {
	sqlDB, err := db.DB() // Get the underlying *sql.DB
	if err != nil {
		errMsg := fmt.Sprintf("[PingDatabase] [Error: Failed to get underlying *sql.DB: %v]", err)
		tm.LoggerManager.LogError(tm.ServiceName, errMsg)
		return err
	}
	if err := sqlDB.Ping(); err != nil {
		errMsg := fmt.Sprintf("[PingDatabase] [Error: Failed to ping database: %v]", err)
		tm.LoggerManager.LogError(tm.ServiceName, errMsg)
		return err
	}
	tm.LoggerManager.LogInfo(tm.ServiceName, "[PingDatabase] Database ping successful")
	return nil
}

func (tm *TransactionManager) FnCommitTran() int {
	switch tm.TranType {
	case LOCAL_TRNSCTN:
		if err := tm.Tran.Commit().Error; err != nil {
			errMsg := fmt.Sprintf("[FnCommitTran] [Error: tpcommit error: %v]", err)
			tm.LoggerManager.LogError(tm.ServiceName, errMsg)
			tm.Tran.Rollback() // Ensure rollback on commit error
			return -1
		}
		tm.LoggerManager.LogInfo(tm.ServiceName, "[FnCommitTran] Transaction committed")
	case REMOTE_TRNSCTN:
		tm.LoggerManager.LogInfo(tm.ServiceName, "[FnCommitTran] Skipping commit for remote transaction")

	default:
		errMsg := fmt.Sprintf("[FnCommitTran] [Error: Invalid Transaction Number |%d|]", tm.TranType)
		tm.LoggerManager.LogError(tm.ServiceName, errMsg)
		return -1
	}
	return 0
}

func (tm *TransactionManager) FnAbortTran() int {
	switch tm.TranType {
	case LOCAL_TRNSCTN:
		if err := tm.Tran.Rollback().Error; err != nil {
			errMsg := fmt.Sprintf("[FnAbortTran] [Error: tpabort error: %v]", err)
			tm.LoggerManager.LogError(tm.ServiceName, errMsg)
			return -1
		}
		tm.LoggerManager.LogInfo(tm.ServiceName, "[FnAbortTran] Transaction aborted")
	case REMOTE_TRNSCTN:
		tm.LoggerManager.LogInfo(tm.ServiceName, "[FnAbortTran] Skipping abort for remote transaction")

	default:
		errMsg := fmt.Sprintf("[FnAbortTran] [Error: Invalid Transaction Number |%d|]", tm.TranType)
		tm.LoggerManager.LogError(tm.ServiceName, errMsg)
		return -1
	}
	return 0
}
