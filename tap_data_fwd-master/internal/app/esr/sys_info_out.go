package esr

import (
	"DATA_FWD_TAP/internal/models"
	"DATA_FWD_TAP/util"
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

func (ESRM *ESRManager) Fn_system_information_out(sys_info_data models.St_system_info_data, C_err_msg *C_err_msg) error {

	var Li_mkt_type int64   // long int
	var Li_stream_no int64  // long int
	var I_ch_val int32      // int
	var I_trnsctn int32     // int
	var C_stream_no [2]byte // char c_stream_no[2]

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

	//Li_mkt_type = EXERCISE_MKT; //Looj into this

	//fn_mrkt_stts_hndlr

	err := ESRM.fn_mrkt_stts_hndlr(si_exr_mkt_stts, Li_mkt_type, sql_c_xchng_cd, C_Servicename, C_err_msg)

	si_pl_mkt_stts = sys_info_data.St_pl_mkt_stts.Si_normal
	si_pl_oddmkt_stts = sys_info_data.St_pl_mkt_stts.Si_oddlot
	si_pl_spotmkt_stts = sys_info_data.St_pl_mkt_stts.Si_spot
	si_pl_auctionmkt_stts = sys_info_data.St_pl_mkt_stts.Si_auction

	log.Printf("The PL si_mkt_stts - Normal is: %d", si_pl_mkt_stts) // Add ServiceName
	log.Printf("The PL si_mkt_stts - oddlot is: %d, : %d", sys_info_data.St_pl_mkt_stts.Si_oddlot, si_pl_oddmkt_stts)
	log.Printf("The PL si_mkt_stts - spot is: %d, : %d", sys_info_data.St_pl_mkt_stts.Si_spot, si_pl_spotmkt_stts)
	log.Printf("The PL si_mkt_stts - auction is: %d, : %d", sys_info_data.St_pl_mkt_stts.Si_auction, si_pl_auctionmkt_stts)

	//Li_mkt_type = PL_MKT//Look into this

	Li_stream_no = int64(sys_info_data.St_hdr.C_alpha_char[0])
	log.Printf("St_sys_info_data->St_hdr.c_alpha_char[0] Is :", sys_info_data.St_hdr.C_alpha_char[0])
	log.Printf("Before Li_stream_no is :", Li_stream_no)

	log.Printf("St_hdr.c_alpha_char Is :", sys_info_data.St_hdr.C_alpha_char)

	log.Printf("The number of streams recieved is :", Li_stream_no)

	return nil
}

func (ESRM *ESRManager) Fn_mrkt_stts_hndlr(siMktStts int16, li_mkt_type int64, cXchngCd string, cServiceName string, cErrMsg *string) error {
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

	}

	return nil

}
