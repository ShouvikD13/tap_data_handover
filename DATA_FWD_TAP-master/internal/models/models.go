package models

import (
	"DATA_FWD_TAP/util"
)

const (
	// ---------------------- for order flags -----------------------
	Flg_ATO          uint16 = 1 << 0  // ATO is the 1st bit
	Flg_Market       uint16 = 1 << 1  // Market is the 2nd bit
	Flg_SL           uint16 = 1 << 2  // SL is the 3rd bit
	Flg_MIT          uint16 = 1 << 3  // MIT is the 4th bit
	Flg_Day          uint16 = 1 << 4  // Day is the 5th bit
	Flg_GTC          uint16 = 1 << 5  // GTC is the 6th bit
	Flg_IOC          uint16 = 1 << 6  // IOC is the 7th bit
	Flg_AON          uint16 = 1 << 7  // AON is the 8th bit
	Flg_MF           uint16 = 1 << 8  // MF is the 9th bit
	Flg_MatchedInd   uint16 = 1 << 9  // MatchedInd is the 10th bit
	Flg_Traded       uint16 = 1 << 10 // Traded is the 11th bit
	Flg_Modified     uint16 = 1 << 11 // Modified is the 12th bit
	Flg_Frozen       uint16 = 1 << 12 // Frozen is the 13th bit
	Flg_OrderFiller1 uint16 = 7 << 13 // Filler1 is the 14th, 15th, and 16th bits (3 bits)

	// ------------------------ for St_broker_eligibility_per_mkt ------------------------
	Flg_NormalMkt     uint16 = 1 << 0    // Normal market is the 1st bit
	Flg_OddlotMkt     uint16 = 1 << 1    // Oddlot market is the 2nd bit
	Flg_SpotMkt       uint16 = 1 << 2    // Spot market is the 3rd bit
	Flg_AuctionMkt    uint16 = 1 << 3    // Auction market is the 4th bit
	Flg_BrokerFiller1 uint16 = 0xF << 4  // Filler1 is the 5th to 8th bits (4 bits)
	Flg_BrokerFiller2 uint16 = 0xFF << 8 // Filler2 is the 9th to 16th bits (8 bits)
)

type Vw_xchngbook struct {
	C_ordr_rfrnc  string // this valiable we are getting from the table
	C_pipe_id     string // this variable we are getting from the args[3]
	C_mod_trd_dt  string // checkout this one, we are getting from table "exg_xchng_mstr"
	L_ord_seq     int32  // this variable we are gettign from the table
	C_slm_flg     string // byte   // this variable we are gettign from the table
	L_dsclsd_qty  int32  // this variable we are gettign from the table
	L_ord_tot_qty int32  // this variable we are gettign from the table
	L_ord_lmt_rt  int32  // this variable we are gettign from the table
	L_stp_lss_tgr int32  // this variable we are gettign from the table
	C_ord_typ     string // byte   // this variable we are gettign from the table
	C_req_typ     string // byte   // this variable we are gettign from the table
	C_valid_dt    string // this variable we are gettign from the table
	C_xchng_cd    string // this we are getting from "opm_ord_pipe_mstr"

	//-----------------------------------------------------------------------------
	C_ex_ordr_type  string  // this one we have to use to use for set we need this for ordinary order type.
	C_oprn_typ      string  // this is we are setting       // this column is not in the 'fxb' table.
	C_plcd_stts     string  // this field is used in "fnRefToOrd" to change the status in "UPDATION_ON_ORDER_FORWARDING"
	L_mdfctn_cntr   int32   // used in same querry as above
	C_frwrd_tm      string  // used in same querry as above
	D_jiffy         float64 //used in "fnRefToOrd" to change the status in "UPDATION_ON_EXCHANGE_RESPONSE"
	L_dwnld_flg     int32   // used in same querry as above // this column is not in the 'fxb' table.
	C_xchng_rmrks   string  // used in last query of "fnRefToOrd"
	C_rms_prcsd_flg string  // used in same query as above
	L_ors_msg_typ   int32   // used in same query as above
	C_ack_tm        string  // used in same querry as above

	//c_ip_addrs, FXB_IP, 'NA'
	//c_prcimpv_flg, FXB_PRCIMPV_FLG 'N'
	// these values are also to be added in the table

}

/*
	value of this structure "Vw_xchngbook" we are extracting from "FXB_FO_XCHNG_BOOK". and all this is happening in "fn_seq_to_omd" in "cln_pack_clnt".
*/

type Vw_orderbook struct {
	C_ordr_rfrnc     string
	C_cln_mtch_accnt string
	C_ordr_flw       string //byte
	L_ord_tot_qty    int32
	L_exctd_qty      int32
	L_exctd_qty_day  int32
	C_settlor        string
	C_spl_flg        string //byte
	C_ack_tm         string
	C_prev_ack_tm    string
	C_pro_cli_ind    string //byte
	C_ctcl_id        string
	C_xchng_ack      string //getting from querry of fetching the order details
	//-----------------------------------------------------

	L_mdfctn_cntr int32  // this is we are changing after we fetch the data from both the tables. (fod and fxb)
	C_ordr_stts   string // it is used in updating the status after the fetcing the orders from the table
	// C_oprn_typ    string // it is used in updating the status after the fetcing the orders from the table

	C_xchng_cd     string
	C_prd_typ      string
	C_undrlyng     string
	C_expry_dt     string
	C_exrc_typ     string
	C_opt_typ      string
	L_strike_prc   int32
	C_ctgry_indstk string
	// L_ca_lvl       int64
}

/*
	values of this structure  "Vw_orderbook". we are extracting from "fod_fo_ordr_dtls" . all this is happening in "fn_ref_to_ord" in "cln_pack_clnt"
*/

