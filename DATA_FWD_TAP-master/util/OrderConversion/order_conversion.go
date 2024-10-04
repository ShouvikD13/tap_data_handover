package OrderConversion

import (
	"DATA_FWD_TAP/internal/models"
	"encoding/binary"
	"math"
)

type OrderConversionManager struct {
	Oe_reqres      *models.St_oe_reqres
	Net_hdr        *models.St_net_hdr
	Int_header     *models.St_int_header
	Contract_desc  *models.St_contract_desc
	St_sign_on_req *models.St_sign_on_req
}

func (OCM *OrderConversionManager) ConvertInt16ToNetworkOrder(i int16) int16 {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(i))
	return int16(binary.BigEndian.Uint16(b))
}

func (OCM *OrderConversionManager) ConvertInt16ToHostOrder(i int16) int16 {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(i))
	return int16(binary.BigEndian.Uint16(b))
}

func (OCM *OrderConversionManager) ConvertInt32ToNetworkOrder(i int32) int32 {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(i))
	return int32(binary.BigEndian.Uint32(b))
}

func (OCM *OrderConversionManager) ConvertInt32ToHostOrder(i int32) int32 {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(i))
	return int32(binary.BigEndian.Uint32(b))
}

func (OCM *OrderConversionManager) ConvertInt64ToNetworkOrder(i int64) int64 {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(i))
	return int64(binary.BigEndian.Uint64(b))
}

func (OCM *OrderConversionManager) ConvertInt64ToHostOrder(i int64) int64 {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(i))
	return int64(binary.BigEndian.Uint64(b))
}

func (OCM *OrderConversionManager) ConvertFloat32ToNetworkOrder(f float32) float32 {
	bits := math.Float32bits(f)
	networkBits := OCM.ConvertInt32ToNetworkOrder(int32(bits))
	return math.Float32frombits(uint32(networkBits))
}

func (OCM *OrderConversionManager) ConvertFloat32ToHostOrder(f float32) float32 {
	bits := math.Float32bits(f)
	hostBits := OCM.ConvertInt32ToHostOrder(int32(bits))
	return math.Float32frombits(uint32(hostBits))
}

func (OCM *OrderConversionManager) ConvertFloat64ToNetworkOrder(f float64) float64 {
	bits := math.Float64bits(f)
	networkBits := OCM.ConvertInt64ToNetworkOrder(int64(bits))
	return math.Float64frombits(uint64(networkBits))
}

func (OCM *OrderConversionManager) ConvertFloat64ToHostOrder(f float64) float64 {
	bits := math.Float64bits(f)
	hostBits := OCM.ConvertInt64ToHostOrder(int64(bits))
	return math.Float64frombits(uint64(hostBits))
}

func (OCM *OrderConversionManager) ConvertOrderReqResToNetworkOrder(st_oe *models.St_oe_reqres, int_header *models.St_int_header, con_desc *models.St_contract_desc) {
	OCM.Oe_reqres = st_oe
	OCM.ConvertIntHeaderToNetworkOrder(int_header) //&OCM.Int_header
	OCM.Oe_reqres.Si_competitor_period = OCM.ConvertInt16ToNetworkOrder(OCM.Oe_reqres.Si_competitor_period)
	OCM.Oe_reqres.Si_solicitor_period = OCM.ConvertInt16ToNetworkOrder(OCM.Oe_reqres.Si_solicitor_period)
	OCM.Oe_reqres.Si_reason_code = OCM.ConvertInt16ToNetworkOrder(OCM.Oe_reqres.Si_reason_code)
	OCM.Oe_reqres.L_token_no = OCM.ConvertInt32ToNetworkOrder(OCM.Oe_reqres.L_token_no)
	OCM.ConvertContractDescToNetworkOrder(con_desc) //&OCM.Contract_desc
	OCM.Oe_reqres.Si_order_type = OCM.ConvertInt16ToNetworkOrder(OCM.Oe_reqres.Si_order_type)
	OCM.Oe_reqres.D_order_number = OCM.ConvertFloat64ToNetworkOrder(OCM.Oe_reqres.D_order_number)
	OCM.Oe_reqres.Si_book_type = OCM.ConvertInt16ToNetworkOrder(OCM.Oe_reqres.Si_book_type)
	OCM.Oe_reqres.Si_buy_sell_indicator = OCM.ConvertInt16ToNetworkOrder(OCM.Oe_reqres.Si_buy_sell_indicator)
	OCM.Oe_reqres.Li_disclosed_volume = OCM.ConvertInt32ToNetworkOrder(OCM.Oe_reqres.Li_disclosed_volume)
	OCM.Oe_reqres.Li_disclosed_volume_remaining = OCM.ConvertInt32ToNetworkOrder(OCM.Oe_reqres.Li_disclosed_volume_remaining)
	OCM.Oe_reqres.Li_total_volume_remaining = OCM.ConvertInt32ToNetworkOrder(OCM.Oe_reqres.Li_total_volume_remaining)
	OCM.Oe_reqres.Li_volume = OCM.ConvertInt32ToNetworkOrder(OCM.Oe_reqres.Li_volume)
	OCM.Oe_reqres.Li_volume_filled_today = OCM.ConvertInt32ToNetworkOrder(OCM.Oe_reqres.Li_volume_filled_today)
	OCM.Oe_reqres.Li_price = OCM.ConvertInt32ToNetworkOrder(OCM.Oe_reqres.Li_price)
	OCM.Oe_reqres.Li_trigger_price = OCM.ConvertInt32ToNetworkOrder(OCM.Oe_reqres.Li_trigger_price)
	OCM.Oe_reqres.Li_good_till_date = OCM.ConvertInt32ToNetworkOrder(OCM.Oe_reqres.Li_good_till_date)
	OCM.Oe_reqres.Li_entry_date_time = OCM.ConvertInt32ToNetworkOrder(OCM.Oe_reqres.Li_entry_date_time)
	OCM.Oe_reqres.Li_minimum_fill_aon_volume = OCM.ConvertInt32ToNetworkOrder(OCM.Oe_reqres.Li_minimum_fill_aon_volume)
	OCM.Oe_reqres.Li_last_modified = OCM.ConvertInt32ToNetworkOrder(OCM.Oe_reqres.Li_last_modified)
	OCM.Oe_reqres.Si_branch_id = OCM.ConvertInt16ToNetworkOrder(OCM.Oe_reqres.Si_branch_id)
	OCM.Oe_reqres.Li_trader_id = OCM.ConvertInt32ToNetworkOrder(OCM.Oe_reqres.Li_trader_id)
	OCM.Oe_reqres.Si_pro_client_indicator = OCM.ConvertInt16ToNetworkOrder(OCM.Oe_reqres.Si_pro_client_indicator)
	OCM.Oe_reqres.Si_settlement_period = OCM.ConvertInt16ToNetworkOrder(OCM.Oe_reqres.Si_settlement_period)
	OCM.Oe_reqres.I_order_seq = OCM.ConvertInt32ToNetworkOrder(OCM.Oe_reqres.I_order_seq)
	OCM.Oe_reqres.D_nnf_field = OCM.ConvertFloat64ToNetworkOrder(OCM.Oe_reqres.D_nnf_field)
	OCM.Oe_reqres.D_filler19 = OCM.ConvertFloat64ToNetworkOrder(OCM.Oe_reqres.D_filler19)
	OCM.Oe_reqres.L_algo_id = OCM.ConvertInt32ToNetworkOrder(OCM.Oe_reqres.L_algo_id)
	OCM.Oe_reqres.Si_algo_category = OCM.ConvertInt16ToNetworkOrder(OCM.Oe_reqres.Si_algo_category)
	OCM.Oe_reqres.Ll_lastactivityref = OCM.ConvertInt64ToNetworkOrder(OCM.Oe_reqres.Ll_lastactivityref)
}

