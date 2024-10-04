package esr

import (
	//"bytes"
	//"encoding/binary"

	"DATA_FWD_TAP/internal/models"
	"DATA_FWD_TAP/util"
	"DATA_FWD_TAP/util/MessageQueue"
	"DATA_FWD_TAP/util/OrderConversion"
	typeconversionutil "DATA_FWD_TAP/util/TypeConversionUtil"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"

	"time"

	"unsafe"

	"gorm.io/gorm"
)

type ESRManager struct {
	Req_q_data             *models.St_req_q_data
	ServiceName            string
	St_exch_msg            *models.St_exch_msg
	St_net_hdr             *models.St_net_hdr
	St_oe_reqres           *models.St_oe_reqres
	St_sign_on_req         *models.St_sign_on_req
	St_exch_msg_Log_on     *models.St_exch_msg_Log_On
	St_exch_msg_Log_on_use *models.St_exch_msg_Log_On_use
	St_exch_msg_resp       *models.St_exch_msg_resp
    OCM                    *OrderConversion.OrderConversionManager
	ENVM                  *util.EnvironmentManager
	Req_q_data1           models.St_req_q_data
	PUM                   *util.PasswordUtilManger
	Message_queue_manager *MessageQueue.MessageQueueManager
	TCUM                  *typeconversionutil.TypeConversionUtilManager
	LoggerManager         *util.LoggerManager
	TM                    *util.TransactionManager
	Mtype                 *int
	Max_Pack_Val          int
	Db                    *gorm.DB
	
}

var li_send_tap_msg_size int64
var li_business_data_size int64
var c_rqst_for_open_ordr byte

// Global Use Variables
var Sql_c_xchng_cd [3 + 1]byte
var Sql_C_nxt_trd_date [util.LEN_DATE]byte
var Sql_c_pipe_id [2 + 1]byte
var C_auto_connect byte
var C_server_flg byte
var Sql_C_trd_ref [8 + 1]byte
var Sql_sec_tm [util.LEN_DATE]byte
var Sql_part_tm [util.LEN_DATE]byte
var Sql_inst_tm [util.LEN_DATE]byte
var Sql_idx_tm [util.LEN_DATE]byte

var (
	messageQueue = make(chan models.St_req_q_data, 20)
	//logonResponseChan = make(chan models.St_exch_msg_so)
	wg        sync.WaitGroup
	exitcount int64
)

type Config struct {
	IP            string
	Port          string
	AutoReconnect bool
}

/*
type exch_msg struct {
	TestData1 int64
	TestData2 byte
	TestData3 [20]byte // Fixed-size byte array for the string
}

type req_q_data struct {
	li_msg_type   int64
	exch_msg_data exch_msg
} */

func (ESRM *ESRManager) loadConfig() (Config, error) {
	// cfg, err := ini.Load("EnvConfig.ini")
	// if err != nil {
	// 	return Config{}, fmt.Errorf("failed to read config file: %v", err)
	// }

	autoReconnect := false //ESRM.ENVM.GetProcessSpaceValue("server", "auto_reconnect")
	// if err != nil {
	// 	return Config{}, fmt.Errorf("failed to read auto_reconnect config: %v", err)
	// }

	config := Config{
		IP:            ESRM.ENVM.GetProcessSpaceValue("server", "ip"),
		Port:          ESRM.ENVM.GetProcessSpaceValue("server", "port"),
		AutoReconnect: autoReconnect, // strconv.ParseBool(autoReconnect)
	}
	return config, nil
}

