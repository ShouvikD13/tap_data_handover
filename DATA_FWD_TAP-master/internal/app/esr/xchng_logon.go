package esr

import (
	//"bytes"
	//"encoding/binary"

	"DATA_FWD_TAP/internal/models"
	"DATA_FWD_TAP/util"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"net"

	"unsafe"
)

func (ESRM *ESRManager) Do_xchng_logon(st_exch_msg models.St_sign_on_req, Conn net.Conn, signalChan chan models.Signal) error {
	li_business_data_size = int64(unsafe.Sizeof(models.St_sign_on_req{}))
	li_send_tap_msg_size = int64(unsafe.Sizeof(models.St_net_hdr{})) + li_business_data_size
	var Li_opm_trdr_id int32

	// Send the logon message
	if err := ESRM.fn_writen(Conn, st_exch_msg, li_send_tap_msg_size); err != nil {
		log.Printf("[%s] []Failed to send exch_msg_data: %v", err)
		return err
	}

	// Wait for the logon response from the receive thread
	log.Printf("Waiting for logon response...")
	// Block until logon response is received
	log.Printf("Logon response received, processing...")

	// // Process the logon response (example)
	// log.Printf("Transaction Code: %v", logonResponse.St_sign_on_req.StHdr.Si_transaction_code)
	// log.Printf("Error Code: %v", logonResponse.St_sign_on_req.StHdr.Si_error_code)

	// // Return if error code in logon response indicates failure
	// if logonResponse.St_sign_on_req.StHdr.Si_error_code != 0 {
	// 	return fmt.Errorf("logon failed with error code: %v", logonResponse.St_sign_on_req.StHdr.Si_error_code)
	// }

	//INSERT THE THREAD HANDLING CHANNEL HERE:
	signal := <-signalChan

	if signal.Message == "LOGON_RESP_RECVD" {
		log.Printf("LOGON Resposnse recived is:", signal.Message)
	} else if signal.Message == "RCV_ERR" || signal.ErrorMessage != "" {
		log.Printf("Error in Recive thread")
	}

	// Populate System Information Request Structure
	var st_sys_inf_req models.St_system_info_req        // Zeroing the struct
	var st_exch_message models.St_exch_msg_sys_info_req // Zeroing the struct
	//digest = Digest{}                       // Zeroing the struct
	var i_send_seq int

	/*** ver 1.7 started ***/
	// Execute SQL Query to fetch OPM_TRDR_ID
	query := `SELECT OPM_TRDR_ID FROM OPM_ORD_PIPE_MSTR 
				  WHERE OPM_PIPE_ID = ? AND OPM_XCHNG_CD = ?`
	err := ESRM.Db.Raw(query, Sql_c_pipe_id, Sql_c_xchng_cd).Scan(&Li_opm_trdr_id)
	if err != nil {
		log.Printf("Error in DB Operation for opm_trdr id in exchng_logo sys_info")
		return nil
	}

	// Assign the fetched trader ID
	st_sys_inf_req.St_Hdr.Li_trader_id = int32(Li_opm_trdr_id)
	/*** ver 1.7 ended ***/

	// Populate other header fields
	st_sys_inf_req.St_Hdr.Li_log_time = 0
	ESRM.TCUM.CopyAndFormatSymbol(st_sys_inf_req.St_Hdr.C_alpha_char[:], util.LEN_ALPHA_CHAR, " ")
	st_sys_inf_req.St_Hdr.Si_transaction_code = util.SYSTEM_INFORMATION_IN
	st_sys_inf_req.St_Hdr.Si_error_code = 0

	ESRM.TCUM.CopyAndFormatSymbol(st_sys_inf_req.St_Hdr.C_filler_2[:], 8, " ")

	// Initialize timestamp fields
	var tempSlice1 = make([]byte, util.LEN_TIME_STAMP)
	copy(st_sys_inf_req.St_Hdr.C_time_stamp_1[:], tempSlice1)

	var tempSlice2 = make([]byte, util.LEN_TIME_STAMP)
	copy(st_sys_inf_req.St_Hdr.C_time_stamp_2[:], tempSlice2)

	// Calculate message length (Assuming size in bytes; adjust as needed)
	st_sys_inf_req.St_Hdr.Si_message_length = int16(unsafe.Sizeof(st_sys_inf_req))
	st_sys_inf_req.Li_last_update_protfolio_time = 0

	// Additional Debug Logging
	// Some log statements are commented out in C; replicate accordingly
	log.Printf("%s st_sys_inf_req.St_Hdr.li_trader_id is :%d:", st_sys_inf_req.St_Hdr.Li_trader_id) //Insert Servicename
	log.Printf("%s st_sys_inf_req.St_Hdr.li_log_time : %d", st_sys_inf_req.St_Hdr.Li_log_time)
	log.Printf("%s st_sys_inf_req.St_Hdr.c_alpha_char : %s", st_sys_inf_req.St_Hdr.C_alpha_char)
	log.Printf("%s st_sys_inf_req.St_Hdr.Si_transaction_code : %d", st_sys_inf_req.St_Hdr.Si_transaction_code)
	log.Printf("%s st_sys_inf_req.St_Hdr.Si_error_code : %d", st_sys_inf_req.St_Hdr.Si_error_code)
	log.Printf("%s st_sys_inf_req.St_Hdr.Si_message_length : %d", st_sys_inf_req.St_Hdr.Si_message_length)
	log.Printf("%s st_sys_inf_req.li_last_update_protfolio_time : %d", st_sys_inf_req.Li_last_update_protfolio_time)

	err3 := ESRM.fn_cnvt_hstnw_sys_info_req(&st_sys_inf_req)
	if err3 != nil {
		log.Printf("Error in converting st_sys_inf_req to network byte order")
	}

	st_exch_message.St_system_info_req = models.St_system_info_req(st_sys_inf_req)

	err1 := ESRM.Fn_get_reset_seq(Sql_c_pipe_id, Sql_C_nxt_trd_date, util.GET_PLACED_SEQ, &i_send_seq)
	if err1 != nil {
		log.Printf("Error in getting sequnece number")
	}

	st_exch_message.St_net_hdr.I_seq_num = int32(i_send_seq)
	log.Printf("The sequence number got is:", i_send_seq)

	// Calculate the MD5 digest for the struct
	digest := md5.Sum([]byte(fmt.Sprintf("%v", st_sys_inf_req)))

	copy(st_exch_message.St_net_hdr.C_checksum[:], digest[:])

	//print the checksum in hexadecimal format, comment out in final implementatiom
	fmt.Printf("MD5 Checksum: %s\n", hex.EncodeToString(st_exch_message.St_net_hdr.C_checksum[:]))

	st_exch_message.St_net_hdr.S_message_length = int16(unsafe.Sizeof(st_exch_message.St_net_hdr) + unsafe.Sizeof(st_sys_inf_req))

	//////////////////////////////////////////////////////////////////////////////////////
	//C CODE MIGHT B EUSED LATER
	/* Ver 1.9 Starts here **/
	/* i_ntwk_hdr_seqno = i_ntwk_hdr_seqno + 1;
	st_exch_message.st_net_header.i_seq_num = i_ntwk_hdr_seqno;
	fn_userlo,"st_net_header.i_seq_num :%d:",st_exch_message.st_net_header.i_seq_num); */
	/** Ver 1.9 Ends Here ***/

	//Business data calcualtions put in here

	//CALL FN WRITEN
	signal = <-signalChan

	if signal.Message == "SID RESP RCVD" {
		log.Printf("System Information Request response Recived:", signal.Message)
	} else if signal.Message == "RCV ERR" || signal.ErrorMessage != "" {
		log.Printf("Error in SID response")
	}

	St_upd_local_db := models.St_update_local_database{}
	st_exch_message_ldb_upd := models.St_exch_msg_ldb_upd{} //make new st_Exh msg for db

	/******* commented in ver 1.7 *****
	  fn_orstonse_char(St_upd_local_db.St_Hdr.c_filler_1[:],
	                  2,
	                  " ",
	                  1)
	*/

	// Assigning fields
	St_upd_local_db.St_Hdr.Li_trader_id = int32(Li_opm_trdr_id) // ver 1.7
	St_upd_local_db.St_Hdr.Li_log_time = 0

	ESRM.TCUM.CopyAndFormatSymbol(St_upd_local_db.St_Hdr.C_alpha_char[:], util.LEN_ALPHA_CHAR, " ")

	St_upd_local_db.St_Hdr.Si_transaction_code = util.UPDATE_LOCALDB_IN
	St_upd_local_db.St_Hdr.Si_error_code = 0

	ESRM.TCUM.CopyAndFormatSymbol(St_upd_local_db.St_Hdr.C_filler_2[:], 8, " ")

	// Zeroing time stamps (already zeroed by struct initialization)
	// If you need to explicitly set them, uncomment the following lines
	/*
		for i := 0; i < LEN_TIME_STAMP; i++ {
			St_upd_local_db.St_Hdr.c_time_stamp_1[i] = 0
			St_upd_local_db.St_Hdr.c_time_stamp_2[i] = 0
		}
	*/

	St_upd_local_db.St_Hdr.Si_message_length = int16(unsafe.Sizeof(St_upd_local_db))

	var dateString string

	dateString = string(Sql_sec_tm[:])
	ESRM.TCUM.TimeArrToLong(dateString, &St_upd_local_db.Li_LastUpdate_Security_Time)

	dateString = string(Sql_part_tm[:])
	ESRM.TCUM.TimeArrToLong(dateString, &St_upd_local_db.Li_LastUpdate_Participant_Time)

	dateString = string(Sql_inst_tm[:])
	ESRM.TCUM.TimeArrToLong(dateString, &St_upd_local_db.Li_LastUpdate_Instrument_Time)

	dateString = string(Sql_idx_tm[:])
	ESRM.TCUM.TimeArrToLong(dateString, &St_upd_local_db.Li_Last_Update_Index_Time)

	St_upd_local_db.C_Request_For_Open_Orders = c_rqst_for_open_ordr
	St_upd_local_db.C_Filler_1 = ' '

	St_upd_local_db.St_Mkt_Stts.Si_normal = si_normal_mkt_stts
	St_upd_local_db.St_Mkt_Stts.Si_oddlot = si_oddmkt_stts
	St_upd_local_db.St_Mkt_Stts.Si_spot = si_spotmkt_stts
	St_upd_local_db.St_Mkt_Stts.Si_auction = si_auctionmkt_stts

	St_upd_local_db.St_ExMkt_Stts.Si_normal = si_exr_mkt_stts
	St_upd_local_db.St_ExMkt_Stts.Si_oddlot = si_exr_oddmkt_stts
	St_upd_local_db.St_ExMkt_Stts.Si_spot = si_exr_spotmkt_stts
	St_upd_local_db.St_ExMkt_Stts.Si_auction = si_exr_auctionmkt_stts

	St_upd_local_db.St_Pl_Mkt_Stts.Si_normal = si_pl_mkt_stts
	St_upd_local_db.St_Pl_Mkt_Stts.Si_oddlot = si_pl_oddmkt_stts
	St_upd_local_db.St_Pl_Mkt_Stts.Si_spot = si_pl_spotmkt_stts
	St_upd_local_db.St_Pl_Mkt_Stts.Si_auction = si_pl_auctionmkt_stts

	/******* commented in ver 1.7 *****
	// fn_userlo, "St_upd_local_db.St_Hdr.c_filler_1 is : %s", St_upd_local_db.St_Hdr.c_filler_1)
	*/

	log.Printf("St_upd_local_db.St_Hdr.li_trader_id is :%d:", St_upd_local_db.St_Hdr.Li_trader_id) // ver 1.7
	log.Printf("St_upd_local_db.St_Hdr.li_log_time : %d", St_upd_local_db.St_Hdr.Li_log_time)
	log.Printf("St_upd_local_db.St_Hdr.c_alpha_char : %s", string(St_upd_local_db.St_Hdr.C_alpha_char[:]))
	log.Printf("St_upd_local_db.St_Hdr.Si_transaction_code : %d", St_upd_local_db.St_Hdr.Si_transaction_code)
	log.Printf("St_upd_local_db.St_Hdr.Si_error_code : %d", St_upd_local_db.St_Hdr.Si_error_code)
	/******* commented in ver 1.7 *****
	// log.Printf("St_upd_local_db.St_Hdr.c_filler_2 : %s", St_upd_local_db.St_Hdr.c_filler_2)
	*/
	log.Printf("St_upd_local_db.St_Hdr.Li_trader_id is :%d:", St_upd_local_db.St_Hdr.Li_trader_id) // ver 1.7
	log.Printf("St_upd_local_db.St_Hdr.Li_log_time : %d", St_upd_local_db.St_Hdr.Li_log_time)
	log.Printf("St_upd_local_db.St_Hdr.C_alpha_char : %s", string(St_upd_local_db.St_Hdr.C_alpha_char[:]))
	log.Printf("St_upd_local_db.St_Hdr.Si_transaction_code : %d", St_upd_local_db.St_Hdr.Si_transaction_code)
	log.Printf("St_upd_local_db.St_Hdr.Si_error_code : %d", St_upd_local_db.St_Hdr.Si_error_code)
	log.Printf("St_upd_local_db.St_Hdr.C_time_stamp_1 : %s", string(St_upd_local_db.St_Hdr.C_time_stamp_1[:]))
	log.Printf("St_upd_local_db.St_Hdr.C_time_stamp_2 : %s", string(St_upd_local_db.St_Hdr.C_time_stamp_2[:]))
	log.Printf("St_upd_local_db.St_Hdr.Si_message_length : %d", St_upd_local_db.St_Hdr.Si_message_length)
	log.Printf("St_upd_local_db.Li_last_update_security_time :%d", St_upd_local_db.Li_LastUpdate_Security_Time)
	log.Printf("St_upd_local_db.Li_last_update_participant_time :%d", St_upd_local_db.Li_LastUpdate_Participant_Time)
	log.Printf("St_upd_local_db.Li_last_update_instrument_time :%d", St_upd_local_db.Li_LastUpdate_Instrument_Time)
	log.Printf("St_upd_local_db.Li_last_update_index_time :%d", St_upd_local_db.Li_Last_Update_Index_Time)
	log.Printf("St_upd_local_db.C_Request_For_Open_Orders :%c", St_upd_local_db.C_Request_For_Open_Orders)
	log.Printf("St_upd_local_db.C_Filler_1 :%c", St_upd_local_db.C_Filler_1)
	log.Printf("NORMAL :%d:", si_normal_mkt_stts)
	log.Printf("St_upd_local_db.St_Mkt_Stts.Si_normal :%d", St_upd_local_db.St_Mkt_Stts.Si_normal)
	log.Printf("St_upd_local_db.St_Mkt_Stts.Si_oddlot :%d", St_upd_local_db.St_Mkt_Stts.Si_oddlot)
	log.Printf("St_upd_local_db.St_Mkt_Stts.Si_spot :%d", St_upd_local_db.St_Mkt_Stts.Si_spot)
	log.Printf("St_upd_local_db.St_Mkt_Stts.Si_auction :%d", St_upd_local_db.St_Mkt_Stts.Si_auction)
	log.Printf("EXER :%d:", si_exr_mkt_stts)
	log.Printf("St_upd_local_db.St_ExMkt_Stts.Si_normal :%d", St_upd_local_db.St_ExMkt_Stts.Si_normal)
	log.Printf("St_upd_local_db.St_ExMkt_Stts.Si_oddlot :%d", St_upd_local_db.St_ExMkt_Stts.Si_oddlot)
	log.Printf("St_upd_local_db.St_ExMkt_Stts.Si_spot :%d", St_upd_local_db.St_ExMkt_Stts.Si_spot)
	log.Printf("St_upd_local_db.St_ExMkt_Stts.Si_auction :%d", St_upd_local_db.St_ExMkt_Stts.Si_auction)
	log.Printf("PL :%d:", si_pl_mkt_stts)
	log.Printf("St_upd_local_db.St_Pl_Mkt_Stts.Si_normal :%d", St_upd_local_db.St_Pl_Mkt_Stts.Si_normal)
	log.Printf("St_upd_local_db.St_Pl_Mkt_Stts.Si_oddlot :%d", St_upd_local_db.St_Pl_Mkt_Stts.Si_oddlot)
	log.Printf("St_upd_local_db.St_Pl_Mkt_Stts.Si_spot :%d", St_upd_local_db.St_Pl_Mkt_Stts.Si_spot)
	log.Printf("St_upd_local_db.St_Pl_Mkt_Stts.Si_auction :%d", St_upd_local_db.St_Pl_Mkt_Stts.Si_auction)

	err4 := ESRM.fn_cnvt_hs_t_nw(&St_upd_local_db)
	if err4 != nil {
		log.Printf("Error in host to network conversion in ldb_upd", err4)
	}

	st_exch_message_ldb_upd.St_ldb_upd = St_upd_local_db

	err2 := ESRM.Fn_get_reset_seq(Sql_c_pipe_id, Sql_C_nxt_trd_date, util.GET_PLACED_SEQ, &i_send_seq)
	if err2 != nil {
		log.Printf("Error in getting sequnece number")
	}

	st_exch_message.St_net_hdr.I_seq_num = int32(i_send_seq)
	log.Printf("The sequence number got is:", i_send_seq)

	// Calculate the MD5 digest for the struct
	digest2 := md5.Sum([]byte(fmt.Sprintf("%v", St_upd_local_db)))

	copy(st_exch_message_ldb_upd.St_net_hdr.C_checksum[:], digest2[:])

	//print the checksum in hexadecimal format, comment out in final implementatiom
	fmt.Printf("MD5 Checksum: %s\n", hex.EncodeToString(st_exch_message_ldb_upd.St_net_hdr.C_checksum[:]))

	st_exch_message_ldb_upd.St_net_hdr.S_message_length = int16(unsafe.Sizeof(st_exch_message_ldb_upd.St_net_hdr) + unsafe.Sizeof(St_upd_local_db))

	ESRM.OCM.ConvertNetHeaderToNetworkOrder(&st_exch_message_ldb_upd.St_net_hdr)

	//Calculate Business data sizes and send off to FN_writen

	signal = <-signalChan

	if signal.Message == "UPD LDB RCVD" {
		log.Printf("Local Database Update Request response Recived:", signal.Message)
	} else if signal.Message == "RCV ERR" || signal.ErrorMessage != "" {
		log.Printf("Error in UPD_LDB response")
	}

	log.Printf("SIGN ON PROCESS COMPLETED")
	log.Printf("SIGNED ON SUCESSFULLY")
	log.Printf("Exiting Fn_do_xchng_logon")

	return nil
}

