/***************************************************************************************
 * Fn_bat_init initializes the client package manager by setting up data structures,
 * fetching configuration values, and preparing the environment for processing.
 * It ensures all prerequisites are met before the service starts.
 *
 * INPUT PARAMETERS:
 *    args[] - Command-line arguments, which must include at least 7 elements.
 *             The fourth argument (args[3]) specifies the pipe ID.
 *    Db     - Pointer to a GORM database connection object used for SQL queries.
 *
 * OUTPUT PARAMETERS:
 *    int    - Returns 0 for success, or -1 if an error occurs.
 *
 * FUNCTIONAL FLOW:
 * 1. Initializes required structures and managers for the service.
 * 2. Validates command-line arguments to ensure all necessary parameters are present.
 * 3. Sets the pipe ID from the command-line argument and creates a message queue.
 * 4. Executes SQL queries to fetch and store necessary data such as exchange code,
 *    trader ID, branch ID, trade date, and broker ID.
 * 5. Reads configuration values from a file, including user type and message limit.
 * 6. Invokes the `CLN_PACK_CLNT` function to start the main service logic.
 * 7. Returns a status code indicating the success or failure of the initialization.
 *
 ***********************************************************************************************/

/***************************************************************************************
 * CLN_PACK_CLNT handles the core operations of the client package manager service.
 * It continuously retrieves newly created orders from the database, manages database
 * transactions, and writes the data to a system queue.
 *
 * INPUT PARAMETERS:
 *    args[] - Command-line arguments used for service configuration and control.
 *    Db     - Pointer to a GORM database connection object for SQL operations.
 *
 * OUTPUT PARAMETERS:
 *    int    - Returns 0 for success, or -1 if an error occurs.
 *
 * FUNCTIONAL FLOW:
 * 1. Enters an infinite loop to continuously process orders.
 * 2. Initiates a local transaction with `FnBeginTran()`.
 * 3. Retrieves the next order record using `fnGetNxtRec()`.
 * 4. Commits the transaction with `FnCommitTran()`.
 * 5. Writes the fetched data to the system queue using `WriteToQueue()`.
 * 6. Increments the message type and reset if it exceeds the maximum value.
 *
 ***********************************************************************************************/

/***************************************************************************************
 * fnGetNxtRec retrieves and processes the next record from the database based on its
 * status and type, then updates relevant structures and statuses. It handles extraction
 * of order data, updates database statuses, and packs the data for further processing.
 *
 * INPUT PARAMETERS:
 *    Db - Pointer to a GORM database connection object for SQL operations.
 *
 * OUTPUT PARAMETERS:
 *    int - Returns 0 for successful execution, or -1 if an error occurs.
 *
 * FUNCTIONAL FLOW:
 * 1. Calls 'fnSeqToOmd' to fetch records from the 'FXB_FO_XCHNG_BOOK' table.
 * 2. Ensures 'C_plcd_stts' is 'REQUESTED'.
 * 3. Processes the order if its type is 'ORDINARY_ORDER':
 *    a. Transfers 'C_ordr_rfrnc' from 'Xchngbook' to 'orderbook'.
 *    b. Calls 'fnRefToOrd' to fetch order details from 'fod_fo_ordr_dtls'.
 *    c. Compares 'L_mdfctn_cntr' between 'Xchngbook' and 'orderbook'.
 *    d. Updates 'Xchngbook' and 'orderbook' statuses to 'QUEUED':
 *       - Calls 'fnUpdXchngbk' and 'fnUpdOrdrbk' to update the database.
 *    e. Sets the contract data using values from 'orderbook'.
 *    f. Checks if 'C_xchng_cd' is "NFO":
 *       - If "NFO", sets the request type and calls 'fnGetExtCnt' to load additional data.
 *    g. Calls 'fnPackOrdnryOrdToNse' to pack data into the final structure 'St_req_q_data'.
 *
 ***********************************************************************************************/

