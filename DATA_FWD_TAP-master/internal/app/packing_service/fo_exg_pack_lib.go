package packing_service

import (
	"DATA_FWD_TAP/internal/models"
	"DATA_FWD_TAP/util"
	"DATA_FWD_TAP/util/OrderConversion"
	typeconversionutil "DATA_FWD_TAP/util/TypeConversionUtil"
	"unsafe"

	"bytes"
	"crypto/md5"

	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

const defaultTimeStamp = "0"

type ExchngPackLibMaster struct {
	serviceName   string
	xchngbook     *models.Vw_xchngbook
	orderbook     *models.Vw_orderbook
	pipe_mstr     *models.St_opm_pipe_mstr
	nse_contract  *models.Vw_nse_cntrct
	oe_reqres     *models.St_oe_reqres
	exch_msg      *models.St_exch_msg
	net_hdr       *models.St_net_hdr
	q_packet      *models.St_req_q_data
	int_header    *models.St_int_header
	contract_desc *models.St_contract_desc
	order_flag    *models.St_order_flags
	OCM           *OrderConversion.OrderConversionManager
	cPanNo        [util.LEN_PAN]byte
	cLstActRef    [22]byte
	cEspID        [51]byte
	cAlgoID       [51]byte
	cSourceFlg    byte
	cPrgmFlg      byte
	cUserTypGlb   byte
	LoggerManager *util.LoggerManager
	TCUM          *typeconversionutil.TypeConversionUtilManager
	Mtype         *int
	Db            *gorm.DB
}

// constructor function
func NewExchngPackLibMaster(serviceName string,
	xchngbook *models.Vw_xchngbook,
	orderbook *models.Vw_orderbook,
	pipe_mstr *models.St_opm_pipe_mstr,
	nseContract *models.Vw_nse_cntrct,
	oe_req *models.St_oe_reqres,
	exch_msg *models.St_exch_msg,
	net_hdr *models.St_net_hdr,
	q_packet *models.St_req_q_data,
	int_header *models.St_int_header,
	contract_desc *models.St_contract_desc,
	order_flag *models.St_order_flags,
	OCM *OrderConversion.OrderConversionManager,
	cPanNo, cLstActRef, cEspID, cAlgoID, cSourceFlg, cPrgmFlg string,
	Log *util.LoggerManager,
	TCUM *typeconversionutil.TypeConversionUtilManager,
	mtype *int,
	Db *gorm.DB) *ExchngPackLibMaster {

	var (
		bPanNo     [10]byte
		bLstActRef [22]byte
		bEspID     [51]byte
		bAlgoID    [51]byte
	)

	copy(bPanNo[:], []byte(cPanNo))
	copy(bLstActRef[:], []byte(cLstActRef))
	copy(bEspID[:], []byte(cEspID))
	copy(bAlgoID[:], []byte(cAlgoID))
	bSourceFlg := cSourceFlg[0]
	bPrgmFlg := cPrgmFlg[0]

	return &ExchngPackLibMaster{
		serviceName:   serviceName,
		xchngbook:     xchngbook,
		orderbook:     orderbook,
		pipe_mstr:     pipe_mstr,
		nse_contract:  nseContract,
		oe_reqres:     oe_req,
		exch_msg:      exch_msg,
		net_hdr:       net_hdr,
		q_packet:      q_packet,
		int_header:    int_header,
		contract_desc: contract_desc,
		order_flag:    order_flag,
		OCM:           OCM,
		cPanNo:        bPanNo,
		cLstActRef:    bLstActRef,
		cEspID:        bEspID,
		cAlgoID:       bAlgoID,
		cSourceFlg:    bSourceFlg,
		cPrgmFlg:      bPrgmFlg,
		LoggerManager: Log,
		TCUM:          TCUM,
		Mtype:         mtype,
		Db:            Db,
	}
}

// There is an issue with the field `c_remark`. it is not being initialized anywhere in the original code.
// In the NNF, `c_remark` is mentioned as 'cOrdFiller CHAR 24', but there is no description provided for it.

func (eplm *ExchngPackLibMaster) FnPackOrdnryOrdToNse() int {

	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] Inside 'fnPackOrdnryOrdToNse' ")

	trimmedEspID := strings.TrimSpace(string(eplm.cEspID[:]))
	copy(eplm.cEspID[:], trimmedEspID)
	for i := len(trimmedEspID); i < len(eplm.cEspID); i++ {
		eplm.cEspID[i] = 0 // null in golnag
	}

	trimmedAlgoID := strings.TrimSpace(string(eplm.cAlgoID[:]))
	copy(eplm.cAlgoID[:], trimmedAlgoID)
	for i := len(trimmedAlgoID); i < len(eplm.cAlgoID); i++ {
		eplm.cAlgoID[i] = 0
	}

	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] Trimmed cEspID: '%s'", eplm.cEspID)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] Trimmed cAlgoID: '%s'", eplm.cAlgoID)

	if eplm.orderbook != nil {
		eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] Order book data...")
		eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] Order book c_pro_cli_ind received: '%s'", eplm.orderbook.C_pro_cli_ind)
		eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] Order book c_cln_mtch_accnt received: '%s'", eplm.orderbook.C_cln_mtch_accnt)
		eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] Order book c_ordr_flw received: '%s'", eplm.orderbook.C_ordr_flw)
		eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] Order book l_ord_tot_qty received: '%d'", eplm.orderbook.L_ord_tot_qty)
		eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] Order book l_exctd_qty received: '%d'", eplm.orderbook.L_exctd_qty)
		eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] Order book l_exctd_qty_day received: '%d'", eplm.orderbook.L_exctd_qty_day)
		eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] Order book c_settlor received: '%s'", eplm.orderbook.C_settlor)
		eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] Order book c_ctcl_id received: '%s'", eplm.orderbook.C_ctcl_id)
		eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] Order book c_ack_tm received: '%s'", eplm.orderbook.C_ack_tm)
		eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] Order book c_prev_ack_tm received: '%s'", eplm.orderbook.C_prev_ack_tm)
	} else {
		eplm.LoggerManager.LogError(eplm.serviceName, " [fnPackOrdnryOrdToNse] No data received in orderbook in 'fnPackOrdnryOrdToNse'")
		eplm.LoggerManager.LogError(eplm.serviceName, " [fnPackOrdnryOrdToNse] Exiting from 'fnPackOrdnryOrdToNse'")
		return -1
	}

	if eplm.xchngbook != nil {
		eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] Exchange book data...")
		eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] Exchange book c_req_typ received: '%s'", eplm.xchngbook.C_req_typ)
		eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] Exchange book l_dsclsd_qty received: '%d'", eplm.xchngbook.L_dsclsd_qty)
		eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] Exchange book c_slm_flg received: '%s'", eplm.xchngbook.C_slm_flg)
		eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] Exchange book l_ord_lmt_rt received: '%d'", eplm.xchngbook.L_ord_lmt_rt)
		eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] Exchange book l_stp_lss_tgr received: '%d'", eplm.xchngbook.L_stp_lss_tgr)
		eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] Exchange book c_ord_typ received: '%s'", eplm.xchngbook.C_ord_typ)
		eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] Exchange book l_ord_tot_qty received: '%d'", eplm.xchngbook.L_ord_tot_qty)
		eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] Exchange book c_valid_dt received: '%s'", eplm.xchngbook.C_valid_dt)
		eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] Exchange book l_ord_seq received: '%d'", eplm.xchngbook.L_ord_seq)

	} else {
		eplm.LoggerManager.LogError(eplm.serviceName, " [fnPackOrdnryOrdToNse] No data received in xchngbook in 'fnPackOrdnryOrdToNse'")
		eplm.LoggerManager.LogError(eplm.serviceName, " [fnPackOrdnryOrdToNse] Exiting from 'fnPackOrdnryOrdToNse'")
		return -1
	}

	if eplm.pipe_mstr != nil {
		eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] OPM data...")
		eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] OPM li_opm_brnch_id received: '%d'", eplm.pipe_mstr.L_opm_brnch_id)
		eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] OPM c_opm_trdr_id received: '%s'", eplm.pipe_mstr.C_opm_trdr_id)
		eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] OPM c_xchng_brkr_id received: '%s'", eplm.pipe_mstr.C_xchng_brkr_id)
	} else {
		eplm.LoggerManager.LogError(eplm.serviceName, " [fnPackOrdnryOrdToNse] No data received in 'OPM_PIPE_MSTR' Struct in 'fnPackOrdnryOrdToNse'")
		eplm.LoggerManager.LogError(eplm.serviceName, " [fnPackOrdnryOrdToNse] Exiting from 'fnPackOrdnryOrdToNse'")
		return -1
	}

	if eplm.nse_contract != nil {
		eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] Exchange message structure data...")
		eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] NSE CONTRACT c_prd_typ received: '%s'", eplm.nse_contract.C_prd_typ)
		eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] NSE CONTRACT c_ctgry_indstk received: '%s'", eplm.nse_contract.C_ctgry_indstk)
		eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] NSE CONTRACT c_symbol received: '%s'", eplm.nse_contract.C_symbol)
		eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] NSE CONTRACT l_ca_lvl received: '%d'", eplm.nse_contract.L_ca_lvl)
		eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] NSE CONTRACT l_token_id received: '%d'", eplm.nse_contract.L_token_id)

	} else {
		eplm.LoggerManager.LogError(eplm.serviceName, " [fnPackOrdnryOrdToNse] No data received in 'fnPackOrdnryOrdToNse'")
		eplm.LoggerManager.LogError(eplm.serviceName, " [fnPackOrdnryOrdToNse] Exiting from 'fnPackOrdnryOrdToNse'")
		return -1
	}

	// Starting the intialization of Header

	value, err := strconv.Atoi(strings.TrimSpace(eplm.pipe_mstr.C_opm_trdr_id))
	if err != nil {
		eplm.LoggerManager.LogError(eplm.serviceName, "[fnPackOrdnryOrdToNse] ERROR: Failed to convert Trader ID to int: %v", err)
		return -1
	}

	eplm.int_header.Li_trader_id = int32(value) // 1 HDR

	eplm.int_header.Li_log_time = 0                                                          // 2 HDR
	eplm.TCUM.CopyAndFormatSymbol(eplm.int_header.C_alpha_char[:], util.LEN_ALPHA_CHAR, " ") // 3 HDR 'orstonse'
	eplm.int_header.Si_transaction_code = util.BOARD_LOT_IN                                  // 4 HDR
	copy(eplm.int_header.C_filler_2[:], "        ")                                          // 5 HDR

	eplm.int_header.Si_error_code = 0 // 6 HDR
	/*
		for i := range eplm.int_header.C_time_stamp_1 {
			eplm.int_header.C_time_stamp_1[i] = 0 //7 HDR
		}
		for i := range eplm.int_header.C_time_stamp_2 {
			eplm.int_header.C_time_stamp_2[i] = 0 //8 HDR
	}*/

	copy(eplm.int_header.C_time_stamp_1[:], defaultTimeStamp)                                 //7 HDR
	copy(eplm.int_header.C_time_stamp_2[:], defaultTimeStamp)                                 //8 HDR
	eplm.int_header.Si_message_length = int16(reflect.TypeOf(eplm.exch_msg.St_oe_res).Size()) //9 HDR

	// Header structure Done Till here ----------------------------------------------------------------

	switch eplm.pipe_mstr.S_user_typ_glb {
	case 0:
		eplm.cUserTypGlb = util.TRADER
	case 4:
		eplm.cUserTypGlb = util.CORPORATE_MANAGER
	case 5:
		eplm.cUserTypGlb = util.BRANCH_MANAGER
	default:
		eplm.LoggerManager.LogError(eplm.serviceName, " [fnPackOrdnryOrdToNse] ERROR: Invalid user type, returning from 'fo_exg_pack_lib' with an error..")
		return -1
	}

	eplm.oe_reqres.Ll_lastactivityref = 0                     //1 BDY
	eplm.oe_reqres.C_participant_type = 'C'                   //2 BDY
	eplm.oe_reqres.C_filler_1 = ' '                           //3 BDY
	eplm.oe_reqres.Si_competitor_period = 0                   //4 BDY
	eplm.oe_reqres.Si_solicitor_period = 0                    //5 BDY
	eplm.oe_reqres.C_modified_cancelled_by = eplm.cUserTypGlb //6 BDY
	eplm.oe_reqres.C_filler_2 = ' '                           //7 BDY
	eplm.oe_reqres.Si_reason_code = 0                         //8 BDY
	eplm.oe_reqres.C_filler_3 = [4]byte{' ', ' ', ' ', ' '}   //9 BDY
	eplm.oe_reqres.L_token_no = eplm.nse_contract.L_token_id  //10 BDY

	if eplm.Fn_orstonse_cntrct_desc() != 0 { // contract_desc
		eplm.LoggerManager.LogError(eplm.serviceName, " [fnPackOrdnryOrdToNse] returned from 'Fn_orstonse_cntrct_desc()' with error ")
		eplm.LoggerManager.LogError(eplm.serviceName, " [fnPackOrdnryOrdToNse] Exiting from 'fnPackOrdnryOrdToNse' ")
		return -1
	}

	eplm.TCUM.CopyAndFormatSymbol(eplm.oe_reqres.C_counter_party_broker_id[:], util.LEN_BROKER_ID, " ") // 12 BDY
	eplm.oe_reqres.C_filler_4 = ' '                                                                     // 13 BDY
	eplm.oe_reqres.C_filler_5 = [2]byte{' ', ' '}                                                       // 14 BDY
	eplm.oe_reqres.C_closeout_flg = ' '                                                                 // 15 BDY
	eplm.oe_reqres.C_filler_6 = ' '                                                                     // 16 BDY
	eplm.oe_reqres.Si_order_type = 0                                                                    // 17 BDY

	if eplm.xchngbook.C_req_typ == string(util.NEW) { // 18 BDY
		eplm.oe_reqres.D_order_number = 0
	} else {
		eplm.oe_reqres.D_order_number, err = strconv.ParseFloat(eplm.orderbook.C_xchng_ack, 64)
		if err != nil {
			eplm.LoggerManager.LogError(eplm.serviceName, " [fnPackOrdnryOrdToNse] [ERROR: Error parsing order number: %v", err)
			eplm.LoggerManager.LogError(eplm.serviceName, " [fnPackOrdnryOrdToNse] Returning from 'fnPackOrdnryOrdToNse' with an error ")
			return -1
		}
	}

	eplm.TCUM.CopyAndFormatSymbol(eplm.oe_reqres.C_account_number[:], len(eplm.oe_reqres.C_account_number), eplm.orderbook.C_cln_mtch_accnt) //19 BDY

	if eplm.xchngbook.C_slm_flg == "M" {
		eplm.oe_reqres.Si_book_type = util.REGULAR_LOT_ORDER //20 BDY
		eplm.oe_reqres.Li_price = 0                          //21 BDY
	} else {
		eplm.LoggerManager.LogError(eplm.serviceName, " [fnPackOrdnryOrdToNse] [ERROR: Invalid slm flag ...")
		return -1
	}

	switch eplm.orderbook.C_ordr_flw { //22 BDY
	case "B":
		eplm.oe_reqres.Si_buy_sell_indicator = util.NSE_BUY
	case "S":
		eplm.oe_reqres.Si_buy_sell_indicator = util.NSE_SELL
	default:
		eplm.LoggerManager.LogError(eplm.serviceName, " [fnPackOrdnryOrdToNse] [ERROR: Invalid order flow flag ...")
		return -1
	}

	eplm.oe_reqres.Li_disclosed_volume = eplm.xchngbook.L_dsclsd_qty                     // 23 BDY
	eplm.oe_reqres.Li_disclosed_volume_remaining = eplm.oe_reqres.Li_disclosed_volume    //24 BDY
	eplm.oe_reqres.Li_volume = eplm.xchngbook.L_ord_tot_qty - eplm.orderbook.L_exctd_qty //25 BDY

	if eplm.xchngbook.L_ord_tot_qty > eplm.orderbook.L_ord_tot_qty {
		eplm.oe_reqres.Li_total_volume_remaining = eplm.orderbook.L_ord_tot_qty - eplm.orderbook.L_exctd_qty //26 BDY
	} else {
		eplm.oe_reqres.Li_total_volume_remaining = eplm.xchngbook.L_ord_tot_qty - eplm.orderbook.L_exctd_qty
	}

	eplm.oe_reqres.Li_volume_filled_today = eplm.orderbook.L_exctd_qty_day //27 BDY
	eplm.oe_reqres.Li_trigger_price = eplm.xchngbook.L_stp_lss_tgr         //28 BDY

	switch eplm.xchngbook.C_ord_typ { // 29 BDY
	case "T", "I":
		eplm.oe_reqres.Li_good_till_date = 0

	case "D":
		eplm.TCUM.TimeArrToLong(eplm.xchngbook.C_valid_dt, &eplm.oe_reqres.Li_good_till_date)

	default:
		eplm.LoggerManager.LogError(eplm.serviceName, "] [fnPackOrdnryOrdToNse] [ERROR: Invalid order type ...")
		return -1
	}

	if eplm.xchngbook.C_req_typ == string(util.NEW) {
		eplm.oe_reqres.Li_entry_date_time = 0 // 30 BDY
		eplm.oe_reqres.Li_last_modified = 0   //31 BDY
	} else {
		eplm.LoggerManager.LogError(eplm.serviceName, "] [fnPackOrdnryOrdToNse] [ERROR: Invalid Request type for setting 'Entry Date Time'  ...")
		eplm.LoggerManager.LogError(eplm.serviceName, "] [fnPackOrdnryOrdToNse] [ERROR: Invalid Request type for setting 'Last Modified'  ...")
		return -1
	}

	eplm.oe_reqres.Li_minimum_fill_aon_volume = 0 // 32 BDY

	eplm.order_flag.ClearFlag(models.Flg_ATO) // 1 FLG

	eplm.order_flag.ClearFlag(models.Flg_Market) // 2 FLG

	if eplm.xchngbook.C_slm_flg == "M" { // 3 FLG
		eplm.order_flag.SetFlag(models.Flg_SL) // Set flag
	} else {
		eplm.LoggerManager.LogError(eplm.serviceName, " [fnPackOrdnryOrdToNse] [ERROR: Invalid slm flag for setting 'Flg_SL'  ...")
		return -1
	}
	eplm.order_flag.ClearFlag(models.Flg_MIT) // 4 FLG

	if eplm.xchngbook.C_ord_typ == "T" {
		eplm.order_flag.SetFlag(models.Flg_Day) // 5 FLG
	} else {
		eplm.order_flag.ClearFlag(models.Flg_Day) // 6 FLG
	}

	eplm.order_flag.ClearFlag(models.Flg_GTC) // 7 FLG

	if eplm.xchngbook.C_ord_typ == "I" {
		eplm.order_flag.SetFlag(models.Flg_IOC) // 8 FLG
	} else {
		eplm.order_flag.ClearFlag(models.Flg_IOC)
	}

	eplm.order_flag.ClearFlag(models.Flg_AON)        // 9 FLG
	eplm.order_flag.ClearFlag(models.Flg_MF)         // 10 FLG
	eplm.order_flag.ClearFlag(models.Flg_MatchedInd) // 11 FLG
	eplm.order_flag.ClearFlag(models.Flg_Traded)     // 12 FLG
	eplm.order_flag.ClearFlag(models.Flg_Modified)   // 13 FLG
	eplm.order_flag.ClearFlag(models.Flg_Frozen)     // 14 FLG

	// Clear bits 13, 14, and 15 in Flags using Flg_Filler1 bitmask.
	eplm.order_flag.Flags &^= models.Flg_OrderFiller1 // 15 FLG

	eplm.oe_reqres.Si_branch_id = int16(eplm.pipe_mstr.L_opm_brnch_id) //33 BDY

	value1, err := strconv.Atoi(eplm.pipe_mstr.C_opm_trdr_id)
	if err != nil {
		eplm.LoggerManager.LogError(eplm.serviceName, " [fnPackOrdnryOrdToNse] [Error converting C_opm_trdr_id to int:  ... %v", err)
		return -1
	}
	eplm.oe_reqres.Li_trader_id = int32(value1) // 34 BDY

	eplm.TCUM.CopyAndFormatSymbol(eplm.oe_reqres.C_broker_id[:], len(eplm.oe_reqres.C_broker_id), eplm.pipe_mstr.C_xchng_brkr_id) //35 BDY

	eplm.oe_reqres.I_order_seq = eplm.xchngbook.L_ord_seq // 36 BDY
	eplm.oe_reqres.C_open_close = 'O'                     // 37 BDY

	for i := range eplm.oe_reqres.C_settlor {
		eplm.oe_reqres.C_settlor[i] = 0 // 38 BDY
	}

	query := `
		SELECT ICD_CUST_TYPE
		FROM ICD_INFO_CLIENT_DTLS
		JOIN IAI_INFO_ACCOUNT_INFO ON ICD_SERIAL_NO = IAI_SERIAL_NO
		WHERE IAI_MATCH_ACCOUNT_NO = ?
	`
	var icdCustType string
	result := eplm.Db.Raw(query, eplm.orderbook.C_cln_mtch_accnt).Scan(&icdCustType)
	if result.Error != nil {
		eplm.LoggerManager.LogError(eplm.serviceName, " [fnPackOrdnryOrdToNse] [Error executing query: ... %v", result.Error)
		return -1
	}

	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] ICD_CUST_TYPE: %s", icdCustType)
	if icdCustType != "NRI" {
		if eplm.orderbook.C_pro_cli_ind == string(util.BRKR_PLCD) {
			eplm.TCUM.CopyAndFormatSymbol(eplm.oe_reqres.C_settlor[:], util.LEN_SETTLOR, " ") // already set to zero now setting value
			eplm.oe_reqres.Si_pro_client_indicator = util.NSE_PRO                             // 39 BDY
		} else {
			eplm.TCUM.CopyAndFormatSymbol(eplm.oe_reqres.C_settlor[:], util.LEN_SETTLOR, eplm.pipe_mstr.C_xchng_brkr_id)
			eplm.oe_reqres.Si_pro_client_indicator = util.NSE_CLIENT
		}
	} else {
		eplm.TCUM.CopyAndFormatSymbol(eplm.oe_reqres.C_settlor[:], util.LEN_SETTLOR, eplm.orderbook.C_settlor)
		eplm.oe_reqres.Si_pro_client_indicator = util.NSE_CLIENT
	}

	switch eplm.cPrgmFlg {
	case 'A':
		eplm.oe_reqres.L_algo_id = util.FO_AUTO_MTM_ALG_ID            //40 BDY
		eplm.oe_reqres.Si_algo_category = util.FO_AUTO_MTM_ALG_CAT_ID //41 BDY

	case 'T':
		eplm.oe_reqres.L_algo_id = util.FO_PRICE_IMP_ALG_ID
		eplm.oe_reqres.Si_algo_category = util.FO_PRICE_IMP_ALG_CAT_ID

	case 'Z':
		eplm.oe_reqres.L_algo_id = util.FO_PRFT_ORD_ALG_ID
		eplm.oe_reqres.Si_algo_category = util.FO_PRFT_ORD_ALG_CAT_ID

	case 'G':
		eplm.oe_reqres.L_algo_id = util.FO_FLASH_TRD_ALG_ID
		eplm.oe_reqres.Si_algo_category = util.FO_FLASH_TRD_ALG_CAT_ID

	default:
		cAlgoIDStr1 := strings.TrimSpace(string(eplm.cAlgoID[:]))
		if cAlgoIDStr1 != "*" {
			algoID, err := strconv.Atoi(cAlgoIDStr1)
			if err != nil {
				eplm.LoggerManager.LogError(eplm.serviceName, " [ProcessAlgorithmFlags] [Error converting cAlgoID ...  %v", eplm.cUserTypGlb, err)
				return -1
			}
			eplm.oe_reqres.L_algo_id = int32(algoID)
			eplm.oe_reqres.Si_algo_category = 0
		} else {
			eplm.oe_reqres.L_algo_id = util.FO_NON_ALG_ID
			eplm.oe_reqres.Si_algo_category = util.FO_NON_ALG_CAT_ID
		}
	}

	if eplm.cSourceFlg == 'G' {
		eplm.oe_reqres.L_algo_id = util.FO_FOGTT_ALG_ID
		eplm.oe_reqres.Si_algo_category = util.FO_FOGTT_ALG_CAT_ID
	}

	eplm.orderbook.C_ctcl_id = strings.TrimSpace(eplm.orderbook.C_ctcl_id)

	cAlgoIDStr := strings.TrimSpace(string(eplm.cAlgoID[:]))
	cSourceFlgStr := strings.TrimSpace(string(eplm.cSourceFlg))
	cEspIDStr := strings.TrimSpace(string(eplm.cEspID[:]))

	if eplm.cPrgmFlg == 'A' || eplm.cPrgmFlg == 'T' || eplm.cPrgmFlg == 'Z' || eplm.cPrgmFlg == 'G' || cSourceFlgStr == "G" {
		eplm.orderbook.C_ctcl_id = eplm.orderbook.C_ctcl_id + "000"
	} else if cAlgoIDStr != "*" && cSourceFlgStr == "M" {
		eplm.orderbook.C_ctcl_id = "333333333333" + "000"
	} else if cAlgoIDStr != "*" && (cSourceFlgStr == "W" || cSourceFlgStr == "E") {
		eplm.orderbook.C_ctcl_id = "111111111111" + "000"
	} else if cAlgoIDStr != "*" && cEspIDStr == "4124" {
		eplm.orderbook.C_ctcl_id = "333333333333" + "000"
	} else if cAlgoIDStr != "*" {
		eplm.orderbook.C_ctcl_id = "111111111111" + "000"
	} else {
		eplm.orderbook.C_ctcl_id = eplm.orderbook.C_ctcl_id + "100"
	}

	// eplm.oe_reqres.D_nnf_field, err = strconv.ParseFloat(eplm.orderbook.C_ctcl_id, 64)
	// if err != nil {
	// 	eplm.LoggerManager.LogError(eplm.serviceName, " [fnPackOrdnryOrdToNse] [Error in converting the 'C_ctcl_id' in float ... ")
	// 	eplm.LoggerManager.GetLogger().Warn(eplm.serviceName, " [fnPackOrdnryOrdToNse] Exiting from the function ")
	// 	return -1
	// }

	cPanNoStr := strings.TrimSpace(strings.ToUpper(string(eplm.cPanNo[:])))

	if cPanNoStr != "*" {
		copy(eplm.oe_reqres.C_pan[:], cPanNoStr)
	} else {
		copy(eplm.oe_reqres.C_pan[:], "          ")
	}

	/*
		cPanNo
		C_cover_uncover
		C_giveup_flag
		C_reserved
	*/

	/***** Since there is no provision for
	settlement period in EBA we set this to  **/
	eplm.oe_reqres.Si_settlement_period = 0

	eplm.oe_reqres.C_cover_uncover = 'V'

	eplm.oe_reqres.C_giveup_flag = 'P'

	eplm.oe_reqres.D_filler19 = 0.0

	var reserved [52]byte
	for i := range reserved {
		reserved[i] = ' '
	}
	eplm.oe_reqres.C_reserved = reserved

	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse]  Printing header structure for Ordinary Order....")
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'Int Header' li_Log_time is :	%d ", eplm.int_header.Li_log_time)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'Int Header' c_alpha_char is :	%c ", eplm.int_header.C_alpha_char)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'Int Header' si_transaction_code is :	%d ", eplm.int_header.Si_transaction_code)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'Int Header' si_error_code is :	%d ", eplm.int_header.Si_error_code)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'Int Header' c_filler_2 is :	%c ", eplm.int_header.C_filler_2)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'Int Header' c_time_stamp_1 is :	%c ", eplm.int_header.C_time_stamp_1)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'Int Header' c_time_stamp_2 is :	%c ", eplm.int_header.C_time_stamp_2)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'Int Header' si_message_length is :	%d ", eplm.int_header.Si_message_length)

	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'St_oe_reqres' c_participant_type is :	%c ", eplm.oe_reqres.C_participant_type)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'St_oe_reqres' c_filler_1 is :	%c ", eplm.oe_reqres.C_filler_1)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'St_oe_reqres' si_competitor_period is :	%d ", eplm.oe_reqres.Si_competitor_period)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'St_oe_reqres' si_solicitor_period is :	%d ", eplm.oe_reqres.Si_solicitor_period)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'St_oe_reqres' c_modified_cancelled_by is :	%c ", eplm.oe_reqres.C_modified_cancelled_by)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'St_oe_reqres' c_filler_2 is :	%c ", eplm.oe_reqres.C_filler_2)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'St_oe_reqres' si_reason_code is :	%d ", eplm.oe_reqres.Si_reason_code)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'St_oe_reqres' c_filler_3 is :	%c ", eplm.oe_reqres.C_filler_3)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'St_oe_reqres' l_token_no is :	%d ", eplm.oe_reqres.L_token_no)

	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] Printing contract_desc structure for Ordinary Order....")
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'contract_desc' c_instrument_name is :	%c ", eplm.contract_desc.C_instrument_name)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'contract_desc' c_symbol is :	%c ", eplm.contract_desc.C_symbol)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'contract_desc' li_expiry_date is :	%d ", eplm.contract_desc.Li_expiry_date)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'contract_desc' li_strike_price is :	%d ", eplm.contract_desc.Li_strike_price)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'contract_desc' c_option_type is :	%c ", eplm.contract_desc.C_option_type)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'contract_desc' si_ca_level is :	%d ", eplm.contract_desc.Si_ca_level)

	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'St_oe_reqres' c_counter_party_broker_id is :	%c ", eplm.oe_reqres.C_counter_party_broker_id)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'St_oe_reqres' c_filler_4 is :	%c ", eplm.oe_reqres.C_filler_4)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'St_oe_reqres' c_filler_5 is :	%c ", eplm.oe_reqres.C_filler_5)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'St_oe_reqres' c_closeout_flg is :	%c ", eplm.oe_reqres.C_closeout_flg)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'St_oe_reqres' c_filler_6 is :	%c ", eplm.oe_reqres.C_filler_6)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'St_oe_reqres' si_order_type is :	%d ", eplm.oe_reqres.Si_order_type)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'St_oe_reqres' d_order_number is :	%f ", eplm.oe_reqres.D_order_number)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'St_oe_reqres' c_account_number is :	%c ", eplm.oe_reqres.C_account_number)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'St_oe_reqres' si_book_type is :	%d ", eplm.oe_reqres.Si_book_type)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'St_oe_reqres' si_buy_sell_indicator is :	%d ", eplm.oe_reqres.Si_buy_sell_indicator)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'St_oe_reqres' li_disclosed_volume is :	%d ", eplm.oe_reqres.Li_disclosed_volume)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'St_oe_reqres' li_disclosed_volume_remaining is :	%d ", eplm.oe_reqres.Li_disclosed_volume_remaining)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'St_oe_reqres' li_total_volume_remaining is :	%d ", eplm.oe_reqres.Li_total_volume_remaining)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'St_oe_reqres' li_volume is :	%d ", eplm.oe_reqres.Li_volume)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'St_oe_reqres' li_volume_filled_today is :	%d ", eplm.oe_reqres.Li_volume_filled_today)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'St_oe_reqres' li_price is :	%d ", eplm.oe_reqres.Li_price)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'St_oe_reqres' li_trigger_price is :	%d ", eplm.oe_reqres.Li_trigger_price)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'St_oe_reqres' li_good_till_date is :	%d ", eplm.oe_reqres.Li_good_till_date)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'St_oe_reqres' li_entry_date_time is :	%d ", eplm.oe_reqres.Li_entry_date_time)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'St_oe_reqres' li_minimum_fill_aon_volume is :	%d ", eplm.oe_reqres.Li_minimum_fill_aon_volume)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'St_oe_reqres' li_last_modified is :	%d ", eplm.oe_reqres.Li_last_modified)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'St_oe_reqres' c_PanNo is :	%c ", eplm.oe_reqres.C_pan)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'St_oe_reqres' C_cover_uncover is :	%c ", eplm.oe_reqres.C_cover_uncover)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'St_oe_reqres' C_giveup_flag is :	%c ", eplm.oe_reqres.C_giveup_flag)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'St_oe_reqres' C_reserved is :	%c ", eplm.oe_reqres.C_reserved)

	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] Printing order flags structure for Ordinary Order....")
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'order flags' flg_ato is : %d", eplm.TCUM.BoolToInt(eplm.order_flag.GetFlagValue(models.Flg_ATO)))
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'order flags' flg_market is : %d", eplm.TCUM.BoolToInt(eplm.order_flag.GetFlagValue(models.Flg_Market)))
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'order flags' flg_sl is : %d", eplm.TCUM.BoolToInt(eplm.order_flag.GetFlagValue(models.Flg_SL)))
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'order flags' flg_mit is : %d", eplm.TCUM.BoolToInt(eplm.order_flag.GetFlagValue(models.Flg_MIT)))
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'order flags' flg_day is : %d", eplm.TCUM.BoolToInt(eplm.order_flag.GetFlagValue(models.Flg_Day)))
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'order flags' flg_gtc is : %d", eplm.TCUM.BoolToInt(eplm.order_flag.GetFlagValue(models.Flg_GTC)))
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'order flags' flg_ioc is : %d", eplm.TCUM.BoolToInt(eplm.order_flag.GetFlagValue(models.Flg_IOC)))
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'order flags' flg_aon is : %d", eplm.TCUM.BoolToInt(eplm.order_flag.GetFlagValue(models.Flg_AON)))
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'order flags' flg_mf is : %d", eplm.TCUM.BoolToInt(eplm.order_flag.GetFlagValue(models.Flg_MF)))
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'order flags' flg_matched_ind is : %d", eplm.TCUM.BoolToInt(eplm.order_flag.GetFlagValue(models.Flg_MatchedInd)))
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'order flags' flg_traded is : %d", eplm.TCUM.BoolToInt(eplm.order_flag.GetFlagValue(models.Flg_Traded)))
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] 'order flags' flg_modified is : %d", eplm.TCUM.BoolToInt(eplm.order_flag.GetFlagValue(models.Flg_Modified)))
	// eplm.LoggerManager.LogInfo(eplm.serviceName , " [fnPackOrdnryOrdToNse] flg_cancelled is :	%d ", eplm.order_flag.Flg_cancelled)
	// eplm.LoggerManager.LogInfo(eplm.serviceName , " [fnPackOrdnryOrdToNse] flg_cancel_pending is :	%d ", eplm.order_flag.Flg_cancel_pending)
	// eplm.LoggerManager.LogInfo(eplm.serviceName , " [fnPackOrdnryOrdToNse] flg_closed is :	%d ", eplm.order_flag.Flg_closed)
	// eplm.LoggerManager.LogInfo(eplm.serviceName , " [fnPackOrdnryOrdToNse] flg_fill_and_kill is :	%d ", eplm.order_flag.Flg_fill_and_kill)

	tmpVar := eplm.TCUM.GetResetSequence(eplm.Db, eplm.xchngbook.C_pipe_id, eplm.xchngbook.C_mod_trd_dt)
	if tmpVar == -1 {
		eplm.LoggerManager.LogError(eplm.serviceName, " [fnPackOrdnryOrdToNse] Error: Failed to retrieve sequence number")
		return -1
	}
	eplm.net_hdr.I_seq_num = tmpVar
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] Set I_seq_num to: %d", eplm.net_hdr.I_seq_num)

	// Convert order request and network header to network order (NSE Order)
	eplm.OCM.ConvertOrderReqResToNetworkOrder(eplm.oe_reqres, eplm.int_header, eplm.contract_desc)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] Converted ST OE REQRES to network order")

	hasher := md5.New()

	oeReqResString, err := json.Marshal(eplm.oe_reqres) // to convert the structure in string
	if err != nil {
		eplm.LoggerManager.LogError(eplm.serviceName, " [ Error: failed to convert oe_reqres to string: %v", err)
	}

	io.WriteString(hasher, string(oeReqResString))

	checksum := hasher.Sum(nil)

	copy(eplm.net_hdr.C_checksum[:], fmt.Sprintf("%x", checksum)) // for Hexadecimal format

	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] Set C_checksum to computed MD5 hash: %x", eplm.net_hdr.C_checksum[:])
	eplm.LoggerManager.LogInfo(eplm.serviceName, "ksum : %s", checksum)
	eplm.OCM.ConvertNetHeaderToNetworkOrder(eplm.net_hdr)
	eplm.LoggerManager.LogInfo(eplm.serviceName, " [fnPackOrdnryOrdToNse] Converted net header to network order")

	eplm.net_hdr.S_message_length = int16(unsafe.Sizeof(eplm.net_hdr) + unsafe.Sizeof(eplm.oe_reqres))
	eplm.LoggerManager.LogInfo(eplm.serviceName, fmt.Sprintf("[fnPackOrdnryOrdToNse] Calculated message length: %d", eplm.net_hdr.S_message_length))

	buf := new(bytes.Buffer)

	// Write int_header and handle error
	if err := eplm.TCUM.WriteAndCopy(buf, *eplm.int_header, eplm.oe_reqres.St_hdr[:]); err != nil {
		eplm.LoggerManager.LogError(eplm.serviceName, " [fnPackOrdnryOrdToNse] [Error: Failed to write int_header: %v", err)
	}

	// Write contract_desc and handle error
	if err := eplm.TCUM.WriteAndCopy(buf, *eplm.contract_desc, eplm.oe_reqres.St_con_desc[:]); err != nil {
		eplm.LoggerManager.LogError(eplm.serviceName, " [fnPackOrdnryOrdToNse] [Error: Failed to write contract_desc: %v", err)
	}

	// Write order_flag and handle error
	if err := eplm.TCUM.WriteAndCopy(buf, *eplm.order_flag, eplm.oe_reqres.St_ord_flg[:]); err != nil {
		eplm.LoggerManager.LogError(eplm.serviceName, " [fnPackOrdnryOrdToNse] [Error: Failed to write order_flag: %v", err)
	}

	// Write oe_reqres and handle error
	if err := eplm.TCUM.WriteAndCopy(buf, *eplm.oe_reqres, eplm.exch_msg.St_oe_res[:]); err != nil {
		eplm.LoggerManager.LogError(eplm.serviceName, " [Error: Failed to write oe_reqres: %v", err)
	}

	// Write net_hdr and handle error
	if err := eplm.TCUM.WriteAndCopy(buf, *eplm.net_hdr, eplm.exch_msg.St_net_header[:]); err != nil {
		eplm.LoggerManager.LogError(eplm.serviceName, " [fnPackOrdnryOrdToNse] [Error: Failed to write net_hdr: %v", err)
	}

	// Write exch_msg and handle error
	if err := eplm.TCUM.WriteAndCopy(buf, *eplm.exch_msg, eplm.q_packet.St_exch_msg_data[:]); err != nil {
		eplm.LoggerManager.LogError(eplm.serviceName, " [fnPackOrdnryOrdToNse] [Error: Failed to write exch_msg: %v", err)
	}

	// Set L_msg_type and log success message
	eplm.q_packet.L_msg_type = int64(util.BOARD_LOT_IN)
	eplm.LoggerManager.LogInfo(eplm.serviceName, "[fnPackOrdnryOrdToNse] Data copied into Queue Packet")

	return 0
}