func (OCM *OrderConversionManager) ConvertIntHeaderToNetworkOrder(int_header *models.St_int_header) {
	OCM.Int_header = int_header
	OCM.Int_header.Si_transaction_code = OCM.ConvertInt16ToNetworkOrder(OCM.Int_header.Si_transaction_code)
	OCM.Int_header.Li_log_time = OCM.ConvertInt32ToNetworkOrder(OCM.Int_header.Li_log_time)
	OCM.Int_header.Li_trader_id = OCM.ConvertInt32ToNetworkOrder(OCM.Int_header.Li_trader_id)
	OCM.Int_header.Si_error_code = OCM.ConvertInt16ToNetworkOrder(OCM.Int_header.Si_error_code)
	OCM.Int_header.Si_message_length = OCM.ConvertInt16ToNetworkOrder(OCM.Int_header.Si_message_length)
}

func (OCM *OrderConversionManager) ConvertContractDescToNetworkOrder(con_desc *models.St_contract_desc) {
	OCM.Contract_desc = con_desc
	OCM.Contract_desc.Li_expiry_date = OCM.ConvertInt32ToNetworkOrder(OCM.Contract_desc.Li_expiry_date)
	OCM.Contract_desc.Li_strike_price = OCM.ConvertInt32ToNetworkOrder(OCM.Contract_desc.Li_strike_price)
	OCM.Contract_desc.Si_ca_level = OCM.ConvertInt16ToNetworkOrder(OCM.Contract_desc.Si_ca_level)
}

func (OCM *OrderConversionManager) ConvertNetHeaderToNetworkOrder(Net_hdr *models.St_net_hdr) {
	OCM.Net_hdr = Net_hdr
	OCM.Net_hdr.I_seq_num = OCM.ConvertInt32ToNetworkOrder(OCM.Net_hdr.I_seq_num)
	OCM.Net_hdr.S_message_length = OCM.ConvertInt16ToNetworkOrder(OCM.Net_hdr.S_message_length)
}

func (OCM *OrderConversionManager) ConvertSignOnReqToNetworkOrder(St_sign *models.St_sign_on_req, int_header *models.St_int_header) {
	OCM.St_sign_on_req = St_sign
	OCM.ConvertIntHeaderToNetworkOrder(int_header)
	OCM.St_sign_on_req.Li_user_id = OCM.ConvertInt64ToNetworkOrder(OCM.St_sign_on_req.Li_user_id)
	OCM.St_sign_on_req.Li_last_password_change_date = OCM.ConvertInt64ToNetworkOrder(OCM.St_sign_on_req.Li_last_password_change_date)
	OCM.St_sign_on_req.Si_branch_id = OCM.ConvertInt16ToNetworkOrder(OCM.St_sign_on_req.Si_branch_id)
	OCM.St_sign_on_req.Li_version_number = OCM.ConvertInt32ToNetworkOrder(OCM.St_sign_on_req.Li_version_number)
	OCM.St_sign_on_req.Li_batch_2_start_time = OCM.ConvertInt32ToNetworkOrder(OCM.St_sign_on_req.Li_batch_2_start_time)
	OCM.St_sign_on_req.Si_user_type = OCM.ConvertInt16ToNetworkOrder(OCM.St_sign_on_req.Si_user_type)
	OCM.St_sign_on_req.D_sequence_number = OCM.ConvertFloat64ToNetworkOrder(OCM.St_sign_on_req.D_sequence_number)
	OCM.St_sign_on_req.Si_member_type = OCM.ConvertInt16ToNetworkOrder(OCM.St_sign_on_req.Si_member_type)
}
