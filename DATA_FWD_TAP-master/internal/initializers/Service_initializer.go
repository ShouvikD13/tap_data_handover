package initializers

import (
	"DATA_FWD_TAP/internal/app/esr"
	"DATA_FWD_TAP/internal/app/exchange_connection"
	"DATA_FWD_TAP/internal/app/pack_clnt"
)

type ServiceInitializationManager struct {
	MainContainer *MainContainer
}

func (SIM *ServiceInitializationManager) ClnPackClntInitialization() int {
	VarClnPack := &pack_clnt.ClnPackClntManager{
		ServiceName:              SIM.MainContainer.UtilContainer.ServiceName,
		Xchngbook:                SIM.MainContainer.ClientContainer.Xchngbook,
		Orderbook:                SIM.MainContainer.ClientContainer.Orderbook,
		Contract:                 SIM.MainContainer.ClientContainer.Contract,
		Nse_contract:             SIM.MainContainer.ClientContainer.NseContract,
		Pipe_mstr:                SIM.MainContainer.ClientContainer.PipeMaster,
		Oe_reqres:                SIM.MainContainer.ClientContainer.OeReqres,
		Exch_msg:                 SIM.MainContainer.ClientContainer.ExchMsg,
		Net_hdr:                  SIM.MainContainer.ClientContainer.NetHdr,
		Q_packet:                 SIM.MainContainer.ClientContainer.QPacket,
		Int_header:               SIM.MainContainer.ClientContainer.IntHeader,
		Contract_desc:            SIM.MainContainer.ClientContainer.ContractDesc,
		Order_flag:               SIM.MainContainer.ClientContainer.OrderFlag,
		Order_conversion_manager: SIM.MainContainer.UtilContainer.OrderConversionManager,
		CPanNo:                   SIM.MainContainer.ClientGlobalValueContainer.CPanNo,
		CLstActRef:               SIM.MainContainer.ClientGlobalValueContainer.CLastActRef,
		CEspID:                   SIM.MainContainer.ClientGlobalValueContainer.CEspID,
		CAlgoID:                  SIM.MainContainer.ClientGlobalValueContainer.CAlgoID,
		CSourceFlg:               SIM.MainContainer.ClientGlobalValueContainer.CSourceFlg,
		CPrgmFlg:                 SIM.MainContainer.ClientGlobalValueContainer.CPrgmFlg,
		Enviroment_manager:       SIM.MainContainer.UtilContainer.EnvironmentManager,
		Message_queue_manager:    SIM.MainContainer.UtilContainer.MessageQueueManager,
		Transaction_manager:      SIM.MainContainer.UtilContainer.TransactionManager,
		Max_Pack_Val:             SIM.MainContainer.UtilContainer.MaxPackVal,
		Config_manager:           SIM.MainContainer.UtilContainer.ConfigManager,
		LoggerManager:            SIM.MainContainer.UtilContainer.LoggerManager,
		TCUM:                     SIM.MainContainer.UtilContainer.TypeConversionUtilManager,
		MtypeWrite:               SIM.MainContainer.UtilContainer.MTypeWrite,
		Args:                     SIM.MainContainer.ClientGlobalValueContainer.Args,
		Db:                       SIM.MainContainer.UtilContainer.DB,
	}

	initResult := VarClnPack.Fn_bat_init()

	if initResult != 0 {
		SIM.MainContainer.UtilContainer.LoggerManager.LogInfo(SIM.MainContainer.UtilContainer.ServiceName, " Fn_bat_init failed with result code: %d", initResult)
		return -1
	}
	return 0
}

func (SIM *ServiceInitializationManager) LogOnToTapInitialization() int {

	VarLogOn := &exchange_connection.LogOnToTapManager{
		ServiceName:                SIM.MainContainer.UtilContainer.ServiceName,
		St_sign_on_req:             SIM.MainContainer.LogOnContainer.StSignOnReq,
		St_req_q_data:              SIM.MainContainer.LogOnContainer.StReqQData,
		St_exch_msg_Log_On:         SIM.MainContainer.LogOnContainer.StExchMsgLogOn,
		Int_header:                 SIM.MainContainer.LogOnContainer.IntHeader,
		St_net_hdr:                 SIM.MainContainer.LogOnContainer.StNetHdr,
		St_BrokerEligibilityPerMkt: SIM.MainContainer.LogOnContainer.StBrokerEligibilityPerMkt,
		EM:                         SIM.MainContainer.UtilContainer.EnvironmentManager,
		PUM:                        SIM.MainContainer.UtilContainer.PasswordUtilManager,
		DB:                         SIM.MainContainer.UtilContainer.DB,
		LoggerManager:              SIM.MainContainer.UtilContainer.LoggerManager,
		OCM:                        SIM.MainContainer.UtilContainer.OrderConversionManager,
		TCUM:                       SIM.MainContainer.UtilContainer.TypeConversionUtilManager,
		Message_queue_manager:      SIM.MainContainer.UtilContainer.MessageQueueManager,
		C_pipe_id:                  SIM.MainContainer.LogOnGlobalValueContainer.CPipeID,
		UserID:                     SIM.MainContainer.LogOnGlobalValueContainer.UserID,
		MtypeWrite:                 SIM.MainContainer.UtilContainer.MTypeWrite,
		Max_Pack_Val:               SIM.MainContainer.UtilContainer.MaxPackVal,
		Opm_loginStatus:            SIM.MainContainer.LogOnGlobalValueContainer.OpmLoginStatus,
		Opm_userID:                 SIM.MainContainer.LogOnGlobalValueContainer.OpmUserID,
		Opm_existingPasswd:         SIM.MainContainer.LogOnGlobalValueContainer.OpmExistingPasswd,
		Opm_newPasswd:              SIM.MainContainer.LogOnGlobalValueContainer.OpmNewPasswd,
		Opm_LstPswdChgDt:           SIM.MainContainer.LogOnGlobalValueContainer.OpmLstPswdChgDt,
		Opm_XchngCd:                SIM.MainContainer.LogOnGlobalValueContainer.OpmXchngCd,
		Opm_TrdrID:                 SIM.MainContainer.LogOnGlobalValueContainer.OpmTrdrID,
		Opm_BrnchID:                SIM.MainContainer.LogOnGlobalValueContainer.OpmBrnchID,
		Exg_BrkrID:                 SIM.MainContainer.LogOnGlobalValueContainer.ExgBrkrID,
		Exg_NxtTrdDate:             SIM.MainContainer.LogOnGlobalValueContainer.ExgNxtTrdDate,
		Exg_BrkrName:               SIM.MainContainer.LogOnGlobalValueContainer.ExgBrkrName,
		Exg_BrkrStts:               SIM.MainContainer.LogOnGlobalValueContainer.ExgBrkrStts,
		Args:                       SIM.MainContainer.LogOnGlobalValueContainer.Args,
	}

	initResult := VarLogOn.LogOnToTap()

	if initResult != 0 {
		SIM.MainContainer.UtilContainer.LoggerManager.LogInfo(SIM.MainContainer.UtilContainer.ServiceName, " LogOnToTap failed with result code: %d", initResult)
		return -1
	}

	return 0
}

