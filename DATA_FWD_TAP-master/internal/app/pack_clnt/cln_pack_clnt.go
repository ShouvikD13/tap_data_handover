package pack_clnt

import (
	"DATA_FWD_TAP/internal/app/packing_service"
	"DATA_FWD_TAP/internal/database"
	"DATA_FWD_TAP/internal/models"
	"DATA_FWD_TAP/util"
	"DATA_FWD_TAP/util/MessageQueue"
	"DATA_FWD_TAP/util/OrderConversion"
	typeconversionutil "DATA_FWD_TAP/util/TypeConversionUtil"
	"fmt"

	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ClnPackClntManager struct {
	ServiceName              string                // Name of the service being managed
	Xchngbook                *models.Vw_xchngbook  // Pointer to the Vw_Xchngbook structure
	Orderbook                *models.Vw_orderbook  // Pointer to the Vw_orderbook structure
	Contract                 *models.Vw_contract   // we are updating it's value from orderbook
	Nse_contract             *models.Vw_nse_cntrct // we are updating it in 'fn_get_ext_cnt'
	Pipe_mstr                *models.St_opm_pipe_mstr
	Oe_reqres                *models.St_oe_reqres
	Exch_msg                 *models.St_exch_msg
	Net_hdr                  *models.St_net_hdr
	Q_packet                 *models.St_req_q_data
	Int_header               *models.St_int_header
	Contract_desc            *models.St_contract_desc
	Order_flag               *models.St_order_flags
	Order_conversion_manager *OrderConversion.OrderConversionManager
	CPanNo                   string // Pan number, initialized in the 'fnRefToOrd' method
	CLstActRef               string // Last activity reference, initialized in the 'fnRefToOrd' method
	CEspID                   string // ESP ID, initialized in the 'fnRefToOrd' method
	CAlgoID                  string // Algorithm ID, initialized in the 'fnRefToOrd' method
	CSourceFlg               string // Source flag, initialized in the 'fnRefToOrd' method
	CPrgmFlg                 string
	Enviroment_manager       *util.EnvironmentManager
	Message_queue_manager    *MessageQueue.MessageQueueManager
	Transaction_manager      *util.TransactionManager
	Max_Pack_Val             int
	Config_manager           *database.ConfigManager
	LoggerManager            *util.LoggerManager
	TCUM                     *typeconversionutil.TypeConversionUtilManager
	MtypeWrite               *int
	Args                     []string
	Db                       *gorm.DB
	QueueId                  int
}

/***************************************************************************************
 * Fn_bat_init initializes the client package manager by setting up data models,
 * fetching configuration values, and preparing the environment for processing.
 * It ensures all prerequisites are met before the service starts.
 *
 * INPUT PARAMETERS:
 *    cpcm.Args[] - Command-line arguments, which must include at least 7 elements.
 *             The fourth argument (cpcm.Args[3]) specifies the pipe ID.
 *    Db     - Pointer to a GORM database connection object used for SQL queries.
 *
 * OUTPUT PARAMETERS:
 *    int    - Returns 0 for success, or -1 if an error occurs.
 *
 * FUNCTIONAL FLOW:
 * 1. Initializes required models and managers for the service.
 * 2. Validates command-line arguments to ensure all necessary parameters are present.
 * 3. Sets the pipe ID from the command-line argument and creates a message queue.
 * 4. Executes SQL queries to fetch and store necessary data such as exchange code,
 *    trader ID, branch ID, trade date, and broker ID.
 * 5. Reads configuration values from a file, including user type and message limit.
 * 6. Invokes the `CLN_PACK_CLNT` function to start the main service logic.
 * 7. Returns a status code indicating the success or failure of the initialization.
 *
 ***********************************************************************************************/
func (cpcm *ClnPackClntManager) Fn_bat_init() int {

	var temp_str string
	var resultTmp int
	cpcm.ServiceName = "cln_pack_clnt"
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [Fn_bat_init] Entering Fn_bat_init")

	// we are getting the 7 cpcm.Args
	if len(cpcm.Args) < 7 {
		cpcm.LoggerManager.LogError(cpcm.ServiceName, " [Fn_bat_init] [Error: insufficient arguments provided %d", len(cpcm.Args))
		return -1
	}

	// setting pipe_id with command line argument (command line argument recieved from main)
	// using Pipe_Id to fetch xchng_code
	cpcm.Xchngbook.C_pipe_id = cpcm.Args[3]

	/* Create a message queue using the CreateQueue function from MessageQueue.go file ().
	This function is responsible for creating a system-level queue on Linux. */

	cpcm.Message_queue_manager.LoggerManager = cpcm.LoggerManager
	if cpcm.Message_queue_manager.CreateQueue(util.ORDINARY_ORDER_QUEUE_ID) != 0 {
		cpcm.LoggerManager.LogError(cpcm.ServiceName, " [Fn_bat_init] [Error: Returning from 'CreateQueue' with an Error... %s")
		cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [Fn_bat_init]  Exiting from function")
	} else {
		cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [Fn_bat_init] Created Message Queue SuccessFully")
	}

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [Fn_bat_init] Copied pipe ID from cpcm.Args[3]: %s", cpcm.Xchngbook.C_pipe_id)

	// Fetch exchange code from 'opm_ord_pipe_mstr'
	queryForOpm_Xchng_Cd := `SELECT opm_xchng_cd
				FROM opm_ord_pipe_mstr
				WHERE opm_pipe_id = ?`

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [Fn_bat_init] Executing query to fetch exchange code with pipe ID: %s", cpcm.Xchngbook.C_pipe_id)

	row := cpcm.Db.Raw(queryForOpm_Xchng_Cd, cpcm.Xchngbook.C_pipe_id).Row()
	temp_str = ""
	err := row.Scan(&temp_str)
	if err != nil {
		cpcm.LoggerManager.LogError(cpcm.ServiceName, " [Fn_bat_init] [Error: scanning row for exchange code: %v", cpcm.Xchngbook.C_xchng_cd, err)
		return -1
	}

	cpcm.Xchngbook.C_xchng_cd = temp_str
	temp_str = ""
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [Fn_bat_init] Exchange code fetched: %s", cpcm.Xchngbook.C_xchng_cd)

	/*here we are fetching the 'OPM_TRDR_ID' and 'OPM_BRNCH_ID' by using 'OPM_XCHNG_CD' and 'OPM_PIPE_ID'
	which is fatched earlier */

	queryForTraderAndBranchID := `SELECT OPM_TRDR_ID, OPM_BRNCH_ID
				FROM OPM_ORD_PIPE_MSTR
				WHERE OPM_XCHNG_CD = ? AND OPM_PIPE_ID = ?`

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [Fn_bat_init] Executing query to fetch trader and branch ID with exchange code: %s and pipe ID: %s", cpcm.Xchngbook.C_xchng_cd, cpcm.Xchngbook.C_pipe_id)

	rowQueryForTraderAndBranchID := cpcm.Db.Raw(queryForTraderAndBranchID, cpcm.Xchngbook.C_xchng_cd, cpcm.Xchngbook.C_pipe_id).Row()

	errQueryForTraderAndBranchID := rowQueryForTraderAndBranchID.Scan(&cpcm.Pipe_mstr.C_opm_trdr_id, &cpcm.Pipe_mstr.L_opm_brnch_id)
	if errQueryForTraderAndBranchID != nil {
		cpcm.LoggerManager.LogError(cpcm.ServiceName, " [Fn_bat_init] Error scanning row for trader and branch ID: %v", errQueryForTraderAndBranchID)
		return -1
	}

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [Fn_bat_init] fetched trader ID: %s and branch ID : %d", cpcm.Pipe_mstr.C_opm_trdr_id, cpcm.Pipe_mstr.L_opm_brnch_id)

	// Fetch modification trade date from 'exg_xchng_mstr'

	queryFor_exg_nxt_trd_dt := `SELECT exg_nxt_trd_dt,
				exg_brkr_id
				FROM exg_xchng_mstr
				WHERE exg_xchng_cd = ?`

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [Fn_bat_init] Fetching trade date and broker ID with exchange code: %s", cpcm.Xchngbook.C_xchng_cd)
	row2 := cpcm.Db.Raw(queryFor_exg_nxt_trd_dt, cpcm.Xchngbook.C_xchng_cd).Row()

	err2 := row2.Scan(&temp_str, &cpcm.Pipe_mstr.C_xchng_brkr_id)
	if err2 != nil {
		cpcm.LoggerManager.LogError(cpcm.ServiceName, " [Fn_bat_init] [Error: scanning trade date and broker ID: %v", err2)
		return -1
	}

	cpcm.Xchngbook.C_mod_trd_dt = temp_str

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [Fn_bat_init] Fetched trade date : %s and broker ID : %s", cpcm.Xchngbook.C_mod_trd_dt, cpcm.Pipe_mstr.C_xchng_brkr_id)

	cpcm.Xchngbook.L_ord_seq = 0 // I am initially setting it to '0' because it was set that way in 'fn_bat_init' and I have not seen it getting changed anywhere. If I find it being changed somewhere, I will update it accordingly.

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [Fn_bat_init] L_ord_seq initialized to %d", cpcm.Xchngbook.L_ord_seq)

	//here we are setting the S_user_typ_glb of 'St_opm_pipe_mstr' from the configuration file
	userTypeStr := cpcm.Enviroment_manager.GetProcessSpaceValue("UserSettings", "UserType")
	userType, err := strconv.Atoi(userTypeStr)
	if err != nil {
		cpcm.LoggerManager.LogError(cpcm.ServiceName, " [Fn_bat_init] [Error: Failed to convert UserType '%s' to integer: %v", userTypeStr, err)

	}
	cpcm.Pipe_mstr.S_user_typ_glb = userType

	// 'Below code is moved to 'Factory.go''
	//this value also we are reding from configuration file "I have set the temp value in the configuration file for now"
	// maxPackValStr := cpcm.Enviroment_manager.GetProcessSpaceValue("PackingLimit", "PACK_VAL")
	// if maxPackValStr == "" {
	// 	cpcm.LoggerManager.LogError(cpcm.ServiceName, " [Fn_bat_init] [Error: 'PACK_VAL' not found in the configuration under 'PackingLimit'")
	// } else {
	// 	maxPackVal, err := strconv.Atoi(maxPackValStr)
	// 	if err != nil {
	// 		cpcm.LoggerManager.LogError(cpcm.ServiceName, " [Fn_bat_init] [Error: Failed to convert 'PACK_VAL' '%s' to integer: %v", maxPackValStr, err)
	// 		maxPackVal = 0
	// 	} else {
	// 		cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [Fn_bat_init] Fetched and converted 'PACK_VAL' from configuration: %d", maxPackVal)
	// 	}
	// 	cpcm.Max_Pack_Val = maxPackVal
	// }

	// finally we are calling the CLN_PACK_CLNT function (service)
	resultTmp = cpcm.CLN_PACK_CLNT()
	if resultTmp != 0 {
		cpcm.LoggerManager.LogError(cpcm.ServiceName, " [Fn_bat_init] [Error: CLN_PACK_CLNT failed with result code: %d", resultTmp)
		cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [Fn_bat_init] [Error: Returning to main from fn_bat_init")
		return -1
	}
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [Fn_bat_init] Exiting Fn_bat_init")
	return 0
}

/***************************************************************************************
 * CLN_PACK_CLNT handles the core operations of the client package manager service.
 * It continuously retrieves newly created orders from the database, manages database
 * transactions, and writes the data to a system queue.
 *
 * INPUT PARAMETERS:
 *    cpcm.Args[] - Command-line arguments used for service configuration and control.
 *    Db     - Pointer to a GORM database connection object for SQL operations.
 *
 * OUTPUT PARAMETERS:
 *    int    - Returns 0 for success, or -1 if an error occurs.
 *
 * FUNCTIONAL FLOW:
 * 1. Enters an infinite loop to continuously process orders.
 * 2. Initiates a local transaction with `FnBeginTran()`.
 * 3. Retrieves the next order record using `fnGetNxtRec()`.
 * 4. Commits the transaction with `FnCommitTran()`.
 * 5. Writes the fetched data to the system queue using `WriteToQueue()`.
 * 6. Increments the message type and reset if it exceeds the maximum value.
 *
 ***********************************************************************************************/

func (cpcm *ClnPackClntManager) CLN_PACK_CLNT() int {
	var resultTmp int

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [CLN_PACK_CLNT] Entering CLN_PACK_CLNT")

	// Setting up an infinite loop to keep the service running and continuously fetch newly created orders.
	for {

		if cpcm.Message_queue_manager.FnCanWriteToQueue() != 0 {
			cpcm.LoggerManager.LogError(cpcm.ServiceName, "[CLN_PACK_CLNT] [Error: Queue is Full ")

			continue
		}

		// Beginning a local transaction for the service to perform fetch and update operations on the database.
		resultTmp = cpcm.Transaction_manager.FnBeginTran()
		if resultTmp == -1 {
			cpcm.LoggerManager.LogError(cpcm.ServiceName, " [CLN_PACK_CLNT] [Error: Transaction begin failed")
			return -1
		}
		cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [CLN_PACK_CLNT] Transaction started with type: %d", cpcm.Transaction_manager.TranType)

		// Calling the function to fetch the next order record.
		resultTmp = cpcm.fnGetNxtRec()
		if resultTmp != 0 {
			cpcm.LoggerManager.LogError(cpcm.ServiceName, " [CLN_PACK_CLNT] [Error: failed in getting the next record returning with result code: %d", resultTmp)
			if cpcm.Transaction_manager.FnAbortTran() == -1 {
				cpcm.LoggerManager.LogError(cpcm.ServiceName, " [CLN_PACK_CLNT] [Error: Transaction abort failed")
				return -1
			}
			cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [CLN_PACK_CLNT] Exiting CLN_PACK_CLNT with error")
			return -1
		}

		// Committing the transaction if it is a local transaction.
		if cpcm.Transaction_manager.FnCommitTran() == -1 {
			cpcm.LoggerManager.LogError(cpcm.ServiceName, " [CLN_PACK_CLNT] [Error: Transaction commit failed")
			return -1
		}

		cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [CLN_PACK_CLNT] Transaction committed successfully")

		// cpcm.LoggerManager.LogInfo( cpcm.ServiceName ,"********* Printing the q_packet from CLN_PACK_CLNT : %v", cpcm.Q_packet)

		// Writing the data fetched from 'fnGetNxtRec' to the system queue (Linux).
		//here setting the data becasues initially i am getting all zeros .
		cpcm.Message_queue_manager.ServiceName = cpcm.ServiceName

		mtypeWrite := *cpcm.MtypeWrite

		if cpcm.Message_queue_manager.WriteToQueue(mtypeWrite, cpcm.Q_packet) != 0 {
			cpcm.LoggerManager.LogError(cpcm.ServiceName, " [CLN_PACK_CLNT] [Error:  Failed to write to queue with message type %d", mtypeWrite)
			return -1
		}

		cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [CLN_PACK_CLNT] Successfully wrote to queue with message type %d", mtypeWrite)

		// testing purposes (this will not be the part of actual code)
		receivedData, exchng_struct, readErr := cpcm.Message_queue_manager.ReadFromQueue(mtypeWrite)
		if readErr != 0 {
			cpcm.LoggerManager.LogError(cpcm.ServiceName, " [CLN_PACK_CLNT] [Error:  Failed to read from queue with message type %d: %d", mtypeWrite, readErr)
			return -1
		}
		cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [CLN_PACK_CLNT] Successfully read from queue with message type %d, received structure as byte array: %d", mtypeWrite, exchng_struct)

		fmt.Println("Li Message Type:", receivedData)

		*cpcm.MtypeWrite++

		if *cpcm.MtypeWrite > cpcm.Max_Pack_Val {
			*cpcm.MtypeWrite = 1
		}

		TemporaryQueryForTesting := `UPDATE FXB_FO_XCHNG_BOOK
		                SET fxb_plcd_stts = 'R'
		                WHERE fxb_plcd_stts = 'Q'`

		cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [CLN_PACK_CLNT] Executing update query to change status from 'Q' to 'R' in FXB_FO_XCHNG_BOOK")

		result := cpcm.Db.Exec(TemporaryQueryForTesting)
		if result.Error != nil {
			cpcm.LoggerManager.LogError(cpcm.ServiceName, " [CLN_PACK_CLNT] [Error: Failed to execute update query: %v", result.Error)
			return -1
		}
		rowsAffected := result.RowsAffected
		cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [CLN_PACK_CLNT]  Update query executed successfully, %d rows affected", rowsAffected)

	}

	// cpcm.LoggerManager.LogInfo( cpcm.ServiceName ," [CLN_PACK_CLNT] Exiting CLN_PACK_CLNT")
	//return 0
}

/***************************************************************************************
 * fnGetNxtRec retrieves and processes the next record from the database based on its
 * status and type, then updates relevant models and statuses. It handles extraction
 * of order data, updates database statuses, and packs the data for further processing.
 *
 * INPUT PARAMETERS:
 *    Db - Pointer to a GORM database connection object for SQL operations.
 *
 * OUTPUT PARAMETERS:
 *    int - Returns 0 for successful execution, or -1 if an error occurs.
 *
 * FUNCTIONAL FLOW:
 * 1. Calls 'fnSeqToOmd' to fetch records from the 'FXB_FO_XCHNG_BOOK' table.
 * 2. Ensures 'C_plcd_stts' is 'REQUESTED'.
 * 3. Processes the order if its type is 'ORDINARY_ORDER':
 *    a. Transfers 'C_ordr_rfrnc' from 'Xchngbook' to 'orderbook'.
 *    b. Calls 'fnRefToOrd' to fetch order details from 'fod_fo_ordr_dtls'.
 *    c. Compares 'L_mdfctn_cntr' between 'Xchngbook' and 'orderbook'.
 *    d. Updates 'Xchngbook' and 'orderbook' statuses to 'QUEUED':
 *       - Calls 'fnUpdXchngbk' and 'fnUpdOrdrbk' to update the database.
 *    e. Sets the contract data using values from 'orderbook'.
 *    f. Checks if 'C_xchng_cd' is "NFO":
 *       - If "NFO", sets the request type and calls 'fnGetExtCnt' to load additional data.
 *    g. Calls 'fnPackOrdnryOrdToNse' to pack data into the final structure 'St_req_q_data'.
 *
 ***********************************************************************************************/

func (cpcm *ClnPackClntManager) fnGetNxtRec() int {
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetNxtRec] Entering fnGetNxtRec")

	cpcm.CPrgmFlg = "0"

	// Calling the function 'fnSeqToOmd' to fetched the Records from the 'FXB_FO_XCHNG_BOOK'
	resultTmp := cpcm.fnSeqToOmd()
	if resultTmp != 0 {
		cpcm.LoggerManager.LogError(cpcm.ServiceName, " [fnGetNxtRec] [Error: Failed to fetch data into eXchngbook structure")
		return -1
	}

	if cpcm.Xchngbook == nil {
		cpcm.LoggerManager.LogError(cpcm.ServiceName, " [fnGetNxtRec] [Error: Xchngbook is nil")
		return -1
	}

	if cpcm.Xchngbook.C_rms_prcsd_flg == "P" {
		cpcm.LoggerManager.GetLogger().Warnf(cpcm.ServiceName, " [fnGetNxtRec] Order ref / Mod number = |%s| / |%d| Already Processed", cpcm.Xchngbook.C_ordr_rfrnc, cpcm.Xchngbook.L_mdfctn_cntr)

		cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetNxtRec] Exiting from 'fnGetNxtRec' ")

		return 0
	}

	if cpcm.Xchngbook.C_plcd_stts != string(util.REQUESTED) {
		cpcm.LoggerManager.LogError(cpcm.ServiceName, " [fnGetNxtRec] [Error: the Placed status is not requested . so we are we are not getting the record ")
		return -1
	}

	// Checking the order type. Only fetch the order if it is of type 'Ordinary_Order'; otherwise, return without fetching.
	if cpcm.Xchngbook.C_ex_ordr_type == string(util.ORDINARY_ORDER) {

		cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetNxtRec] Assigning C_ordr_rfrnc from eXchngbook to orderbook")
		cpcm.Orderbook.C_ordr_rfrnc = cpcm.Xchngbook.C_ordr_rfrnc

		// Calling the function 'fnRefToOrd' to fetch orders from the 'fod_fo_ordr_dtls' table.

		resultTmp = cpcm.fnRefToOrd()
		if resultTmp != 0 {
			cpcm.LoggerManager.LogError(cpcm.ServiceName, " [fnGetNxtRec] [Error: Failed to fetch data into orderbook structure")
			return -1
		}

		// After fetching the records from 'order_book' and 'xchng_book', we check whether the values of 'L_mdfctn_cntr' from both tables are equal.
		if cpcm.Xchngbook.L_mdfctn_cntr != cpcm.Orderbook.L_mdfctn_cntr {
			cpcm.LoggerManager.LogError(cpcm.ServiceName, " [fnGetNxtRec] [Error: L_mdfctn_cntr of both Xchngbook and order are not same")
			cpcm.LoggerManager.LogInfo(cpcm.ServiceName, "GetNxtRec]  Exiting fnGetNxtRec")
			return -1

		}

		// Here, we are updating the status of 'xchngbook' to reflect the change in 'FXB_FO_XCHNG_BOOK', indicating that the record has been read.
		cpcm.Xchngbook.C_plcd_stts = string(util.QUEUED)
		cpcm.Xchngbook.C_oprn_typ = string(util.UPDATION_ON_ORDER_FORWARDING)

		resultTmp = cpcm.fnUpdXchngbk()

		if resultTmp != 0 {
			cpcm.LoggerManager.LogError(cpcm.ServiceName, " [fnGetNxtRec] [Error: failed to update the status in Xchngbook")
			return -1
		}

		cpcm.Orderbook.C_ordr_stts = string(util.QUEUED)
		// Here, we are updating the status of the 'fod_fo_ordr_dtls' table to indicate that this record has been read and queued.
		resultTmp = cpcm.fnUpdOrdrbk()

		if resultTmp != 0 {
			cpcm.LoggerManager.LogError(cpcm.ServiceName, " [fnGetNxtRec] [Error: failed to update the status in Order book")
			return -1
		}

		// here we are setting the contract data from orderbook structure
		cpcm.LoggerManager.LogInfo(cpcm.ServiceName, "  [fnGetNxtRec] Setting contract data from orderbook structure")
		cpcm.Contract.C_xchng_cd = cpcm.Orderbook.C_xchng_cd
		cpcm.Contract.C_prd_typ = cpcm.Orderbook.C_prd_typ
		cpcm.Contract.C_undrlyng = cpcm.Orderbook.C_undrlyng
		cpcm.Contract.C_expry_dt = cpcm.Orderbook.C_expry_dt
		cpcm.Contract.C_exrc_typ = cpcm.Orderbook.C_exrc_typ
		cpcm.Contract.C_opt_typ = cpcm.Orderbook.C_opt_typ
		cpcm.Contract.L_strike_prc = cpcm.Orderbook.L_strike_prc
		cpcm.Contract.C_ctgry_indstk = cpcm.Orderbook.C_ctgry_indstk
		// cpcm.Contract.L_ca_lvl = cpcm.Orderbook.L_ca_lvl

		cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetNxtRec] contract.C_xchng_cd: %s ", cpcm.Contract.C_xchng_cd)
		cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetNxtRec] contract.C_prd_typ: %s ", cpcm.Contract.C_prd_typ)
		cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetNxtRec] contract.C_undrlyng: %s ", cpcm.Contract.C_undrlyng)
		cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetNxtRec] contract.C_expry_dt: %s ", cpcm.Contract.C_expry_dt)
		cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetNxtRec] contract.C_exrc_typ: %s ", cpcm.Contract.C_exrc_typ)
		cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetNxtRec] contract.C_opt_typ: %s ", cpcm.Contract.C_opt_typ)
		cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetNxtRec] contract.L_strike_prc: %d ", cpcm.Contract.L_strike_prc)
		cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetNxtRec] contract.C_ctgry_indstk: %s ", cpcm.Contract.C_ctgry_indstk)
		//cpcm.LoggerManager.LogInfo( cpcm.ServiceName ," contract.L_ca_lvl: |%d|",  cpcm.Contract.L_ca_lvl)

		if strings.TrimSpace(cpcm.Xchngbook.C_xchng_cd) == "NFO" {

			cpcm.Contract.C_rqst_typ = string(util.CONTRACT_TO_NSE_ID)

			// Calling this function to extract data from various tables and store it in the contract structure.
			resultTmp = cpcm.fnGetExtCnt()

			if resultTmp != 0 {
				cpcm.LoggerManager.LogError(cpcm.ServiceName, "  [fnGetNxtRec] [Error: failed to load data in NSE CNT ")
				return -1
			}
		} else {
			cpcm.LoggerManager.LogError(cpcm.ServiceName, " [fnGetNxtRec] [Error: returning from 'fnGetNxtRec' because 'Xchngbook.C_xchng_cd' is not 'NFO' --> %s ", cpcm.Xchngbook.C_xchng_cd)
			cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetNxtRec] returning with Error")
			return -1
		}

		cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetNxtRec] Packing Ordinary Order STARTS ")

		// Checking if 'Token_id' is received from 'fnGetExtCnt'. If not, the record will be rejected.

		if cpcm.Nse_contract.L_token_id == 0 {
			cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetNxtRec] Token id for ordinary order is: %d", cpcm.Nse_contract.L_token_id)

			cpcm.LoggerManager.LogError(cpcm.ServiceName, " [fnGetNxtRec] [Error: token id is not avialable so we are calling 'fn_rjct_rcrd' ")

			// Calling 'fnRjctRcrd' to reject the record.
			resultTmp = cpcm.fnRjctRcrd()

			if resultTmp != 0 {
				cpcm.LoggerManager.LogError(cpcm.ServiceName, " [fnGetNxtRec] [Error: returned from 'fnRjctRcrd' with Error")
				cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetNxtRec] Exiting from 'fnGetNxtRec' ")
				return -1
			}

		}

		// here we are initialising the "ExchngPackLibMaster"
		eplm := packing_service.NewExchngPackLibMaster(
			cpcm.ServiceName,
			cpcm.Xchngbook,
			cpcm.Orderbook,
			cpcm.Pipe_mstr,
			cpcm.Nse_contract,
			cpcm.Oe_reqres,
			cpcm.Exch_msg,
			cpcm.Net_hdr,
			cpcm.Q_packet,
			cpcm.Int_header,
			cpcm.Contract_desc,
			cpcm.Order_flag,
			cpcm.Order_conversion_manager,
			cpcm.CPanNo,
			cpcm.CLstActRef,
			cpcm.CEspID,
			cpcm.CAlgoID,
			cpcm.CSourceFlg,
			cpcm.CPrgmFlg,
			cpcm.LoggerManager,
			cpcm.TCUM,
			cpcm.MtypeWrite,
			cpcm.Db,
		)

		// Calling the function 'fnPackOrdnryOrdToNse' to pack the entire data into the final structure 'St_req_q_data'.
		resultTmp = eplm.FnPackOrdnryOrdToNse()

		if resultTmp != 0 {
			cpcm.LoggerManager.LogError(cpcm.ServiceName, "  [fnGetNxtRec] [Error: 'fnPackOrdnryOrdToNse' returned an error.")
			cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetNxtRec] Exiting 'fnGetNxtRec'.")
		} else {
			cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetNxtRec] Data successfully packed.")
		}

	}

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetNxtRec] Exiting fnGetNxtRec")
	return 0
}

