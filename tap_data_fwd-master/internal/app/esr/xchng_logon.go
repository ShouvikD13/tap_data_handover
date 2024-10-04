package esr

/*
import (
	//"bytes"
	//"encoding/binary"

	"DATA_FWD_TAP/internal/models"
	"fmt"
	"log"
	"net"

	"unsafe"
)

func (ESRM *ESRManger) Do_xchng_logon(st_exch_msg models.St_exch_msg_so, Conn net.Conn) error {
	li_business_data_size = int64(unsafe.Sizeof(models.St_sign_on_req{}))
	li_send_tap_msg_size = int64(unsafe.Sizeof(models.St_net_hdr{})) + li_business_data_size

	// Send the logon message
	if err := ESRM.fn_writen(Conn, st_exch_msg, li_send_tap_msg_size); err != nil {
		log.Printf("[%s] []Failed to send exch_msg_data: %v", err)
		return err
	}

	// Wait for the logon response from the receive thread
	log.Printf("Waiting for logon response...")
	logonResponse := <-logonResponseChan // Block until logon response is received
	log.Printf("Logon response received, processing...")

	// Process the logon response (example)
	log.Printf("Transaction Code: %v", logonResponse.St_sign_on_req.StHdr.Si_transaction_code)
	log.Printf("Error Code: %v", logonResponse.St_sign_on_req.StHdr.Si_error_code)

	// Return if error code in logon response indicates failure
	if logonResponse.St_sign_on_req.StHdr.Si_error_code != 0 {
		return fmt.Errorf("logon failed with error code: %v", logonResponse.St_sign_on_req.StHdr.Si_error_code)
	}

	// Logon success
	log.Printf("Logon successful, continuing...")

	return nil
}
*/