func (SIM *ServiceInitializationManager) LogOffFromTapInitialization() int {

	VarLogOff := &exchange_connection.LogOffFromTapManager{
		DB:                    SIM.MainContainer.UtilContainer.DB,
		LoggerManager:         SIM.MainContainer.UtilContainer.LoggerManager,
		ServiceName:           SIM.MainContainer.UtilContainer.ServiceName,
		MtypeWrite:            SIM.MainContainer.UtilContainer.MTypeWrite,
		Int_header:            SIM.MainContainer.LogOffContainer.IntHeader,
		St_net_hdr:            SIM.MainContainer.LogOffContainer.StNetHdr,
		Exch_msg_Log_Off:      SIM.MainContainer.LogOffContainer.StExchMsgLogOff,
		St_req_q_data:         SIM.MainContainer.LogOffContainer.StReqQData,
		TCUM:                  SIM.MainContainer.UtilContainer.TypeConversionUtilManager,
		Message_queue_manager: SIM.MainContainer.UtilContainer.MessageQueueManager,
		Opm_XchngCd:           SIM.MainContainer.LogOffGlobalValueContainer.OpmXchngCd,
		Opm_TrdrID:            SIM.MainContainer.LogOffGlobalValueContainer.OpmTrdrID,
		Exg_NxtTrdDate:        SIM.MainContainer.LogOffGlobalValueContainer.ExgNxtTrdDate,
		C_pipe_id:             SIM.MainContainer.LogOffGlobalValueContainer.C_pipe_id,
		Args:                  SIM.MainContainer.LogOffGlobalValueContainer.Args,
		Max_Pack_Val:          SIM.MainContainer.UtilContainer.MaxPackVal,
	}

	initResult := VarLogOff.Fn_logoff_from_TAP()

	if initResult != 0 {
		SIM.MainContainer.UtilContainer.LoggerManager.LogInfo(SIM.MainContainer.UtilContainer.ServiceName, " Fn_logoff_from_TAP() failed with result code: %d", initResult)
		return -1
	}

	return 0

}
func (SIM *ServiceInitializationManager) ESRInitialization() int {

	VarEsr := &esr.ESRManager{
		Req_q_data:             SIM.MainContainer.ESRContainer.Req_q_data,
		St_exch_msg:            SIM.MainContainer.ESRContainer.St_exch_msg,
		St_net_hdr:             SIM.MainContainer.ESRContainer.St_net_hdr,
		St_oe_reqres:           SIM.MainContainer.ESRContainer.St_oe_reqres,
		St_sign_on_req:         SIM.MainContainer.ESRContainer.St_sign_on_req,
		St_exch_msg_Log_on:     SIM.MainContainer.ESRContainer.St_exch_msg_Log_on,
		St_exch_msg_Log_on_use: SIM.MainContainer.ESRContainer.St_exch_msg_Log_on_use,
		St_exch_msg_resp:       SIM.MainContainer.ESRContainer.St_exch_msg_resp,
		ENVM:                   SIM.MainContainer.UtilContainer.EnvironmentManager,
		Req_q_data1:            *SIM.MainContainer.ESRContainer.Req_q_data,
		PUM:                    SIM.MainContainer.UtilContainer.PasswordUtilManager,
		Message_queue_manager:  SIM.MainContainer.UtilContainer.MessageQueueManager,
		TCUM:                   SIM.MainContainer.UtilContainer.TypeConversionUtilManager,
		LoggerManager:          SIM.MainContainer.UtilContainer.TransactionManager.LoggerManager,
		TM:                     SIM.MainContainer.UtilContainer.TransactionManager,
		Mtype:                  SIM.MainContainer.UtilContainer.MTypeRead,
		Max_Pack_Val:           SIM.MainContainer.ESRGlobalValueContainer.Max_Pack_Val,
		Db:                     SIM.MainContainer.UtilContainer.DB,
	}

	initResult := VarEsr.ClnEsrClnt()
	if initResult != 0 {
		SIM.MainContainer.UtilContainer.LoggerManager.LogInfo(SIM.MainContainer.UtilContainer.ServiceName, " Fn_logoff_from_TAP() failed with result code: %d", initResult)
		return -1
	}

	return 0
}