/***************************************************************************************
 * fnSeqToOmd fetches the next record from the 'FXB_FO_XCHNG_BOOK' table in the database
 * based on specified criteria (exchange code, pipe ID, trade date, and order sequence).
 * It retrieves order information and stores it in the 'Xchngbook' structure
 * within the 'ClnPackClntManager' instance.
 *
 * INPUT PARAMETERS:
 *    db - Pointer to a GORM database connection object for SQL operations.
 *
 * OUTPUT PARAMETERS:
 *    int - Returns 0 for successful data retrieval, or -1 if an error occurs.
 *
 * FUNCTIONAL FLOW:
 * 1. Constructs and executes a SQL query using GORM's 'Raw' method to fetch order details
 *    from the 'FXB_FO_XCHNG_BOOK' table:
 *    a. The query selects various columns, including order reference, quantity, order type,
 *       sequence number, etc., based on specified conditions.
 *    b. Utilizes 'TO_DATE' to format date fields and 'COALESCE' to handle null values for
 *       specific fields.
 * 2. Scans the retrieved row into the fields of the 'Xchngbook' structure and additional
 *    local variables ('c_ip_addrs', 'c_prcimpv_flg').
 *
 ***********************************************************************************************/

func (cpcm *ClnPackClntManager) fnSeqToOmd() int {
	var c_ip_addrs string
	var c_prcimpv_flg string

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, "[fnSeqToOmd] Entering fnSeqToOmd")
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, "[fnSeqToOmd] Before extracting the data from the 'fxb_ordr_rfrnc' and storing it in the 'Xchngbook' structure")

	query := `
		SELECT fxb_ordr_rfrnc,
		fxb_lmt_mrkt_sl_flg AS C_slm_flg,
		fxb_dsclsd_qty AS L_dsclsd_qty,
		fxb_ordr_tot_qty AS L_ord_tot_qty,
		fxb_lmt_rt AS L_ord_lmt_rt,
		fxb_stp_lss_tgr AS L_stp_lss_tgr,
		TO_CHAR(TO_DATE(fxb_ordr_valid_dt, 'YYYY-MM-DD'), 'DD-Mon-YYYY') AS C_valid_dt,
		CASE
			WHEN fxb_ordr_type = 'V' THEN 'T'
			ELSE fxb_ordr_type
		END AS C_ord_typ,
		fxb_rqst_typ AS C_req_typ,
		fxb_ordr_sqnc AS L_ord_seq,
		COALESCE(fxb_ip, 'NA') AS C_ip_addrs,
		COALESCE(fxb_prcimpv_flg, 'N') AS C_prcimpv_flg,
		fxb_ex_ordr_typ AS C_ex_ordr_type,
		fxb_plcd_stts AS C_plcd_stts,
		fxb_mdfctn_cntr AS L_mdfctn_cntr,
		fxb_frwd_tm AS C_frwrd_tm,
		fxb_jiffy AS D_jiffy,
		fxb_xchng_rmrks AS C_xchng_rmrks,
		fxb_rms_prcsd_flg AS C_rms_prcsd_flg,
		fxb_ors_msg_typ AS L_ors_msg_typ,
		fxb_ack_tm AS C_ack_tm
		FROM FXB_FO_XCHNG_BOOK
		WHERE fxb_xchng_cd = ?
		AND fxb_pipe_id = ?
		AND TO_DATE(fxb_mod_trd_dt, 'YYYY-MM-DD') = TO_DATE(?, 'YYYY-MM-DD')
		AND fxb_ordr_sqnc = (
		SELECT MIN(fxb_b.fxb_ordr_sqnc)
		FROM FXB_FO_XCHNG_BOOK fxb_b
		WHERE fxb_b.fxb_xchng_cd = ?
         AND TO_DATE(fxb_b.fxb_mod_trd_dt, 'YYYY-MM-DD') = TO_DATE(?, 'YYYY-MM-DD')
         AND fxb_b.fxb_pipe_id = ?
         AND fxb_b.fxb_plcd_stts = 'R'
		);


	`

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, "[fnSeqToOmd] Executing query to fetch order details")

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, "[fnSeqToOmd] C_xchng_cd: %s", cpcm.Xchngbook.C_xchng_cd)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, "[fnSeqToOmd] C_pipe_id: %s", cpcm.Xchngbook.C_pipe_id)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, "[fnSeqToOmd] C_mod_trd_dt: %s", cpcm.Xchngbook.C_mod_trd_dt)

	row := cpcm.Db.Raw(query,
		cpcm.Xchngbook.C_xchng_cd,
		cpcm.Xchngbook.C_pipe_id,
		cpcm.Xchngbook.C_mod_trd_dt,
		cpcm.Xchngbook.C_xchng_cd,
		cpcm.Xchngbook.C_mod_trd_dt,
		cpcm.Xchngbook.C_pipe_id).Row()

	err := row.Scan(
		&cpcm.Xchngbook.C_ordr_rfrnc,    // fxb_ordr_rfrnc
		&cpcm.Xchngbook.C_slm_flg,       // fxb_lmt_mrkt_sl_flg
		&cpcm.Xchngbook.L_dsclsd_qty,    // fxb_dsclsd_qty
		&cpcm.Xchngbook.L_ord_tot_qty,   // fxb_ordr_tot_qty
		&cpcm.Xchngbook.L_ord_lmt_rt,    // fxb_lmt_rt
		&cpcm.Xchngbook.L_stp_lss_tgr,   // fxb_stp_lss_tgr
		&cpcm.Xchngbook.C_valid_dt,      // C_valid_dt (formatted as 'DD-Mon-YYYY')
		&cpcm.Xchngbook.C_ord_typ,       // C_ord_typ (transformed 'V' to 'T')
		&cpcm.Xchngbook.C_req_typ,       // C_req_typ
		&cpcm.Xchngbook.L_ord_seq,       // L_ord_seq
		&c_ip_addrs,                     // COALESCE(fxb_ip, 'NA')
		&c_prcimpv_flg,                  // COALESCE(fxb_prcimpv_flg, 'N')
		&cpcm.Xchngbook.C_ex_ordr_type,  // C_ex_ordr_type
		&cpcm.Xchngbook.C_plcd_stts,     // C_plcd_stts
		&cpcm.Xchngbook.L_mdfctn_cntr,   // L_mdfctn_cntr
		&cpcm.Xchngbook.C_frwrd_tm,      // C_frwrd_tm
		&cpcm.Xchngbook.D_jiffy,         // D_jiffy
		&cpcm.Xchngbook.C_xchng_rmrks,   // C_xchng_rmrks
		&cpcm.Xchngbook.C_rms_prcsd_flg, // C_rms_prcsd_flg
		&cpcm.Xchngbook.L_ors_msg_typ,   // L_ors_msg_typ
		&cpcm.Xchngbook.C_ack_tm,        // C_ack_tm
	)

	if err != nil {
		cpcm.LoggerManager.LogError(cpcm.ServiceName, "[fnSeqToOmd] [Error: Error scanning row: %v", err)
		cpcm.LoggerManager.LogInfo(cpcm.ServiceName, "[fnSeqToOmd] Exiting fnSeqToOmd with error")
		return -1
	}

	if c_prcimpv_flg == "Y" {
		cpcm.CPrgmFlg = "T"
	} else {
		cpcm.CPrgmFlg = c_ip_addrs
	}

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, "[fnSeqToOmd] Data extracted and stored in the 'Xchngbook' structure:")
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, "[fnSeqToOmd]   C_ordr_rfrnc:     %s", cpcm.Xchngbook.C_ordr_rfrnc)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, "[fnSeqToOmd]   C_slm_flg:        %s", cpcm.Xchngbook.C_slm_flg)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, "[fnSeqToOmd]   L_dsclsd_qty:     %d", cpcm.Xchngbook.L_dsclsd_qty)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, "[fnSeqToOmd]   L_ord_tot_qty:    %d", cpcm.Xchngbook.L_ord_tot_qty)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, "[fnSeqToOmd]   L_ord_lmt_rt:     %d", cpcm.Xchngbook.L_ord_lmt_rt)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, "[fnSeqToOmd]   L_stp_lss_tgr:    %d", cpcm.Xchngbook.L_stp_lss_tgr)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, "[fnSeqToOmd]   C_valid_dt:       %s", cpcm.Xchngbook.C_valid_dt)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, "[fnSeqToOmd]   C_ord_typ:        %s", cpcm.Xchngbook.C_ord_typ)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, "[fnSeqToOmd]   C_req_typ:        %s", cpcm.Xchngbook.C_req_typ)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, "[fnSeqToOmd]   L_ord_seq:        %d", cpcm.Xchngbook.L_ord_seq)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, "[fnSeqToOmd]   IpAddrs:          %s", c_ip_addrs)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, "[fnSeqToOmd]   PrcimpvFlg:       %s", c_prcimpv_flg)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, "[fnSeqToOmd]   C_ex_ordr_type:   %s", cpcm.Xchngbook.C_ex_ordr_type)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, "[fnSeqToOmd]   C_plcd_stts:      %s", cpcm.Xchngbook.C_plcd_stts)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, "[fnSeqToOmd]   L_mdfctn_cntr:    %d", cpcm.Xchngbook.L_mdfctn_cntr)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, "[fnSeqToOmd]   C_frwrd_tm:       %s", cpcm.Xchngbook.C_frwrd_tm)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, "[fnSeqToOmd]   D_jiffy:          %f", cpcm.Xchngbook.D_jiffy)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, "[fnSeqToOmd]   C_xchng_rmrks:    %s", cpcm.Xchngbook.C_xchng_rmrks)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, "[fnSeqToOmd]   C_rms_prcsd_flg:  %s", cpcm.Xchngbook.C_rms_prcsd_flg)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, "[fnSeqToOmd]   L_ors_msg_typ:    %d", cpcm.Xchngbook.L_ors_msg_typ)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, "[fnSeqToOmd]   C_ack_tm:         %s", cpcm.Xchngbook.C_ack_tm)

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, "[fnSeqToOmd] Data extracted and stored in the 'Xchngbook' structure successfully")

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, "[fnSeqToOmd] Exiting fnSeqToOmd")
	return 0
}