func (ESRM *ESRManager) Fn_get_reset_seq(pipe_id [3]byte, nxt_trd_date [21]byte, SEQutil rune, send_seq *int) error {
	var query string

	log.Printf(" [GetResetSequence] [Executing increment sequence query for pipe ID: %s, trade date: %s]", pipe_id, nxt_trd_date)

	query = `
		UPDATE fsp_fo_seq_plcd
		SET fsp_seq_num = fsp_seq_num + 1
		WHERE fsp_pipe_id = ? AND fsp_trd_dt = TO_DATE(?, 'YYYY-MM-DD')
		RETURNING fsp_seq_num;
	`
	result1 := ESRM.Db.Raw(query, pipe_id, nxt_trd_date).Scan(send_seq)

	if result1.Error != nil {
		log.Printf(" [GetResetSequence] [Error: executing increment sequence query: %v]", result1.Error)
		return nil
	}
	log.Printf(" [GetResetSequence] [Incremented sequence number: %d]", *send_seq)

	if *send_seq == util.MAX_SEQ_NUM {
		log.Printf(" [GetResetSequence] [Sequence number reached MAX_SEQ_NUM, resetting...]")

		query = `
			UPDATE fsp_fo_seq_plcd
			SET fsp_seq_num = 0
			WHERE fsp_pipe_id = ? AND fsp_trd_dt = TO_DATE(?, 'YYYY-MM-DD')
			RETURNING fsp_seq_num;
		`
		result2 := ESRM.Db.Raw(query, pipe_id, nxt_trd_date).Scan(send_seq)

		if result2.Error != nil {
			log.Printf(" [GetResetSequence] [Error: executing reset sequence query: %v]", result2.Error)
			return nil
		}

		log.Printf(" [GetResetSequence] [Sequence number reset to: %d]", *send_seq)
	}

	return nil
}

