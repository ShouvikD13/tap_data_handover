package foexgconlib

import (
	"DATA_FWD_TAP/internal/models"
	"DATA_FWD_TAP/util"
	"DATA_FWD_TAP/util/MessageQueue"
	typeconversionutil "DATA_FWD_TAP/util/TypeConversionUtil"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
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
	C_pipe_id             string
	Mtype                 *int
	int_header            *models.St_int_header
	St_net_hdr            *models.St_net_hdr
	St_req_q_data         *models.St_req_q_data
	TCUM                  *typeconversionutil.TypeConversionUtilManager
	Message_queue_manager *MessageQueue.MessageQueueManager
	//--------------------------------
	Opm_XchngCd string
	Opm_TrdrID  string
	//---------------------------------------
	Exg_NxtTrdDate string
}

func (LOFTM *LogOffFromTapManager) Fn_logoff_from_TAP() int {

	// Fetch opm_xchng_cd and opm_trdr_id
	query1 := `
		SELECT opm_xchng_cd, opm_trdr_id
		FROM opm_ord_pipe_mstr
		WHERE opm_pipe_id = ?`
	err1 := LOFTM.DB.Raw(query1, LOFTM.C_pipe_id).Row().Scan(&LOFTM.Opm_XchngCd, &LOFTM.Opm_TrdrID).Error
	if err1 != nil {
		LOFTM.LoggerManager.LogError(LOFTM.ServiceName, "[Fn_logoff_from_TAP] Error fetching opm_xchng_cd: %v", err1)
		return -1
	}

	// Fetch exg_nxt_trd_date
	query2 := `
		SELECT exg_nxt_trd_dt
		FROM exg_xchng_mstr
		WHERE exg_xchng_cd = ?`
	err2 := LOFTM.DB.Raw(query2, LOFTM.Opm_XchngCd).Scan(&LOFTM.Exg_NxtTrdDate).Error
	if err2 != nil {
		LOFTM.LoggerManager.LogError(LOFTM.ServiceName, "[Fn_logoff_from_TAP] Error fetching exg_nxt_trd_dt: %v", err2)
		return -1
	}

	LOFTM.LoggerManager.LogInfo(LOFTM.ServiceName, "[Fn_logoff_from_TAP] The Trade date is: %s", LOFTM.Exg_NxtTrdDate)

	value, err := strconv.Atoi(strings.TrimSpace(LOFTM.Opm_TrdrID))
	if err != nil {
		LOFTM.LoggerManager.LogError(LOFTM.ServiceName, "[fFn_logoff_from_TAP] ERROR: Failed to convert Trader ID to int: %v", err)
		return -1
	}

	/********************** Header Starts ********************/
	LOFTM.int_header.Li_trader_id = int32(value)                                               // 1 HDR
	LOFTM.int_header.Li_log_time = 0                                                           // 2 HDR
	LOFTM.TCUM.CopyAndFormatSymbol(LOFTM.int_header.C_alpha_char[:], util.LEN_ALPHA_CHAR, " ") // 3 HDR
	LOFTM.int_header.Si_transaction_code = util.SIGN_OFF_REQUEST_IN                            // 4 HDR
	LOFTM.int_header.Si_error_code = 0                                                         // 5 HDR
	copy(LOFTM.int_header.C_filler_2[:], "        ")                                           // 6 HDR
	copy(LOFTM.int_header.C_time_stamp_1[:], defaultTimeStamp)                                 // 7 HDR
	copy(LOFTM.int_header.C_time_stamp_2[:], defaultTimeStamp)                                 // 8 HDR
	LOFTM.int_header.Si_message_length = int16(reflect.TypeOf(LOFTM.int_header).Size())        // 9 HDR

	LOFTM.LoggerManager.LogInfo(LOFTM.ServiceName, " [LogOffFromTap] `int_header` Li_trader_id: %d", LOFTM.int_header.Li_trader_id)
	LOFTM.LoggerManager.LogInfo(LOFTM.ServiceName, " [LogOffFromTap] `int_header` Li_log_time: %d", LOFTM.int_header.Li_log_time)
	LOFTM.LoggerManager.LogInfo(LOFTM.ServiceName, " [LogOffFromTap] `int_header` C_alpha_char: %s", LOFTM.int_header.C_alpha_char)
	LOFTM.LoggerManager.LogInfo(LOFTM.ServiceName, " [LogOffFromTap] `int_header` Si_transaction_code: %d", LOFTM.int_header.Si_transaction_code)
	LOFTM.LoggerManager.LogInfo(LOFTM.ServiceName, " [LogOffFromTap] `int_header` Si_error_code: %d", LOFTM.int_header.Si_error_code)
	LOFTM.LoggerManager.LogInfo(LOFTM.ServiceName, " [LogOffFromTap] `int_header` C_filler_2: %s", LOFTM.int_header.C_filler_2)
	LOFTM.LoggerManager.LogInfo(LOFTM.ServiceName, " [LogOffFromTap] `int_header` C_time_stamp_1: %s", LOFTM.int_header.C_time_stamp_1)
	LOFTM.LoggerManager.LogInfo(LOFTM.ServiceName, " [LogOffFromTap] `int_header` C_time_stamp_2: %s", LOFTM.int_header.C_time_stamp_2)
	LOFTM.LoggerManager.LogInfo(LOFTM.ServiceName, " [LogOffFromTap] `int_header` Si_message_length: %d", LOFTM.int_header.Si_message_length)
	/********************** Header Done ********************/
	/********************** Net_Hdr Starts ********************/
	// Generate the sequence number
	LOFTM.Exg_NxtTrdDate = strings.TrimSpace(LOFTM.Exg_NxtTrdDate)

	LOFTM.St_net_hdr.S_message_length = int16(unsafe.Sizeof(LOFTM.St_net_hdr) + unsafe.Sizeof(LOFTM.int_header)) // 1 NET_FDR
	LOFTM.St_net_hdr.I_seq_num = LOFTM.TCUM.GetResetSequence(LOFTM.DB, LOFTM.C_pipe_id, LOFTM.Exg_NxtTrdDate)    // 2 NET_HDR

	hasher := md5.New()
	int_headerString, err := json.Marshal(LOFTM.int_header) // to convert the structure in string
	if err != nil {
		LOFTM.LoggerManager.LogError(LOFTM.ServiceName, " [ Error: failed to convert int_header to string: %v", err)
	}
	io.WriteString(hasher, string(int_headerString))

	checksum := hasher.Sum(nil)
	copy(LOFTM.St_net_hdr.C_checksum[:], fmt.Sprintf("%x", checksum)) // 3 NET_HDR

	LOFTM.LoggerManager.LogInfo(LOFTM.ServiceName, " [LogOnToTap] `St_net_hdr` S_message_length: %d", LOFTM.St_net_hdr.S_message_length)
	LOFTM.LoggerManager.LogInfo(LOFTM.ServiceName, " [LogOnToTap] `St_net_hdr` I_seq_num: %d", LOFTM.St_net_hdr.I_seq_num)
	LOFTM.LoggerManager.LogInfo(LOFTM.ServiceName, " [LogOnToTap] `St_net_hdr` C_checksum: %s", string(LOFTM.St_net_hdr.C_checksum[:]))
	/********************** Net_Hdr Done ********************/
	// Fill request queue data
	LOFTM.St_req_q_data.L_msg_type = util.SIGN_OFF_REQUEST_IN

	// Here I have to write Another field on St_Req_data for Log_Off

	// Write to message queue
	LOFTM.Message_queue_manager.LoggerManager = LOFTM.LoggerManager
	if LOFTM.Message_queue_manager.CreateQueue(util.ORDINARY_ORDER_QUEUE_ID) != 0 { // here we are giving the value of queue id as 'util.ORDINARY_ORDER_QUEUE_ID' because we are going to create only one queue for all the services. (Ordinary order , LogOn , LogOff)
		LOFTM.LoggerManager.LogError(LOFTM.ServiceName, " [LogOffFromTap] [Error: Returning from 'CreateQueue' with an Error ")
		LOFTM.LoggerManager.LogInfo(LOFTM.ServiceName, " [LogOffFromTap]  Exiting from function")
	} else {
		LOFTM.LoggerManager.LogInfo(LOFTM.ServiceName, " [LogOffFromTap] Created Message Queue SuccessFully")
	}

	for {
		if LOFTM.Message_queue_manager.FnCanWriteToQueue() != 0 {
			LOFTM.LoggerManager.LogError(LOFTM.ServiceName, "[LogOffFromTap] [Error: Queue is Full ")

			continue
		} else {
			break
		}
	}

	mtype := *LOFTM.Mtype

	if LOFTM.Message_queue_manager.WriteToQueue(mtype, LOFTM.St_req_q_data) != 0 {
		LOFTM.LoggerManager.LogError(LOFTM.ServiceName, " [LogOffFromTap] [Error:  Failed to write to queue with message type %d", mtype)
		return -1
	}

	LOFTM.LoggerManager.LogInfo(LOFTM.ServiceName, " [LogOffFromTap] Successfully wrote to queue with message type %d", mtype)

	return 0
}