type Vw_contract struct {
	L_eba_cntrct_id int64  // null = -1
	C_xchng_cd      string // null = "*"
	C_prd_typ       string // null = '*'
	C_undrlyng      string // null = "*"
	C_expry_dt      string // null = "*"
	C_exrc_typ      string // null = '*'
	C_opt_typ       string // null = '\0'
	L_strike_prc    int32  // null = -1
	C_ctgry_indstk  string // null = '*'
	// L_ca_lvl        int64  // null = -1
	C_rqst_typ string // null = '*'
	C_rout_crt string // null = "*"
}

type Vw_nse_cntrct struct {
	C_prd_typ      string //byte
	C_ctgry_indstk string //byte
	C_symbol       string // this field we are initialising from first query in "fn_get_ext_cnt"
	L_ca_lvl       int32  //this field we are initialising from  query in "fn_get_ext_cnt"
	L_token_id     int32  //this field we are initialising from  query in "fn_get_ext_cnt"

	//----------------------------
	// these all fields are initialising from "fn_get_ext_cnt"
	C_xchng_cd   string
	C_expry_dt   string
	C_exrc_typ   string
	C_opt_typ    string
	L_strike_prc int32
	C_rqst_typ   string
	C_series     string
}

/*
	this value of this structure we are getting from "Vw_contract". which is in "fn_get_ext_cnt" in "cln_pack_clnt"

	 **** problem here is i am not able find the source like from where we are getting the value of the structure "Vw_contract".
*/

type St_opm_pipe_mstr struct {
	/*
		long li_opm_brnch_id;	 // --------------------- (this one is used) from (OPM_ORD_PIPE_MSTR)
		char c_xchng_brkr_id[6]; //--------------------- (this one is used)	 from (exg_xchng_mstr)
		char c_opm_trdr_id[6];	 //--------------------- (this one is used) from (OPM_ORD_PIPE_MSTR)
		int si_user_typ_glb;	 //----------------------(this one is used) "this one we are getting from "cofiguration file" "GetProcessSpaceValue()""

	*/
	L_opm_brnch_id  int64  // null=-1
	C_xchng_brkr_id string // null="*"
	C_opm_trdr_id   string // null="*"
	S_user_typ_glb  int    // null=0   (i think , it can be 0 (for trader) , 4 (for CORPORATE_MANAGER) , 5 (for BRANCH_MANAGER) )

	/* Table name "OPM_ORD_PIPE_MSTR"
	opm_xchng_cd,
	opm_max_pnd_ord,
	opm_stream_no
	OPM_TRDR_ID,
	OPM_BRNCH_ID
	OPM_XCHNG_CD
	OPM_PIPE_ID
	*/

	/* Table Name "exg_xchng_mstr"
	exg_nxt_trd_dt,
	exg_brkr_id,
	exg_ctcl_id
	exg_xchng_cd
	*/
}

/*
	fields of this structure we are getting from different sources.
*/

// ================================ till now all the strctures we are getting as parameters ========================

type St_req_q_data struct {
	L_msg_type       int64
	St_exch_msg_data [338]byte
}

type St_exch_msg struct {
	St_net_header [22]byte  //St_net_hdr
	St_oe_res     [316]byte //St_oe_reqres
}

type St_net_hdr struct { //correct size
	S_message_length int16
	I_seq_num        int32
	C_checksum       [16]byte
}

/* Message Formats (above structure)
Change to Packet Format

• Max length will be the predefined value of 1024 bytes.

Length = size of length field (2 bytes) +
         size of sequence number field (4 bytes) +
         size of the checksum field (16 bytes) +
         size of Message data (variable number of bytes as per the transcode)
*/

type St_oe_reqres struct {
	St_hdr                        [40]byte //St_hdr                        St_int_header
	C_participant_type            byte     // byte
	C_filler_1                    byte     //byte
	Si_competitor_period          int16
	Si_solicitor_period           int16
	C_modified_cancelled_by       byte //byte
	C_filler_2                    byte // byte
	Si_reason_code                int16
	C_filler_3                    [4]byte
	L_token_no                    int32
	St_con_desc                   [28]byte //St_con_desc                        St_contract_desc
	C_counter_party_broker_id     [util.LEN_BROKER_ID]byte
	C_filler_4                    byte //byte
	C_filler_5                    [2]byte
	C_closeout_flg                byte //byte
	C_filler_6                    byte //byte
	Si_order_type                 int16
	D_order_number                float64
	C_account_number              [util.LEN_ACCOUNT_NUMBER]byte
	Si_book_type                  int16
	Si_buy_sell_indicator         int16
	Li_disclosed_volume           int32
	Li_disclosed_volume_remaining int32
	Li_total_volume_remaining     int32
	Li_volume                     int32
	Li_volume_filled_today        int32
	Li_price                      int32
	Li_trigger_price              int32
	Li_good_till_date             int32
	Li_entry_date_time            int32
	Li_minimum_fill_aon_volume    int32
	Li_last_modified              int32
	St_ord_flg                    [2]byte //St_ord_flg                    St_order_flags
	Si_branch_id                  int16
	Li_trader_id                  int32
	C_broker_id                   [util.LEN_BROKER_ID]byte
	C_remarks                     [util.LEN_REMARKS]byte
	C_open_close                  byte // byte
	C_settlor                     [util.LEN_SETTLOR]byte
	Si_pro_client_indicator       int16
	Si_settlement_period          int16
	//================================================
	C_cover_uncover byte // byte   Replaced with 'St_additional_ord_flags'
	C_giveup_flag   byte // byte   Replaced with 'C_reserved_7 '

	/*
		Filler1                       1 bit                             // Size: 1 bit (Offset: 220)
		Filler2                       1 bit                            // Size: 1 bit (Offset: 220)
		Filler3                       1 bit                            // Size: 1 bit (Offset: 220)
		Filler4                       1 bit                            // Size: 1 bit (Offset: 220)
		Filler5                       1 bit                            // Size: 1 bit (Offset: 220)
		Filler6                       1 bit                            // Size: 1 bit (Offset: 220)
		Filler7                       1 bit                            // Size: 1 bit (Offset: 220)
		Filler8                       1 bit                            // Size: 1 bit (Offset: 220)
		Filler9                       1 bit                            // Size: 1 bit (Offset: 221)
		Filler10                      1 bit                            // Size: 1 bit (Offset: 221)
		Filler11                      1 bit                            // Size: 1 bit (Offset: 221)
		Filler12                      1 bit                            // Size: 1 bit (Offset: 221)
		Filler13                      1 bit                            // Size: 1 bit (Offset: 221)
		Filler14                      1 bit                            // Size: 1 bit (Offset: 221)
		Filler15                      1 bit                            // Size: 1 bit (Offset: 221)
		Filler16                      1 bit                            // Size: 1 bit (Offset: 221)
		Filler17                      1 byte                            // Size: 1 byte (Offset: 222)
		Filler18                      1 byte                            // Size: 1 byte (Offset: 223)

	*/
	// All the fillers replace with below field (4 bytes total)
	I_order_seq int32
	//==================================================
	D_nnf_field        float64
	D_filler19         float64 // Replaced with 'MktReplay                     int64 '
	C_pan              [util.LEN_PAN]byte
	L_algo_id          int32
	Si_algo_category   int16
	Ll_lastactivityref int64
	C_reserved         [52]byte
}