func (ESRM *ESRManager) fn_cnvt_hs_t_nw(st_upd_ldb *models.St_update_local_database) error {

	ESRM.OCM.ConvertIntHeaderToNetworkOrder(&st_upd_ldb.St_Hdr)

	st_upd_ldb.Li_LastUpdate_Security_Time = ESRM.OCM.ConvertInt32ToNetworkOrder(st_upd_ldb.Li_LastUpdate_Security_Time)
	st_upd_ldb.Li_LastUpdate_Participant_Time = ESRM.OCM.ConvertInt32ToNetworkOrder(st_upd_ldb.Li_LastUpdate_Participant_Time)
	st_upd_ldb.Li_LastUpdate_Instrument_Time = ESRM.OCM.ConvertInt32ToNetworkOrder(st_upd_ldb.Li_LastUpdate_Instrument_Time)
	st_upd_ldb.Li_Last_Update_Index_Time = ESRM.OCM.ConvertInt32ToNetworkOrder(st_upd_ldb.Li_Last_Update_Index_Time)

	st_upd_ldb.St_Mkt_Stts.Si_normal = ESRM.OCM.ConvertInt16ToNetworkOrder(st_upd_ldb.St_Mkt_Stts.Si_normal)
	st_upd_ldb.St_Mkt_Stts.Si_oddlot = ESRM.OCM.ConvertInt16ToNetworkOrder(st_upd_ldb.St_Mkt_Stts.Si_oddlot)
	st_upd_ldb.St_Mkt_Stts.Si_spot = ESRM.OCM.ConvertInt16ToNetworkOrder(st_upd_ldb.St_Mkt_Stts.Si_spot)
	st_upd_ldb.St_Mkt_Stts.Si_auction = ESRM.OCM.ConvertInt16ToNetworkOrder(st_upd_ldb.St_Mkt_Stts.Si_auction)

	st_upd_ldb.St_ExMkt_Stts.Si_normal = ESRM.OCM.ConvertInt16ToNetworkOrder(st_upd_ldb.St_ExMkt_Stts.Si_normal)
	st_upd_ldb.St_ExMkt_Stts.Si_oddlot = ESRM.OCM.ConvertInt16ToNetworkOrder(st_upd_ldb.St_ExMkt_Stts.Si_oddlot)
	st_upd_ldb.St_ExMkt_Stts.Si_spot = ESRM.OCM.ConvertInt16ToNetworkOrder(st_upd_ldb.St_ExMkt_Stts.Si_spot)
	st_upd_ldb.St_ExMkt_Stts.Si_auction = ESRM.OCM.ConvertInt16ToNetworkOrder(st_upd_ldb.St_ExMkt_Stts.Si_auction)

	st_upd_ldb.St_Pl_Mkt_Stts.Si_normal = ESRM.OCM.ConvertInt16ToNetworkOrder(st_upd_ldb.St_Pl_Mkt_Stts.Si_normal)
	st_upd_ldb.St_Pl_Mkt_Stts.Si_oddlot = ESRM.OCM.ConvertInt16ToNetworkOrder(st_upd_ldb.St_Pl_Mkt_Stts.Si_oddlot)
	st_upd_ldb.St_Pl_Mkt_Stts.Si_spot = ESRM.OCM.ConvertInt16ToNetworkOrder(st_upd_ldb.St_Pl_Mkt_Stts.Si_spot)
	st_upd_ldb.St_Pl_Mkt_Stts.Si_auction = ESRM.OCM.ConvertInt16ToNetworkOrder(st_upd_ldb.St_Pl_Mkt_Stts.Si_auction)

	return nil
}

func (ESRM *ESRManager) fn_cnvt_hstnw_sys_info_req(st_sys_inf_req *models.St_system_info_req) error {

	ESRM.OCM.ConvertIntHeaderToNetworkOrder(&st_sys_inf_req.St_Hdr)

	st_sys_inf_req.Li_last_update_protfolio_time = ESRM.OCM.ConvertInt64ToNetworkOrder(st_sys_inf_req.Li_last_update_protfolio_time)

	return nil
}