/***********************************************************************************************
 * fnRefToOrd retrieves and processes order data from the 'fod_fo_ordr_dtls' and 'CLM_CLNT_MSTR'
 * tables in the database, and stores it in the `orderbook` structure within the
 * 'ClnPackClntManager' instance.
 *
 * INPUT PARAMETERS:
 *    None
 *
 * OUTPUT PARAMETERS:
 *    int - Returns 0 for success, or -1 if an error occurs.
 *
 * FUNCTIONAL FLOW:
 * 1. Constructs and executes a SQL query using GORM's 'Raw' method to fetch order details
 *    from the 'fod_fo_ordr_dtls' table:
 *    a. Selects various columns, including client match account, order flow, total order quantity,
 *       executed quantity, etc., based on specified conditions.
 *    b. Uses 'FOR UPDATE NOWAIT' to lock selected rows for update and prevent other transactions
 *       from waiting for the lock.
 * 2. Scans the retrieved row into the fields of the `orderbook` structure within the
 *    'ClnPackClntManager' instance.
 * 3. Constructs and executes a second SQL query to fetch client details from the 'CLM_CLNT_MSTR' table
 *    where the client match account matches the previously fetched value.
 *
 ***********************************************************************************************/

func (cpcm *ClnPackClntManager) fnRefToOrd() int {
	/*
		c_pan_no --> we are simply setting it to 0 using memset in the
		MEMSET(c_ordr_rfrnc);
		MEMSET(c_pan_no);
		MEMSET(c_lst_act_ref);
		MEMSET(c_esp_id);
		MEMSET(c_algo_id);
		char c_source_flg = '\0';
	*/

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd]  Entering fnFetchOrderDetails")
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd] Before extracting the data from the 'fod_fo_ordr_dtls' and storing it in the 'orderbook' structure")

	query := `
    SELECT
    fod_clm_mtch_accnt AS C_cln_mtch_accnt,
    fod_ordr_flw AS C_ordr_flw,
    fod_ordr_tot_qty AS L_ord_tot_qty,
    fod_exec_qty AS L_exctd_qty,
    COALESCE(fod_exec_qty_day, 0) AS L_exctd_qty_day,
    fod_settlor AS C_settlor,
    fod_spl_flag AS C_spl_flg,
    TO_CHAR(fod_ord_ack_tm, 'YYYY-MM-DD HH24:MI:SS') AS C_ack_tm,
    TO_CHAR(fod_lst_rqst_ack_tm, 'YYYY-MM-DD HH24:MI:SS') AS C_prev_ack_tm,
    fod_pro_cli_ind AS C_pro_cli_ind,
    COALESCE(fod_ctcl_id, ' ') AS C_ctcl_id,
    COALESCE(fod_pan_no, '*') AS C_pan_no,
    COALESCE(fod_lst_act_ref, '0') AS C_lst_act_ref,
    COALESCE(fod_esp_id, '*') AS C_esp_id,
    COALESCE(fod_algo_id, '*') AS C_algo_id,
    COALESCE(fod_source_flg, '*') AS C_source_flg,
    fod_mdfctn_cntr AS L_mdfctn_cntr,
    fod_ordr_stts AS C_ordr_stts,
    fod_xchng_cd AS C_xchng_cd,
    fod_prdct_typ AS C_prd_typ,
    fod_undrlying AS C_undrlyng,
    fod_expry_dt AS C_expry_dt,
    fod_exer_typ AS C_exrc_typ,
    fod_opt_typ AS C_opt_typ,
    fod_strk_prc AS L_strike_prc,
    fod_indstk AS C_ctgry_indstk,
	fod_ack_nmbr As C_xchng_ack
	FROM
		fod_fo_ordr_dtls
	WHERE
		fod_ordr_rfrnc =  ?
	FOR UPDATE NOWAIT;

    `

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd] Executing query to fetch order details")

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd] Order Reference: %s", cpcm.Orderbook.C_ordr_rfrnc)

	row := cpcm.Db.Raw(query, strings.TrimSpace(cpcm.Orderbook.C_ordr_rfrnc)).Row()

	err := row.Scan(
		&cpcm.Orderbook.C_cln_mtch_accnt,
		&cpcm.Orderbook.C_ordr_flw,
		&cpcm.Orderbook.L_ord_tot_qty,
		&cpcm.Orderbook.L_exctd_qty,
		&cpcm.Orderbook.L_exctd_qty_day,
		&cpcm.Orderbook.C_settlor,
		&cpcm.Orderbook.C_spl_flg,
		&cpcm.Orderbook.C_ack_tm,
		&cpcm.Orderbook.C_prev_ack_tm,
		&cpcm.Orderbook.C_pro_cli_ind,
		&cpcm.Orderbook.C_ctcl_id,
		&cpcm.CPanNo,
		&cpcm.CLstActRef,
		&cpcm.CEspID,
		&cpcm.CAlgoID,
		&cpcm.CSourceFlg,
		&cpcm.Orderbook.L_mdfctn_cntr,
		&cpcm.Orderbook.C_ordr_stts,
		&cpcm.Orderbook.C_xchng_cd,
		&cpcm.Orderbook.C_prd_typ,
		&cpcm.Orderbook.C_undrlyng,
		&cpcm.Orderbook.C_expry_dt,
		&cpcm.Orderbook.C_exrc_typ,
		&cpcm.Orderbook.C_opt_typ,
		&cpcm.Orderbook.L_strike_prc,
		&cpcm.Orderbook.C_ctgry_indstk,
		&cpcm.Orderbook.C_xchng_ack,
	)

	if err != nil {
		cpcm.LoggerManager.LogError(cpcm.ServiceName, " [fnRefToOrd] [Error: scanning row: %v", err)
		cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd] Exiting fnFetchOrderDetails with error")
		return -1
	}

	cpcm.CPanNo = strings.TrimSpace(cpcm.CPanNo)
	cpcm.CLstActRef = strings.TrimSpace(cpcm.CLstActRef)
	cpcm.CEspID = strings.TrimSpace(cpcm.CEspID)
	cpcm.CAlgoID = strings.TrimSpace(cpcm.CAlgoID)
	cpcm.CSourceFlg = strings.TrimSpace(cpcm.CSourceFlg)

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd] Data extracted and stored in the 'orderbook' structure:")

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd]   C_cln_mtch_accnt:   %s", cpcm.Orderbook.C_cln_mtch_accnt)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd]   C_ordr_flw:        %s", cpcm.Orderbook.C_ordr_flw)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd]   L_ord_tot_qty:     %d", cpcm.Orderbook.L_ord_tot_qty)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd]   L_exctd_qty:       %d", cpcm.Orderbook.L_exctd_qty)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd]   L_exctd_qty_day:   %d", cpcm.Orderbook.L_exctd_qty_day)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd]   C_settlor:         %s", cpcm.Orderbook.C_settlor)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd]   C_spl_flg:         %s", cpcm.Orderbook.C_spl_flg)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd]   C_ack_tm:          %s", cpcm.Orderbook.C_ack_tm)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd]   C_prev_ack_tm:     %s", cpcm.Orderbook.C_prev_ack_tm)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd]   C_pro_cli_ind:     %s", cpcm.Orderbook.C_pro_cli_ind)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd]   C_ctcl_id:         %s", cpcm.Orderbook.C_ctcl_id)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd]   C_pan_no:          %s", cpcm.CPanNo)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd]   C_lst_act_ref_tmp: %s", cpcm.CLstActRef)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd]   C_esp_id:          %s", cpcm.CEspID)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd]   C_algo_id:         %s", cpcm.CAlgoID)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd]   C_source_flg_tmp:  %s", cpcm.CSourceFlg)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd]   L_mdfctn_cntr:     %d", cpcm.Orderbook.L_mdfctn_cntr)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd]   C_ordr_stts:       %s", cpcm.Orderbook.C_ordr_stts)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd]   C_xchng_cd:        %s", cpcm.Orderbook.C_xchng_cd)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd]   C_prd_typ:         %s", cpcm.Orderbook.C_prd_typ)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd]   C_undrlyng:        %s", cpcm.Orderbook.C_undrlyng)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd]   C_expry_dt:        %s", cpcm.Orderbook.C_expry_dt)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd]   C_exrc_typ:        %s", cpcm.Orderbook.C_exrc_typ)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd]   C_opt_typ:         %s", cpcm.Orderbook.C_opt_typ)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd]   L_strike_prc:      %d", cpcm.Orderbook.L_strike_prc)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd]   C_ctgry_indstk:    %s", cpcm.Orderbook.C_ctgry_indstk)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd]   C_xchng_ack	: 	   %s", cpcm.Orderbook.C_xchng_ack)

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd] Data extracted and stored in the 'orderbook' structure successfully")

	clmQuery := `
    SELECT
    COALESCE(TRIM(CLM_CP_CD), ' '),
    COALESCE(TRIM(CLM_CLNT_CD), CLM_MTCH_ACCNT)
	FROM
    CLM_CLNT_MSTR
	WHERE
    CLM_MTCH_ACCNT = ?
    `
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd] Executing query to fetch data from 'CLM_CLNT_MSTR' ")

	row = cpcm.Db.Raw(clmQuery, strings.TrimSpace(cpcm.Orderbook.C_cln_mtch_accnt)).Row()

	var cCpCode, cUccCd string
	err = row.Scan(&cCpCode, &cUccCd)

	if err != nil {
		cpcm.LoggerManager.LogError(cpcm.ServiceName, " [fnRefToOrd] [Error: scanning client details: %v", err)
		cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd] Exiting fnRefToOrd with error")
		return -1
	}

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd] cCpCode : %s", cCpCode)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd] cUccCd : %s", cUccCd)

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd] Updating orderbook with client details")

	cpcm.Orderbook.C_cln_mtch_accnt = strings.TrimSpace(cUccCd)
	cpcm.Orderbook.C_settlor = strings.TrimSpace(cCpCode)
	cpcm.Orderbook.C_ctcl_id = strings.TrimSpace(cpcm.Orderbook.C_ctcl_id)

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRefToOrd] Exiting fnFetchOrderDetails")
	return 0
}