/***************************************************************************************
 * fnSeqToOmd fetches the next record from the 'FXB_FO_XCHNG_BOOK' table in the database
 * based on specified criteria (exchange code, pipe ID, trade date, and order sequence).
 * It retrieves order information and stores it in the 'Xchngbook' structure
 * within the 'ClnPackClntManager' instance.
 *
 * INPUT PARAMETERS:
 *    db - Pointer to a GORM database connection object for SQL operations.
 *
 * OUTPUT PARAMETERS:
 *    int - Returns 0 for successful data retrieval, or -1 if an error occurs.
 *
 * FUNCTIONAL FLOW:
 * 1. Constructs and executes a SQL query using GORM's 'Raw' method to fetch order details
 *    from the 'FXB_FO_XCHNG_BOOK' table:
 *    a. The query selects various columns, including order reference, quantity, order type,
 *       sequence number, etc., based on specified conditions.
 *    b. Utilizes 'TO_DATE' to format date fields and 'COALESCE' to handle null values for
 *       specific fields.
 * 2. Scans the retrieved row into the fields of the 'Xchngbook' structure and additional
 *    local variables ('c_ip_addrs', 'c_prcimpv_flg').
 *
 ***********************************************************************************************/

/***********************************************************************************************
 * fnRefToOrd retrieves and processes order data from the 'fod_fo_ordr_dtls' and 'CLM_CLNT_MSTR'
 * tables in the database, and stores it in the `orderbook` structure within the
 * 'ClnPackClntManager' instance.
 *
 * INPUT PARAMETERS:
 *    None
 *
 * OUTPUT PARAMETERS:
 *    int - Returns 0 for success, or -1 if an error occurs.
 *
 * FUNCTIONAL FLOW:
 * 1. Constructs and executes a SQL query using GORM's 'Raw' method to fetch order details
 *    from the 'fod_fo_ordr_dtls' table:
 *    a. Selects various columns, including client match account, order flow, total order quantity,
 *       executed quantity, etc., based on specified conditions.
 *    b. Uses 'FOR UPDATE NOWAIT' to lock selected rows for update and prevent other transactions
 *       from waiting for the lock.
 * 2. Scans the retrieved row into the fields of the `orderbook` structure within the
 *    'ClnPackClntManager' instance.
 * 3. Constructs and executes a second SQL query to fetch client details from the 'CLM_CLNT_MSTR' table
 *    where the client match account matches the previously fetched value.
 *
 ***********************************************************************************************/

/***********************************************************************************************
 * fnUpdXchngbk updates the status of a record in the 'fxb_fo_xchng_book' table based on the
 * operation type specified in the `Xchngbook` structure within the 'ClnPackClntManager' instance.
 * which are set in the 'fnGetNxtRec' .
 * The function handles two main operations:
 * 1. **Order Forwarding Update**: Updates the 'fxb_plcd_stts' and 'fxb_frwd_tm' fields for a record
 *    where 'fxb_ordr_rfrnc' and 'fxb_mdfctn_cntr' match the values in the `Xchngbook` structure.
 * 2. **Exchange Response Update**: If the download flag indicates the record should be processed, it
 *    checks if the record exists using the provided 'fxb_jiffy', 'fxb_xchng_cd', and 'fxb_pipe_id'.
 *    If the record exists, it logs that it has already been processed. Otherwise, it updates the record
 *    with the provided status, remarks, and other details.
 *
 * INPUT PARAMETERS:
 *    - db *gorm.DB: The database connection instance used for executing SQL queries.
 *
 * OUTPUT PARAMETERS:
 *    int - Returns 0 for success or -1 if an error occurs during the update operations.
 *
 * FUNCTIONAL FLOW:

 * 1. Based on the `C_oprn_typ` value in the `Xchngbook` structure:
 *    a. For **Order Forwarding** (`models.UPDATION_ON_ORDER_FORWARDING`):
 *       - Constructs and executes a SQL query to update 'fxb_plcd_stts' and 'fxb_frwd_tm' fields
 *         for the record that matches 'fxb_ordr_rfrnc' and 'fxb_mdfctn_cntr'.
 *    b. For **Exchange Response** (`models.UPDATION_ON_EXCHANGE_RESPONSE`):
 *       - Checks if the record should be processed based on the `L_dwnld_flg` value.
 *       - If processing is required, it verifies if a record already exists using the provided
 *         'fxb_jiffy', 'fxb_xchng_cd', and 'fxb_pipe_id'.
 *       - If the record is not found, updates the 'fxb_fo_xchng_book' table with new values including
 *         'fxb_plcd_stts', 'fxb_rms_prcsd_flg', 'fxb_ors_msg_typ', and 'fxb_ack_tm'.
 ***********************************************************************************************/

