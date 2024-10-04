package exchange_connection

import (
	"DATA_FWD_TAP/internal/models"
	"DATA_FWD_TAP/util"
	"DATA_FWD_TAP/util/MessageQueue"
	typeconversionutil "DATA_FWD_TAP/util/TypeConversionUtil"
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	"gorm.io/gorm"
)

type LogOffFromTapManager struct {
	DB                    *gorm.DB
	LoggerManager         *util.LoggerManager
	ServiceName           string
	MtypeWrite            *int
	Int_header            *models.St_int_header
	St_net_hdr            *models.St_net_hdr
	Exch_msg_Log_Off      *models.St_exch_msg_Log_Off
	St_req_q_data         *models.St_req_q_data_Log_Off
	TCUM                  *typeconversionutil.TypeConversionUtilManager
	Message_queue_manager *MessageQueue.MessageQueueManager
	C_pipe_id             string
	//--------------------------------
	Opm_XchngCd string
	Opm_TrdrID  string
	//---------------------------------------
	Exg_NxtTrdDate string
	Args           []string
	Max_Pack_Val   int // Maximum number of message that can be written to a queue .
}

func (LOFTM *LogOffFromTapManager) Fn_logoff_from_TAP() int {

	LOFTM.C_pipe_id = LOFTM.Args[3] // temporary

	query1 := `
	SELECT opm_xchng_cd, opm_trdr_id
	FROM opm_ord_pipe_mstr
	WHERE opm_pipe_id = ?`

	var localOpmXchngCd string
	var localOpmTrdrID string

	err1 := LOFTM.DB.Raw(query1, LOFTM.C_pipe_id).Row().Scan(&localOpmXchngCd, &localOpmTrdrID)

	if err1 != nil {
		LOFTM.LoggerManager.LogError(LOFTM.ServiceName, "[Fn_logoff_from_TAP] Error fetching opm_xchng_cd: %v", err1)
		return -1
	}

	LOFTM.Opm_XchngCd = localOpmXchngCd
	LOFTM.Opm_TrdrID = localOpmTrdrID

	query2 := `
	SELECT exg_nxt_trd_dt
	FROM exg_xchng_mstr
	WHERE exg_xchng_cd = ?`

	var localExgNxtTrdDate string

	err2 := LOFTM.DB.Raw(query2, localOpmXchngCd).Scan(&localExgNxtTrdDate).Error
	if err2 != nil {
		LOFTM.LoggerManager.LogError(LOFTM.ServiceName, "[Fn_logoff_from_TAP] Error fetching exg_nxt_trd_dt: %v", err2)
		return -1
	}

	LOFTM.Exg_NxtTrdDate = localExgNxtTrdDate

	LOFTM.LoggerManager.LogInfo(LOFTM.ServiceName, "[Fn_logoff_from_TAP] The Trade date is: %s", LOFTM.Exg_NxtTrdDate)

	value, err := strconv.Atoi(strings.TrimSpace(LOFTM.Opm_TrdrID))
	if err != nil {
		LOFTM.LoggerManager.LogError(LOFTM.ServiceName, "[fFn_logoff_from_TAP] ERROR: Failed to convert Trader ID to int: %v", err)
		return -1
	}

	/********************** Header Starts ********************/

	LOFTM.Int_header.Li_trader_id = int32(value)                                               // 1 HDR
	LOFTM.Int_header.Li_log_time = 0                                                           // 2 HDR
	LOFTM.TCUM.CopyAndFormatSymbol(LOFTM.Int_header.C_alpha_char[:], util.LEN_ALPHA_CHAR, " ") // 3 HDR
	LOFTM.Int_header.Si_transaction_code = util.SIGN_OFF_REQUEST_IN                            // 4 HDR
	LOFTM.Int_header.Si_error_code = 0                                                         // 5 HDR
	copy(LOFTM.Int_header.C_filler_2[:], "        ")                                           // 6 HDR
	copy(LOFTM.Int_header.C_time_stamp_1[:], defaultTimeStamp)                                 // 7 HDR
	copy(LOFTM.Int_header.C_time_stamp_2[:], defaultTimeStamp)                                 // 8 HDR
	LOFTM.Int_header.Si_message_length = int16(reflect.TypeOf(LOFTM.Int_header).Size())        // 9 HDR

	LOFTM.LoggerManager.LogInfo(LOFTM.ServiceName, " [LogOffFromTap] `Int_header` Li_trader_id: %d", LOFTM.Int_header.Li_trader_id)
	LOFTM.LoggerManager.LogInfo(LOFTM.ServiceName, " [LogOffFromTap] `Int_header` Li_log_time: %d", LOFTM.Int_header.Li_log_time)
	LOFTM.LoggerManager.LogInfo(LOFTM.ServiceName, " [LogOffFromTap] `Int_header` C_alpha_char: %s", LOFTM.Int_header.C_alpha_char)
	LOFTM.LoggerManager.LogInfo(LOFTM.ServiceName, " [LogOffFromTap] `Int_header` Si_transaction_code: %d", LOFTM.Int_header.Si_transaction_code)
	LOFTM.LoggerManager.LogInfo(LOFTM.ServiceName, " [LogOffFromTap] `Int_header` Si_error_code: %d", LOFTM.Int_header.Si_error_code)
	LOFTM.LoggerManager.LogInfo(LOFTM.ServiceName, " [LogOffFromTap] `Int_header` C_filler_2: %s", LOFTM.Int_header.C_filler_2)
	LOFTM.LoggerManager.LogInfo(LOFTM.ServiceName, " [LogOffFromTap] `Int_header` C_time_stamp_1: %s", LOFTM.Int_header.C_time_stamp_1)
	LOFTM.LoggerManager.LogInfo(LOFTM.ServiceName, " [LogOffFromTap] `Int_header` C_time_stamp_2: %s", LOFTM.Int_header.C_time_stamp_2)
	LOFTM.LoggerManager.LogInfo(LOFTM.ServiceName, " [LogOffFromTap] `Int_header` Si_message_length: %d", LOFTM.Int_header.Si_message_length)

	/********************** Header Done ********************/

	/********************** Net_Hdr Starts ********************/

	// Generate the sequence number
	LOFTM.Exg_NxtTrdDate = strings.TrimSpace(LOFTM.Exg_NxtTrdDate)

	LOFTM.St_net_hdr.S_message_length = int16(unsafe.Sizeof(LOFTM.St_net_hdr) + unsafe.Sizeof(LOFTM.Int_header)) // 1 NET_FDR
	LOFTM.St_net_hdr.I_seq_num = LOFTM.TCUM.GetResetSequence(LOFTM.DB, LOFTM.C_pipe_id, LOFTM.Exg_NxtTrdDate)    // 2 NET_HDR

	hasher := md5.New()
	Int_headerString, err := json.Marshal(LOFTM.Int_header) // to convert the structure in string
	if err != nil {
		LOFTM.LoggerManager.LogError(LOFTM.ServiceName, " [ Error: failed to convert Int_header to string: %v", err)
	}
	io.WriteString(hasher, string(Int_headerString))

	checksum := hasher.Sum(nil)
	copy(LOFTM.St_net_hdr.C_checksum[:], fmt.Sprintf("%x", checksum)) // 3 NET_HDR

	LOFTM.LoggerManager.LogInfo(LOFTM.ServiceName, " [Fn_logoff_from_TAP] `St_net_hdr` S_message_length: %d", LOFTM.St_net_hdr.S_message_length)
	LOFTM.LoggerManager.LogInfo(LOFTM.ServiceName, " [Fn_logoff_from_TAP] `St_net_hdr` I_seq_num: %d", LOFTM.St_net_hdr.I_seq_num)
	LOFTM.LoggerManager.LogInfo(LOFTM.ServiceName, " [Fn_logoff_from_TAP] `St_net_hdr` C_checksum: %s", string(LOFTM.St_net_hdr.C_checksum[:]))

	/********************** Net_Hdr Done ********************/

	// Fill request queue data
	LOFTM.St_req_q_data.L_msg_type = util.SIGN_OFF_REQUEST_IN

	buf := new(bytes.Buffer)

	// Write Int_header
	if err := LOFTM.TCUM.WriteAndCopy(buf, *LOFTM.Int_header, LOFTM.Exch_msg_Log_Off.St_int_header[:]); err != nil {
		LOFTM.LoggerManager.LogError(LOFTM.ServiceName, "[Fn_logoff_from_TAP] [Error: Failed to write Int_header: %v", err)
		return -1
	}

	// Write Net_header
	if err := LOFTM.TCUM.WriteAndCopy(buf, *LOFTM.St_net_hdr, LOFTM.Exch_msg_Log_Off.St_net_header[:]); err != nil {
		LOFTM.LoggerManager.LogError(LOFTM.ServiceName, "[Fn_logoff_from_TAP] [Error: Failed to write Net_header: %v", err)
		return -1
	}

	// Write exchange message
	if err := LOFTM.TCUM.WriteAndCopy(buf, *LOFTM.Exch_msg_Log_Off, LOFTM.St_req_q_data.St_exch_msg_Log_Off[:]); err != nil {
		LOFTM.LoggerManager.LogError(LOFTM.ServiceName, "[Fn_logoff_from_TAP] [Error: Failed to write exch_msg: %v", err)
		return -1
	}

	// Validate Message Queue Manager
	if LOFTM.Message_queue_manager == nil {
		LOFTM.LoggerManager.LogError(LOFTM.ServiceName, "[Error: LOFTM.Message_queue_manager is nil]")
		return -1
	}

	// Create message queue
	LOFTM.Message_queue_manager.LoggerManager = LOFTM.LoggerManager
	if LOFTM.Message_queue_manager.CreateQueue(util.ORDINARY_ORDER_QUEUE_ID) != 0 {
		LOFTM.LoggerManager.LogError(LOFTM.ServiceName, "[LogOffFromTap] [Error: Returning from 'CreateQueue' with an Error")
		LOFTM.LoggerManager.LogInfo(LOFTM.ServiceName, "[LogOffFromTap] Exiting from function")
		return -1
	} else {
		LOFTM.LoggerManager.LogInfo(LOFTM.ServiceName, "[LogOffFromTap] Created Message Queue Successfully")
	}

	// Wait for queue to be writable
	for {
		if LOFTM.Message_queue_manager.FnCanWriteToQueue() != 0 {
			LOFTM.LoggerManager.LogError(LOFTM.ServiceName, "[LogOffFromTap] [Error: Queue is Full")
			continue
		} else {
			break
		}
	}

	// Check if mTypeWrite is nil
	if LOFTM.MtypeWrite == nil {
		LOFTM.LoggerManager.LogError(LOFTM.ServiceName, "[Error: LOFTM.MtypeWrite is nil]")
		return -1
	}

	mtype := *LOFTM.MtypeWrite

	// Write to queue
	if LOFTM.Message_queue_manager.WriteToQueue(mtype, LOFTM.St_req_q_data) != 0 {
		LOFTM.LoggerManager.LogError(LOFTM.ServiceName, "[LogOffFromTap] [Error: Failed to write to queue with message type %d", mtype)
		return -1
	}

	log.Println(LOFTM.St_req_q_data)
	LOFTM.LoggerManager.LogInfo(LOFTM.ServiceName, "[LogOffFromTap] Successfully wrote to queue with message type %d", mtype)

	/********************* Below Code is only for Testing remove that after Testing*************/

	R_L_msg_type, receivedType, readErr := LOFTM.Message_queue_manager.ReadFromQueue(mtype)
	if readErr != 0 {
		LOFTM.LoggerManager.LogError(LOFTM.ServiceName, " [Fn_logoff_from_TAP] [Error:  Failed to read from queue with message type %d: %d", mtype, readErr)
		return -1
	}
	LOFTM.LoggerManager.LogInfo(LOFTM.ServiceName, " [Fn_logoff_from_TAP] Successfully read from queue with message type %d, received type: %d", mtype, receivedType)

	fmt.Println("Li Message Type:", R_L_msg_type)

	*LOFTM.MtypeWrite++

	if *LOFTM.MtypeWrite > LOFTM.Max_Pack_Val {
		*LOFTM.MtypeWrite = 1
	}

	return 0
}