/***********************************************************************************************
 * fnUpdXchngbk updates the status of a record in the 'fxb_fo_xchng_book' table based on the
 * operation type specified in the `Xchngbook` structure within the 'ClnPackClntManager' instance.
 * which are set in the 'fnGetNxtRec' .
 * The function handles two main operations:
 * 1. **Order Forwarding Update**: Updates the 'fxb_plcd_stts' and 'fxb_frwd_tm' fields for a record
 *    where 'fxb_ordr_rfrnc' and 'fxb_mdfctn_cntr' match the values in the `Xchngbook` structure.
 * 2. **Exchange Response Update**: If the download flag indicates the record should be processed, it
 *    checks if the record exists using the provided 'fxb_jiffy', 'fxb_xchng_cd', and 'fxb_pipe_id'.
 *    If the record exists, it logs that it has already been processed. Otherwise, it updates the record
 *    with the provided status, remarks, and other details.
 *
 * INPUT PARAMETERS:
 *    - db *gorm.DB: The database connection instance used for executing SQL queries.
 *
 * OUTPUT PARAMETERS:
 *    int - Returns 0 for success or -1 if an error occurs during the update operations.
 *
 * FUNCTIONAL FLOW:

 * 1. Based on the `C_oprn_typ` value in the `Xchngbook` structure:
 *    a. For **Order Forwarding** (`util.UPDATION_ON_ORDER_FORWARDING`):
 *       - Constructs and executes a SQL query to update 'fxb_plcd_stts' and 'fxb_frwd_tm' fields
 *         for the record that matches 'fxb_ordr_rfrnc' and 'fxb_mdfctn_cntr'.
 *    b. For **Exchange Response** (`util.UPDATION_ON_EXCHANGE_RESPONSE`):
 *       - Checks if the record should be processed based on the `L_dwnld_flg` value.
 *       - If processing is required, it verifies if a record already exists using the provided
 *         'fxb_jiffy', 'fxb_xchng_cd', and 'fxb_pipe_id'.
 *       - If the record is not found, updates the 'fxb_fo_xchng_book' table with new values including
 *         'fxb_plcd_stts', 'fxb_rms_prcsd_flg', 'fxb_ors_msg_typ', and 'fxb_ack_tm'.
 ***********************************************************************************************/