/*    Structure from NSE Document of 316 bytes
type St_oe_reqres struct {
St_hdr                        *St_int_header                  // Size: 40 bytes (Offset: 0)
C_participant_type            byte                            // Size: 1 byte (Offset: 40)
C_filler_1                    byte                            // Size: 1 byte (Offset: 41)
Si_competitor_period          int16                           // Size: 2 bytes (Offset: 42)
Si_solicitor_period           int16                           // Size: 2 bytes (Offset: 44)
C_modified_cancelled_by       byte                            // Size: 1 byte (Offset: 46)
C_filler_2                    byte                            // Size: 1 byte (Offset: 47)
Si_reason_code                int16                           // Size: 2 bytes (Offset: 48)
C_filler_3                    [4]byte                         // Size: 4 bytes (Offset: 50)
L_token_no                    int32                           // Size: 4 bytes (Offset: 54)
St_con_desc                   *St_contract_desc               // Size: 28 bytes (Offset: 58)
C_counter_party_broker_id     [util.LEN_BROKER_ID]byte      // Size: 5 bytes (Offset: 86)
C_filler_4                    byte                            // Size: 1 byte (Offset: 91)
C_filler_5                    [2]byte                         // Size: 2 bytes (Offset: 92)
C_closeout_flg                byte                            // Size: 1 byte (Offset: 94)
C_filler_6                    byte                            // Size: 1 byte (Offset: 95)
Si_order_type                 int16                           // Size: 2 bytes (Offset: 96)
D_order_number                float64                         // Size: 8 bytes (Offset: 98)
C_account_number              [util.LEN_ACCOUNT_NUMBER]byte // Size: 10 bytes (Offset: 106)
Si_book_type                  int16                           // Size: 2 bytes (Offset: 116)
Si_buy_sell_indicator         int16                           // Size: 2 bytes (Offset: 118)
Li_disclosed_volume           int32                           // Size: 4 bytes (Offset: 120)
Li_disclosed_volume_remaining int32                           // Size: 4 bytes (Offset: 124)
Li_total_volume_remaining     int32                           // Size: 4 bytes (Offset: 128)
Li_volume                     int32                           // Size: 4 bytes (Offset: 132)
Li_volume_filled_today        int32                           // Size: 4 bytes (Offset: 136)
Li_price                      int32                           // Size: 4 bytes (Offset: 140)
Li_trigger_price              int32                           // Size: 4 bytes (Offset: 144)
Li_good_till_date             int32                           // Size: 4 bytes (Offset: 148)
Li_entry_date_time            int32                           // Size: 4 bytes (Offset: 152)
Li_minimum_fill_aon_volume    int32                           // Size: 4 bytes (Offset: 156)
Li_last_modified              int32                           // Size: 4 bytes (Offset: 160)
St_ord_flg                    *St_order_flags                 // Size: 2 bytes (Offset: 164)
Si_branch_id                  int16                           // Size: 2 bytes (Offset: 166)
Li_trader_id                  int32                           // Size: 4 bytes (Offset: 168)
C_broker_id                   [util.LEN_BROKER_ID]byte      // Size: 5 bytes (Offset: 172)
C_ord_filler                  [util.LEN_REMARKS]byte        // Size: 24 bytes (Offset: 177)
C_open_close                  byte                            // Size: 1 byte (Offset: 201)
C_settlor                     [util.LEN_SETTLOR]byte        // Size: 12 bytes (Offset: 202)
Si_pro_client_indicator       int16                           // Size: 2 bytes (Offset: 214)
Si_settlement_period          int16                           // Size: 2 bytes (Offset: 216)
St_additional_ord_flags       *St_order_flags                 // Size: 1 byte (Offset: 218)
C_reserved_7                  byte                            // Size: 1 byte (Offset: 219)
Filler1                       byte                             // Size: 1 bit (Offset: 220)
Filler2                       byte                            // Size: 1 bit (Offset: 220)
Filler3                       byte                            // Size: 1 bit (Offset: 220)
Filler4                       byte                            // Size: 1 bit (Offset: 220)
Filler5                       byte                            // Size: 1 bit (Offset: 220)
Filler6                       byte                            // Size: 1 bit (Offset: 220)
Filler7                       byte                            // Size: 1 bit (Offset: 220)
Filler8                       byte                            // Size: 1 bit (Offset: 220)
Filler9                       byte                            // Size: 1 bit (Offset: 221)
Filler10                      byte                            // Size: 1 bit (Offset: 221)
Filler11                      byte                            // Size: 1 bit (Offset: 221)
Filler12                      byte                            // Size: 1 bit (Offset: 221)
Filler13                      byte                            // Size: 1 bit (Offset: 221)
Filler14                      byte                            // Size: 1 bit (Offset: 221)
Filler15                      byte                            // Size: 1 bit (Offset: 221)
Filler16                      byte                            // Size: 1 bit (Offset: 221)
Filler17                      byte                            // Size: 1 byte (Offset: 222)
Filler18                      byte                            // Size: 1 byte (Offset: 223)
D_nnf_field                   float64                         // Size: 8 bytes (Offset: 224)
MktReplay                     int64                           // <-------------                    // Size: 8 bytes (Offset: 232)
C_pan                         [util.LEN_PAN]byte            // Size: 10 bytes (Offset: 240)
L_algo_id                     int32                           // Size: 4 bytes (Offset: 250)
Si_algo_category              int16                           // Size: 2 bytes (Offset: 254)
Ll_lastactivityref            int64                           // Size: 8 bytes (Offset: 256)
C_reserved_8                  [52]byte                        // Size: 52 bytes (Offset: 264)
}
*/

