package app

import (
	//"bytes"
	//"encoding/binary"

	"DATA_FWD_TAP/internal/models"
	"DATA_FWD_TAP/util"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"unsafe"

	_ "github.com/lib/pq"
)

func (ESRM *ESRManger) Fn_chng_exp_passwd_req(db *sql.DB, sign_on_resp models.St_sign_on_res, pipe_id sql_c_pipe_id, new_gen_passwd [9]byte) error {

	// Integer variables
	var chVal, tranID int

	// Character arrays (slices) and single character variables
	var curntPasswd [util.LEN_PASS_WD]byte
	var newPasswd [util.LEN_PASS_WD]byte
	var char byte

	// Query to get the current password
	query := "SELECT OPM_EXG_PWD FROM OPM_ORD_PIPE_MSTR WHERE OPM_PIPE_ID = ?"
	err := db.QueryRow(query, pipe_id).Scan(&curntPasswd)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("[%s] Error in getting current password", pipe_id)
			return fmt.Errorf("error in getting current password: %v", err)
		}
		log.Printf("[%s] L31130: %s", pipe_id, err.Error())
		return err
	}

	// Convert the byte array to a string, trim spaces, and copy it back to the byte array
	trimmedPasswd := strings.TrimSpace(string(curntPasswd[:]))
	copy(curntPasswd[:], trimmedPasswd)

	// Log the current password (in real applications, be careful with logging passwords)
	log.Printf("[%s] The decrypted current password: %s", pipe_id, trimmedPasswd)

	// Decrypt the current password
	decryptedPasswd := ESRM.PUM.Fndecrypt(string(curntPasswd[:]))
	copy(curntPasswd[:], decryptedPasswd)

	CRNT_PSSWD_LEN := int(unsafe.Sizeof(curntPasswd))

	// Generate the new password
	newPasswordStr, err := ESRM.PUM.FngenerateNewPassword(string(curntPasswd[:]), CRNT_PSSWD_LEN)
	if err != nil {
		return fmt.Errorf("error generating new password: %v", err)
	}

	// Copy the generated password to the byte array
	newPasswordBytes := []byte(newPasswordStr)
	copy(newPasswd[:], newPasswordBytes)

	// Trim spaces from the new password
	trimmedPasswd = strings.TrimSpace(string(newPasswd[:]))
	copy(newPasswd[:], trimmedPasswd)

	fnUserlog(serviceName, fmt.Sprintf("New Password Generated on old Password Expiry: %s", trimmedPasswd))

	// Encrypt the new password
	encryptedPasswd := ESRM.PUM.Fnencrypt(string(newPasswd[:]))
	copy(newPasswd[:], encryptedPasswd)

	// Log the new password
	fnUserlog(serviceName, fmt.Sprintf("New Password: %s", newPasswd))
	// Begin transaction
	iTrnsctn := ESRM.TM.FnBeginTran()
	if iTrnsctn == -1 {
		cErrMsg := fmt.Sprintf("[Error in FnBeginTran]")
		fnErrLog(cServiceName, "L31090", "LIBMSG", &cErrMsg)
		return nil
	}

	// Execute the first SQL update statement
	updateStmt := `
	UPDATE OPM_ORD_PIPE_MSTR
	SET OPM_EXG_PWD = OPM_NEW_EXG_PWD, OPM_NEW_EXG_PWD = NULL
	WHERE OPM_PIPE_ID = ? AND OPM_XCHNG_CD = ? AND OPM_NEW_EXG_PWD IS NOT NULL`
	_, err = ESRM.TM.Tran.Exec(updateStmt, sqlCPipeID, sqlCXchngCd)
	if err != nil {
		cErrMsg := fmt.Sprintf("[Error in first SQL update: %v]", err)
		fnErrLog(cServiceName, "L31095", "SQLMSG", &cErrMsg)
		ESRM.TM.FnAbortTran()
		return -1
	}

	// Check if broker status is ACTIVE and perform the second update
	if sign_on_resp.C_broker_status == 'A' {
		updateStmt = `
		UPDATE OPM_ORD_PIPE_MSTR
		SET OPM_LST_PSWD_CHG_DT = TO_TIMESTAMP(?, 'DD-Mon-YYYY HH24:MI:SS'),
			OPM_LOGIN_STTS = 1
		WHERE OPM_PIPE_ID = ?`
		_, err = ESRM.TM.Tran.Exec(updateStmt, string(sqlPasswdLstUpdtDt[:]), sqlCPipeID)
		if err != nil {
			cErrMsg := fmt.Sprintf("[Error in second SQL update: %v]", err)
			fnErrLog(cServiceName, "L31100", "SQLMSG", &cErrMsg)
			ESRM.TM.FnAbortTran()
			return -1
		}
	}

	// Perform the third update if broker status matches any of the specified statuses
	if sign_on_resp.C_broker_status == 'A' || sign_on_resp.C_broker_status == 'C' || sign_on_resp.C_broker_status == 'S' {
		updateStmt = `
		UPDATE EXG_XCHNG_MSTR
		SET EXG_BRKR_STTS = ?
		WHERE EXG_XCHNG_CD = ?`
		_, err = tm.Tran.Exec(updateStmt, sign_on_resp.C_broker_status, sqlCXchngCd)
		if err != nil {
			cErrMsg := fmt.Sprintf("[Error in third SQL update: %v]", err)
			fnErrLog(cServiceName, "L31105", "SQLMSG", &cErrMsg)
			ESRM.TM.FnAbortTran()
			return -1
		}
	}

	// Commit transaction
	iChVal := ESRM.TM.FnCommitTran()
	if iChVal == -1 {
		cErrMsg := fmt.Sprintf("[Error in FnCommitTran]")
		fnErrLog(cServiceName, "L31110", "LIBMSG", &cErrMsg)
		return -1
	}

	/* // Decrypt the new password
	decryptedNewPasswd := fnDecrypt(newPasswd[:])
	copy(newPasswd[:], decryptedNewPasswd)

	// Create the command for sending the password change email
	char = '"'
	commandStr := fmt.Sprintf(`ksh snd_passwd_chng_mail.sh %s '%c%s%c'`, pipeID, char, newPasswd, char)
	cmd := exec.Command("sh", "-c", commandStr)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute command: %v", err)
	} */ //Email functionality TBD

	// Copy the new password to newGenPasswd
	copy(newGenPasswd[:], newPasswd[:9])

	return nil
}