func (cpcm *ClnPackClntManager) fnUpdXchngbk() int {

	var iRecExists int64

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnUpdXchngbk] Entering fnUpdXchngbk")
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnUpdXchngbk]: %s", cpcm.Xchngbook.C_xchng_cd)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnUpdXchngbk]: %s", cpcm.Xchngbook.C_ordr_rfrnc)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnUpdXchngbk]: %s", cpcm.Xchngbook.C_pipe_id)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnUpdXchngbk]: %s", cpcm.Xchngbook.C_mod_trd_dt)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnUpdXchngbk]: %d", cpcm.Xchngbook.L_ord_seq)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnUpdXchngbk]: %s", cpcm.Xchngbook.C_slm_flg)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnUpdXchngbk]: %d", cpcm.Xchngbook.L_dsclsd_qty)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnUpdXchngbk]: %d", cpcm.Xchngbook.L_ord_tot_qty)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnUpdXchngbk]: %d", cpcm.Xchngbook.L_ord_lmt_rt)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnUpdXchngbk]: %d", cpcm.Xchngbook.L_stp_lss_tgr)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnUpdXchngbk]: %d", cpcm.Xchngbook.L_mdfctn_cntr)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnUpdXchngbk]: %s", cpcm.Xchngbook.C_valid_dt)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnUpdXchngbk]: %s", cpcm.Xchngbook.C_ord_typ)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnUpdXchngbk]: %s", cpcm.Xchngbook.C_req_typ)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnUpdXchngbk]: %s", cpcm.Xchngbook.C_plcd_stts)

	switch cpcm.Xchngbook.C_oprn_typ {
	case string(util.UPDATION_ON_ORDER_FORWARDING):

		query1 := `UPDATE fxb_fo_xchng_book 
				SET fxb_plcd_stts = ?, fxb_frwd_tm = CURRENT_TIMESTAMP
				WHERE fxb_ordr_rfrnc = ? 
				AND fxb_mdfctn_cntr = ?`

		result := cpcm.Db.Exec(query1, cpcm.Xchngbook.C_plcd_stts, cpcm.Xchngbook.C_ordr_rfrnc, cpcm.Xchngbook.L_mdfctn_cntr)
		if result.Error != nil {
			cpcm.LoggerManager.LogError(cpcm.ServiceName, " [fnUpdXchngbk] [Error: updating order forwarding: %v", result.Error)
			return -1
		}

	case string(util.UPDATION_ON_EXCHANGE_RESPONSE):
		if cpcm.Xchngbook.L_dwnld_flg == util.DOWNLOAD {

			result := cpcm.Db.Raw(
				`SELECT COUNT(*) 
					FROM fxb_fo_xchng_book 
					WHERE fxb_jiffy = ? 
					AND fxb_xchng_cd = ? 
					AND fxb_pipe_id = ?`,
				cpcm.Xchngbook.D_jiffy, cpcm.Xchngbook.C_xchng_cd, cpcm.Xchngbook.C_pipe_id,
			).Scan(&iRecExists)
			if result.Error != nil {
				cpcm.LoggerManager.LogError(cpcm.ServiceName, " [fnUpdXchngbk] [Error: SQL error: %v", result.Error)
				return -1
			}

			if iRecExists > 0 {
				cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnUpdXchngbk] Record already processed")
				return -1
			}
		}

		c_xchng_rmrks := strings.TrimSpace(cpcm.Xchngbook.C_xchng_rmrks)

		result := cpcm.Db.Exec(
			`UPDATE fxb_fo_xchng_book 
				SET fxb_plcd_stts = ?, 
					fxb_rms_prcsd_flg = ?, 
					fxb_ors_msg_typ = ?, 
					fxb_ack_tm = TO_DATE(?, 'DD-Mon-yyyy hh24:mi:ss'), 
					fxb_xchng_rmrks = RTRIM(fxb_xchng_rmrks) || ?, 
					fxb_jiffy = ? 
				WHERE fxb_ordr_rfrnc = ? 
				AND fxb_mdfctn_cntr = ?`,
			cpcm.Xchngbook.C_plcd_stts,
			cpcm.Xchngbook.C_rms_prcsd_flg,
			cpcm.Xchngbook.L_ors_msg_typ,
			cpcm.Xchngbook.C_ack_tm,
			c_xchng_rmrks,
			cpcm.Xchngbook.D_jiffy,
			cpcm.Xchngbook.C_ordr_rfrnc,
			cpcm.Xchngbook.L_mdfctn_cntr,
		)
		if result.Error != nil {
			cpcm.LoggerManager.LogError(cpcm.ServiceName, " [fnUpdXchngbk] [Error: updating exchange response: %v", result.Error)
			return -1
		}

	default:
		cpcm.LoggerManager.LogError(cpcm.ServiceName, " [fnUpdXchngbk] [Error: Invalid Operation Type")
		return -1
	}

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnUpdXchngbk] Exiting fnUpdXchngbk")

	return 0
}

