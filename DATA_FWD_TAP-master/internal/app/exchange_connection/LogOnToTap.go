package exchange_connection

import (
	"bytes"
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	"DATA_FWD_TAP/internal/models"
	"DATA_FWD_TAP/util"
	"DATA_FWD_TAP/util/MessageQueue"
	"DATA_FWD_TAP/util/OrderConversion"
	typeconversionutil "DATA_FWD_TAP/util/TypeConversionUtil"

	"gorm.io/gorm"
)

const defaultTimeStamp = "0"

type LogOnToTapManager struct {
	ServiceName                string
	St_sign_on_req             *models.St_sign_on_req
	St_req_q_data              *models.St_req_q_data_Log_On
	St_exch_msg_Log_On         *models.St_exch_msg_Log_On
	Int_header                 *models.St_int_header
	St_net_hdr                 *models.St_net_hdr
	St_BrokerEligibilityPerMkt *models.St_broker_eligibility_per_mkt
	EM                         *util.EnvironmentManager
	PUM                        *util.PasswordUtilManger // initialize it where ever you are initializing 'LOTTM'
	DB                         *gorm.DB
	LoggerManager              *util.LoggerManager
	OCM                        *OrderConversion.OrderConversionManager
	TCUM                       *typeconversionutil.TypeConversionUtilManager
	Message_queue_manager      *MessageQueue.MessageQueueManager
	C_pipe_id                  string
	UserID                     int64
	MtypeWrite                 *int
	Max_Pack_Val               int
	//----------------------------------
	Opm_loginStatus    int
	Opm_userID         int64
	Opm_existingPasswd string
	Opm_newPasswd      sql.NullString
	Opm_LstPswdChgDt   string
	Opm_XchngCd        string
	Opm_TrdrID         string
	Opm_BrnchID        int64
	//---------------------------------------
	Exg_BrkrID     string
	Exg_NxtTrdDate string // <--
	Exg_BrkrName   string
	Exg_BrkrStts   string
	//-----------------------
	Args []string
}