type St_int_header struct { // correct size
	Si_transaction_code int16
	Li_log_time         int32
	C_alpha_char        [util.LEN_ALPHA_CHAR]byte // const util.LEN_ALPHA_CHAR untyped int = 2
	Li_trader_id        int32
	Si_error_code       int16
	C_filler_2          [8]byte
	C_time_stamp_1      [util.LEN_TIME_STAMP]byte // const util.LEN_TIME_STAMP untyped int = 8
	C_time_stamp_2      [util.LEN_TIME_STAMP]byte // const util.LEN_TIME_STAMP untyped int = 8
	Si_message_length   int16                     // 316 size
}

/* Structure Name: MESSAGE_HEADER
Packet Length: 40 bytes

| Field Name      | Data Type | Size (bytes) | Offset |
|-----------------|-----------|--------------|--------|
| TransactionCode | SHORT     | 2            | 0      |
| LogTime         | LONG      | 4            | 2      |
| AlphaChar       | CHAR      | 2            | 6      |
| UserId          | LONG      | 4            | 8      |
| ErrorCode       | SHORT     | 2            | 12     |
| Timestamp       | LONGLONG  | 8            | 14     |
| TimeStamp1      | CHAR      | 8            | 22     |
| TimeStamp2      | CHAR      | 8            | 30     |
| MessageLength   | SHORT     | 2            | 38     |
*/

type St_contract_desc struct { // correct size
	C_instrument_name [util.LEN_INSTRUMENT_NAME]byte
	C_symbol          [util.LEN_SYMBOL_NSE]byte
	Li_expiry_date    int32
	Li_strike_price   int32
	C_option_type     [util.LEN_OPTION_TYPE]byte
	Si_ca_level       int16
}

/*Structure Name: CONTRACT_DESC
Packet Length: 28 bytes

| Field Name      | Data Type | Size (bytes) | Offset |
|-----------------|-----------|--------------|--------|
| InstrumentName  | CHAR      | 6            | 0      |
| Symbol          | CHAR      | 10           | 6      |
| ExpiryDate      | LONG      | 4            | 16     |
| StrikePrice     | LONG      | 4            | 20     |
| OptionType      | CHAR      | 2            | 24     |
| CALevel         | SHORT     | 2            | 26     |
*/

type St_order_flags struct { //correct size
	Flags uint16
}

/* Structure Name: ST_ORDER_FLAGS
Packet Length: 2 bytes

| Field Name  | Data Type | Size (bits) | Offset |
|-------------|-----------|-------------|--------|
| ATO         | BIT       | 1           | 0      |
| Market      | BIT       | 1           | 0      |
| SL          | BIT       | 1           | 0      |
| MIT         | BIT       | 1           | 0      |
| Day         | BIT       | 1           | 0      |
| GTC         | BIT       | 1           | 0      |
| IOC         | BIT       | 1           | 0      |
| AON         | BIT       | 1           | 0      |
| MF          | BIT       | 1           | 1      |
| MatchedInd  | BIT       | 1           | 1      |
| Traded      | BIT       | 1           | 1      |
| Modified    | BIT       | 1           | 1      |
| Frozen      | BIT       | 1           | 1      |
| Reserved    | BIT       | 3           | 1      |
*/

func (o *St_order_flags) SetFlag(flag uint16) {
	o.Flags |= flag
}

func (o *St_order_flags) ClearFlag(flag uint16) {
	o.Flags &^= flag
}

func (o *St_order_flags) GetFlagValue(flag uint16) bool {
	return o.Flags&flag != 0
}

//-------------------------------------------------------------------------------------------------------------------

// type St_pk_sequence struct { Not Used
// 	C_pipe_id  byte
// 	C_trd_dt   string
// 	C_rqst_typ string
// 	C_seq_num  int32
// }

