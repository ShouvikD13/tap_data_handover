package esr

/* import (
	//"bytes"
	//"encoding/binary"

	"DATA_FWD_TAP/util"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	//_ "github.com/lib/pq"
)

func (ESRM *ESRManager) Fn_chng_exp_passwd_req(pipe_id Sql_c_pipe_id, new_gen_passwd *[9]byte, c_err_msg [256]byte, xchng_cd sql_c_xchng_cd) error {

	// Integer variables
	//var chVal, tranID int

	// Character arrays (slices) and single character variables
	var C_curnt_psswd [util.LEN_PASS_WD]byte
	var C_new_passwd [util.LEN_PASS_WD]byte

	// Query to get the current password
	query := "SELECT OPM_EXG_PWD FROM OPM_ORD_PIPE_MSTR WHERE OPM_PIPE_ID = ?"
	result := ESRM.Db.Raw(query, pipe_id)

	// Check if an error occurred
	if err := result.Error; err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("[%s] Error in getting current password", pipe_id)
			return fmt.Errorf("error in getting current password: %v", err)
		}
		log.Printf("[%s] L31130: %s", pipe_id, err.Error())
		return err
	}

	// If no error, scan the result
	if err := result.Scan(&C_curnt_psswd); err != nil { // Reassign to the same err variable
		log.Printf("[%s] L31140: %s", pipe_id, err)
		return nil
	} //Problematic?

	// Convert the byte array to a string, trim spaces, and copy it back to the byte array
	trimmedPasswd := strings.TrimSpace(string(C_curnt_psswd[:]))
	copy(C_curnt_psswd[:], trimmedPasswd)

	// Log the current password (in real applications, be careful with logging passwords)
	log.Printf("[%s] The decrypted current password: %s", pipe_id, trimmedPasswd)

	// Decrypt the current password
	decryptedPasswd := ESRM.PUM.Fndecrypt(string(C_curnt_psswd[:]))
	copy(C_curnt_psswd[:], decryptedPasswd)

	//CRNT_PSSWD_LEN := int(unsafe.Sizeof(C_curnt_psswd))

	// Generate the new password
	newPasswordStr, errr := ESRM.PUM.FngenerateNewPassword(string(C_curnt_psswd[:]))
	if errr != nil {
		// Handle the error appropriately
		log.Printf("Error generating new password: %v", errr)
		return nil
	}

	// Copy the generated password to the byte array
	newPasswordBytes := []byte(newPasswordStr)
	copy(C_new_passwd[:], newPasswordBytes)

	// Trim spaces from the new password
	trimmedPasswd = strings.TrimSpace(string(C_new_passwd[:]))
	copy(C_new_passwd[:], trimmedPasswd)

	fnUserlog(serviceName, fmt.Sprintf("New Password Generated on old Password Expiry: %s", trimmedPasswd))

	// Encrypt the new password
	encryptedPasswd := ESRM.PUM.Fnencrypt(string(C_new_passwd[:]))
	copy(C_new_passwd[:], encryptedPasswd)

	// Log the new password
	fnUserlog(serviceName, fmt.Sprintf("New Password: %s", C_new_passwd))
	// Begin transaction
	iTrnsctn := ESRM.TM.FnBeginTran()
	if iTrnsctn == -1 {
		cErrMsg := fmt.Sprintf("[Error in FnBeginTran]")
		fnErrLog(cServiceName, "L31090", "LIBMSG", &cErrMsg)
		return -1
	}

	// Execute the first SQL update statement
	updateStmt := `
	UPDATE OPM_ORD_PIPE_MSTR
	SET OPM_EXG_PWD = OPM_NEW_EXG_PWD, OPM_NEW_EXG_PWD = NULL
	WHERE OPM_PIPE_ID = ? AND OPM_XCHNG_CD = ? AND OPM_NEW_EXG_PWD IS NOT NULL`
	tx := ESRM.Db.Exec(updateStmt, pipe_id, xchng_cd)
	if tx.Error != nil {
		cErrMsg := fmt.Sprintf("[Error in first SQL update: %v]", tx.Error)
		fnErrLog(cServiceName, "L31095", "SQLMSG", &cErrMsg)
		ESRM.TM.FnAbortTran()
		return nil
	}

	// Commit transaction
	iChVal := ESRM.TM.FnCommitTran()
	if iChVal == -1 {
		cErrMsg := fmt.Sprintf("[Error in FnCommitTran]")
		fnErrLog(cServiceName, "L31110", "LIBMSG", &cErrMsg)
		return -1
	}

	/*	// Decrypt the new password
		decryptedC_new_passwd := fnDecrypt(C_new_passwd[:])
		copy(C_new_passwd[:], decryptedC_new_passwd)

		// Create the command for sending the password change email
		char = '"'
		commandStr := fmt.Sprintf(`ksh snd_passwd_chng_mail.sh %s '%c%s%c'`, pipeID, char, C_new_passwd, char)
		cmd := exec.Command("sh", "-c", commandStr)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to execute command: %v", err)
		} */

// Copy the new password to newGenPasswd
/*copy(new_gen_passwd[:], C_new_passwd[:9])

	return nil
}
*/