func (LOTTM *LogOnToTapManager) LogOnToTap() int {
	/*
		what variables i have to find 'from where they are getting value'
		 1. C_pipe_id (in c code it is 'ql_c_ip_pipe_id'. in con_clnt we are setting it from the `args`)
		 2. UserID

		 ** If these values are not found the refer to the Service .
	*/

	//	Step 1: Check the ESR Status.
	//	Step 2: Check password expiration and prompt for a password change if required.
	//	Step 3: Fetch the login status, user ID, password, and new password from the database and validate them.
	//	Step 4: Fetch four fields from opm_ord_pipe_mstr (XchngCd, LstPswdChgDt, TrdrID, BrnchID) and four fields from exg_xchng_mstr (exg_brkr_id, exg_nxt_trd_dt, exg_brkr_name, exg_brkr_stts).
	//	Step 5: Assign the values to the St_sign_on_req structure.
	//	Step 6: Pack all the structures into the final structure, `St_req_q_data``.
	//	Step 7: Write the final structure to the message queue.

	//	Step 1:  Query for checking if ESR is running

	LOTTM.C_pipe_id = LOTTM.Args[3] // temporary
	queryForBPS_Stts := `
		SELECT bps_stts
		FROM bps_btch_pgm_stts
		WHERE bps_pgm_name = ?
		AND bps_xchng_cd = 'NA'
		AND bps_pipe_id = ?
	`

	var bpsStts string
	result := LOTTM.DB.Raw(queryForBPS_Stts, "cln_esr_clnt", LOTTM.C_pipe_id).Scan(&bpsStts)

	if result.Error != nil {
		errorMsg := fmt.Sprintf("Error executing query: %v", result.Error)
		LOTTM.LoggerManager.LogError(LOTTM.ServiceName, "[LogOnToTap] %s", errorMsg)
		return -1
	}

	if bpsStts == "N" {
		errorMsg := fmt.Sprintf("ESR client is not running for Pipe ID: %s and User ID: %d", LOTTM.C_pipe_id, LOTTM.UserID)
		LOTTM.LoggerManager.LogError(LOTTM.ServiceName, "[LogOnToTap] %s", errorMsg)
		return -1
	}

	successMsg := fmt.Sprintf("ESR client is running for Pipe ID: %s and User ID: %d", LOTTM.C_pipe_id, LOTTM.UserID)
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, "[LogOnToTap] %s", successMsg)

	/**** Step2: Checking if user is already logged on in console or not ****/

	var exgMaxPswdVldDays int
	queryForExgMaxPswdVldDays := `
		SELECT exg_max_pswd_vld_days
		FROM opm_ord_pipe_mstr, exg_xchng_mstr
		WHERE opm_pipe_id = ?
		AND opm_xchng_cd = exg_xchng_cd
		AND exg_max_pswd_vld_days > (current_date - opm_lst_pswd_chg_dt);

	`
	errForExgMaxPswdVldDays := LOTTM.DB.Raw(queryForExgMaxPswdVldDays, LOTTM.C_pipe_id).Scan(&exgMaxPswdVldDays).Error

	if errForExgMaxPswdVldDays != nil {
		/**** There is certain time limit of Days for the password after that password will be updated ****/
		if errForExgMaxPswdVldDays == gorm.ErrRecordNotFound {

			LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, "[LogOnToTap] No record found, changing password")

			var currentPassword string
			queryFor_opm_exg_pwd :=
				`SELECT opm_exg_pwd
				FROM opm_ord_pipe_mstr
				WHERE opm_pipe_id = ?
				AND opm_xchng_cd = 'NFO' `

			errFor_opm_exg_pwd := LOTTM.DB.Raw(queryFor_opm_exg_pwd, LOTTM.C_pipe_id).Scan(&currentPassword).Error

			if errFor_opm_exg_pwd != nil {
				LOTTM.LoggerManager.LogError(LOTTM.ServiceName, "[LogOnToTap] Error fetching current password: %v", errFor_opm_exg_pwd)
				return -1
			}

			LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, "Existing password: %s", currentPassword)

			decryptedPassword := LOTTM.PUM.Fndecrypt(currentPassword)

			newPassword, errFor_generateNewPassword := LOTTM.PUM.FngenerateNewPassword(decryptedPassword)
			if errFor_generateNewPassword != nil {

				LOTTM.LoggerManager.LogError(LOTTM.ServiceName, "Failed to generate new password: %v", errFor_generateNewPassword)
				return -1
			}

			LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, "New password generated: %s", newPassword)

			encryptedPassword := LOTTM.PUM.Fnencrypt(newPassword)
			LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, "Encrypted new password: %s", encryptedPassword)

			tx := LOTTM.DB.Begin()

			queryforSET_opm_new_exg_pwd :=
				`	UPDATE opm_ord_pipe_mstr
					SET opm_new_exg_pwd = ?
					WHERE opm_pipe_id = ?
					AND opm_xchng_cd = 'NFO'
				`

			errforSET_opm_new_exg_pwd := tx.Exec(queryforSET_opm_new_exg_pwd, encryptedPassword, LOTTM.C_pipe_id).Error

			if errforSET_opm_new_exg_pwd != nil {
				LOTTM.LoggerManager.LogError(LOTTM.ServiceName, "[LogOnToTap] Error updating new password: %v", errforSET_opm_new_exg_pwd)
				tx.Rollback()
				return -1
			}

			if err := tx.Commit().Error; err != nil {
				LOTTM.LoggerManager.LogError(LOTTM.ServiceName, "[LogOnToTap] Transaction commit failed: %v", err)
				return -1
			}

			LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, "Decrypted new password: %s", LOTTM.PUM.Fndecrypt(encryptedPassword))
			errForPasswordChange := LOTTM.PUM.FnwritePasswordChangeToFile(LOTTM.C_pipe_id, newPassword)
			if errForPasswordChange != nil {
				LOTTM.LoggerManager.LogError("PasswordChange", "Failed to log password change: %v", errForPasswordChange)
			}

			successMsg := fmt.Sprintf("Password changed successfully for Pipe ID: %s", LOTTM.C_pipe_id)
			LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, "[LogOnToTap] %s", successMsg)

		} else {
			LOTTM.LoggerManager.LogError(LOTTM.ServiceName, "[LogOnToTap] SQL error: %v", errForExgMaxPswdVldDays)
			return -1
		}
	}

	// step 3: Fetch login status, user ID, and passwords from the database and validating these values
	queryFor_opm_ord_pipe_mstr_Vars :=
		`SELECT opm_login_stts, opm_user_id, opm_exg_pwd, opm_new_exg_pwd
												FROM opm_ord_pipe_mstr
												WHERE opm_pipe_id = ?
											`
	row := LOTTM.DB.Raw(queryFor_opm_ord_pipe_mstr_Vars, LOTTM.C_pipe_id).Row()

	errFor_opm_ord_pipe_mstr_Vars := row.Scan(&LOTTM.Opm_loginStatus, &LOTTM.Opm_userID, &LOTTM.Opm_existingPasswd, &LOTTM.Opm_newPasswd)

	if errFor_opm_ord_pipe_mstr_Vars != nil {
		LOTTM.LoggerManager.LogError(LOTTM.ServiceName, "SQL Error: %v", errFor_opm_ord_pipe_mstr_Vars)
		return -1
	}

	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, "Fetched new password: %s", LOTTM.Opm_newPasswd.String)

	if LOTTM.UserID != LOTTM.Opm_userID {
		errorMsg := fmt.Sprintf("Incorrect User ID for Pipe ID: %s, expected: %d, got: %d", LOTTM.C_pipe_id, LOTTM.Opm_userID, LOTTM.UserID)
		LOTTM.LoggerManager.LogError(LOTTM.ServiceName, errorMsg)
		return -1
	}

	if LOTTM.Opm_loginStatus == 1 {
		errorMsg := fmt.Sprintf("User is already logged in for Pipe ID: %s, User ID: %d", LOTTM.C_pipe_id, LOTTM.Opm_userID)
		LOTTM.LoggerManager.LogError(LOTTM.ServiceName, errorMsg)
		return -1
	}

	LOTTM.Opm_existingPasswd = strings.TrimSpace(LOTTM.Opm_existingPasswd)
	if LOTTM.Opm_newPasswd.Valid {
		LOTTM.Opm_newPasswd.String = strings.TrimSpace(LOTTM.Opm_newPasswd.String)
	}

	decryptedExistingPasswd := LOTTM.PUM.Fndecrypt(LOTTM.Opm_existingPasswd)
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, "Decrypted existing password: %s", decryptedExistingPasswd)

	if LOTTM.Opm_newPasswd.Valid {
		decryptedNewPasswd := LOTTM.PUM.Fndecrypt(LOTTM.Opm_newPasswd.String)
		LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, "Decrypted new password: %s", decryptedNewPasswd)
	}

	// Step4: we are fetching 4 feilds from `opm_ord_pipe_mstr` (XchngCd, LstPswdChgDt, TrdrID, BrnchID)
	//	and 4 fields from `exg_xchng_mstr` (exg_brkr_id, exg_nxt_trd_dt, exg_brkr_name, exg_brkr_stts)

	queryForOpmOrdPipeMstr := `
		SELECT opm_xchng_cd, 
		       to_char(opm_lst_pswd_chg_dt, 'dd-mon-yyyy'),
		       opm_trdr_id,
		       opm_brnch_id
		FROM opm_ord_pipe_mstr
		WHERE opm_pipe_id = ?
	`
	rowForOpmOrdPipeMstr := LOTTM.DB.Raw(queryForOpmOrdPipeMstr, LOTTM.C_pipe_id).Row()

	errForOpmOrdPipeMstr := rowForOpmOrdPipeMstr.Scan(&LOTTM.Opm_XchngCd, &LOTTM.Opm_LstPswdChgDt, &LOTTM.Opm_TrdrID, &LOTTM.Opm_BrnchID)
	if errForOpmOrdPipeMstr != nil {
		LOTTM.LoggerManager.LogError(LOTTM.ServiceName, "SQL Error: %v", errForOpmOrdPipeMstr)
		return -1
	}

	queryForExgXchngMstr := `
		SELECT exg_brkr_id, 
		       exg_nxt_trd_dt, 
		       COALESCE(exg_brkr_name, ' ') AS exg_brkr_name, 
		       exg_brkr_stts 
		FROM exg_xchng_mstr 
		WHERE exg_xchng_cd = ?
		`
	rowExg := LOTTM.DB.Raw(queryForExgXchngMstr, LOTTM.Opm_XchngCd).Row()

	errForExgXchngMstr := rowExg.Scan(&LOTTM.Exg_BrkrID, &LOTTM.Exg_NxtTrdDate, &LOTTM.Exg_BrkrName, &LOTTM.Exg_BrkrStts)
	if errForExgXchngMstr != nil {
		LOTTM.LoggerManager.LogError(LOTTM.ServiceName, "SQL Error: %v", errForExgXchngMstr)
		return -1
	}

	if LOTTM.Exg_BrkrName == "" {
		LOTTM.Exg_BrkrName = " "
	}

	// Step 5 : assigning the values to the the 'St_sign_on_req'

	/********************** Header Starts ********************/
	traderID, err := strconv.ParseInt(LOTTM.Opm_TrdrID, 10, 64)
	if err != nil {
		LOTTM.LoggerManager.LogError(LOTTM.ServiceName, "Failed to convert Opm_TrdrID to int32: %v", err)
		return -1
	}
	LOTTM.Int_header.Li_trader_id = int32(traderID)                                            // 1 HDR
	LOTTM.Int_header.Li_log_time = 0                                                           // 2 HDR
	LOTTM.TCUM.CopyAndFormatSymbol(LOTTM.Int_header.C_alpha_char[:], util.LEN_ALPHA_CHAR, " ") // 3 HDR
	LOTTM.Int_header.Si_transaction_code = util.SIGN_ON_REQUEST_IN                             // 4 HDR
	LOTTM.Int_header.Si_error_code = 0                                                         // 5 HDR
	copy(LOTTM.Int_header.C_filler_2[:], "        ")                                           // 6 HDR
	copy(LOTTM.Int_header.C_time_stamp_1[:], defaultTimeStamp)                                 // 7 HDR
	copy(LOTTM.Int_header.C_time_stamp_2[:], defaultTimeStamp)                                 // 8 HDR
	LOTTM.Int_header.Si_message_length = int16(reflect.TypeOf(LOTTM.St_sign_on_req).Size())    // 9 HDR

	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `Int_header` Li_trader_id: %d", LOTTM.Int_header.Li_trader_id)
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `Int_header` Li_log_time: %d", LOTTM.Int_header.Li_log_time)
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `Int_header` C_alpha_char: %s", LOTTM.Int_header.C_alpha_char)
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `Int_header` Si_transaction_code: %d", LOTTM.Int_header.Si_transaction_code)
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `Int_header` Si_error_code: %d", LOTTM.Int_header.Si_error_code)
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `Int_header` C_filler_2: %s", LOTTM.Int_header.C_filler_2)
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `Int_header` C_time_stamp_1: %s", LOTTM.Int_header.C_time_stamp_1)
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `Int_header` C_time_stamp_2: %s", LOTTM.Int_header.C_time_stamp_2)
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `Int_header` Si_message_length: %d", LOTTM.Int_header.Si_message_length)

	/********************** Header Done ********************/

	/********************** Body Starts ********************/
	LOTTM.St_sign_on_req.Li_user_id = LOTTM.Opm_userID                                                                     // 1 BDY
	LOTTM.TCUM.CopyAndFormatSymbol(LOTTM.St_sign_on_req.C_reserved_1[:], 8, " ")                                           // 2 BDY
	LOTTM.PUM.CopyAndFormatPassword(LOTTM.St_sign_on_req.C_password[:], util.LEN_PASSWORD, LOTTM.Opm_existingPasswd)       // 3 BDY
	LOTTM.TCUM.CopyAndFormatSymbol(LOTTM.St_sign_on_req.C_reserved_2[:], 8, " ")                                           // 4 BDY
	LOTTM.PUM.CopyAndFormatPassword(LOTTM.St_sign_on_req.C_new_password[:], util.LEN_PASSWORD, LOTTM.Opm_newPasswd.String) // 5 BDY
	LOTTM.TCUM.CopyAndFormatSymbol(LOTTM.St_sign_on_req.C_trader_name[:], util.LEN_TRADER_NAME, LOTTM.Opm_TrdrID)          // 6 BDY
	LOTTM.St_sign_on_req.Li_last_password_change_date = 0                                                                  // 7 BDY
	LOTTM.TCUM.CopyAndFormatSymbol(LOTTM.St_sign_on_req.C_broker_id[:], util.LEN_BROKER_ID, LOTTM.Exg_BrkrID)              // 8 BDY
	LOTTM.St_sign_on_req.C_filler_1 = ' '                                                                                  // 9 BDY
	LOTTM.St_sign_on_req.Si_branch_id = int16(LOTTM.Opm_BrnchID)                                                           // 10 BDY

	versionStr := LOTTM.EM.GetProcessSpaceValue("version", "VERSION_NUMBER")
	if versionStr == "" {
		LOTTM.LoggerManager.LogError(LOTTM.ServiceName, "Failed to retrieve VERSION_NUMBER")
		return -1
	}

	parsedValue, err := strconv.ParseInt(versionStr, 10, 64)
	if err != nil {
		LOTTM.LoggerManager.LogError(LOTTM.ServiceName, "[LogOnToTap] [Error: Failed to parse version number: %v", err)
		return -1
	}

	LOTTM.St_sign_on_req.Li_version_number = int32(parsedValue)                            // 11 BDY
	LOTTM.St_sign_on_req.Li_batch_2_start_time = 0                                         // 12 BDY
	LOTTM.St_sign_on_req.C_host_switch_context = ' '                                       // 13 BDY
	LOTTM.TCUM.CopyAndFormatSymbol(LOTTM.St_sign_on_req.C_colour[:], util.LEN_COLOUR, " ") // 14 BDY
	LOTTM.St_sign_on_req.C_filler_2 = ' '                                                  // 15 BDY

	userTypeStr := LOTTM.EM.GetProcessSpaceValue("UserSettings", "USER_TYPE")
	userType, err := strconv.Atoi(userTypeStr)
	if err != nil {
		LOTTM.LoggerManager.LogError(LOTTM.ServiceName, "[LogOnToTap] [Error: Failed to convert USER_TYPE to int: %v", err)
		return -1
	}

	LOTTM.St_sign_on_req.Si_user_type = int16(userType) // 16 BDY
	LOTTM.St_sign_on_req.D_sequence_number = 0          // 17 BDY

	wsClassName := LOTTM.EM.GetProcessSpaceValue("Service", "WSC_NAME")
	if wsClassName == "" {
		LOTTM.LoggerManager.LogError(LOTTM.ServiceName, "[LogOnToTap] [Error: Failed to retrieve WSC_NAME")
		return -1
	}

	LOTTM.TCUM.CopyAndFormatSymbol(LOTTM.St_sign_on_req.C_ws_class_name[:], util.LEN_WS_CLASS_NAME, wsClassName) // 18 BDY
	LOTTM.St_sign_on_req.C_broker_status = ' '                                                                   // 19 BDY
	LOTTM.St_sign_on_req.C_show_index = 'T'                                                                      // 20 BDY

	/***************************** St_broker_eligibility_per_mkt Configuration ***************************************/

	LOTTM.St_BrokerEligibilityPerMkt.ClearFlag(models.Flg_NormalMkt)
	LOTTM.St_BrokerEligibilityPerMkt.ClearFlag(models.Flg_OddlotMkt)
	LOTTM.St_BrokerEligibilityPerMkt.ClearFlag(models.Flg_SpotMkt)
	LOTTM.St_BrokerEligibilityPerMkt.ClearFlag(models.Flg_AuctionMkt)
	LOTTM.St_BrokerEligibilityPerMkt.ClearFlag(models.Flg_BrokerFiller1)
	LOTTM.St_BrokerEligibilityPerMkt.ClearFlag(models.Flg_BrokerFiller2)

	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap]  Printing broker eligibility flags structure...")
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] 'broker eligibility' Flg_NormalMkt is : %d", LOTTM.TCUM.BoolToInt(LOTTM.St_BrokerEligibilityPerMkt.GetFlagValue(models.Flg_NormalMkt)))
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] 'broker eligibility' Flg_OddlotMkt is : %d", LOTTM.TCUM.BoolToInt(LOTTM.St_BrokerEligibilityPerMkt.GetFlagValue(models.Flg_OddlotMkt)))
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] 'broker eligibility' Flg_SpotMkt is : %d", LOTTM.TCUM.BoolToInt(LOTTM.St_BrokerEligibilityPerMkt.GetFlagValue(models.Flg_SpotMkt)))
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] 'broker eligibility' Flg_AuctionMkt is : %d", LOTTM.TCUM.BoolToInt(LOTTM.St_BrokerEligibilityPerMkt.GetFlagValue(models.Flg_AuctionMkt)))
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] 'broker eligibility' Flg_BrokerFiller1 is : %d", LOTTM.TCUM.BoolToInt(LOTTM.St_BrokerEligibilityPerMkt.GetFlagValue(models.Flg_BrokerFiller1)))
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] 'broker eligibility' Flg_BrokerFiller2 is : %d", LOTTM.TCUM.BoolToInt(LOTTM.St_BrokerEligibilityPerMkt.GetFlagValue(models.Flg_BrokerFiller2)))

	/***************************** End of St_broker_eligibility_per_mkt Configuration *********************************/

	LOTTM.St_sign_on_req.Si_member_type = 0      // 21 BDY
	LOTTM.St_sign_on_req.C_clearing_status = ' ' // 22 BDY

	LOTTM.TCUM.CopyAndFormatSymbol(LOTTM.St_sign_on_req.C_broker_name[:], util.LEN_BROKER_NAME, LOTTM.Exg_BrkrName) // 23 BDY
	LOTTM.TCUM.CopyAndFormatSymbol(LOTTM.St_sign_on_req.C_reserved_3[:], 16, " ")                                   // 24 BDY
	LOTTM.TCUM.CopyAndFormatSymbol(LOTTM.St_sign_on_req.C_reserved_4[:], 16, " ")                                   // 25 BDY
	LOTTM.TCUM.CopyAndFormatSymbol(LOTTM.St_sign_on_req.C_reserved_5[:], 16, " ")                                   // 26 BDY

	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `St_sign_on_req` Li_user_id: %d", LOTTM.St_sign_on_req.Li_user_id)
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `St_sign_on_req` C_reserved_1: %s", string(LOTTM.St_sign_on_req.C_reserved_1[:]))
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `St_sign_on_req` C_password: %s", string(LOTTM.St_sign_on_req.C_password[:]))
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `St_sign_on_req` C_reserved_2: %s", string(LOTTM.St_sign_on_req.C_reserved_2[:]))
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `St_sign_on_req` C_new_password: %s", string(LOTTM.St_sign_on_req.C_new_password[:]))
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `St_sign_on_req` C_trader_name: %s", string(LOTTM.St_sign_on_req.C_trader_name[:]))
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `St_sign_on_req` Li_last_password_change_date: %d", LOTTM.St_sign_on_req.Li_last_password_change_date)
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `St_sign_on_req` C_broker_id: %s", string(LOTTM.St_sign_on_req.C_broker_id[:]))
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `St_sign_on_req` C_filler_1: %c", LOTTM.St_sign_on_req.C_filler_1)
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `St_sign_on_req` Si_branch_id: %d", LOTTM.St_sign_on_req.Si_branch_id)
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `St_sign_on_req` Li_version_number: %d", LOTTM.St_sign_on_req.Li_version_number)
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `St_sign_on_req` Li_batch_2_start_time: %d", LOTTM.St_sign_on_req.Li_batch_2_start_time)
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `St_sign_on_req` C_host_switch_context: %c", LOTTM.St_sign_on_req.C_host_switch_context)
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `St_sign_on_req` C_colour: %s", string(LOTTM.St_sign_on_req.C_colour[:]))
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `St_sign_on_req` C_filler_2: %c", LOTTM.St_sign_on_req.C_filler_2)
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `St_sign_on_req` Si_user_type: %d", LOTTM.St_sign_on_req.Si_user_type)
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `St_sign_on_req` D_sequence_number: %f", LOTTM.St_sign_on_req.D_sequence_number)
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `St_sign_on_req` C_ws_class_name: %s", string(LOTTM.St_sign_on_req.C_ws_class_name[:]))
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `St_sign_on_req` C_broker_status: %c", LOTTM.St_sign_on_req.C_broker_status)
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `St_sign_on_req` C_show_index: %c", LOTTM.St_sign_on_req.C_show_index)
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `St_sign_on_req` Si_member_type: %d", LOTTM.St_sign_on_req.Si_member_type)
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `St_sign_on_req` C_clearing_status: %c", LOTTM.St_sign_on_req.C_clearing_status)
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `St_sign_on_req` C_broker_name: %s", string(LOTTM.St_sign_on_req.C_broker_name[:]))
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `St_sign_on_req` C_reserved_3: %s", string(LOTTM.St_sign_on_req.C_reserved_3[:]))
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `St_sign_on_req` C_reserved_4: %s", string(LOTTM.St_sign_on_req.C_reserved_4[:]))
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `St_sign_on_req` C_reserved_5: %s", string(LOTTM.St_sign_on_req.C_reserved_5[:]))

	/********************** Body Done ********************/

	/********************** Net_Hdr Starts ********************/
	LOTTM.OCM.ConvertSignOnReqToNetworkOrder(LOTTM.St_sign_on_req, LOTTM.Int_header) // Here we are converting all the numbers to Network Order

	LOTTM.St_net_hdr.S_message_length = int16(unsafe.Sizeof(LOTTM.St_sign_on_req) + unsafe.Sizeof(LOTTM.St_net_hdr)) // 1 NET_FDR
	LOTTM.St_net_hdr.I_seq_num = LOTTM.TCUM.GetResetSequence(LOTTM.DB, LOTTM.C_pipe_id, LOTTM.Exg_NxtTrdDate)        // 2 NET_HDR

	hasher := md5.New()

	oeReqResString, err := json.Marshal(LOTTM.St_sign_on_req) // to convert the structure in string
	if err != nil {
		LOTTM.LoggerManager.LogError(LOTTM.ServiceName, " [LogOnToTap] [Error: failed to convert St_sign_on_req to string: %v", err)
	}
	io.WriteString(hasher, string(oeReqResString))

	checksum := hasher.Sum(nil)
	copy(LOTTM.St_net_hdr.C_checksum[:], fmt.Sprintf("%x", checksum)) // 3 NET_FDR

	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `St_net_hdr` S_message_length: %d", LOTTM.St_net_hdr.S_message_length)
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `St_net_hdr` I_seq_num: %d", LOTTM.St_net_hdr.I_seq_num)
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] `St_net_hdr` C_checksum: %s", string(LOTTM.St_net_hdr.C_checksum[:]))

	/********************** Net_Hdr Done ********************/

	//Step 6. Packing all the structures in the final structure 'St_sign_on_req'.

	buf := new(bytes.Buffer)

	// Write Int_header and handle error
	if err := LOTTM.TCUM.WriteAndCopy(buf, *LOTTM.Int_header, LOTTM.St_sign_on_req.St_hdr[:]); err != nil {
		LOTTM.LoggerManager.LogError(LOTTM.ServiceName, " [LogOnToTap] [Error: Failed to write Int_header: %v", err)
	}

	if err := LOTTM.TCUM.WriteAndCopy(buf, *LOTTM.St_BrokerEligibilityPerMkt, LOTTM.St_sign_on_req.St_mkt_allwd_lst[:]); err != nil {
		LOTTM.LoggerManager.LogError(LOTTM.ServiceName, " [LogOnToTap] [Error: Failed to write St_mkt_allwd_lst: %v", err)
	}

	// Write St_sign_on_req and handle error
	if err := LOTTM.TCUM.WriteAndCopy(buf, *LOTTM.St_sign_on_req, LOTTM.St_exch_msg_Log_On.St_sign_on_req[:]); err != nil {
		LOTTM.LoggerManager.LogError(LOTTM.ServiceName, " [LogOnToTap] [Error: Failed to write St_sign_on_req: %v", err)
	}

	// Write net_hdr and handle error
	if err := LOTTM.TCUM.WriteAndCopy(buf, *LOTTM.St_net_hdr, LOTTM.St_exch_msg_Log_On.St_net_header[:]); err != nil {
		LOTTM.LoggerManager.LogError(LOTTM.ServiceName, " [LogOnToTap] [Error: Failed to write net_hdr: %v", err)
	}

	// Write exch_msg and handle error
	if err := LOTTM.TCUM.WriteAndCopy(buf, *LOTTM.St_exch_msg_Log_On, LOTTM.St_req_q_data.St_exch_msg_Log_On[:]); err != nil {
		LOTTM.LoggerManager.LogError(LOTTM.ServiceName, " [LogOnToTap] [Error: Failed to write exch_msg: %v", err)
	}

	// Set L_msg_type and log success message
	LOTTM.St_req_q_data.L_msg_type = util.LOGIN_WITHOUT_OPEN_ORDR_DTLS // 1 St_req_q
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, "[LogOnToTap] Data copied into Queue Packet")

	/* Create a message queue using the CreateQueue function from MessageQueue.go file ().
	This function is responsible for creating a system-level queue on Linux. */

	LOTTM.Message_queue_manager.LoggerManager = LOTTM.LoggerManager
	if LOTTM.Message_queue_manager.CreateQueue(util.ORDINARY_ORDER_QUEUE_ID) != 0 { // here we are giving the value of queue id as 'util.ORDINARY_ORDER_QUEUE_ID' because we are going to create only one queue for all the services. (Ordinary order , LogOn , LogOff)
		LOTTM.LoggerManager.LogError(LOTTM.ServiceName, " [LogOnToTap] [Error: Returning from 'CreateQueue' with an Error ")
		LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap]  Exiting from function")
	} else {
		LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] Created Message Queue SuccessFully")
	}

	for {
		if LOTTM.Message_queue_manager.FnCanWriteToQueue() != 0 {
			LOTTM.LoggerManager.LogError(LOTTM.ServiceName, "[LogOnToTap] [Error: Queue is Full ")

			continue
		} else {
			break
		}
	}

	mtype := *LOTTM.MtypeWrite

	if LOTTM.Message_queue_manager.WriteToQueue(mtype, LOTTM.St_req_q_data) != 0 {
		LOTTM.LoggerManager.LogError(LOTTM.ServiceName, " [LogOnToTap] [Error:  Failed to write to queue with message type %d", mtype)
		return -1
	}

	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [LogOnToTap] Successfully wrote to queue with message type %d", mtype)

	/********************* Below Code is only for Testing remove that after Testing*************/

	R_L_msg_type, receivedType, readErr := LOTTM.Message_queue_manager.ReadFromQueue(mtype)
	if readErr != 0 {
		LOTTM.LoggerManager.LogError(LOTTM.ServiceName, " [CLN_PACK_CLNT] [Error:  Failed to read from queue with message type %d: %d", mtype, readErr)
		return -1
	}
	LOTTM.LoggerManager.LogInfo(LOTTM.ServiceName, " [CLN_PACK_CLNT] Successfully read from queue with message type %d, received type: %d", mtype, receivedType)

	fmt.Println("Li Message Type:", R_L_msg_type)

	*LOTTM.MtypeWrite++

	if *LOTTM.MtypeWrite > LOTTM.Max_Pack_Val {
		*LOTTM.MtypeWrite = 1
	}

	return 0
}
