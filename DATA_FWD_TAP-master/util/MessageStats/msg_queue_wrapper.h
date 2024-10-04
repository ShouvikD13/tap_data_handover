#ifndef MSG_QUEUE_WRAPPER_H
#define MSG_QUEUE_WRAPPER_H

#include <sys/types.h>
#include <time.h> // Include this for time_t
#include <unistd.h> // Optional, include if you use pid_t

#ifdef __cplusplus
extern "C" {
#endif

typedef struct {
    time_t msg_stime;
    time_t msg_rtime;
    time_t msg_ctime;
    unsigned long __msg_cbytes;
    unsigned long msg_qnum;
    unsigned long msg_qbytes;
    pid_t msg_lspid;
    pid_t msg_lrpid;
} MsgQueueStats;

int get_msg_queue_stats(key_t key, MsgQueueStats* stats);

#ifdef __cplusplus
}
#endif

#endif // MSG_QUEUE_WRAPPER_H