/***********************************************************************************************
 * fnUpdOrdrbk updates the 'order status' of a record in the 'fod_fo_ordr_dtls' table based on
 * the details provided in the `orderbook` structure within the 'ClnPackClntManager' instance.
 * The function performs the following operations:
 * 1. **Order Status Update**: Updates the 'fod_ordr_stts' field for a record where the
 *    'fod_ordr_rfrnc' matches the value specified in the `orderbook` structure.
 *
 * INPUT PARAMETERS:
 *    - db *gorm.DB: The database connection instance used for executing SQL queries.
 *
 * OUTPUT PARAMETERS:
 *    int - Returns 0 for success or -1 if an error occurs during the update operation.
 *
 * FUNCTIONAL FLOW:
 * 1. Constructs and executes a SQL query to update the 'fod_ordr_stts' field in the
 *    'fod_fo_ordr_dtls' table where 'fod_ordr_rfrnc' matches the value from the `orderbook` structure.
 ***********************************************************************************************/

func (cpcm *ClnPackClntManager) fnUpdOrdrbk() int {

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnUpdOrdrbk] Starting 'fnUpdOrdbk' function")

	query := `
    UPDATE fod_fo_ordr_dtls
    SET fod_ordr_stts = ?
    WHERE fod_ordr_rfrnc = ?;
	`

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnUpdOrdrbk] Executing query to update order status")

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnUpdOrdrbk] Order Reference: %s", cpcm.Orderbook.C_ordr_rfrnc)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnUpdOrdrbk]  Order Status: %s", cpcm.Orderbook.C_ordr_stts)

	result := cpcm.Db.Exec(query, cpcm.Orderbook.C_ordr_stts, strings.TrimSpace(cpcm.Orderbook.C_ordr_rfrnc))

	if result.Error != nil {
		cpcm.LoggerManager.LogError(cpcm.ServiceName, " [fnUpdOrdrbk] [Error: executing update query: %v", result.Error)
		cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnUpdOrdrbk] Exiting 'fnUpdOrdbk' with error")
		return -1
	}

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnUpdOrdrbk] Order status updated successfully")

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnUpdOrdrbk] returning from 'fnUpdOrdbk' function")

	return 0
}

