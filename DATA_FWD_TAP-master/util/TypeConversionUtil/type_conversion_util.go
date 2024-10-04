package typeconversionutil

import (
	"DATA_FWD_TAP/util"
	"bytes"
	"encoding/binary"
	"strings"
	"time"

	"gorm.io/gorm"
)

type TypeConversionUtilManager struct {
	LoggerManager *util.LoggerManager
	ServiceName   string
}

func (TCUM *TypeConversionUtilManager) CopyAndFormatSymbol(dest []byte, destLen int, src string) { //'fn_orstonse_char'

	TCUM.ServiceName = "Type_Conversion_Util"

	srcUpper := strings.ToUpper(src)

	copyLen := len(srcUpper)
	if copyLen > destLen {
		copyLen = destLen
	}

	copy(dest, srcUpper[:copyLen])

	for i := copyLen; i < destLen; i++ {
		dest[i] = ' '
	}
}

func (TCUM *TypeConversionUtilManager) LongToTimeArr(C_expry_dt string) (int, int32) {
	TCUM.ServiceName = "Type_Conversion_Util"
	var Li_expiry_date int32
	// Parse the ISO 8601 date-time string to a time.Time object
	layout := time.RFC3339
	t, err := time.Parse(layout, C_expry_dt)
	if err != nil {
		TCUM.LoggerManager.LogError(TCUM.ServiceName, "[LongToTimeArr] ERROR: error in parsing expiry date: %v", err)
		return -1, Li_expiry_date // Return -1 and the unmodified expiry date in case of an error
	}

	liSrc := t.Unix() + util.TEN_YRS_IN_SEC

	Li_expiry_date = int32(liSrc)

	return 0, Li_expiry_date
}

func (TCUM *TypeConversionUtilManager) TimeArrToLong(C_valid_dt string, date *int32) int {
	TCUM.ServiceName = "Type_Conversion_Util"
	layout := time.RFC3339

	t, err := time.Parse(layout, C_valid_dt)
	if err != nil {
		TCUM.LoggerManager.LogError(TCUM.ServiceName, " [TimeArrToLong] [ERROR: error in parsing valid date: %v", err)
		return -1
	}

	*date = int32(t.Unix())

	return 0
}

func (TCUM *TypeConversionUtilManager) GetResetSequence(db *gorm.DB, C_pipe_id string, C_mod_trd_dt string) int32 {
	TCUM.ServiceName = "Type_Conversion_Util"
	var seqNum int32
	var query string

	TCUM.LoggerManager.LogInfo(TCUM.ServiceName, " [GetResetSequence] [Executing increment sequence query for pipe ID: %s, trade date: %s]", C_pipe_id, C_mod_trd_dt)

	query = `
		UPDATE fsp_fo_seq_plcd
		SET fsp_seq_num = fsp_seq_num + 1
		WHERE fsp_pipe_id = ? AND fsp_trd_dt = TO_DATE(?, 'YYYY-MM-DD')
		RETURNING fsp_seq_num;
	`
	result1 := db.Raw(query, C_pipe_id, C_mod_trd_dt).Scan(&seqNum)

	if result1.Error != nil {
		TCUM.LoggerManager.LogError(TCUM.ServiceName, " [GetResetSequence] [Error: executing increment sequence query: %v]", result1.Error)
		return -1
	}
	TCUM.LoggerManager.LogInfo(TCUM.ServiceName, " [GetResetSequence] [Incremented sequence number: %d]", seqNum)

	if seqNum == util.MAX_SEQ_NUM {
		TCUM.LoggerManager.LogInfo(TCUM.ServiceName, " [GetResetSequence] [Sequence number reached MAX_SEQ_NUM, resetting...]")

		query = `
			UPDATE fsp_fo_seq_plcd
			SET fsp_seq_num = 0
			WHERE fsp_pipe_id = ? AND fsp_trd_dt = TO_DATE(?, 'YYYY-MM-DD')
			RETURNING fsp_seq_num;
		`
		result2 := db.Raw(query, C_pipe_id, C_mod_trd_dt).Scan(&seqNum)

		if result2.Error != nil {
			TCUM.LoggerManager.LogError(TCUM.ServiceName, " [GetResetSequence] [Error: executing reset sequence query: %v]", result2.Error)
			return -1
		}
		TCUM.LoggerManager.LogInfo(TCUM.ServiceName, " [GetResetSequence] [Sequence number reset to: %d]", seqNum)
	}
	return seqNum
}

func (TCUM *TypeConversionUtilManager) BoolToInt(value bool) int {

	if value {
		return 1
	}
	return 0
}

func (TCUM *TypeConversionUtilManager) WriteAndCopy(buf *bytes.Buffer, data interface{}, dest []byte) error {
	buf.Reset()
	err := binary.Write(buf, binary.LittleEndian, data)
	if err != nil {
		return err
	}
	copy(dest, buf.Bytes())
	return nil
}