func (ESRM *ESRManager) Fn_bat_init(args []string) error {

	if len(args) < 4 {
		log.Printf("Usage - cln_esr_clnt <tag qualifier> <Exchange cd> <Pipe id> ")
	}

	//C_Servicename = args[0]
	//Sql_c_pipe_id = args[3]
	if len(args[3]) >= 3 {
		// Copy the first 3 bytes from the string
		copy(Sql_c_pipe_id[:], args[3])
	} else {
		log.Fatalf("Argument is too short. Expected at least 3 characters in args[3]")
	}
	C_auto_connect = args[4][0]
	C_server_flg = args[5][0]

	err := ESRM.Fn_init_ddr_pop(args[3], util.TRADING_SECTION, util.COMMON)
	if err != nil {
		log.Printf("Error In ddr population") // Use like panic or fatal here.
	}

	log.Printf("Sql_c_pipe_id :", Sql_c_pipe_id)
	log.Printf("C_auto_connect:", C_auto_connect)
	log.Printf("C_Server_flg:", C_server_flg)

	query := "SELECT opm_xchng_cd FROM opm_ord_pipe_mstr WHERE opm_pipe_id = ?"
	err1 := ESRM.Db.Raw(query, Sql_c_pipe_id).Scan(&Sql_c_xchng_cd)
	if err1 != nil {
		log.Printf("%s: Error while fetching exchange code: %v", cServiceName, err1)
		fnErrLog(cServiceName, "S31010", "SQLMSG", cErrMsg)
		return fmt.Errorf("system error: %d", SYSTEM_ERROR)
	}

	log.Printf("Sql_c_xchng_cd :", Sql_c_xchng_cd)

	query = `
    SELECT 
        to_char(exg_nxt_trd_dt, 'dd-mon-yyyy'), 
        to_char(exg_nxt_trd_dt, 'yyyymmdd'),
        to_char(exg_sec_tm, 'dd-mon-yyyy'),
        to_char(exg_part_tm, 'dd-mon-yyyy'),
        to_char(exg_inst_tm, 'dd-mon-yyyy'),
        to_char(exg_idx_tm, 'dd-mon-yyyy')
    FROM exg_xchng_mstr 
    WHERE exg_xchng_cd = ?
`

	result := ESRM.Db.Raw(query, Sql_c_xchng_cd).Scan(&Sql_C_nxt_trd_date, &Sql_C_trd_ref, &Sql_sec_tm, &Sql_part_tm, &Sql_inst_tm, &Sql_idx_tm)
	if result.Error != nil {
		log.Printf("%s: Error while fetching next trade date: %v", cServiceName, result.Error)
		fnErrLog(cServiceName, "S31015", "SQLMSG", cErrMsg)
		return fmt.Errorf("system error: %d", SYSTEM_ERROR)
	}

	return nil
}

func (ESRM *ESRManager) crt_tap_con() (net.Conn, error) {
	config, err := ESRM.loadConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading config: %v", err)
	}
	log.Printf("[%s] []IP for TAP socket: %v", ESRM.ServiceName, config.IP)
	log.Printf("[%s] []Port for TAP socket: %v", ESRM.ServiceName, config.Port)
	log.Printf("[%s] []Auto Reconnect status: %v", ESRM.ServiceName, config.AutoReconnect)

	log.Printf("Starting Connection to TAP")

	address := net.JoinHostPort(config.IP, config.Port)
	log.Printf("After join host port")
	conn, err := net.Dial("tcp", address)
	log.Printf("after dialing to")
	if err != nil {
		log.Printf("Error connecting:", err)
		if config.AutoReconnect {
			log.Printf("Disconnected. Auto Reconnecting...")
			return nil, nil
		} else {
			log.Printf("Auto Reconnect is disabled.")
			return nil, err
		}
	}

	log.Printf("Connected to TAP at", address)
	return conn, nil
}