/***********************************************************************************************
 * fnGetExtCnt retrieves and sets external contract details based on the values provided
 * in the `contract` structure within the 'ClnPackClntManager' instance. This function performs
 * multiple operations to populate the `nse_contract` structure with relevant data.
 *
 * The function performs the following operations:
 * 1. **Determine Entity Type**: Based on the value of `C_xchng_cd` in the `contract` structure,
 *    sets the `i_sem_entity` variable to either `util.NFO_ENTTY` or `util.BFO_ENTTY`.
 *
 * 2. **Fetch Symbol**: Executes a query to fetch the stock symbol (`C_symbol`) from the
 *    'sem_stck_map' table using the stock code (`C_undrlyng`) and entity type.
 *
 * 3. **Determine Exchange Code**: Sets the exchange code (`c_exg_cd`) based on the value of
 *    `C_xchng_cd` in the `contract` structure.
 *
 * 4. **Fetch Series**: Executes a query to fetch the series (`C_series`) from the 'ESS_SGMNT_STCK'
 *    table using the underlying stock code (`C_undrlyng`) and exchange code (`c_exg_cd`).
 *
 * 5. **Fetch Token ID and CA Level**: Executes a query to fetch token ID (`L_token_id`) and
 *    CA level (`L_ca_lvl`) from the 'ftq_fo_trd_qt' table using various contract details including
 *    exchange code, product type, underlying stock code, expiry date, exercise type, option type,
 *    and strike price. The expiry date is formatted to 'dd-Mon-yyyy'.
 *
 * INPUT PARAMETERS:
 *    - db *gorm.DB: The database connection instance used for executing SQL queries.
 *
 * OUTPUT PARAMETERS:
 *    int - Returns 0 for success or -1 if an error occurs .
 *
 * FUNCTIONAL FLOW:
 * 1. Determines the entity type and fetches the symbol from 'sem_stck_map'.
 * 2. Based on the `C_xchng_cd`, sets the appropriate exchange code and fetches the series from
 *    'ESS_SGMNT_STCK'.
 * 3. Retrieves token ID and CA level from 'ftq_fo_trd_qt' using the formatted expiry date and
 *    other contract details.
 * 4. Updates the `nse_contract` structure with the fetched values.
 ***********************************************************************************************/

