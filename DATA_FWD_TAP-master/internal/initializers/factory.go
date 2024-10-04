package initializers

import (
	"DATA_FWD_TAP/internal/database"
	"DATA_FWD_TAP/internal/models"
	"DATA_FWD_TAP/util"
	"DATA_FWD_TAP/util/MessageQueue"
	MessageStat "DATA_FWD_TAP/util/MessageStats"
	"DATA_FWD_TAP/util/OrderConversion"
	typeconversionutil "DATA_FWD_TAP/util/TypeConversionUtil"
	"log"
	"strconv"
)

func NewMainContainer(serviceName string, args []string, mTypeRead *int, mTypeWrite *int) *MainContainer {

	log.Printf("[%s] Program %s starts", serviceName, args[0])

	utilContainer := NewUtilContainer(serviceName, args, mTypeRead, mTypeWrite)
	clientContainer := NewClientContainer()
	logOnContainer := NewLogOnContainer()
	logOffContainer := NewLogOffContainer()
	clientGlobalValueContainer := NewClientGlobalValueContainer(args)
	logOnGlobalValueContainer := NewLogOnGlobalValueContainer(args)
	logOffGlobalValueContainer := NewLogOffGlobalValueContainer(args)
	ESRGlobalValueContainer := NewESRGlobalValueContainer()
	ESRContainer := NewESRContainer()

	return &MainContainer{
		UtilContainer:              utilContainer,
		ClientContainer:            clientContainer,
		LogOnContainer:             logOnContainer,
		LogOffContainer:            logOffContainer,
		ClientGlobalValueContainer: clientGlobalValueContainer,
		LogOnGlobalValueContainer:  logOnGlobalValueContainer,
		LogOffGlobalValueContainer: logOffGlobalValueContainer,
		ESRContainer:               ESRContainer,
		ESRGlobalValueContainer:    ESRGlobalValueContainer,
	}
}

func NewUtilContainer(serviceName string, args []string, mTypeRead *int, mTypeWrite *int) *UtilContainer {

	log.Printf("[%s] Program %s starts", serviceName, args[0])

	environmentManager := util.NewEnvironmentManager(serviceName, "/mnt/c/Users/deysh/OneDrive/Desktop/IRRA_FWD_Integrate/DATA_FWD_TAP-master/internal/config/env_config.ini")

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

	TCUM := &typeconversionutil.TypeConversionUtilManager{
		LoggerManager: loggerManager,
	}

	//this value also we are reding from configuration file "I have set the temp value in the configuration file for now"
	maxPackValStr := environmentManager.GetProcessSpaceValue("PackingLimit", "PACK_VAL")
	var maxPackVal int
	if maxPackValStr == "" {
		loggerManager.LogError(serviceName, " [Factory] [Error: 'PACK_VAL' not found in the configuration under 'PackingLimit'")
	} else {
		maxPackVal, err := strconv.Atoi(maxPackValStr)
		if err != nil {
			loggerManager.LogError(serviceName, " [Factory] [Error: Failed to convert 'PACK_VAL' '%s' to integer: %v", maxPackValStr, err)
			maxPackVal = 0
		} else {
			loggerManager.LogInfo(serviceName, " [Factory] Fetched and converted 'PACK_VAL' from configuration: %d", maxPackVal)
		}

	}
	loggerManager.LogInfo(serviceName, " [Factory]****************** TESTING 1 **************************")

	return &UtilContainer{
		ServiceName:               serviceName,
		OrderConversionManager:    &OrderConversion.OrderConversionManager{},
		EnvironmentManager:        environmentManager,
		MaxPackVal:                maxPackVal,
		ConfigManager:             configManager,
		LoggerManager:             loggerManager,
		TypeConversionUtilManager: TCUM,
		TransactionManager:        util.NewTransactionManager(serviceName, configManager, loggerManager),
		PasswordUtilManager:       &util.PasswordUtilManger{LM: loggerManager},
		MessageQueueManager: &MessageQueue.MessageQueueManager{
			ServiceName:   serviceName,
			LoggerManager: loggerManager,
			MSM:           &MessageStat.MessageStatManager{},
		},
		DB:         DB,
		MTypeRead:  mTypeRead,
		MTypeWrite: mTypeWrite,
	}

}

func NewClientContainer() *ClientContainer {
	return &ClientContainer{
		Xchngbook:    &models.Vw_xchngbook{},
		Orderbook:    &models.Vw_orderbook{},
		Contract:     &models.Vw_contract{},
		NseContract:  &models.Vw_nse_cntrct{},
		PipeMaster:   &models.St_opm_pipe_mstr{},
		OeReqres:     &models.St_oe_reqres{},
		ExchMsg:      &models.St_exch_msg{},
		NetHdr:       &models.St_net_hdr{},
		QPacket:      &models.St_req_q_data{},
		IntHeader:    &models.St_int_header{},
		ContractDesc: &models.St_contract_desc{},
		OrderFlag:    &models.St_order_flags{},
	}
}

func NewClientGlobalValueContainer(args []string) *ClientGlobalValueContainer {

	return &ClientGlobalValueContainer{
		Args: args,
	}
}

func NewLogOnContainer() *LogOnContainer {
	return &LogOnContainer{
		StSignOnReq:               &models.St_sign_on_req{},
		StReqQData:                &models.St_req_q_data_Log_On{},
		StExchMsgLogOn:            &models.St_exch_msg_Log_On{},
		IntHeader:                 &models.St_int_header{},
		StNetHdr:                  &models.St_net_hdr{},
		StBrokerEligibilityPerMkt: &models.St_broker_eligibility_per_mkt{},
	}
}

func NewLogOnGlobalValueContainer(args []string) *LogOnGlobalValueContainer {
	return &LogOnGlobalValueContainer{
		Args: args,
	}
}

func NewLogOffContainer() *LogOffContainer {
	return &LogOffContainer{
		IntHeader:       &models.St_int_header{},
		StNetHdr:        &models.St_net_hdr{},
		StExchMsgLogOff: &models.St_exch_msg_Log_Off{},
		StReqQData:      &models.St_req_q_data_Log_Off{},
	}
}

func NewLogOffGlobalValueContainer(args []string) *LogOffGlobalValueContainer {
	return &LogOffGlobalValueContainer{
		Args: args,
	}
}

func NewESRContainer() *ESRContainer {
	return &ESRContainer{
		Req_q_data:             &models.St_req_q_data{},
		St_exch_msg:            &models.St_exch_msg{},
		St_net_hdr:             &models.St_net_hdr{},
		St_oe_reqres:           &models.St_oe_reqres{},
		St_sign_on_req:         &models.St_sign_on_req{},
		St_exch_msg_Log_on:     &models.St_exch_msg_Log_On{},
		St_exch_msg_Log_on_use: &models.St_exch_msg_Log_On_use{},
		St_exch_msg_resp:       &models.St_exch_msg_resp{},
	}
}

func NewESRGlobalValueContainer() *ESRGlobalValueContainer {
	return &ESRGlobalValueContainer{}
}