func (ESRM *ESRManager) send_thrd(conn net.Conn) {
	//defer wg.Done()
	mtype := *ESRM.Mtype
	log.Printf("Inside Send Routine STARTS HERE")

	for {
		// Log before reading from the queue
		log.Printf("Attempting to read from queue with message type: %d", mtype)

		R_L_msg_type, receivedexchngMsg, readErr := ESRM.Message_queue_manager.ReadFromQueue(mtype)
		if readErr != 0 {
			ESRM.LoggerManager.LogError(ESRM.ServiceName, "[send_thrd] [Error: Failed to read from queue with message type %d: %d", mtype, readErr)
			continue // Skip to the next iteration to avoid potential nil dereference
		}

		// Log received message details
		log.Printf("[send_thrd] Successfully read from queue: message type: %d, received data: %d", mtype, receivedexchngMsg)

		fmt.Println("Message Type:", R_L_msg_type)

		*ESRM.Mtype++
		if *ESRM.Mtype > ESRM.Max_Pack_Val {
			*ESRM.Mtype = 1
		}

		// Check if R_L_msg_type is "2000"
		if R_L_msg_type == util.LOGIN_WITH_OPEN_ORDR_DTLS || R_L_msg_type == util.LOGIN_WITHOUT_OPEN_ORDR_DTLS {
			log.Printf("[send_thrd] Message Type is LOGIN")

			if R_L_msg_type == util.LOGIN_WITH_OPEN_ORDR_DTLS {
				c_rqst_for_open_ordr = 'G'
			} else {
				c_rqst_for_open_ordr = 'N'
			}

			log.Printf("[send_thrd] c_rqst_for_open_ordr: %c", c_rqst_for_open_ordr)

			signalChan := make(chan models.Signal)

			// Call Do_xchng_logon to start the sign-on process
			 if err := ESRM.Do_xchng_logon(msg.St_exch_msg_so, conn, signalChan); err != nil {
				log.Printf("[%s] Failed to send exch_msg_data: %v", err)
				return
			} 

		} else if R_L_msg_type == util.BOARD_LOT_IN {
			log.Printf("[send_thrd] Message type is BOARD LOT IN")

			li_business_data_size := int64(unsafe.Sizeof(models.St_oe_reqres{}))
			li_send_tap_msg_size := /*int64(unsafe.Sizeof(models.St_net_hdr{})) +*/ int64(unsafe.Sizeof(receivedexchngMsg))
			log.Printf("[send_thrd] Business Data Size: %v", li_business_data_size)
			log.Printf("[send_thrd] TAP message size: %v", li_send_tap_msg_size)

			// Call fn_writen to write exch_msg_data to TCP connection
			if err := ESRM.fn_writen(conn, receivedexchngMsg, li_send_tap_msg_size); err != nil {
				log.Printf("[%s] Failed to send exch_msg_data: %v", err)
				return
			}

		} else {
			log.Printf("[%s] [send_thrd] Skipped message with message type: %d", ESRM.ServiceName, R_L_msg_type)
		}

		log.Printf("Inside SEND Routine ENDS HERE")
	}

	log.Printf("Message queue closed, send thread exiting")
	exitcount += 1
}

func (ESRM *ESRManager) fn_writen(conn net.Conn, msg interface{}, msg_size int64) error {
	// Calculate the size of the message
	msgSize := int(unsafe.Sizeof(msg))

	// Ensure the message is not larger than msg_size
	if int64(msgSize) > msg_size {
		return errors.New("message size exceeds allowed size")
	}

	// Convert the struct to a byte slice without a buffer
	data := (*[1 << 30]byte)(unsafe.Pointer(&msg))[:msgSize:msgSize] // Create a byte slice from the struct

	totalWritten := 0

	for totalWritten < len(data) {
		log.Printf("[%s] [fn_writen] Attempting to write: totalWritten=%d, data_len=%d", totalWritten, len(data))

		// Attempt to write to the connection
		written, err := conn.Write(data[totalWritten:])
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				log.Printf("Temporary network error, retrying...")
				time.Sleep(100 * time.Millisecond) // Retry after a brief wait
				continue
			}
			log.Printf("[%s] Write failed: %v\n", err)
			return err
		}

		if written == 0 {
			log.Printf("No bytes written; connection may be stalled or closed.")
			return net.ErrWriteToConnected
		}

		totalWritten += written
		log.Printf("[%s] [fn_writen] Written: %d bytes, Total Written: %d bytes", written, totalWritten)

		if totalWritten < len(data) {
			// Small delay to help with buffer recovery
			time.Sleep(50 * time.Millisecond)
		}
	}

	log.Printf("[%s] Successfully sent exch_msg", ESRM.ServiceName)
	return nil
}

