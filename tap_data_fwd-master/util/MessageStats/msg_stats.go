package MessageStat

/*
#cgo CFLAGS: -I.
#cgo LDFLAGS: -L. -lmsgqueue
#include "msg_queue_wrapper.h"
*/
import "C"
import (
	"fmt"
	"time"
)

type MessageStatManager struct {
}

// MsgQueueStats wraps the C MsgQueueStats struct
type MsgQueueStats struct {
	MsgStime  time.Time
	MsgRtime  time.Time
	MsgCtime  time.Time
	MsgCbytes uint64
	MsgQnum   uint64
	MsgQbytes uint64
	MsgLspid  int32
	MsgLrpid  int32
}

// GetMsgQueueStats gets the message queue statistics
func (MSM *MessageStatManager) GetMsgQueueStats(key int) (*MsgQueueStats, error) {
	var stats C.MsgQueueStats

	// Call the C function to fetch message queue stats
	res := C.get_msg_queue_stats(C.key_t(key), &stats)
	if res != 0 {
		return nil, fmt.Errorf("failed to get message queue stats")
	}

	return &MsgQueueStats{
		MsgStime:  time.Unix(int64(stats.msg_stime), 0),
		MsgRtime:  time.Unix(int64(stats.msg_rtime), 0),
		MsgCtime:  time.Unix(int64(stats.msg_ctime), 0),
		MsgCbytes: uint64(stats.__msg_cbytes),
		MsgQnum:   uint64(stats.msg_qnum),
		MsgQbytes: uint64(stats.msg_qbytes),
		MsgLspid:  int32(stats.msg_lspid),
		MsgLrpid:  int32(stats.msg_lrpid),
	}, nil
}