// type St_addtnal_order_flags struct {  Not Used
// 	C_cover_uncover string //byte
// }

//***************************************************************************************************************
//******************************************* LogOn Structures **************************************************

type St_sign_on_req struct {
	St_hdr                       [40]byte //St_int_header
	Li_user_id                   int64
	C_reserved_1                 [8]byte
	C_password                   [util.LEN_PASSWORD]byte
	C_reserved_2                 [8]byte
	C_new_password               [util.LEN_PASSWORD]byte
	C_trader_name                [util.LEN_TRADER_NAME]byte
	Li_last_password_change_date int64
	C_broker_id                  [util.LEN_BROKER_ID]byte
	C_filler_1                   byte
	Si_branch_id                 int16
	Li_version_number            int32
	Li_batch_2_start_time        int32
	C_host_switch_context        byte
	C_colour                     [util.LEN_COLOUR]byte
	C_filler_2                   byte
	Si_user_type                 int16
	D_sequence_number            float64
	C_ws_class_name              [util.LEN_WS_CLASS_NAME]byte
	C_broker_status              byte
	C_show_index                 byte
	St_mkt_allwd_lst             [2]byte //St_broker_eligibility_per_mkt // 2byte array
	Si_member_type               int16
	C_clearing_status            byte
	C_broker_name                [util.LEN_BROKER_NAME]byte
	C_reserved_3                 [16]byte
	C_reserved_4                 [16]byte
	C_reserved_5                 [16]byte
}

/*
Structure Name: MS_SIGNON
Packet Length: 278 bytes
Transaction Code: SIGN_ON_REQUEST_IN (2300)

Field Breakdown:
------------------------------------------------
Field Name                       	Data Type   	Size (Bytes)   Offset
------------------------------------------------
MESSAGE_HEADER (Refer to Message Header)     STRUCT           40             0
UserID                                LONG             4             40
Reserved                              CHAR             8             44
Password                              CHAR             8             52
Reserved                              CHAR             8             60
NewPassword                           CHAR             8             68
TraderName                            CHAR             26            76
LastPasswordChangeDate                LONG             4             102
BrokerID                              CHAR             5             106
Reserved                              CHAR             1             111
BranchID                              SHORT            2             112
VersionNumber                         LONG             4             114
Batch2StartTime                       LONG             4             118
HostSwitchContext                     CHAR             1             122
Colour                                CHAR             50            123
Reserved                              CHAR             1             173
UserType                              SHORT            2             174
SequenceNumber                        DOUBLE           8             176
WsClassName                           CHAR             14            184
BrokerStatus                          CHAR             1             198
ShowIndex                             CHAR             1             199
ST_BROKER_ELIGIBILITY_PER_MKT         STRUCT           2             200
MemberType                            SHORT            2             202
ClearingStatus                        CHAR             1             204
BrokerName                            CHAR             25            205
Reserved                              CHAR             16            230
Reserved                              CHAR             16            246
Reserved                              CHAR             16            262

Total Packet Length: 278 bytes
*/

type St_sign_on_res struct {
	St_hdr                       St_int_header
	Li_user_id                   int64
	C_reserved_1                 [8]byte
	C_password                   [util.LEN_PASSWORD]byte
	C_reserved_2                 [8]byte
	C_new_password               [util.LEN_PASSWORD]byte
	C_trader_name                [util.LEN_TRADER_NAME]byte
	Li_last_password_change_date int64
	C_broker_id                  [util.LEN_BROKER_ID]byte
	C_filler_1                   byte
	Si_branch_id                 int16
	Li_version_number            int64
	Li_end_time                  int64
	C_filler_2                   [52]byte
	Si_user_type                 int16
	D_sequence_number            float64
	C_filler_3                   [14]byte
	C_broker_status              byte
	C_show_index                 byte
	St_mkt_allwd_lst             St_broker_eligibility_per_mkt
	Si_member_type               int16
	C_clearing_status            byte
	C_broker_name                [util.LEN_BROKER_NAME]byte
	C_reserved_3                 [16]byte
	C_reserved_4                 [16]byte
	C_reserved_5                 [16]byte
}

type St_broker_eligibility_per_mkt struct {
	Flags uint16
}

func (b *St_broker_eligibility_per_mkt) SetFlag(flag uint16) {
	b.Flags |= flag
}

func (b *St_broker_eligibility_per_mkt) ClearFlag(flag uint16) {
	b.Flags &^= flag
}

func (b *St_broker_eligibility_per_mkt) GetFlagValue(flag uint16) bool {
	return b.Flags&flag != 0
}

type St_exch_snd_msg struct {
	St_signon_req St_sign_on_req
	St_oe_res     St_oe_reqres
}

/*
Structure Name: ST_BROKER_ELIGIBILITY_PER_MKT
Packet Length: 2 bytes

Field Breakdown (For Small Endian Machines):
------------------------------------------------
Field Name                     Data Type   Size (Bytes)   Offset
------------------------------------------------
Reserved                        BIT         4              0
Auction Market                  BIT         1              0
Spot Market                     BIT         1              0
Oddlot Market                   BIT         1              0
Normal Market                   BIT         1              0
Reserved Byte                   BYTE        1              1

Field Breakdown (For Big Endian Machines):
------------------------------------------------
Field Name                     Data Type   Size (Bytes)   Offset
------------------------------------------------
Normal Market                   BIT         1              0
Oddlot Market                   BIT         1              0
Spot Market                     BIT         1              0
Auction Market                  BIT         1              0
Reserved                        BIT         4              0
Reserved Byte                   BYTE        1              1

Total Packet Length: 2 bytes
*/

