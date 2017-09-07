package main

import (
	"net/http"
	"time"

	"bytes"
	"io"

	"sync"

	"io/ioutil"

	"github.com/cihub/seelog"
)

// 缩小临界区
func waitSignal(c *sync.Cond) {

	c.L.Lock()
	defer c.L.Unlock()

	c.Wait()

}

// 进行http/https请求
func doRequest(worker *RequestWorker) error {

	var err error
	var req *http.Request

	defer worker.waitGroup.Done()
	//defer worker.cond.L.Unlock()

	request_info := worker.requestInfo

	// 等待执行信号
	waitSignal(worker.cond)

	body_reader := bytes.NewReader(request_info.requestBody)

	if len(request_info.requestBody) > 0 {
		req, err =
			http.NewRequest("POST", request_info.requestUrl, body_reader)
	} else {
		req, err = http.NewRequest("GET", request_info.requestUrl, nil)
	}

	if err != nil {
		seelog.Debugf("Create NewRequest Error!%s", err.Error())
		return err
	}

	// 进行请求处理
	for i := 0; i < request_info.requestCount; i++ {

		// client := &http.Client{}

		client := &http.Client{Timeout: time.Second * 0}

		body_reader.Seek(0, io.SeekStart)

		resp, err := client.Do(req)
		if err != nil {
			worker.failCount++
			seelog.Debugf("Call Do Method Error!%s", err.Error())

			if request_info.failContinueFlag {
				continue
			} else {
				break
			}
		}

		ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != 200 {
			seelog.Debugf("Resp Status:%s", resp.Status)
			worker.failCount++
			continue
		} else {
			worker.successCount++
		}

	}

	return nil
}
