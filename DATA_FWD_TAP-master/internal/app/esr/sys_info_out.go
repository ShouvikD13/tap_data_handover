package esr

import (
	"DATA_FWD_TAP/internal/models"
	"DATA_FWD_TAP/util"
	"fmt"
	"log"
)

var si_normal_mkt_stts int16
var si_exr_mkt_stts int16
var si_pl_mkt_stts int16

var si_oddmkt_stts int16
var si_spotmkt_stts int16
var si_auctionmkt_stts int16

var si_exr_oddmkt_stts int16
var si_exr_spotmkt_stts int16
var si_exr_auctionmkt_stts int16

var si_pl_oddmkt_stts int16
var si_pl_spotmkt_stts int16
var si_pl_auctionmkt_stts int16

func (ESRM *ESRManager) Fn_system_information_out(sys_info_data models.St_system_info_data, C_err_msg models.C_err_msg) error {

	var Li_mkt_type int64   // long int
	var Li_stream_no int64  // long int
	// var I_ch_val int32      // int
	// var I_trnsctn int32     // int
	// var C_stream_no [2]byte // char c_stream_no[2]

	si_normal_mkt_stts = sys_info_data.St_mkt_stts.Si_normal
	si_oddmkt_stts = sys_info_data.St_mkt_stts.Si_oddlot
	si_spotmkt_stts = sys_info_data.St_mkt_stts.Si_spot
	si_auctionmkt_stts = sys_info_data.St_mkt_stts.Si_auction

	log.Printf("The Normal  si_mkt_stts - Normal is:", si_normal_mkt_stts) //Servicename needs to be added
	log.Printf("The Normal  si_mkt_stts - oddlot is:", sys_info_data.St_mkt_stts.Si_oddlot, ", :", si_oddmkt_stts)
	log.Printf("The Normal  si_mkt_stts - spot is:", sys_info_data.St_mkt_stts.Si_spot, ", :", si_spotmkt_stts)
	log.Printf("The Normal  si_mkt_stts - auction is:", sys_info_data.St_mkt_stts.Si_auction, ", :", si_auctionmkt_stts)

	si_exr_mkt_stts = sys_info_data.St_ex_mkt_stts.Si_normal
	si_exr_oddmkt_stts = sys_info_data.St_ex_mkt_stts.Si_oddlot
	si_exr_spotmkt_stts = sys_info_data.St_ex_mkt_stts.Si_spot
	si_exr_auctionmkt_stts = sys_info_data.St_ex_mkt_stts.Si_auction

	log.Printf("The Exer  si_mkt_stts - Normal is:", si_exr_mkt_stts) //Servicename needs to be added
	log.Printf("The Exer  si_mkt_stts - oddlot is:", sys_info_data.St_ex_mkt_stts.Si_oddlot, ", :", si_exr_oddmkt_stts)
	log.Printf("The Exer  si_mkt_stts - spot is:", sys_info_data.St_ex_mkt_stts.Si_spot, ", :", si_exr_spotmkt_stts)
	log.Printf("The Exer  si_mkt_stts - auction is:", sys_info_data.St_ex_mkt_stts.Si_auction, ", :", si_exr_auctionmkt_stts)

	Li_mkt_type = EXERCISE_MKT //Looj into this

	//fn_mrkt_stts_hndlr

	err := ESRM.Fn_mrkt_stts_hndlr(si_exr_mkt_stts, Li_mkt_type, Sql_c_xchng_cd, C_Servicename, &C_err_msg)

	si_pl_mkt_stts = sys_info_data.St_pl_mkt_stts.Si_normal
	si_pl_oddmkt_stts = sys_info_data.St_pl_mkt_stts.Si_oddlot
	si_pl_spotmkt_stts = sys_info_data.St_pl_mkt_stts.Si_spot
	si_pl_auctionmkt_stts = sys_info_data.St_pl_mkt_stts.Si_auction

	log.Printf("The PL si_mkt_stts - Normal is: %d", si_pl_mkt_stts) // Add ServiceName
	log.Printf("The PL si_mkt_stts - oddlot is: %d, : %d", sys_info_data.St_pl_mkt_stts.Si_oddlot, si_pl_oddmkt_stts)
	log.Printf("The PL si_mkt_stts - spot is: %d, : %d", sys_info_data.St_pl_mkt_stts.Si_spot, si_pl_spotmkt_stts)
	log.Printf("The PL si_mkt_stts - auction is: %d, : %d", sys_info_data.St_pl_mkt_stts.Si_auction, si_pl_auctionmkt_stts)

	Li_mkt_type = PL_MKT //Look into this

	Li_stream_no = int64(sys_info_data.St_hdr.C_alpha_char[0])
	log.Printf("St_sys_info_data->St_hdr.c_alpha_char[0] Is :", sys_info_data.St_hdr.C_alpha_char[0])
	log.Printf("Before Li_stream_no is :", Li_stream_no)

	log.Printf("St_hdr.c_alpha_char Is :", sys_info_data.St_hdr.C_alpha_char)

	log.Printf("The number of streams recieved is :", Li_stream_no)

	ESRM.TM.FnBeginTran()

	query := `
		UPDATE exg_xchng_mstr
		SET exg_def_stlmnt_pd   = $1,
			exg_wrn_pcnt        = $2,
			exg_vol_frz_pcnt    = $3,
			exg_brd_lot_qty     = $4,
			exg_tck_sz          = $5,
			exg_gtc_dys         = $6,
			exg_dsclsd_qty_pcnt = $7
		WHERE exg_xchng_cd = $8
	`

	erro := ESRM.Db.Exec(query,
		sys_info_data.Si_default_sttlmnt_period_nm,
		sys_info_data.Si_warning_percent,
		sys_info_data.Si_volume_freeze_percent,
		sys_info_data.Li_board_lot_quantity,
		sys_info_data.Li_tick_size,
		sys_info_data.Si_maximum_gtc_days,
		sys_info_data.Si_disclosed_quantity_percent_allowed,
		Sql_c_xchng_cd,
	)
	if erro != nil {
		log.Printf("Error updating exg_xchng_mstr: %v", erro)
		return fmt.Errorf("error updating exg_xchng_mstr: %w", erro)
		ESRM.TM.FnAbortTran()
	}

	// Update order pipe master
	queryPipe := `
		UPDATE OPM_ORD_PIPE_MSTR
		SET OPM_STREAM_NO = $1
		WHERE OPM_PIPE_ID = $2
		AND   OPM_XCHNG_CD= $3
	`

	result := ESRM.Db.Exec(queryPipe, Li_stream_no, Sql_c_pipe_id, Sql_c_xchng_cd)

	// Check for errors
	if result.Error != nil {
		log.Printf("Error updating OPM_ORD_PIPE_MSTR: %v", result.Error)
		return fmt.Errorf("error updating OPM_ORD_PIPE_MSTR: %w", result.Error)
		ESRM.TM.FnAbortTran()
	}

	ESRM.TM.FnCommitTran()

	return nil
}

