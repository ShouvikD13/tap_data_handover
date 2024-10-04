package app

import (
	//"bytes"
	//"encoding/binary"

	"DATA_FWD_TAP/internal/models"
	"DATA_FWD_TAP/util"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"unsafe"

	_ "github.com/lib/pq"
)

var C_msg [256]byte                      // Equivalent to char[256], represented as a Go string
var CMessage [256]byte                   // Equivalent to char[256], also a Go string
var CBrkrStts byte                       // Equivalent to char (single byte)
var CTmStmp [30]byte                     // Equivalent to char[30], represented as a Go string
var CClrMemStts byte                     // Equivalent to char (single byte)
var CNewGenPasswd [util.LEN_PASS_WD]byte // Equivalent to char[LEN_PASS_WD], represented as a fixed-size byte array
var IChVal int                           // Equivalent to int
var ITrnsctn int
var C_err_msg [256]byte

var sql_passwd_lst_updt_dt [util.LEN_DT]byte
var sql_lst_mkt_cls_dt [util.LEN_DT]byte
var sql_exch_tm [util.LEN_DT]byte

func (ESRM *ESRManger) Fn_sign_on_request_out(sign_on_resp models.St_sign_on_res) error {
	st_msg := models.VwMktMsg{}
	ACTIVE := util.ACTIVE
	CLOSE_OUT := util.CLOSE_OUT
	SYSTEM_ERROR := util.SYSTEM_ERROR

	// Connection string
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
	}

	if sign_on_resp.St_hdr.Si_error_code != 0 {

		/* ptr_sign_on_resp := &sign_on_resp

		ptr_st_err_res := (*&StErrorResponse{})(unsafe.Pointer(ptr_sign_on_resp)) */

		ptr_sign_on_resp := unsafe.Pointer(&sign_on_resp)
		ptr_st_err_res := (*models.StErrorResponse)(ptr_sign_on_resp)

		//C_err_msg = ptr_st_err_res.CErrorMessage
		copy(C_err_msg[:], ptr_st_err_res.CErrorMessage[:])

		err = db.QueryRow("SELECT to_char(current_timestamp, 'dd-Mon-yyyy hh24:mi:ss')").Scan(&st_msg.CTmStmp)
		if err != nil {
			log.Fatal(err)
		}

		C_msg := fmt.Sprintf("|%s|%d- %s|", st_msg.CTmStmp, sign_on_resp.St_hdr.Si_error_code, C_err_msg)

		//---------------------------------------------------------------------------------------------------//
		//st_msg.CXchngCd = sql_c_xchng_cd// need to find out what is in xchng_cd and how it gets that

		copy(st_msg.CMsg[:], []byte(C_msg))

		//---------------------------------------------------------------------------------------------------//
		//st_msg.c_msg_id = TRADER_MSG// need to find

		query := "SELECT exg_brkr_id FROM exg_xchng_mstr WHERE exg_xchng_cd = $1"
		err = db.QueryRow(query, st_msg.CXchngCd).Scan(&st_msg.CBrkrId)
		if err != nil {
			if err == sql.ErrNoRows {
				// Handle no rows returned
				log.Printf("S31190", "No broker found", err.Error())
			} else {
				log.Printf("S31190", "SQL execution error", err.Error())
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

		log.Printf("st_sign_on_resp->St_Hdr.Si_error_code::", sign_on_resp.St_hdr.Si_error_code)

		if sign_on_resp.St_hdr.Si_error_code == 16053 {

			// Auto password change functionality
			iChVal1 := ESRM.Fn_chng_exp_passwd_req(db, sign_on_resp, sql_c_pipe_id, CNewGenPasswd, cServiceName, cErrMsg)
			//sql_c_pipe_id another global variable to be found.
			if iChVal1 == -1 {
				log.Printf("L31085", "LIBMSG", *&C_err_msg)
				log.Printf("######################################")
				log.Printf(" Error in request of password change ")
				log.Printf("######################################")
				return nil
			}

			// Message processing
			copy(st_msg.CMsg[:], fmt.Sprintf("|%s- %s %s|", st_msg.CTmStmp, "Your password has been regenerated, Please relogin. New pwd is", CNewGenPasswd))

			iChVal2 := fnAcallSvc(cServiceName, C_err_msg, &st_msg, "vw_mkt_msg", unsafe.Sizeof(st_msg), true, "SFO_UPD_MSG")

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

		if sign_on_resp.C_broker_status == CLOSE_OUT {
			cMsg := fmt.Sprintf("|%s| - Logon successful, Broker ClosedOut|", st_msg.CTmStmp)
			cMessage.Reset()
			cMessage.WriteString("Logon Successful, Broker ClosedOut")
			log.Printf("Message Is: %s", cMsg)
		}

		if sign_on_resp.C_broker_status != ACTIVE && sign_on_resp.C_broker_status != CLOSE_OUT {
			cMsg := fmt.Sprintf("|%s| - Logon successful, Broker suspended|", st_msg.CTmStmp)
			cMessage.Reset()
			cMessage.WriteString("Logon Successful, Broker Suspended")
			log.Printf("Message Is: %s", cMsg)
		}

		if sign_on_resp.C_clearing_status != ACTIVE {
			cMsg := fmt.Sprintf("|%s| - Logon successful, Clearing member suspended|", st_msg.CTmStmp)
			cMessage.Reset()
			cMessage.WriteString("Logon Successful, Clearing member suspended")
			log.Printf("Message Is: %s", cMsg)
		}
		err := db.QueryRow("SELECT to_char(current_timestamp, 'dd-Mon-yyyy hh24:mi:ss')").Scan(&st_msg.CTmStmp)
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
		err = db.QueryRow(query, st_msg.CXchngCd).Scan(&st_msg.CBrkrId)
		if err != nil {
			log.Printf("%s: Error while fetching broker ID: %v", cServiceName, err)
			fnErrLog(cServiceName, "S31190", "SQLMSG", cErrMsg)
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
		fnCpyDdr(st_msg.CRoutCrt[:])//What even is this?
		//-------------------------------------------------------------------------------------//

		// Service call
		iChVal := fnAcallSvc(cServiceName, cErrMsg, &st_msg, "vw_mkt_msg", unsafe.Sizeof(st_msg), true, "SFO_UPD_MSG")
		if iChVal == fmt.Errorf("system error: %d", SYSTEM_ERROR) {
			fnErrLog(cServiceName, "S31005", "LIBMSG", cErrMsg)
			return fmt.Errorf("system error: %d", SYSTEM_ERROR)
		}

		// Post to trigger
		//fnPstTrg(cServiceName, "TRG_ORS_CON", cMsg, sqlCPipeID)// NOT TO BE IMPLENETED

		// Convert timestamps to time arrays
		fnLongToTimeArr(sqlPasswdLstUpdtDt[:], ptrStLogonRes.LiLastPasswordChangeDate)
		fnLongToTimeArr(sqlLstMktClsDt[:], int64(ptrStLogonRes.St_hdr.Li_log_time))
		fnLongToTimeArr(sqlExchTm[:], ptrStLogonRes.LiEndTime)

		// Log the converted times
		log.Printf("%s: Password Last Change Date Is: %s", cServiceName, sqlPasswdLstUpdtDt)
		log.Printf("%s: Market Closing Date Is: %s", cServiceName, sqlLstMktClsDt)
		log.Printf("%s: Exchange Time Is: %s", cServiceName, sqlExchTm)

	}
	// Start transaction
	iTrnsctn := fnBeginTran(cServiceName, &cErrMsg)
	if iTrnsctn == -1 {
		fnErrLog(cServiceName, "L31090", "LIBMSG", &cErrMsg)
		return -1
	}

	// Execute the first SQL update statement
	updateStmt := `
		UPDATE OPM_ORD_PIPE_MSTR
		SET OPM_EXG_PWD = OPM_NEW_EXG_PWD, OPM_NEW_EXG_PWD = NULL
		WHERE OPM_PIPE_ID = $1 AND OPM_XCHNG_CD = $2 AND OPM_NEW_EXG_PWD IS NOT NULL`
	_, err := db.Exec(updateStmt, sqlCPipeID, sqlCXchngCd)
	if err != nil {
		fnErrLog(cServiceName, "L31095", "SQLMSG", &cErrMsg)
		fnAbortTran(cServiceName, iTrnsctn, &cErrMsg)
		return -1
	}

	// Check if broker status is ACTIVE and perform the second update
	if sign_on_resp.C_broker_status == 'A' {
		updateStmt = `
			UPDATE OPM_ORD_PIPE_MSTR
			SET OPM_LST_PSWD_CHG_DT = TO_TIMESTAMP($1, 'DD-Mon-YYYY HH24:MI:SS'),
				OPM_LOGIN_STTS = 1
			WHERE OPM_PIPE_ID = $2`
		_, err = db.Exec(updateStmt, string(sqlPasswdLstUpdtDt[:]), sqlCPipeID)
		if err != nil {
			fnErrLog(cServiceName, "L31100", "SQLMSG", &cErrMsg)
			fnAbortTran(cServiceName, iTrnsctn, &cErrMsg)
			return -1
		}
	}

	// Perform the third update if broker status matches any of the specified statuses
	if sign_on_resp.C_broker_status == 'A' || sign_on_resp.C_broker_status == 'C' || sign_on_resp.C_broker_status == 'S' {
		updateStmt = `
			UPDATE EXG_XCHNG_MSTR
			SET EXG_BRKR_STTS = $1
			WHERE EXG_XCHNG_CD = $2`
		_, err = db.Exec(updateStmt, sign_on_resp.C_broker_status, sqlCXchngCd)
		if err != nil {
			fnErrLog(cServiceName, "L31105", "SQLMSG", &cErrMsg)
			fnAbortTran(cServiceName, iTrnsctn, &cErrMsg)
			return nil
		}
	}

	// Commit transaction
	iChVal := fnCommitTran(cServiceName, iTrnsctn, &cErrMsg)
	if iChVal == -1 {
		fnErrLog(cServiceName, "L31110", "LIBMSG", &cErrMsg)
		return -1
	}

	// Log success message
	fnUserLog(cServiceName, "Sign On Successful: Checking Other Details Finished")
	return nil
}