/*
	 func (ESRM *ESRManager) fn_writen(conn net.Conn, msg interface{}, msg_size int64) error {
		var buf bytes.Buffer

		// Get the size of the message
		msgSize := int(unsafe.Sizeof(msg))

		// Use a switch to handle different message sizes
		switch msgSize {
		case 338:
			// Serialize St_exch_msg
			if err := binary.Write(&buf, binary.BigEndian, msg); err != nil {
				return err
			}
		case 314:
			// Serialize St_exch_msg_Log_On
			if err := binary.Write(&buf, binary.BigEndian, msg); err != nil {
				return err
			}
		default:
			return errors.New("unsupported message type")
		}
		dsize := buf.Len()

		log.Printf("Size of binary data: %v", dsize)

		data := buf.Bytes()

		// Restrict the message size if it exceeds the given msg_size
		if int64(len(data)) > msg_size {
			data = data[:msg_size]
		}

		totalWritten := 0

		for totalWritten < len(data) {
			log.Printf("[%s] [fn_writen]Attempting to write: totalWritten=%d, data_len=%d", totalWritten, len(data))
			written, err := conn.Write(data[totalWritten:])
			if err != nil {
				if ne, ok := err.(net.Error); ok && ne.Temporary() {
					log.Printf("Temporary network error, retrying...")
					time.Sleep(100 * time.Millisecond)
					continue
				}
				log.Printf("[%s] []Write failed: %v\n", err)
				return err
			}

			if written == 0 {
				// If no bytes are written, the connection might be stalled or closed.
				log.Printf("No bytes written; connection may be stalled or closed.")
				return net.ErrWriteToConnected
			}

			totalWritten += written

			log.Printf("[%s] [fn_writen]Written: %d bytes, Total Written: %d bytes", written, totalWritten)

			if totalWritten < len(data) {
				// Small delay to help with buffer recovery
				time.Sleep(50 * time.Millisecond)
			}
		}

		//log.Printf("[%s] []Successfully sent exch_msg: %+v", msg)
		log.Printf("[%s]Successfully sent exch_msg")
		return nil
	}
*/
func (ESRM *ESRManager) rcv_thrd(conn net.Conn) {
	log.Printf("Inside Rcv Routine STARTS HERE")
	defer wg.Done()

	for {
		// Read the message from the socket using readXchngSocket function
		msgData, err := ESRM.readXchngSocket(conn, 0) // Pass appropriate timeout here
		if err != nil {
			log.Printf("[rcv_thrd] Error reading from socket: %v", err)
			return
		}

		// Log the length of the received message
		log.Printf("[rcv_thrd] Received message of length: %d bytes", len(msgData))

		// Dynamically determine the structure based on the message length
		switch len(msgData) {
		case int(unsafe.Sizeof(models.St_exch_msg_Log_On{})):
			var tmp_buf_dat models.St_exch_msg_Log_On_use
			buf := bytes.NewBuffer(msgData)
			if err := binary.Read(buf, binary.BigEndian, &tmp_buf_dat); err != nil {
				log.Printf("[rcv_thrd] Failed to decode exch_msg_data: %v", err)
				return
			}
			log.Println("[rcv_thrd] Sign On response received")

			//Call_Do_xchng_logon to start the sign on process
			switch tmp_buf_dat.St_sign_on_req.St_hdr.Si_transaction_code {

				case util.SIGN_ON_REQUEST_OUT {
					_,err := ESRM.Fn_sign_on_request_out(tmp_buf_dat.St_sign_on_req); err != nil {
						log.Printf("[rcv_thrd] Failed to send exch_msg_data: %v", err)
						return nil
					}
				}

			 } 

		case int(unsafe.Sizeof(models.St_exch_msg{})):
			var tmp_buf_dat models.St_exch_msg_resp
			buf := bytes.NewBuffer(msgData)
			if err := binary.Read(buf, binary.BigEndian, &tmp_buf_dat); err != nil {
				log.Printf("[rcv_thrd] Failed to decode exch_msg_so_data: %v", err)
				return
			}

			log.Printf("[rcv_thrd] Message type: %d", tmp_buf_dat.St_oe_res.St_hdr.Si_transaction_code)

			switch tmp_buf_dat.St_oe_res.St_hdr.Si_transaction_code {
			case util.PRICE_CONFIRMATION:
				log.Println("[rcv_thrd] PRICE RESPONSE")

			case util.ORDER_CONFIRMATION_OUT:
				log.Println("[rcv_thrd] CONFIRM RESPONSE")
				log.Println("[rcv_thrd] Exchange Response Received:")
				log.Printf("[rcv_thrd] Transaction Id: %d", tmp_buf_dat.St_oe_res.St_hdr.Si_transaction_code)
				log.Printf("[rcv_thrd] IOC Flag Value: %d", ESRM.TCUM.BoolToInt(tmp_buf_dat.St_oe_res.St_ord_flg.GetFlagValue(models.Flg_IOC)))
				log.Printf("[rcv_thrd] Error Code: %d", tmp_buf_dat.St_oe_res.St_hdr.Si_error_code)

			default:
				log.Printf("[rcv_thrd] Unexpected transaction code: %d", tmp_buf_dat.St_oe_res.St_hdr.Si_transaction_code)
			}

		default:
			log.Printf("[rcv_thrd] Unexpected message size: %v bytes", len(msgData))
			return
		}

		if exitcount == 1 {
			log.Printf("[rcv_thrd] Exit condition met, breaking the loop.")
			break
		}
	}

	log.Printf("Inside Rcv Routine ENDS HERE")
}