func (ESRM *ESRManager) Fn_mrkt_stts_hndlr(siMktStts int16, li_mkt_type int64, cXchngCd [4]byte, cServiceName string, cErrMsg *models.C_err_msg) error {
	var C_mrkt_stts byte
	var C_mkt_typ byte

	log.Printf("Inside Function Fn_mrkt_stts_hndlr")

	if li_mkt_type != PL_MKT {
		log.Printf("Inside Market Type", "\n", "Inside Market Status Check")

		switch siMktStts {
		case util.PRE_OPEN:
		case util.PRE_OPEN_ENDED:
		case util.POST_CLOSE:
			return nil

		case util.OPEN:
			log.Printf("Inside OPEN Case")
			C_mrkt_stts = 'O'
			log.Printf("Market Status Is :", C_mrkt_stts)
			break

		case util.CLOSED:
			log.Printf("Inside CLOSED Case")
			C_mrkt_stts = 'C'
			log.Printf("Market Status Is :", C_mrkt_stts)
			break

		case 'X':
			log.Printf("Inside Expired Case")
			C_mrkt_stts = 'X'
			log.Printf("Market Status Is :", C_mrkt_stts)
			break
		}

		switch (li_mkt_type); {
		case ORDER_MKT:
			log.Printf("Inside ORDER_MKT Case")
			C_mkt_typ = ORDER_MARKET
			log.Printf("MARKET Type Is:", C_mkt_typ)

		case EXTND_MKT:
			log.Printf("Inside EXTND_MKT Case")
			C_mkt_typ = EXTENDED_SEGMENT
			log.Printf("MARKET Type Is:", C_mkt_typ)

		case EXERCISE_MKT:
			log.Printf("Inside EXERCISE_MKT Case")
			C_mkt_typ = EXERCISE_MARKET
			log.Printf("MARKET Type Is:", C_mkt_typ)

		case PL_MKT:
			log.Printf("Inside PL_MKT Case")
			C_mkt_typ = PL_MARKET
			log.Printf("MARKET Type Is:", C_mkt_typ)

		}

	}

	err := ESRM.fn_mrkt_stts(C_mrkt_stts, C_mkt_typ, Sql_c_xchng_cd, servicename, cErrMsg)
	if err != nil {
		log.Printf("Failed While calling Function fn_mrkt_stts")
	}

	log.Printf("After Compeletion of fn_mrkt_stts_hndlr")

	return nil

}

