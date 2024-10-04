package esr

import (
	"DATA_FWD_TAP/internal/models"
	"DATA_FWD_TAP/util"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"strings"
	"unsafe"

	"gorm.io/gorm"
)

var C_msg [256]byte    // Equivalent to char[256], represented as a Go string
var CMessage [256]byte // Equivalent to char[256], also a Go string
var CBrkrStts byte     // Equivalent to char (single byte)
var CTmStmp [30]byte   // Equivalent to char[30], represented as a Go string
var CClrMemStts byte   // Equivalent to char (single byte)
// var CNewGenPasswd [util.LEN_PASS_WD]byte // Equivalent to char[LEN_PASS_WD], represented as a fixed-size byte array
var IChVal int // Equivalent to int
var ITrnsctn int
var C_err_msg [256]byte
var C_new_gen_password [util.LEN_PASS_WD]byte

var sql_passwd_lst_updt_dt [util.LEN_DT]byte
var sql_lst_mkt_cls_dt [util.LEN_DT]byte
var sql_exch_tm [util.LEN_DT]byte

func (ESRM *ESRManager) Fn_sign_on_request_out(sign_on_resp models.St_sign_on_req_use) error {
	st_msg := models.Vw_mkt_msg{}
	ACTIVE := util.ACTIVE
	CLOSE_OUT := util.CLOSE_OUT
	SYSTEM_ERROR := util.SYSTEM_ERROR

	/* // Connection string
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Open the connection
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Ping the database to ensure a connection is established
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	} */

	if sign_on_resp.St_hdr.Si_error_code != 0 {

		ptr_sign_on_resp := &sign_on_resp

		//ptr_st_err_res := (*&models.St_err_rspns{})(unsafe.Pointer(ptr_sign_on_resp))
		ptr_st_err_res := (*models.St_err_rspns)(unsafe.Pointer(ptr_sign_on_resp))

		//C_err_msg = ptr_st_err_res.CErrorMessage
		copy(C_err_msg[:], ptr_st_err_res.CErrorMessage[:])

		err := ESRM.Db.Raw("SELECT to_char(current_timestamp, 'dd-Mon-yyyy hh24:mi:ss')").Scan(&st_msg.CTmStmp)
		if err != nil {
			log.Printf("Failed to fetch current timestamp: %v", err)
			return fmt.Errorf("system error: %d", fmt.Errorf("system error: %d", SYSTEM_ERROR))
		}

		C_msg := fmt.Sprintf("|%s|%d- %s|", st_msg.CTmStmp, sign_on_resp.St_hdr.Si_error_code, C_err_msg)

		//---------------------------------------------------------------------------------------------------//
		//st_msg.CXchngCd = sql_c_xchng_cd// need to find out what is in xchng_cd and how it gets that

		copy(st_msg.CMsg[:], []byte(C_msg))

		//---------------------------------------------------------------------------------------------------//
		//st_msg.c_msg_id = TRADER_MSG// need to find

		query := "SELECT exg_brkr_id FROM exg_xchng_mstr WHERE exg_xchng_cd = ?"
		result := ESRM.Db.Raw(query, st_msg.CXchngCd).Scan(&st_msg.CBrkrId)

		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				// Handle no rows returned
				log.Printf("S31190: No broker found")
			} else {
				log.Printf("S31190: SQL execution error: %v", result.Error)
				return result.Error // Return the error for further handling
			}

			return nil
		}
		//---------------------------------------------------------------------------------------------------//
		/* if ( SQLCODE != 0 )
		    {
		      fn_errlog(c_ServiceName,"S31190", SQLMSG, c_err_msg);
		      return ( -1 );

			} */
		// Look into this what is SQLCODE?
		//---------------------------------------------------------------------------------------------------//
		//fn_cpy_ddr
		//fn_acall_svc
		iChVal := ESRM.Fn_AcallSvc_Sfo_Udp_Msg(&st_msg)
		if iChVal != nil {
			fnErrLog("SFO_UPD_MSG", "S31005", "LIBMSG", iChVal.Error())
			return iChVal
		}

		log.Printf("st_sign_on_resp->St_Hdr.Si_error_code::", sign_on_resp.St_hdr.Si_error_code)

		if sign_on_resp.St_hdr.Si_error_code == 16053 {

			// Auto password change functionality
			iChVal1 := ESRM.Fn_chng_exp_passwd_req(sql_c_pipe_id, &C_new_gen_password, C_err_msg, sql_c_xchng_cd)
			//sql_c_pipe_id another global variable to be found.
			if iChVal1 == -1 {
				log.Printf("L31085", "LIBMSG", *&C_err_msg)
				log.Printf("######################################")
				log.Printf(" Error in request of password change ")
				log.Printf("######################################")
				return nil
			}

			// Message processing
			copy(st_msg.CMsg[:], fmt.Sprintf("|%s- %s %s|", st_msg.CTmStmp, "Your password has been regenerated, Please relogin. New pwd is", C_new_gen_password))

			iChVal2 := ESRM.Fn_AcallSvc_Sfo_Udp_Msg(&st_msg)
			if iChVal2 == fmt.Errorf("system error: %d", SYSTEM_ERROR) {
				fnErrLog(cServiceName, "S31005", "LIBMSG", cErrMsg)
				return nil
			}
		}

	} else {
		log.Printf("Sign On Successful: Checking Other Details")
		log.Printf("Broker Status Is: %c", sign_on_resp.C_broker_status)
		log.Printf("Clearing Status Is: %c", sign_on_resp.C_clearing_status)

		var cMessage strings.Builder
		cMessage.WriteString("Logon Successful")

		if sign_on_resp.C_broker_status == byte(CLOSE_OUT) {
			cMsg := fmt.Sprintf("|%s| - Logon successful, Broker ClosedOut|", st_msg.CTmStmp)
			cMessage.Reset()
			cMessage.WriteString("Logon Successful, Broker ClosedOut")
			log.Printf("Message Is: %s", cMsg)
		}

		if sign_on_resp.C_broker_status != byte(ACTIVE) && sign_on_resp.C_broker_status != byte(CLOSE_OUT) {
			cMsg := fmt.Sprintf("|%s| - Logon successful, Broker suspended|", st_msg.CTmStmp)
			cMessage.Reset()
			cMessage.WriteString("Logon Successful, Broker Suspended")
			log.Printf("Message Is: %s", cMsg)
		}

		if sign_on_resp.C_clearing_status != byte(ACTIVE) {
			cMsg := fmt.Sprintf("|%s| - Logon successful, Clearing member suspended|", st_msg.CTmStmp)
			cMessage.Reset()
			cMessage.WriteString("Logon Successful, Clearing member suspended")
			log.Printf("Message Is: %s", cMsg)
		}
		err := ESRM.Db.Raw("SELECT to_char(current_timestamp, 'dd-Mon-yyyy hh24:mi:ss')").Scan(&st_msg.CTmStmp)
		if err != nil {
			log.Printf("Failed to fetch current timestamp: %v", err)
			return fmt.Errorf("system error: %d", fmt.Errorf("system error: %d", SYSTEM_ERROR))
		}
		var C_msg strings.Builder

		// Build the string using strings.Builder
		C_msg.WriteString(fmt.Sprintf("|%s|%d- %s|", st_msg.CTmStmp, sign_on_resp.St_hdr.Si_error_code, cMessage.String()))

		// Copy the resulting string into the byte array
		copy(st_msg.CMsg[:], C_msg.String())

		// Set exchange code (assuming sqlCXchngCd is a global or passed variable)
		copy(st_msg.CXchngCd[:], sqlCXchngCd)

		// Set message ID
		//st_msg.CMsgId = TRADER_MSG **find this**
		//----------------------------------------------------------------------------------//

		// Clear broker ID
		for i := range st_msg.CBrkrId {
			st_msg.CBrkrId[i] = 0
		}

		// Query for broker ID
		query := "SELECT exg_brkr_id FROM exg_xchng_mstr WHERE exg_xchng_cd = $1"
		err = ESRM.Db.Raw(query, st_msg.CXchngCd).Scan(&st_msg.CBrkrId)
		if err != nil {
			log.Printf("%s: Error while fetching broker ID: %v", cServiceName, err)
			fnErrLog(cServiceName, "S31190", "SQLMSG", C_err_msg)
			return fmt.Errorf("system error: %d", SYSTEM_ERROR)
		}

		// Log the exchange details
		log.Printf("%s: Exchange Code Is: %s", cServiceName, st_msg.CXchngCd)
		log.Printf("%s: Broker ID Is: %s", cServiceName, st_msg.CBrkrId)
		log.Printf("%s: Message Is: %s", cServiceName, st_msg.CMsg)
		log.Printf("%s: Time Stamp Is: %s", cServiceName, st_msg.CTmStmp)
		log.Printf("%s: Message ID Is: %c", cServiceName, st_msg.CMsgId)
		log.Printf("%s: Before Call to SFO_UPD_MSG", cServiceName)

		// Copy routing criteria
		fnCpyDdr(st_msg.CRoutCrt[:]) //What even is this?
		//-------------------------------------------------------------------------------------//

		// Service call
		iChVal := ESRM.Fn_AcallSvc_Sfo_Udp_Msg(&st_msg)
		if iChVal == fmt.Errorf("system error: %d", SYSTEM_ERROR) {
			fnErrLog(cServiceName, "S31005", "LIBMSG", cErrMsg)
			return fmt.Errorf("system error: %d", SYSTEM_ERROR)
		}

		// Post to trigger
		//fnPstTrg(cServiceName, "TRG_ORS_CON", cMsg, sqlCPipeID)// NOT TO BE IMPLENETED

		// Convert timestamps to int32
		_, temp_lst_updt_dt := ESRM.TCUM.LongToTimeArr(string(sign_on_resp.Li_last_password_change_date))
		_, temp_mkt_cls_dt := ESRM.TCUM.LongToTimeArr(string(sign_on_resp.St_hdr.Li_log_time))
		_, temp_sql_exch_tm := ESRM.TCUM.LongToTimeArr(string(sign_on_resp.Li_end_time))

		//convert to bianry arrays for actual use
		binary.BigEndian.PutUint32(sql_passwd_lst_updt_dt[:], uint32(temp_lst_updt_dt))
		binary.BigEndian.PutUint32(sql_lst_mkt_cls_dt[:], uint32(temp_mkt_cls_dt))
		binary.BigEndian.PutUint32(sql_exch_tm[:], uint32(temp_sql_exch_tm))

		// Log the converted times
		log.Printf("%s: Password Last Change Date Is: %s", cServiceName, sql_passwd_lst_updt_dt)
		log.Printf("%s: Market Closing Date Is: %s", cServiceName, sql_lst_mkt_cls_dt)
		log.Printf("%s: Exchange Time Is: %s", cServiceName, sql_exch_tm)

	}
	// Start transaction
	iTrnsctn := ESRM.TM.FnBeginTran(&cErrMsg)
	if iTrnsctn == -1 {
		fnErrLog(cServiceName, "L31090", "LIBMSG", &cErrMsg)
		return -1
	}

	// Execute the first SQL update statement
	updateStmt := `
	UPDATE OPM_ORD_PIPE_MSTR
	SET OPM_EXG_PWD = OPM_NEW_EXG_PWD, OPM_NEW_EXG_PWD = NULL
	WHERE OPM_PIPE_ID = ? AND OPM_XCHNG_CD = ? AND OPM_NEW_EXG_PWD IS NOT NULL`

	result := ESRM.Db.Exec(updateStmt, sql_c_pipe_id, sql_c_xchng_cd)
	if result.Error != nil {
		fnErrLog(cServiceName, "L31095", "SQLMSG", &cErrMsg)
		ESRM.TM.FnAbortTran(cServiceName, iTrnsctn, &cErrMsg)
		return -1
	}

	// Check if broker status is ACTIVE and perform the second update
	if sign_on_resp.C_broker_status == 'A' {
		updateStmt = `
			UPDATE OPM_ORD_PIPE_MSTR
			SET OPM_LST_PSWD_CHG_DT = TO_TIMESTAMP(?, 'DD-Mon-YYYY HH24:MI:SS'),
				OPM_LOGIN_STTS = 1
			WHERE OPM_PIPE_ID = ?`
		result = ESRM.Db.Exec(updateStmt, string(sql_passwd_lst_updt_dt[:]), sql_c_pipe_id)
		if result.Error != nil {
			fnErrLog(cServiceName, "L31095", "SQLMSG", &cErrMsg)
			ESRM.TM.FnAbortTran(cServiceName, iTrnsctn, &cErrMsg)
			return -1
		}
	}

	// Perform the third update if broker status matches any of the specified statuses
	if sign_on_resp.C_broker_status == 'A' || sign_on_resp.C_broker_status == 'C' || sign_on_resp.C_broker_status == 'S' {
		updateStmt = `
			UPDATE EXG_XCHNG_MSTR
			SET EXG_BRKR_STTS = ?
			WHERE EXG_XCHNG_CD = ?`
		result = ESRM.Db.Exec(updateStmt, sign_on_resp.C_broker_status, sql_c_xchng_cd)
		if result.Error != nil {
			fnErrLog(cServiceName, "L31095", "SQLMSG", &cErrMsg)
			ESRM.TM.FnAbortTran(cServiceName, iTrnsctn, &cErrMsg)
			return -1
		}
	}

	// Commit transaction
	iChVal := ESRM.TM.FnCommitTran(cServiceName, iTrnsctn, &cErrMsg)
	if iChVal == -1 {
		fnErrLog(cServiceName, "L31110", "LIBMSG", &cErrMsg)
		return -1
	}

	// Log success message
	fnUserLog(cServiceName, "Sign On Successful: Checking Other Details Finished")
	return nil
}

func (ESRM *ESRManager) Fn_AcallSvc_Sfo_Udp_Msg(mkt_msg *models.Vw_mkt_msg) error {

	// Start a transaction

	// Insert the record into FTM_FO_TRD_MSG table
	query := `
		INSERT INTO FTM_FO_TRD_MSG
			(FTM_XCHNG_CD, FTM_BRKR_CD, FTM_MSG_ID, FTM_MSG, FTM_TM)
		VALUES (:1, :2, :3, :4, TO_DATE(:5, 'DD-Mon-YYYY HH24:MI:SS'))
	`
	result := ESRM.Db.Exec(query, mkt_msg.CXchngCd, mkt_msg.CBrkrId, mkt_msg.CMsgId, mkt_msg.CMsg, mkt_msg.CTmStmp)
	if result.Error != nil {
		fnErrLog(cServiceName, "L31095", "SQLMSG", &cErrMsg)
		ESRM.TM.FnAbortTran(cServiceName, iTrnsctn, &cErrMsg)
		return -1
	}

	// Commit the transaction

	log.Println("Transaction committed successfully")

	log.Printf("Sign On Successful : Checking Other Details Finished")

	return nil
}