func (cpcm *ClnPackClntManager) fnGetExtCnt() int {
	var i_sem_entity int
	var c_stck_cd, c_symbl, c_exg_cd string

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetExtCnt] Entering 'fnGetExtCnt' ")

	if cpcm.Contract.C_xchng_cd == "NFO" {
		i_sem_entity = util.NFO_ENTTY
	} else {
		i_sem_entity = util.BFO_ENTTY
	}

	c_stck_cd = cpcm.Contract.C_undrlyng

	c_stck_cd = strings.ToUpper(c_stck_cd)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetExtCnt] Converted 'c_stck_cd' to uppercase: %s", c_stck_cd)

	query1 := `
		SELECT TRIM(sem_map_vl)
		FROM sem_stck_map
		WHERE sem_entty = ?
		AND sem_stck_cd = ?;
	`
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetExtCnt] Executing query to fetch 'C_symbol'")

	row1 := cpcm.Db.Raw(query1, i_sem_entity, c_stck_cd).Row()

	err := row1.Scan(&c_symbl)

	if err != nil {
		cpcm.LoggerManager.LogError(cpcm.ServiceName, " [fnGetExtCnt] [Error:  scanning row: %v", err)
		cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetExtCnt] Exiting 'fnGetExtCnt' due to error")
		return -1
	}

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetExtCnt] Value successfully fetched from the table 'sem_stck_map'")
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetExtCnt] cpcm.Nse_contract.C_symbol is: %s", c_symbl)

	// Assign values to the NSE contract structure
	cpcm.Nse_contract.C_xchng_cd = cpcm.Contract.C_xchng_cd
	cpcm.Nse_contract.C_prd_typ = cpcm.Contract.C_prd_typ
	cpcm.Nse_contract.C_expry_dt = cpcm.Contract.C_expry_dt
	cpcm.Nse_contract.C_exrc_typ = cpcm.Contract.C_exrc_typ
	cpcm.Nse_contract.C_opt_typ = cpcm.Contract.C_opt_typ
	cpcm.Nse_contract.L_strike_prc = cpcm.Contract.L_strike_prc
	cpcm.Nse_contract.C_symbol = c_symbl
	cpcm.Nse_contract.C_rqst_typ = cpcm.Contract.C_rqst_typ
	cpcm.Nse_contract.C_ctgry_indstk = cpcm.Contract.C_ctgry_indstk

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetExtCnt] cpcm.Nse_contract.C_xchng_cd is: %s", cpcm.Nse_contract.C_xchng_cd)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetExtCnt] cpcm.Nse_contract.C_prd_typ is: %s", cpcm.Nse_contract.C_prd_typ)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetExtCnt] cpcm.Nse_contract.C_expry_dt is: %s", cpcm.Nse_contract.C_expry_dt)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetExtCnt] cpcm.Nse_contract.C_exrc_typ is: %s", cpcm.Nse_contract.C_exrc_typ)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetExtCnt] cpcm.Nse_contract.C_opt_typ is: %s", cpcm.Nse_contract.C_opt_typ)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetExtCnt] cpcm.Nse_contract.L_strike_prc is: %d", cpcm.Nse_contract.L_strike_prc)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetExtCnt] cpcm.Nse_contract.C_symbol is: %s", cpcm.Nse_contract.C_symbol)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetExtCnt] cpcm.Nse_contract.C_ctgry_indstk is: %s", cpcm.Nse_contract.C_ctgry_indstk)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetExtCnt] cpcm.Nse_contract.C_rqst_typ is: %s", cpcm.Nse_contract.C_rqst_typ)

	if cpcm.Contract.C_xchng_cd == "NFO" {
		c_exg_cd = "NSE"
	} else if cpcm.Contract.C_xchng_cd == "BFO" {
		c_exg_cd = "BSE"
	} else {
		cpcm.LoggerManager.LogError(cpcm.ServiceName, " [fnGetExtCnt] [Error: Invalid option '%s' for 'cpcm.Contract.C_xchng_cd'. Exiting 'fnGetExtCnt'", cpcm.Contract.C_xchng_cd)
		return -1
	}

	query2 := `
		SELECT COALESCE(ESS_XCHNG_SUB_SERIES, ' ')
		FROM ESS_SGMNT_STCK
		WHERE ess_stck_cd = ?
		AND ess_xchng_cd = ?;
	`

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetExtCnt] Executing query to fetch 'C_series'")

	row2 := cpcm.Db.Raw(query2, cpcm.Contract.C_undrlyng, c_exg_cd).Row()

	err2 := row2.Scan(&cpcm.Nse_contract.C_series)

	if err2 != nil {
		cpcm.LoggerManager.LogError(cpcm.ServiceName, " [fnGetExtCnt] [Error: scanning row: %v", err2)
		cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetExtCnt] Exiting 'fnGetExtCnt' due to error")
		return -1
	}

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetExtCnt] Value successfully fetched from the table 'ESS_SGMNT_STCK'")
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetExtCnt] cpcm.Nse_contract.C_series is: %s", cpcm.Nse_contract.C_series)

	query3 := `
		SELECT COALESCE(ftq_token_no, 0),
		       COALESCE(ftq_ca_lvl, 0)
		FROM ftq_fo_trd_qt
		WHERE ftq_xchng_cd = ?
		  AND ftq_prdct_typ = ?
		  AND ftq_undrlyng = ?
		  AND ftq_expry_dt = TO_DATE(?, 'dd-Mon-yyyy')
		  AND ftq_exer_typ = ?
		  AND ftq_opt_typ = ?
		  AND ftq_strk_prc = ?;
	`

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetExtCnt] Executing query to fetch 'L_token_id' and 'L_ca_lvl'")

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetExtCnt] C_xchng_cd: %s", cpcm.Contract.C_xchng_cd)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetExtCnt] C_prd_typ: %s", cpcm.Contract.C_prd_typ)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetExtCnt] C_undrlyng: %s", cpcm.Contract.C_undrlyng)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetExtCnt] C_expry_dt: %s", cpcm.Contract.C_expry_dt)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetExtCnt] C_exrc_typ: %s", cpcm.Contract.C_exrc_typ)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetExtCnt] C_opt_typ: %s", cpcm.Contract.C_opt_typ)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetExtCnt] L_strike_prc: %d", cpcm.Contract.L_strike_prc)

	parsedDate, err := time.Parse(time.RFC3339, cpcm.Contract.C_expry_dt)
	if err != nil {
		cpcm.LoggerManager.LogError(cpcm.ServiceName, " [fnGetExtCnt] [Error: parsing date: %v", err)
		cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetExtCnt] Exiting 'fnGetExtCnt' due to error")
		return -1
	}

	// Format the parsed date as 'dd-Mon-yyyy'
	formattedDate := parsedDate.Format("02-Jan-2006")

	row3 := cpcm.Db.Raw(query3,
		cpcm.Contract.C_xchng_cd,
		cpcm.Contract.C_prd_typ,
		cpcm.Contract.C_undrlyng,
		formattedDate,
		cpcm.Contract.C_exrc_typ,
		cpcm.Contract.C_opt_typ,
		cpcm.Contract.L_strike_prc,
	).Row()

	err3 := row3.Scan(&cpcm.Nse_contract.L_token_id, &cpcm.Nse_contract.L_ca_lvl)

	if err3 != nil {
		cpcm.LoggerManager.LogError(cpcm.ServiceName, " [fnGetExtCnt] [Error: scanning row: %v", err3)
		cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetExtCnt] Exiting 'fnGetExtCnt' due to error")
		return -1
	}

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetExtCnt] Values successfully fetched from the table 'ftq_fo_trd_qt'")
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetExtCnt] cpcm.Nse_contract.L_token_id is: %d", cpcm.Nse_contract.L_token_id)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnGetExtCnt] cpcm.Nse_contract.L_ca_lvl is: %d", cpcm.Nse_contract.L_ca_lvl)

	return 0
}

func (cpcm *ClnPackClntManager) fnRjctRcrd() int {

	var resultTmp int
	var c_tm_stmp string

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRjctRcrd] Function 'fnRjctRcrd' starts")

	/*
		// cpcm.LoggerManager.LogInfo( cpcm.ServiceName ," [fnRjctRcrd] Product type is : %s",  cpcm.Orderbook.C_prd_typ)

		// if cpcm.Orderbook.C_prd_typ == util.FUTURES {
		// 	svc = "SFO_FUT_ACK"
		// } else {
		// 	svc = "SFO_OPT_ACK"
		// }
	*/

	query := `SELECT TO_CHAR(NOW(), 'DD-Mon-YYYY HH24:MI:SS') AS c_tm_stmp`

	row := cpcm.Db.Raw(query).Row()
	err := row.Scan(&c_tm_stmp)

	if err != nil {
		cpcm.LoggerManager.LogError(cpcm.ServiceName, " [fnRjctRcrd] [Error: getting the system time: %v", err)
		return -1
	}

	c_tm_stmp = strings.TrimSpace(c_tm_stmp)

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRjctRcrd] Current timestamp: %s", c_tm_stmp)

	cpcm.Xchngbook.C_plcd_stts = "REJECT"
	cpcm.Xchngbook.C_rms_prcsd_flg = "NOT_PROCESSED"
	cpcm.Xchngbook.L_ors_msg_typ = util.ORS_NEW_ORD_RJCT
	cpcm.Xchngbook.C_ack_tm = c_tm_stmp
	cpcm.Xchngbook.C_xchng_rmrks = "Token id not available"
	cpcm.Xchngbook.D_jiffy = 0
	cpcm.Xchngbook.C_oprn_typ = "UPDATION_ON_EXCHANGE_RESPONSE"

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRjctRcrd] Before calling 'fnUpdXchngbk' on 'Reject Record' ")

	resultTmp = cpcm.fnUpdXchngbk()

	if resultTmp != 0 {
		cpcm.LoggerManager.LogError(cpcm.ServiceName, " [fnRjctRcrd] [Error: returned from 'fnUpdXchngbk' with an Error")
		cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRjctRcrd] exiting from 'fnRjctRcrd'")
		return -1
	}

	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRjctRcrd] Time Before calling 'fnUpdXchngbk' : %s ", c_tm_stmp)
	cpcm.LoggerManager.LogInfo(cpcm.ServiceName, " [fnRjctRcrd] Function 'fnRjctRcrd' ends successfully")
	return 0
}