type St_exch_msg_Log_On struct {
	St_net_header  [22]byte  //St_net_hdr
	St_sign_on_req [278]byte //St_sign_on_req
}

type St_exch_msg_Log_Off struct {
	St_net_header [22]byte //St_net_hdr
	St_int_header [40]byte //St_int_header
}

type St_req_q_data_Log_On struct {
	L_msg_type         int64
	St_exch_msg_Log_On [300]byte
}

type St_req_q_data_Log_Off struct {
	L_msg_type          int64
	St_exch_msg_Log_Off [62]byte
}

//----------------------- Functions For implementation of Message Queue---------------------

// For St_req_q_data
func (data *St_req_q_data) GetMsgType() int64 {
	return data.L_msg_type
}

// func (data *St_req_q_data) SetMsgType(msgType int64) {
// 	data.L_msg_type = msgType
// }

// For St_req_q_data_Log_On
func (data *St_req_q_data_Log_On) GetMsgType() int64 {
	return data.L_msg_type
}

// func (data *St_req_q_data_Log_On) SetMsgType(msgType int64) {
// 	data.L_msg_type = msgType
// }

// For St_req_q_data_Log_Off
func (data *St_req_q_data_Log_Off) GetMsgType() int64 {
	return data.L_msg_type
}

// func (data *St_req_q_data_Log_Off) SetMsgType(msgType int64) {
// 	data.L_msg_type = msgType
// }

// SHOUVIK
type St_exch_msg_resp struct {
	St_net_header St_net_hdr
	St_oe_res     St_oe_reqresp
}

type St_exch_msg_Log_On_use struct {
	St_net_header  [22]byte //St_net_hdr
	St_sign_on_req St_sign_on_req_use
}

type Vw_mkt_msg struct {
	CXchngCd [4]byte   `json:"c_xchng_cd"` // Equivalent to `char[4]`, 3 characters + null terminator
	CBrkrId  [6]byte   `json:"c_brkr_id"`  // Equivalent to `char[6]`, 5 characters + null terminator
	CMsgId   byte      `json:"c_msg_id"`   // Equivalent to `char`, single byte
	CMsg     [256]byte `json:"c_msg"`      // Equivalent to `char[256]`
	CTmStmp  [23]byte  `json:"c_tm_stmp"`  // Equivalent to `char[23]`
	CRoutCrt [4]byte   `json:"c_rout_crt"` // Equivalent to `char[4]`, 3 characters + null terminator
}

type St_err_rspns struct {
	StHdr         St_int_header                // Nested struct
	CKey          [util.LEN_ERROR_KEY]byte     // Equivalent to char[LEN_ERROR_KEY]
	CErrorMessage [util.LEN_ERROR_MESSAGE]byte // Equivalent to char[LEN_ERROR_MESSAGE]
}

//By Shouvik

type St_oe_reqresp struct {
	St_hdr                        St_int_header
	C_participant_type            byte // byte
	C_filler_1                    byte //byte
	Si_competitor_period          int16
	Si_solicitor_period           int16
	C_modified_cancelled_by       byte //byte
	C_filler_2                    byte // byte
	Si_reason_code                int16
	C_filler_3                    [4]byte
	L_token_no                    int32
	St_con_desc                   [28]byte //St_con_desc                        St_contract_desc
	C_counter_party_broker_id     [util.LEN_BROKER_ID]byte
	C_filler_4                    byte //byte
	C_filler_5                    [2]byte
	C_closeout_flg                byte //byte
	C_filler_6                    byte //byte
	Si_order_type                 int16
	D_order_number                float64
	C_account_number              [util.LEN_ACCOUNT_NUMBER]byte
	Si_book_type                  int16
	Si_buy_sell_indicator         int16
	Li_disclosed_volume           int32
	Li_disclosed_volume_remaining int32
	Li_total_volume_remaining     int32
	Li_volume                     int32
	Li_volume_filled_today        int32
	Li_price                      int32
	Li_trigger_price              int32
	Li_good_till_date             int32
	Li_entry_date_time            int32
	Li_minimum_fill_aon_volume    int32
	Li_last_modified              int32
	St_ord_flg                    St_order_flags
	Si_branch_id                  int16
	Li_trader_id                  int32
	C_broker_id                   [util.LEN_BROKER_ID]byte
	C_remarks                     [util.LEN_REMARKS]byte
	C_open_close                  byte // byte
	C_settlor                     [util.LEN_SETTLOR]byte
	Si_pro_client_indicator       int16
	Si_settlement_period          int16
	//================================================
	C_cover_uncover byte // byte   Replaced with 'St_additional_ord_flags'
	C_giveup_flag   byte // byte   Replaced with 'C_reserved_7 '

	/*
		Filler1                       1 bit                             // Size: 1 bit (Offset: 220)
		Filler2                       1 bit                            // Size: 1 bit (Offset: 220)
		Filler3                       1 bit                            // Size: 1 bit (Offset: 220)
		Filler4                       1 bit                            // Size: 1 bit (Offset: 220)
		Filler5                       1 bit                            // Size: 1 bit (Offset: 220)
		Filler6                       1 bit                            // Size: 1 bit (Offset: 220)
		Filler7                       1 bit                            // Size: 1 bit (Offset: 220)
		Filler8                       1 bit                            // Size: 1 bit (Offset: 220)
		Filler9                       1 bit                            // Size: 1 bit (Offset: 221)
		Filler10                      1 bit                            // Size: 1 bit (Offset: 221)
		Filler11                      1 bit                            // Size: 1 bit (Offset: 221)
		Filler12                      1 bit                            // Size: 1 bit (Offset: 221)
		Filler13                      1 bit                            // Size: 1 bit (Offset: 221)
		Filler14                      1 bit                            // Size: 1 bit (Offset: 221)
		Filler15                      1 bit                            // Size: 1 bit (Offset: 221)
		Filler16                      1 bit                            // Size: 1 bit (Offset: 221)
		Filler17                      1 byte                            // Size: 1 byte (Offset: 222)
		Filler18                      1 byte                            // Size: 1 byte (Offset: 223)

	*/
	// All the fillers replace with below field (4 bytes total)
	I_order_seq int32
	//==================================================
	D_nnf_field        float64
	D_filler19         float64 // Replaced with 'MktReplay                     int64 '
	C_pan              [util.LEN_PAN]byte
	L_algo_id          int32
	Si_algo_category   int16
	Ll_lastactivityref int64
	C_reserved         [52]byte
}