/***********************************************************************************************
 * fnUpdOrdrbk updates the 'order status' of a record in the 'fod_fo_ordr_dtls' table based on
 * the details provided in the `orderbook` structure within the 'ClnPackClntManager' instance.
 * The function performs the following operations:
 * 1. **Order Status Update**: Updates the 'fod_ordr_stts' field for a record where the
 *    'fod_ordr_rfrnc' matches the value specified in the `orderbook` structure.
 *
 * INPUT PARAMETERS:
 *    - db *gorm.DB: The database connection instance used for executing SQL queries.
 *
 * OUTPUT PARAMETERS:
 *    int - Returns 0 for success or -1 if an error occurs during the update operation.
 *
 * FUNCTIONAL FLOW:
 * 1. Constructs and executes a SQL query to update the 'fod_ordr_stts' field in the
 *    'fod_fo_ordr_dtls' table where 'fod_ordr_rfrnc' matches the value from the `orderbook` structure.
 ***********************************************************************************************/

/***********************************************************************************************
 * fnGetExtCnt retrieves and sets external contract details based on the values provided
 * in the `contract` structure within the 'ClnPackClntManager' instance. This function performs
 * multiple operations to populate the `nse_contract` structure with relevant data.
 *
 * The function performs the following operations:
 * 1. **Determine Entity Type**: Based on the value of `C_xchng_cd` in the `contract` structure,
 *    sets the `i_sem_entity` variable to either `models.NFO_ENTTY` or `models.BFO_ENTTY`.
 *
 * 2. **Fetch Symbol**: Executes a query to fetch the stock symbol (`C_symbol`) from the
 *    'sem_stck_map' table using the stock code (`C_undrlyng`) and entity type.
 *
 * 3. **Determine Exchange Code**: Sets the exchange code (`c_exg_cd`) based on the value of
 *    `C_xchng_cd` in the `contract` structure.
 *
 * 4. **Fetch Series**: Executes a query to fetch the series (`C_series`) from the 'ESS_SGMNT_STCK'
 *    table using the underlying stock code (`C_undrlyng`) and exchange code (`c_exg_cd`).
 *
 * 5. **Fetch Token ID and CA Level**: Executes a query to fetch token ID (`L_token_id`) and
 *    CA level (`L_ca_lvl`) from the 'ftq_fo_trd_qt' table using various contract details including
 *    exchange code, product type, underlying stock code, expiry date, exercise type, option type,
 *    and strike price. The expiry date is formatted to 'dd-Mon-yyyy'.
 *
 * INPUT PARAMETERS:
 *    - db *gorm.DB: The database connection instance used for executing SQL queries.
 *
 * OUTPUT PARAMETERS:
 *    int - Returns 0 for success or -1 if an error occurs .
 *
 * FUNCTIONAL FLOW:
 * 1. Determines the entity type and fetches the symbol from 'sem_stck_map'.
 * 2. Based on the `C_xchng_cd`, sets the appropriate exchange code and fetches the series from
 *    'ESS_SGMNT_STCK'.
 * 3. Retrieves token ID and CA level from 'ftq_fo_trd_qt' using the formatted expiry date and
 *    other contract details.
 * 4. Updates the `nse_contract` structure with the fetched values.
 ***********************************************************************************************/