func (eplm *ExchngPackLibMaster) Fn_orstonse_cntrct_desc() int {
	eplm.LoggerManager.LogInfo(eplm.serviceName, " Entered in the function 'Fn_orstonse_cntrct_desc'")

	var instrumentName [6]byte

	switch eplm.nse_contract.C_prd_typ {
	case "F":
		copy(instrumentName[:], []byte{'F', 'U', 'T'})
	case "O":
		copy(instrumentName[:], []byte{'O', 'P', 'T'})
	default:
		eplm.LoggerManager.LogError(eplm.serviceName, " [Error: Invalid product type: %s", eplm.nse_contract.C_prd_typ)
		return -1
	}

	switch eplm.nse_contract.C_ctgry_indstk {
	case "I":
		instrumentName[3] = 'I'
		if strings.HasPrefix(eplm.nse_contract.C_symbol, "INDIAVIX") {
			instrumentName[4] = 'V'
		} else {
			instrumentName[4] = 'D'
		}
		instrumentName[5] = 'X'
	case "S":
		copy(instrumentName[:], []byte{'S', 'T', 'K'})
	default:
		eplm.LoggerManager.LogError(eplm.serviceName, " [Error: Should be Index/stock, got: %s", eplm.nse_contract.C_ctgry_indstk)
		return -1
	}

	eplm.TCUM.CopyAndFormatSymbol(eplm.contract_desc.C_symbol[:], len(eplm.contract_desc.C_symbol), eplm.nse_contract.C_symbol) // 1 contract_desc

	status, Li_expiry_date := eplm.TCUM.LongToTimeArr(eplm.nse_contract.C_expry_dt)

	if status != 0 {
		eplm.LoggerManager.LogError(eplm.serviceName, "[Error: Returned from 'LongToTimeArr()' with an Error]")
		eplm.LoggerManager.LogError(eplm.serviceName, "[Error: Exiting from 'fn_orstone_cntrct_desc']")
		return -1
	}
	eplm.contract_desc.Li_expiry_date = Li_expiry_date // 3 contract_desc

	if eplm.nse_contract.C_prd_typ == "F" {
		eplm.contract_desc.Li_strike_price = -1 // 3 contract_desc
	} else {
		eplm.contract_desc.Li_strike_price = eplm.nse_contract.L_strike_prc
	}

	if eplm.nse_contract.C_prd_typ == "O" {
		if len(eplm.nse_contract.C_opt_typ) > 0 {
			eplm.contract_desc.C_option_type[0] = eplm.nse_contract.C_opt_typ[0] //here i have to assume the the string will be of size 1. This problems happens because i replace all the byte array with string
		} else {
			eplm.contract_desc.C_option_type[0] = 'X' //4 contract_desc
		}

		if len(eplm.nse_contract.C_exrc_typ) > 0 {
			eplm.contract_desc.C_option_type[1] = eplm.nse_contract.C_exrc_typ[0]
		} else {
			eplm.contract_desc.C_option_type[1] = 'X'
		}
	} else {
		eplm.contract_desc.C_option_type[0] = 'X'
		eplm.contract_desc.C_option_type[1] = 'X'
	}

	eplm.contract_desc.Si_ca_level = int16(eplm.nse_contract.L_ca_lvl) // 5 contract_desc

	copy(eplm.contract_desc.C_instrument_name[:], instrumentName[:]) // 6 contract_desc

	eplm.LoggerManager.LogInfo(eplm.serviceName, " Exiting from 'Fn_orstonse_cntrct_desc' after successfully packing the 'St_contract_desc'")

	return 0
}