func (ESRM *ESRManager) readXchngSocket(conn net.Conn, timeout time.Duration) ([]byte, error) {
	// Set the header size based on the size of the NetHdr struct
	headerSize := int(unsafe.Sizeof(models.St_net_hdr{}))
	headerData, err := ESRM.readNBytes(conn, headerSize, timeout)
	if err != nil {
		return nil, err
	}

	var header models.St_net_hdr
	buf := bytes.NewBuffer(headerData)
	if err := binary.Read(buf, binary.BigEndian, &header); err != nil {
		return nil, err
	}

	// Calculate the length of the remainder of the message
	remainingLength := int(header.S_message_length) - headerSize

	// Read the remaining message
	var messageData []byte
	if remainingLength > 0 {
		messageData, err = ESRM.readNBytes(conn, remainingLength, timeout)
		if err != nil {
			return nil, err
		}
	}

	// Combine header and message into one slice
	fullMessage := append(headerData, messageData...)

	return fullMessage, nil
}

func (ESRM *ESRManager) readNBytes(conn net.Conn, li_len int, timeout time.Duration) ([]byte, error) {
	buffer := make([]byte, li_len)
	read := 0

	// Set the read deadline based on the timeout value
	if timeout > 0 {
		conn.SetReadDeadline(time.Now().Add(timeout))
	} else {
		conn.SetReadDeadline(time.Time{}) // No timeout
	}

	for read < li_len {
		n, err := conn.Read(buffer[read:])
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// Handle timeout
				return nil, errors.New("read timeout")
			}
			return nil, err
		}

		if n == 0 {
			// Socket closed by the other end
			return nil, errors.New("socket closed by peer")
		}

		read += n
	}

	// Clear the deadline
	if timeout > 0 {
		conn.SetReadDeadline(time.Time{}) // Remove timeout after successful read
	}

	return buffer, nil
}