type St_sign_on_req_use struct {
	St_hdr                       St_int_header
	Li_user_id                   int64
	C_reserved_1                 [8]byte
	C_password                   [util.LEN_PASSWORD]byte
	C_reserved_2                 [8]byte
	C_new_password               [util.LEN_PASSWORD]byte
	C_trader_name                [util.LEN_TRADER_NAME]byte
	Li_last_password_change_date int64
	C_broker_id                  [util.LEN_BROKER_ID]byte
	C_filler_1                   byte
	Si_branch_id                 int16
	Li_version_number            int32
	Li_batch_2_start_time        int32
	C_host_switch_context        byte
	C_colour                     [util.LEN_COLOUR]byte
	C_filler_2                   byte
	Si_user_type                 int16
	D_sequence_number            float64
	C_ws_class_name              [util.LEN_WS_CLASS_NAME]byte
	C_broker_status              byte
	C_show_index                 byte
	St_mkt_allwd_lst             [2]byte //St_broker_eligibility_per_mkt // 2byte array
	Si_member_type               int16
	C_clearing_status            byte
	C_broker_name                [util.LEN_BROKER_NAME]byte
	C_reserved_3                 [16]byte
	C_reserved_4                 [16]byte
	C_reserved_5                 [16]byte
}

/*    Structure from NSE Document of 316 bytes
type St_oe_reqres struct {
St_hdr                        *St_int_header                  // Size: 40 bytes (Offset: 0)
C_participant_type            byte                            // Size: 1 byte (Offset: 40)
C_filler_1                    byte                            // Size: 1 byte (Offset: 41)
Si_competitor_period          int16                           // Size: 2 bytes (Offset: 42)
Si_solicitor_period           int16                           // Size: 2 bytes (Offset: 44)
C_modified_cancelled_by       byte                            // Size: 1 byte (Offset: 46)
C_filler_2                    byte                            // Size: 1 byte (Offset: 47)
Si_reason_code                int16                           // Size: 2 bytes (Offset: 48)
C_filler_3                    [4]byte                         // Size: 4 bytes (Offset: 50)
L_token_no                    int32                           // Size: 4 bytes (Offset: 54)
St_con_desc                   *St_contract_desc               // Size: 28 bytes (Offset: 58)
C_counter_party_broker_id     [util.LEN_BROKER_ID]byte      // Size: 5 bytes (Offset: 86)
C_filler_4                    byte                            // Size: 1 byte (Offset: 91)
C_filler_5                    [2]byte                         // Size: 2 bytes (Offset: 92)
C_closeout_flg                byte                            // Size: 1 byte (Offset: 94)
C_filler_6                    byte                            // Size: 1 byte (Offset: 95)
Si_order_type                 int16                           // Size: 2 bytes (Offset: 96)
D_order_number                float64                         // Size: 8 bytes (Offset: 98)
C_account_number              [util.LEN_ACCOUNT_NUMBER]byte // Size: 10 bytes (Offset: 106)
Si_book_type                  int16                           // Size: 2 bytes (Offset: 116)
Si_buy_sell_indicator         int16                           // Size: 2 bytes (Offset: 118)
Li_disclosed_volume           int32                           // Size: 4 bytes (Offset: 120)
Li_disclosed_volume_remaining int32                           // Size: 4 bytes (Offset: 124)
Li_total_volume_remaining     int32                           // Size: 4 bytes (Offset: 128)
Li_volume                     int32                           // Size: 4 bytes (Offset: 132)
Li_volume_filled_today        int32                           // Size: 4 bytes (Offset: 136)
Li_price                      int32                           // Size: 4 bytes (Offset: 140)
Li_trigger_price              int32                           // Size: 4 bytes (Offset: 144)
Li_good_till_date             int32                           // Size: 4 bytes (Offset: 148)
Li_entry_date_time            int32                           // Size: 4 bytes (Offset: 152)
Li_minimum_fill_aon_volume    int32                           // Size: 4 bytes (Offset: 156)
Li_last_modified              int32                           // Size: 4 bytes (Offset: 160)
St_ord_flg                    *St_order_flags                 // Size: 2 bytes (Offset: 164)
Si_branch_id                  int16                           // Size: 2 bytes (Offset: 166)
Li_trader_id                  int32                           // Size: 4 bytes (Offset: 168)
C_broker_id                   [util.LEN_BROKER_ID]byte      // Size: 5 bytes (Offset: 172)
C_ord_filler                  [util.LEN_REMARKS]byte        // Size: 24 bytes (Offset: 177)
C_open_close                  byte                            // Size: 1 byte (Offset: 201)
C_settlor                     [util.LEN_SETTLOR]byte        // Size: 12 bytes (Offset: 202)
Si_pro_client_indicator       int16                           // Size: 2 bytes (Offset: 214)
Si_settlement_period          int16                           // Size: 2 bytes (Offset: 216)
St_additional_ord_flags       *St_order_flags                 // Size: 1 byte (Offset: 218)
C_reserved_7                  byte                            // Size: 1 byte (Offset: 219)
Filler1                       byte                             // Size: 1 bit (Offset: 220)
Filler2                       byte                            // Size: 1 bit (Offset: 220)
Filler3                       byte                            // Size: 1 bit (Offset: 220)
Filler4                       byte                            // Size: 1 bit (Offset: 220)
Filler5                       byte                            // Size: 1 bit (Offset: 220)
Filler6                       byte                            // Size: 1 bit (Offset: 220)
Filler7                       byte                            // Size: 1 bit (Offset: 220)
Filler8                       byte                            // Size: 1 bit (Offset: 220)
Filler9                       byte                            // Size: 1 bit (Offset: 221)
Filler10                      byte                            // Size: 1 bit (Offset: 221)
Filler11                      byte                            // Size: 1 bit (Offset: 221)
Filler12                      byte                            // Size: 1 bit (Offset: 221)
Filler13                      byte                            // Size: 1 bit (Offset: 221)
Filler14                      byte                            // Size: 1 bit (Offset: 221)
Filler15                      byte                            // Size: 1 bit (Offset: 221)
Filler16                      byte                            // Size: 1 bit (Offset: 221)
Filler17                      byte                            // Size: 1 byte (Offset: 222)
Filler18                      byte                            // Size: 1 byte (Offset: 223)
D_nnf_field                   float64                         // Size: 8 bytes (Offset: 224)
MktReplay                     int64                           // <-------------                    // Size: 8 bytes (Offset: 232)
C_pan                         [util.LEN_PAN]byte            // Size: 10 bytes (Offset: 240)
L_algo_id                     int32                           // Size: 4 bytes (Offset: 250)
Si_algo_category              int16                           // Size: 2 bytes (Offset: 254)
Ll_lastactivityref            int64                           // Size: 8 bytes (Offset: 256)
C_reserved_8                  [52]byte                        // Size: 52 bytes (Offset: 264)
}
*/