//*********************************************************************//
//* As of current implementation SFO_UPD_XSTTS will update exchange master
//* only when c_mkt_type is EXERCISE_SEGMENT. Due to it being the only
//* case where C_rqst_typ has any value.
//********************************************************************//

func (ESRM *ESRManager) fn_mrkt_stts(C_mrkt_stts byte, C_mkt_typ byte, cxchngcd [4]byte, servicename string, C_err_msg *models.C_err_msg) error {

	var st_stts models.Vw_xchngstts

	log.Printf("Inside Function fn_mrkt_stts")

	st_stts.C_xchng_cd = cxchngcd

	log.Printf("Exchange Code Is :", st_stts.C_xchng_cd)
	log.Printf("Here Market Type is:", C_mkt_typ)

	if C_mkt_typ == EXERCISE_SEGMENT {

		log.Printf("Inside EXERCISE_SEGMENT Check")
		st_stts.C_exrc_mkt_stts = C_mrkt_stts
		st_stts.C_rqst_typ = util.UPD_EXER_MKT_STTS
	}

	ESRM.Fn_cpy_ddr(&st_stts.C_rout_crt)

	err2 := ESRM.Fn_AcallSvc_Sfo_Udp_Xstts(st_stts)
	if err2 != nil {

	}

	return nil
}