func (ESRM *ESRManager) ClnEsrClnt() int {

	ESRM.Req_q_data = &models.St_req_q_data{}
	ESRM.St_exch_msg = &models.St_exch_msg{}
	ESRM.St_net_hdr = &models.St_net_hdr{}
	ESRM.St_oe_reqres = &models.St_oe_reqres{}
	ESRM.St_sign_on_req = &models.St_sign_on_req{}
	ESRM.PUM = &util.PasswordUtilManger{}

	ESRM.TM = &util.TransactionManager{}

	conn, err := ESRM.crt_tap_con()
	if err != nil {
		log.Fatalf("Error in crt_tap_con: %v", err)
	}
	defer conn.Close()

	// time.Sleep(100 * time.Millisecond)

	wg.Add(1)
	go ESRM.send_thrd(conn)
	log.Printf("Send Routine Initialized")
	go ESRM.rcv_thrd(conn)
	log.Printf("Recieve Routine Initialized")

	/* 	// Create three distinct messages
	   	for i := 1; i <= 5; i++ {
	   		var transactionCode int16

	   		if i%2 == 1 {
	   			transactionCode = 2073 // Odd i
	   		} else {
	   			transactionCode = 2012 // Even i
	   		}
	   		// Initialize the St_int_header structure locally

	   		stIntHeader := models.St_int_header{
	   			Si_transaction_code: transactionCode,
	   			Li_log_time:         int32(i * 1000),
	   			C_alpha_char:        [models.LEN_ALPHA_CHAR]byte{'A' + byte(i), 'B' + byte(i)},
	   			Li_trader_id:        int32(i * 200),
	   			Si_error_code:       int16(i),
	   			C_filler_2:          byte(0),
	   			C_time_stamp_1:      [models.LEN_TIME_STAMP]byte{'T', 'S', '1', byte(i + '0'), ' ', ' ', ' ', ' '},
	   			C_time_stamp_2:      [models.LEN_TIME_STAMP]byte{'T', 'S', '2', byte(i + '0'), ' ', ' ', ' ', ' '},
	   			Si_message_length:   int32(i * 100),
	   		}

	   		// Initialize the St_contract_desc structure locally
	   		stContractDesc := models.St_contract_desc{
	   			C_instrument_name: [models.LEN_INSTRUMENT_NAME]byte{'I', 'N', 'S', byte(i + '0')},
	   			C_symbol:          [models.LEN_SYMBOL_NSE]byte{'S', 'Y', 'M', byte(i + '0')},
	   			Li_expiry_date:    int32(i * 10000),
	   			Li_strike_price:   int64(i * 50000),
	   			C_option_type:     [models.LEN_OPTION_TYPE]byte{'O', 'T'},
	   			Si_ca_level:       int16(i * 10),
	   		}

	   		// Initialize the St_order_flags structure locally
	   		stOrderFlags := models.St_order_flags{
	   			Flg_ato:         uint32(1),
	   			Flg_market:      uint32(0),
	   			Flg_sl:          uint32(1),
	   			Flg_mit:         uint32(0),
	   			Flg_day:         uint32(1),
	   			Flg_gtc:         uint32(0),
	   			Flg_ioc:         uint32(1),
	   			Flg_aon:         uint32(0),
	   			Flg_mf:          uint32(1),
	   			Flg_matched_ind: uint32(0),
	   			Flg_traded:      uint32(1),
	   			Flg_modified:    uint32(0),
	   			Flg_frozen:      uint32(1),
	   			Flg_filler1:     uint32(0),
	   		}

	   		// Initialize the main St_req_q_data structure using the local variables
	   		reqData := models.St_req_q_data{
	   			L_msg_type: 2000,
	   			St_exch_msg_data: models.St_exch_msg{
	   				/* St_net_header: models.St_net_hdr{
	   					S_message_length: 336,
	   					I_seq_num:        int32(i),
	   					C_checksum:       [16]byte{},
	   				} */
	/*St_oe_res: models.St_oe_reqres{
				St_hdr:                  stIntHeader,
				C_participant_type:      byte('A' + i),
				Si_competitor_period:    int16(i),
				Si_solicitor_period:     int16(i),
				C_modified_cancelled_by: byte(0),
				C_filler_2:              byte(0),
				Si_reason_code:          int16(i * 2),
				L_token_no:              int32(i * 3000),
				St_con_desc:             stContractDesc,
				St_ord_flg:              stOrderFlags,
				Si_order_type:           int16(i),
				D_order_number:          float64(i) * 1.1,
				Li_volume:               int32(i * 400),
				Si_branch_id:            int16(i * 5),
				Li_trader_id:            int32(i * 500),
				C_broker_id:             [models.LEN_BROKER_ID]byte{'B', 'R', 'K', byte('0' + i)},
				C_account_number:        [models.LEN_ACCOUNT_NUMBER]byte{'A', 'C', 'C', byte('0' + i)},
				L_algo_id:               int32(i * 6000),
				Ll_lastactivityref:      int64(i * 7000),
				C_reserved:              byte(0),
			},
		},
	}

	// Enqueue the message

	// Enqueue the message */
	//messageQueue <- reqData
	//log.Printf("[Message %d] Queued message:", i, reqData)
	//time.Sleep(50 * time.Millisecond)

	// Optionally close the queue when done
	//close(messageQueue)

	wg.Wait()
	return 0
}
