package util

import "log"

const (
	/*upd xchng book*/
	UPDATION_ON_ORDER_FORWARDING  = 'P'
	UPDATION_ON_EXCHANGE_RESPONSE = 'R'

	/*upd ordr book*/
	UPDATE_ORDER_STATUS = 'S'

	/**/
	ORDINARY_ORDER = 'O'

	FOR_NORM = 'N'

	/*placed status*/
	REQUESTED = 'R'
	QUEUED    = 'Q'
	CANCELLED = 'C'

	DOWNLOAD     = 10
	NOT_DOWNLOAD = 20

	CONTRACT_TO_NSE_ID = 'N'

	/**** SFO_GT_EXT_CNT ****/
	NFO_ENTTY = 1
	BFO_ENTTY = 2

	/* product type */
	FUTURES = 'F'

	// on reject order
	ORS_NEW_ORD_RJCT = 5304

	//length
	LEN_PASSWORD        = 8
	LEN_TRADER_NAME     = 26
	LEN_BROKER_NAME     = 25
	LEN_COLOUR          = 50
	LEN_WS_CLASS_NAME   = 14
	LEN_ALPHA_CHAR      = 2
	LEN_SETTLOR         = 12
	LEN_BROKER_ID       = 5
	LEN_FILLER_OPTIONS  = 3
	LEN_ACCOUNT_NUMBER  = 10
	LEN_REMARKS         = 24
	LEN_TIME_STAMP      = 8
	LEN_INSTRUMENT_NAME = 6
	LEN_SYMBOL_NSE      = 10
	LEN_OPTION_TYPE     = 2
	LEN_PAN             = 10 // this value i have deduce from original C code "char c_pan[10];				  /*** Ver 2.7 ***/"
	// Order & Trade management
	BOARD_LOT_IN                 = 2000
	LOGIN_WITH_OPEN_ORDR_DTLS    = 5104 //Test use only NON FUNCTIONAL
	LOGIN_WITHOUT_OPEN_ORDR_DTLS = 5105
	SIGN_OFF_REQUEST_IN          = 2320
	//Added by Shouvik
	PRICE_CONFIRMATION     = 2012
	ORDER_CONFIRMATION_OUT = 2073
	LEN_PASS_WD            = 8 + 1
	LEN_DT                 = 12
	LEN_ERROR_MESSAGE      = 128
	LEN_ERROR_KEY          = 14
	ACTIVE                 = 'A'
	CLOSE_OUT              = 'C'
	SYSTEM_ERROR           = -100
	SIGN_ON_REQUEST_OUT    = 2301
	PRE_OPEN               = 0
	OPEN                   = 1
	CLOSED                 = 2
	PRE_OPEN_ENDED         = 3
	POST_CLOSE             = 4

	TRADER            = 'T'
	CORPORATE_MANAGER = 'M'
	BRANCH_MANAGER    = 'B'

	//for Expiry DT
	TEN_YRS_IN_SEC = 315513000

	/**** Order request type ****/
	NEW = 'N'
	// MODIFY = "M"
	// CANCEL = "C"

	//Book type

	REGULAR_LOT_ORDER   = 1
	STOP_LOSS_MIT_ORDER = 3

	//Buy Sell indicator
	NSE_BUY  = 1
	NSE_SELL = 2

	// Order placed by Broker / Customer Indicator

	BRKR_PLCD  = 'P'
	CSTMR_PLCD = 'C'

	//
	NSE_CLIENT = 1
	NSE_PRO    = 2

	FO_AUTO_MTM_ALG_ID      = 24264
	FO_AUTO_MTM_ALG_CAT_ID  = 99
	FO_PRICE_IMP_ALG_ID     = 37383
	FO_PRICE_IMP_ALG_CAT_ID = 1
	FO_PRFT_ORD_ALG_CAT_ID  = 1
	FO_PRFT_ORD_ALG_ID      = 82686
	FO_FLASH_TRD_ALG_CAT_ID = 1
	FO_FLASH_TRD_ALG_ID     = 89385
	FO_NON_ALG_ID           = 0
	FO_NON_ALG_CAT_ID       = 0
	FO_FOGTT_ALG_ID         = 98816
	FO_FOGTT_ALG_CAT_ID     = 1

	MAX_SEQ_NUM = 2147483647

	// LOGON Shouvik
	MAX_PSSWD_LEN      = 10
	CRNT_PSSWD_LEN     = 8
	SIGN_ON_REQUEST_IN = 2300

	/**** SFO_GET_SEQ ****/
	GET_PLACED_SEQ = 'P'

	/**** Temporary Message Queue ID's ****/
	ORDINARY_ORDER_QUEUE_ID = 1
	LOGON_QUEUE_ID          = 2
	LOGOFF_QUEUE_ID         = 3
)

/*********************************************************************/
/*                                                                   */
/*  Description : Sets a null character ('\0') at a specified        */
/*                position in a byte slice, marking the end of the   */
/*                string in the slice.                               */
/*                                                                   */
/*********************************************************************/

type FO_Manager struct {
}

func (FOM *FO_Manager) SETNULL(serviceName string, a []byte, length int) {
	if length < len(a) {
		a[length] = 0
		log.Printf("[%s] Set null character at position %d in byte slice", serviceName, length)
	} else {
		log.Printf("[%s] Length %d exceeds byte slice size %d. No action taken.", serviceName, length, len(a))
	}
}