func (ESRM *ESRManager) Fn_AcallSvc_Sfo_Udp_Xstts(st_stts models.Vw_xchngstts) error {

	var ptr_st_xchng_stts *models.Vw_xchngstts
	var i_record_exists int = 0
	var c_raise_trigger_type [30]byte

	ptr_st_xchng_stts = &st_stts

	ESRM.TM.FnBeginTran()

	switch ptr_st_xchng_stts.C_rqst_typ {
	case util.UPD_PARTICIPANT_STTS:
		i_record_exists = 0
		query := `
		SELECT 1 
		INTO $1
		FROM EXG_XCHNG_MSTR
		WHERE EXG_XCHNG_CD = $2
		AND EXG_SETTLOR_ID = $3`

		err := ESRM.Db.Exec(query, st_stts.C_xchng_cd, st_stts.C_settlor).Scan(&i_record_exists)
		if err != nil {
			log.Printf("%s: Error querying participant status: %v", serviceName, err)
			ESRM.TM.FnAbortTran()
			return nil
		}

		if i_record_exists == 1 {
			query = "UPDATE EXG_XCHNG_MSTR SET exg_settlor_stts = $1 WHERE exg_xchng_cd = $2 AND exg_settlor_id = $3"
			err = ESRM.Db.Exec(query, st_stts.C_settlor_stts, st_stts.C_xchng_cd, st_stts.C_settlor)
			if err != nil {
				log.Printf("%s: Error updating participant status: %v", serviceName, err)
				ESRM.TM.FnAbortTran()
				return nil
			}

			triggerName := ""
			if st_stts.C_settlor_stts == util.SUSPEND {
				triggerName = "TRG_PART_SUS"
			} else {
				triggerName = "TRG_PART_ACT"
			}
			copy(c_raise_trigger_type[:], triggerName)
		}

	case util.UPD_BRK_LMT_EXCD:
		query := "SELECT 1 FROM EXG_XCHNG_MSTR WHERE exg_xchng_cd = $1 AND exg_brkr_id = $2"
		err := ESRM.Db.Raw(query, st_stts.C_xchng_cd, st_stts.C_brkr_id).Scan(&i_record_exists)
		if err != nil {
			log.Printf("%s: Error querying broker status: %v", serviceName, err)
			ESRM.TM.FnAbortTran()
			return nil
		}

		if i_record_exists == 1 {
			query = "UPDATE EXG_XCHNG_MSTR SET exg_brkr_stts = $1 WHERE exg_xchng_cd = $2 AND exg_brkr_id = $3"
			err = ESRM.Db.Exec(query, st_stts.C_brkr_stts, st_stts.C_xchng_cd, st_stts.C_brkr_id)
			if err != nil {
				log.Printf("%s: Error updating broker status: %v", serviceName, err)
				ESRM.TM.FnAbortTran()
				return nil
			}

			triggerName := ""
			if st_stts.C_brkr_stts == util.SUSPEND {
				triggerName = "TRG_BRKR_SUS"
			} else {
				triggerName = "TRG_BRKR_ACT"
			}
			copy(c_raise_trigger_type[:], triggerName)
		}

	case util.UPD_EXER_MKT_STTS:
		query := "UPDATE EXG_XCHNG_MSTR SET exg_exrc_mkt_stts = $1 WHERE exg_xchng_cd = $2"
		err := ESRM.Db.Exec(query, st_stts.C_exrc_mkt_stts, st_stts.C_xchng_cd)
		if err != nil {
			log.Printf("%s: Error updating exercise market status: %v", serviceName, err)
			ESRM.TM.FnAbortTran()
			return nil
		}

		triggerName := ""
		if st_stts.C_exrc_mkt_stts == util.EXCHANGE_OPEN {
			triggerName = "TRG_EXR_OPN"
		} else {
			triggerName = "TRG_EXR_CLS"
		}
		copy(c_raise_trigger_type[:], triggerName)

	case util.UPD_NORMAL_MKT_STTS:
		query := "UPDATE EXG_XCHNG_MSTR SET exg_crrnt_stts = $1 WHERE exg_xchng_cd = $2"
		err := ESRM.Db.Exec(query, st_stts.C_crrnt_stts, st_stts.C_xchng_cd)
		if err != nil {
			log.Printf("%s: Error updating normal market status: %v", serviceName, err)
			ESRM.TM.FnAbortTran()
			return nil
		}

		if st_stts.C_crrnt_stts == 'X' {
			query = "UPDATE EXG_XCHNG_MSTR SET exg_tmp_mkt_stts = $1 WHERE exg_xchng_cd = $2"
			err = ESRM.Db.Exec(query, st_stts.C_crrnt_stts, st_stts.C_xchng_cd)
			if err != nil {
				log.Printf("%s: Error updating temporary market status: %v", serviceName, err)
				ESRM.TM.FnAbortTran()
				return nil
			}
		}

		triggerName := ""
		if st_stts.C_crrnt_stts == util.EXCHANGE_OPEN {
			triggerName = "TRG_ORD_OPN"
		} else {
			triggerName = "TRG_ORD_CLS"
		}
		copy(c_raise_trigger_type[:], triggerName)

	case util.UPD_EXTND_MKT_STTS:
		query := "UPDATE EXG_XCHNG_MSTR SET exg_extnd_mrkt_stts = $1 WHERE exg_xchng_cd = $2"
		err := ESRM.Db.Exec(query, st_stts.C_crrnt_stts, st_stts.C_xchng_cd)
		if err != nil {
			log.Printf("%s: Error updating extended market status: %v", serviceName, err)
			ESRM.TM.FnAbortTran()
			return nil
		}

		triggerName := ""
		if st_stts.C_crrnt_stts == util.EXCHANGE_OPEN {
			triggerName = "TRG_EXT_OPN"
		} else {
			triggerName = "TRG_EXT_CLS"
		}
		copy(c_raise_trigger_type[:], triggerName)

	default:
		log.Printf("%s: Unknown request type", serviceName)
		return fmt.Errorf("unknown request type")
	}

	ESRM.TM.FnCommitTran()

	// Raise trigger if applicable
	/* if c_raise_trigger_type != "" {
		err = ESRM.raiseTrigger(serviceName, c_raise_trigger_type, st_stts.C_xchng_cd)
		if err != nil {
			log.Printf("%s: Error raising trigger: %v", serviceName, err)
			return nil
		}
	} */

	return nil
}