/***********************************************************************************************
 * fnRjctRcrd processes record rejection by updating the 'Xchngbook' structure within the
 * 'ClnPackClntManager'. It performs the following steps:
 *
 * 1. **Fetch Timestamp**: Retrieves the current system timestamp in 'DD-Mon-YYYY HH24:MI:SS'
 *    format and logs it.
 *
 * 2. **Update Record**: Sets the status of the record to 'REJECT', marks it as 'NOT_PROCESSED',
 *    and updates relevant fields with rejection details.
 *
 * 3. **Persist Updates**: Calls 'fnUpdXchngbk' to save the updated status and details to the
 *    'xchng_book' tble.
 *
 * INPUT PARAMETERS:
 *    - db *gorm.DB: The database connection for executing queries.
 *
 * OUTPUT PARAMETERS:
 *    - int: Returns 0 on success; -1 on error.
 *
 * FUNCTIONAL FLOW:
 * 1. Logs the current timestamp retrieved from the database.
 * 2. Updates the 'Xchngbook' structure with rejection information.
 * 3. Calls 'fnUpdXchngbk' to apply and save the changes, handling any errors.
 ***********************************************************************************************/




2024/09/05 15:16:12 ---- St_req_q_data ----
2024/09/05 15:16:12 Size of St_req_q_data: 8 bytes
2024/09/05 15:16:12 ---- St_req_q_data fields ----
2024/09/05 15:16:12 L_msg_type: 8 bytes
2024/09/05 15:16:12 St_exch_msg_data: 376 bytes
2024/09/05 15:16:12 ---- St_exch_msg ----
2024/09/05 15:16:12 Size of St_exch_msg: 376 bytes
2024/09/05 15:16:12 ---- St_exch_msg fields ----
2024/09/05 15:16:12 St_net_header: 24 bytes
2024/09/05 15:16:12 St_oe_res: 352 bytes
2024/09/05 15:16:12 ---- St_net_hdr ----
2024/09/05 15:16:12 Size of St_net_hdr: 24 bytes
2024/09/05 15:16:12 ---- St_net_hdr fields ----
2024/09/05 15:16:12 S_message_length: 2 bytes
2024/09/05 15:16:12 I_seq_num: 4 bytes
2024/09/05 15:16:12 C_checksum: 16 bytes
2024/09/05 15:16:12 ---- St_oe_reqres ----
2024/09/05 15:16:12 Size of St_oe_reqres: 352 bytes
2024/09/05 15:16:12 ---- St_oe_reqres fields ----
2024/09/05 15:16:12 St_hdr: 56 bytes
2024/09/05 15:16:12 C_participant_type: 1 bytes
2024/09/05 15:16:12 C_filler_1: 1 bytes
2024/09/05 15:16:12 Si_competitor_period: 2 bytes
2024/09/05 15:16:12 Si_solicitor_period: 2 bytes
2024/09/05 15:16:12 C_modified_cancelled_by: 1 bytes
2024/09/05 15:16:12 C_filler_2: 1 bytes
2024/09/05 15:16:12 Si_reason_code: 2 bytes
2024/09/05 15:16:12 C_filler_3: 4 bytes
2024/09/05 15:16:12 L_token_no: 4 bytes
2024/09/05 15:16:12 St_con_desc: 28 bytes
2024/09/05 15:16:12 C_counter_party_broker_id: 5 bytes
2024/09/05 15:16:12 C_filler_4: 1 bytes
2024/09/05 15:16:12 C_filler_5: 2 bytes
2024/09/05 15:16:12 C_closeout_flg: 1 bytes
2024/09/05 15:16:12 C_filler_6: 1 bytes
2024/09/05 15:16:12 Si_order_type: 2 bytes
2024/09/05 15:16:12 D_order_number: 8 bytes
2024/09/05 15:16:12 C_account_number: 10 bytes
2024/09/05 15:16:12 Si_book_type: 2 bytes
2024/09/05 15:16:12 Si_buy_sell_indicator: 2 bytes
2024/09/05 15:16:12 Li_disclosed_volume: 4 bytes
2024/09/05 15:16:12 Li_disclosed_volume_remaining: 4 bytes
2024/09/05 15:16:12 Li_total_volume_remaining: 4 bytes
2024/09/05 15:16:12 Li_volume: 4 bytes
2024/09/05 15:16:12 Li_volume_filled_today: 4 bytes
2024/09/05 15:16:12 Li_price: 4 bytes
2024/09/05 15:16:12 Li_trigger_price: 4 bytes
2024/09/05 15:16:12 Li_good_till_date: 4 bytes
2024/09/05 15:16:12 Li_entry_date_time: 4 bytes
2024/09/05 15:16:12 Li_minimum_fill_aon_volume: 4 bytes
2024/09/05 15:16:12 Li_last_modified: 4 bytes
2024/09/05 15:16:12 St_ord_flg: 2 bytes
2024/09/05 15:16:12 Si_branch_id: 2 bytes
2024/09/05 15:16:12 Li_trader_id: 4 bytes
2024/09/05 15:16:12 C_broker_id: 5 bytes
2024/09/05 15:16:12 C_remarks: 24 bytes
2024/09/05 15:16:12 C_open_close: 1 bytes
2024/09/05 15:16:12 C_settlor: 12 bytes
2024/09/05 15:16:12 Si_pro_client_indicator: 2 bytes
2024/09/05 15:16:12 Si_settlement_period: 2 bytes
2024/09/05 15:16:12 C_cover_uncover: 1 bytes
2024/09/05 15:16:12 C_giveup_flag: 1 bytes
2024/09/05 15:16:12 I_order_seq: 4 bytes
2024/09/05 15:16:12 D_nnf_field: 8 bytes
2024/09/05 15:16:12 D_filler19: 8 bytes
2024/09/05 15:16:12 C_pan: 10 bytes
2024/09/05 15:16:12 L_algo_id: 4 bytes
2024/09/05 15:16:12 Si_algo_category: 2 bytes
2024/09/05 15:16:12 Ll_lastactivityref: 8 bytes
2024/09/05 15:16:12 C_reserved: 52 bytes
2024/09/05 15:16:12 ---- St_int_header ----
2024/09/05 15:16:12 Size of St_int_header: 56 bytes
2024/09/05 15:16:12 ---- St_int_header fields ----
2024/09/05 15:16:12 Si_transaction_code: 2 bytes
2024/09/05 15:16:12 Li_log_time: 4 bytes
2024/09/05 15:16:12 C_alpha_char: 2 bytes
2024/09/05 15:16:12 Li_trader_id: 4 bytes
2024/09/05 15:16:12 Si_error_code: 2 bytes
2024/09/05 15:16:12 C_filler_2: 8 bytes
2024/09/05 15:16:12 C_time_stamp_1: 8 bytes
2024/09/05 15:16:12 C_time_stamp_2: 8 bytes
2024/09/05 15:16:12 Si_message_length: 2 bytes
2024/09/05 15:16:12 ---- St_contract_desc ----
2024/09/05 15:16:12 Size of St_contract_desc: 28 bytes
2024/09/05 15:16:12 ---- St_contract_desc fields ----
2024/09/05 15:16:12 C_instrument_name: 6 bytes
2024/09/05 15:16:12 C_symbol: 10 bytes
2024/09/05 15:16:12 Li_expiry_date: 4 bytes
2024/09/05 15:16:12 Li_strike_price: 4 bytes
2024/09/05 15:16:12 C_option_type: 2 bytes
2024/09/05 15:16:12 Si_ca_level: 2 bytes
2024/09/05 15:16:12 ---- St_order_flags ----
2024/09/05 15:16:12 Size of St_order_flags: 2 bytes
2024/09/05 15:16:12 ---- St_order_flags fields ----
2024/09/05 15:16:12 Flags: 2 bytes