package main

import (
	//"DATA_FWD_TAP/internal/app/esr"
	"DATA_FWD_TAP/internal/app/pack_clnt"
	"DATA_FWD_TAP/internal/database"
	"DATA_FWD_TAP/util"
	"DATA_FWD_TAP/util/MessageQueue"
	typeconversionutil "DATA_FWD_TAP/util/TypeConversionUtil"
	"log"
)

func main() {
	// args := os.Args
	args := []string{"main", "arg1", "arg2", "arg3", "P1", "arg5", "arg6", "arg7"}
	serviceName := args[0]
	var mtype int
	mtype = 1

	// var resultTmp int

	log.Printf("[%s] Program %s starts", serviceName, args[0])

	environmentManager := util.NewEnvironmentManager(serviceName, "/mnt/c/Users/deysh/OneDrive/Desktop/IRRA_FWD_Integrate/tap_data_fwd-master/internal/config/env_config.ini")

	environmentManager.LoadIniFile()

	//Logger file

	loggerManager := &util.LoggerManager{}
	loggerManager.InitLogger(environmentManager)

	// Get the logger instance
	logger := loggerManager.GetLogger()

	loggerManager.LogInfo(serviceName, "Program %s starts", args[0])

	configManager := database.NewConfigManager(serviceName, environmentManager, loggerManager)

	if configManager.LoadPostgreSQLConfig() != 0 {
		loggerManager.LogError(serviceName, "Failed to load PostgreSQL configuration")
		logger.Fatalf("[%s] Failed to load PostgreSQL configuration", serviceName)
	}

	if configManager.GetDatabaseConnection() != 0 {
		loggerManager.LogError(serviceName, "Failed to connect to the database")
		logger.Fatalf("[%s] Failed to connect to database", serviceName)
	}

	DB := configManager.GetDB()
	if DB == nil {
		loggerManager.LogError(serviceName, "Database connection is nil. Failed to get the database instance.")
		logger.Fatalf("[%s] Database connection is nil.", serviceName)
	}

	//=======================================================================================================

	messageQueueManager := &MessageQueue.MessageQueueManager{
		ServiceName:   serviceName,
		LoggerManager: loggerManager,
		// Initialize other fields as necessary
	}

	mq, errCode := messageQueueManager.CreateQueue(33) // Use appropriate key
	if errCode != 0 {
		logger.Fatalf("[%s] Failed to create message queue", serviceName)
	}
	messageQueueManager.MQ = mq
	loggerManager.LogInfo(serviceName, "[inside Main] MQ <----------------------: %v", messageQueueManager.MQ)
	//===================================================================================================
	//
	TCUM := &typeconversionutil.TypeConversionUtilManager{
		LoggerManager: loggerManager,
	}
	//var wg sync.WaitGroup

	// Testing "cln_pack_clnt.go"

	VarClnPack := &pack_clnt.ClnPackClntManager{
		Enviroment_manager:    environmentManager,
		Config_manager:        configManager,
		LoggerManager:         loggerManager,
		TCUM:                  TCUM,
		Message_queue_manager: messageQueueManager,
		Mtype:                 &mtype,
	}
	/* 	wg.Add(1)
	   	go func() {
	   		defer wg.Done() // Signal that this goroutine is done
	   		resultTmp := VarClnPack.Fn_bat_init(args[1:], DB)

	   		if resultTmp != 0 {
	   			loggerManager.LogInfo(serviceName, " Fn_bat_init failed with result code: %d", resultTmp)
	   			logger.Fatalf("[Fatal: Shutting down due to error in %s: result code %d", serviceName, resultTmp)
	   		}
	   	}() */

	resultTmp := VarClnPack.Fn_bat_init(args[1:], DB)

	if resultTmp != 0 {
		loggerManager.LogInfo(serviceName, " Fn_bat_init failed with result code: %d", resultTmp)
		logger.Fatalf("[Fatal: Shutting down due to error in %s: result code %d", serviceName, resultTmp)
	}

	//=======================================================================================================

	//=======================================================================================================
	/* // Test ESRManger

	log.Println("Starting ESR Service")

	VarESR := &esr.ESRManger{
		ServiceName:           serviceName,
		ENVM:                  environmentManager,
		LoggerManager:         loggerManager,
		Message_queue_manager: messageQueueManager,
		TCUM:                  TCUM,
		Mtype:                 &mtype,
		Db: DB
	}

	wg.Add(1)
	go func() {
		defer wg.Done() // Signal that this goroutine is done
		// Start the ESR client
		VarESR.ClnEsrClnt()
		log.Println("ESR Service Ended Successfully")
	}()
	*/
	// Wait for both goroutines to finish
	//wg.Wait()
	//=======================================================================================================

	loggerManager.LogInfo(serviceName, " Main Ended Here...")

}