// Shouvik
type St_system_info_data struct {
	St_hdr                                St_int_header
	St_mkt_stts                           St_market_status
	St_ex_mkt_stts                        St_ex_market_status
	St_pl_mkt_stts                        St_pl_market_status
	C_update_portfolio                    byte
	Li_market_index                       int64
	Si_default_sttlmnt_period_nm          int16
	Si_default_sttlmnt_period_sp          int16
	Si_default_sttlmnt_period_au          int16
	Si_competitor_period                  int16
	Si_solicitor_period                   int16
	Si_warning_percent                    int16
	Si_volume_freeze_percent              int16
	C_filler_1                            [4]byte
	Li_board_lot_quantity                 int64
	Li_tick_size                          int64
	Si_maximum_gtc_days                   int16
	St_stk_elg_ind                        St_stock_eligible_indicators
	Si_disclosed_quantity_percent_allowed int16
	Li_risk_free_interest_rate            int64
}

type St_system_info_req struct {
	St_Hdr                        St_int_header
	Li_last_update_protfolio_time int64
}

type St_market_status struct {
	Si_normal  int16
	Si_oddlot  int16
	Si_spot    int16
	Si_auction int16
}

type St_ex_market_status struct {
	Si_normal  int16
	Si_oddlot  int16
	Si_spot    int16
	Si_auction int16
}

type St_pl_market_status struct {
	Si_normal  int16
	Si_oddlot  int16
	Si_spot    int16
	Si_auction int16
}

type St_stock_eligible_indicators struct {
	Flg_aon          uint32 `json:"flg_aon"`
	Flg_minimum_fill uint32 `json:"flg_minimum_fill"`
	Flg_books_merged uint32 `json:"flg_books_merged"`
	Flg_filler1      uint32 `json:"flg_filler1"`
	Flg_filler2      uint32 `json:"flg_filler2"`
}

type Vw_xchngstts struct {
	C_xchng_cd       [4]byte  // null="*"
	C_nxt_trd_dt     [23]byte // null="*"
	L_opn_tm         int64    // null=-1
	L_cls_tm         int64    // null=-1
	C_crrnt_stts     byte     // null='*'
	L_exrc_rq_opn_tm int64    // null=-1
	L_exrc_rq_cls_tm int64    // null=-1
	C_exrc_mkt_stts  byte     // null='*'
	C_mkt_typ        byte     // null='*'
	C_brkr_id        [6]byte  // null="*"
	C_brkr_stts      byte     // null='*'
	C_settlor        [6]byte  // null="*"
	C_settlor_stts   byte     // null='*'
	C_rqst_typ       byte     // null='*'
	C_rout_crt       [4]byte  // null="*"
}

type C_err_msg [256]byte

type St_exch_msg_sys_info_req struct {
	St_net_hdr         St_net_hdr
	St_system_info_req St_system_info_req
}

type St_exch_msg_ldb_upd struct {
	St_net_hdr St_net_hdr
	St_ldb_upd St_update_local_database
}

type Signal struct {
	Message      string
	ErrorMessage string
}

type St_update_local_database struct {
	St_Hdr                         St_int_header
	Li_LastUpdate_Security_Time    int32
	Li_LastUpdate_Participant_Time int32
	Li_LastUpdate_Instrument_Time  int32
	Li_Last_Update_Index_Time      int32
	C_Request_For_Open_Orders      byte // Use byte for char fields in Go
	C_Filler_1                     byte // Use byte for filler
	St_Mkt_Stts                    St_market_status
	St_ExMkt_Stts                  St_ex_market_status
	St_Pl_Mkt_Stts                 St_pl_market_status
}
