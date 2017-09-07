// 用到的结构体定义
package main

import "sync"

// 请求信息
type RequestInfo struct {
	requestUrl       string // 请求url"
	requestBody      []byte // 请求bodyPOST方法使用"
	requestCount     int    // 请求次数"
	failContinueFlag bool   // 遇到失败时是否继续"
}

// 统计信息
type SummaryInfo struct {
	failCount    uint64 // 失败次数"
	successCount uint64 // 成功次数"
}

type RequestWorker struct {
	requestInfo  *RequestInfo    // 请求信息
	waitGroup    *sync.WaitGroup // work执行完成时调用
	cond         *sync.Cond      // 启动条件变量
	failCount    uint64          // 统计失败次数"
	successCount uint64          // 统计成功次数"
}
