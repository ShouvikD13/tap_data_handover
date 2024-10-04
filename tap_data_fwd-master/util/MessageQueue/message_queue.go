package MessageQueue

import (
	"DATA_FWD_TAP/internal/models"
	"DATA_FWD_TAP/util"
	MessageStat "DATA_FWD_TAP/util/MessageStats"
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/Shopify/sysv_mq"
)

type MessageQueueManager struct {
	ServiceName        string
	Req_q_data         models.St_req_q_data
	Req_q_data_Log_On  models.St_req_q_data_Log_On
	Req_q_data_Log_Off models.St_req_q_data_Log_Off
	MQ                 *sysv_mq.MessageQueue
	LoggerManager      *util.LoggerManager
	MSM                *MessageStat.MessageStatManager
}

type WriteQueueMessage interface {
	GetMsgType() int64
}

func (MQM *MessageQueueManager) pack(message WriteQueueMessage) ([]byte, error) {
	var buf bytes.Buffer

	if err := binary.Write(&buf, binary.LittleEndian, message.GetMsgType()); err != nil {
		return nil, err
	}

	// Packind Second value of the structure
	switch data := message.(type) {
	case *models.St_req_q_data:
		if err := binary.Write(&buf, binary.LittleEndian, data.St_exch_msg_data); err != nil {
			return nil, err
		}
	case *models.St_req_q_data_Log_On:
		if err := binary.Write(&buf, binary.LittleEndian, data.St_exch_msg_Log_On); err != nil {
			return nil, err
		}
	case *models.St_req_q_data_Log_Off:
		if err := binary.Write(&buf, binary.LittleEndian, data.St_exch_msg_Log_Off); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported message type")
	}

	return buf.Bytes(), nil
}

func (MQM *MessageQueueManager) CreateQueue(key int) (*sysv_mq.MessageQueue, int) {
	// Initialize MessageStatManager instance and assign it to MQM.MSM
	// This sets up the MessageStatManager for managing message queue statistics
	MQM.MSM = &MessageStat.MessageStatManager{}
	mq, err := sysv_mq.NewMessageQueue(&sysv_mq.QueueConfig{
		Key:     key,
		MaxSize: 1024 * 3,
		Mode:    (sysv_mq.IPC_CREAT | 0777),
	})
	if err != nil {
		MQM.LoggerManager.LogError(MQM.ServiceName, "[CreateQueue] [Error: failed to Create Message Queue: %v", err)
		return nil, -1
	}

	return mq, 0
}

func (MQM *MessageQueueManager) WriteToQueue(mId int, message WriteQueueMessage) int {
	packedData, err := MQM.pack(message)
	if err != nil {
		MQM.LoggerManager.LogError(MQM.ServiceName, "[writeToQueue] [Error: failed to pack data in Message Queue: %v", err)
		return -1
	}

	MQM.LoggerManager.LogInfo(MQM.ServiceName, "[writeToQueue] Packed data size: %d bytes", len(packedData))
	MQM.LoggerManager.LogInfo(MQM.ServiceName, "[writeToQueue] MQM <----------------------: %v", MQM)
	MQM.LoggerManager.LogInfo(MQM.ServiceName, "[writeToQueue] MQ <----------------------: %v", MQM.MQ)
	err = MQM.MQ.SendBytes(packedData, mId, 0)
	if err != nil {
		MQM.LoggerManager.LogError(MQM.ServiceName, "[writeToQueue] [Error: failed to send message in Message Queue: %v", err)
		return -1
	}
	MQM.LoggerManager.LogInfo(MQM.ServiceName, "[writeToQueue] Packed data size again: %d bytes", len(packedData))
	return 0
}

func (MQM *MessageQueueManager) ReadFromQueue(mId int) (int64, []byte, int) {

	MQM.LoggerManager.LogInfo(MQM.ServiceName, "[readFromQueue] **************** INSIDE THE  Message Queue: ****************")

	receivedData, receivedType, err := MQM.MQ.ReceiveBytes(mId, 0)
	if err != nil {
		MQM.LoggerManager.LogError(MQM.ServiceName, "[readFromQueue] [Error: failed to receive message from Message Queue: %v", err)
		return 0, nil, -1
	}

	MQM.LoggerManager.LogInfo(MQM.ServiceName, "[readFromQueue] Received message with type: %d", receivedType)
	MQM.LoggerManager.LogInfo(MQM.ServiceName, "[readFromQueue] Length of Received Data: %d", len(receivedData))

	var Li_msg_type int64
	buf := bytes.NewReader(receivedData)

	if err := binary.Read(buf, binary.LittleEndian, &Li_msg_type); err != nil {
		MQM.LoggerManager.LogError(MQM.ServiceName, "[readFromQueue] [Error: failed to decode Li_msg_type: %v", err)
		return 0, nil, -1
	}

	byteArray := make([]byte, buf.Len())
	if err := binary.Read(buf, binary.LittleEndian, &byteArray); err != nil {
		MQM.LoggerManager.LogError(MQM.ServiceName, "[readFromQueue] [Error: failed to decode byte array: %v", err)
		return 0, nil, -1
	}
	MQM.LoggerManager.LogError(MQM.ServiceName, "[readFromQueue] **************** RETURNING FROM THE  Message Queue SUCESSFULLY ****************")

	return Li_msg_type, byteArray, 0
}

func (MQM *MessageQueueManager) DestroyQueue() int {
	MQM.LoggerManager.LogInfo(MQM.ServiceName, "[destroyQueue] destroy message queue: ")
	err := MQM.MQ.Destroy()
	if err != nil {
		MQM.LoggerManager.LogError(MQM.ServiceName, "[destroyQueue] [Error:  failed to destroy message queue: %v", err)
		return -1
	}
	return 0
}

func (MQM *MessageQueueManager) FnCanWriteToQueue() int {

	key := util.ORDINARY_ORDER_QUEUE_ID

	// Call GetMsgQueueStats to get the message queue stats
	stats, err := MQM.MSM.GetMsgQueueStats(key)
	if err != nil {
		MQM.LoggerManager.LogError(MQM.ServiceName, "[canWriteToQueue] [Error: failed to get message queue stats: %v", err)
		return -1
	}

	if stats.MsgQnum < 1000 { //Here, I am setting the maximum number of messages that can be present in the queue. Based on my system, this number is 10.
		return 0
	}

	MQM.LoggerManager.LogInfo(MQM.ServiceName, "[canWriteToQueue] Queue is full, current message count: %d", stats.MsgQnum)
	return -1
}
