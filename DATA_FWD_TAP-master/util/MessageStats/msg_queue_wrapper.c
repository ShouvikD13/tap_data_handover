// msg_queue_wrapper.c
#include <stdio.h>
#include <sys/ipc.h>
#include <sys/msg.h>
#include <sys/types.h>
#include <time.h>

// Structure for the message queue statistics
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

// Function to get message queue statistics
int get_msg_queue_stats(key_t key, MsgQueueStats* stats) {
    int msqid;
    struct msqid_ds buf;

    // Get the message queue ID using the key
    if ((msqid = msgget(key, 0666)) == -1) {
        perror("msgget");
        return -1;
    }

    // Fetch the stats of the message queue using msgctl and IPC_STAT
    if (msgctl(msqid, IPC_STAT, &buf) == -1) {
        perror("msgctl");
        return -1;
    }

    // Copy the stats to the output structure
    stats->msg_stime = buf.msg_stime;
    stats->msg_rtime = buf.msg_rtime;
    stats->msg_ctime = buf.msg_ctime;
    stats->__msg_cbytes = buf.__msg_cbytes;
    stats->msg_qnum = buf.msg_qnum;
    stats->msg_qbytes = buf.msg_qbytes;
    stats->msg_lspid = buf.msg_lspid;
    stats->msg_lrpid = buf.msg_lrpid;

    return 0;
}
