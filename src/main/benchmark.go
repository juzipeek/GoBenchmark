package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"sync"

	"html/template"

	"bytes"

	"github.com/cihub/seelog"
)

var ReportTemplate string = `
测试结果
	并发数:{{ .Concurrency}} 每线程请求数:{{ .Count}}
	失败次数:{{ .FailCount}} 成功次数:{{ .SuccessCount}}
	耗时:{{ .Cost | printf "%.4f"}}s Tps:{{ .Tps | printf "%.3f"}}
	成功次数TPS:{{ .SuccessTps | printf "%.3f"}}
	失败次数TPS:{{ .FailedTps | printf "%.3f"}}
`

/*
 *
 */
func LogInit() (ok bool) {

	ok = true
	log_conf := `
	<seelog type="asynctimer" asyncinterval="1000" minlevel="debug"
		maxlevel="error">
		<outputs formatid="main">
			<buffered size="10000" flushperiod="1000">
				<rollingfile type="date" filename="run.log"
					datepattern="20170113" maxrolls="30"/>
			</buffered>
		</outputs>
		<formats>
			<format id="main" format="%Time|%LEV|%File|%Line|%Msg%n"/>
		</formats>
	</seelog>`

	logger, err := seelog.LoggerFromConfigAsBytes([]byte(log_conf))
	if err != nil {
		ok = false
		return
	}

	err = seelog.ReplaceLogger(logger)
	if err != nil {
		ok = false
		fmt.Printf("ReplaceLogger Error!%s\n", err.Error())
		return
	}

	return
}

func main() {

	succ := LogInit()
	if !succ {
		fmt.Println("Init Log Error,Exit Now!")
		os.Exit(0)
	}

	defer seelog.Flush()

	seelog.Debugf("Log Init Success!")

	var dataFile, requestUrl string
	var concurrency, totalCount int
	var failContinue bool

	flag.StringVar(&dataFile, "p", "", "Http body file")
	flag.StringVar(&requestUrl, "u", "", "Request Url")
	flag.IntVar(&concurrency, "c", 10, "Concurrency")
	flag.IntVar(&totalCount, "n", 10, "Total Count")
	flag.BoolVar(&failContinue, "r", false, "Continue if error occour."+
		"Default is false")

	flag.Parse()
	seelog.Debugf("param args:[%v][%v][%v][%v][%v]",
		dataFile, requestUrl, concurrency, totalCount, failContinue)

	if len(requestUrl) == 0 {
		flag.Usage()
		return
	}

	if concurrency > totalCount {
		flag.Usage()
		return
	}

	var err error
	var requestBody []byte
	var count int

	count = totalCount / concurrency
	if len(dataFile) > 0 {
		requestBody, err = ioutil.ReadFile(dataFile)
		if err != nil {
			seelog.Errorf("Read File %s Error! %s", dataFile, err.Error())
			return
		}
		seelog.Debugf("Read File %s Length %d", dataFile, len(requestBody))
	}

	requestInfo := RequestInfo{}

	requestInfo.requestUrl = requestUrl
	requestInfo.failContinueFlag = failContinue
	requestInfo.requestCount = count
	requestInfo.requestBody = requestBody

	seelog.Debugf("Routine will dorequest use param %v", requestInfo)

	waitGroup := new(sync.WaitGroup)

	cond := sync.NewCond(new(sync.Mutex))

	workers := make([]RequestWorker, concurrency)
	for i := 0; i < concurrency; i++ {
		workers[i] = RequestWorker{requestInfo: &requestInfo,
			waitGroup: waitGroup, cond: cond}
		go doRequest(&workers[i])
	}
	waitGroup.Add(concurrency)

	// Broadcast 前需要必须要等待段时间，不知道为什么
	time.Sleep(time.Second * 2)

	beginTime := time.Now()
	cond.Broadcast()
	waitGroup.Wait()
	endTime := time.Now()

	summaryInfo := SummaryInfo{}
	for i := 0; i < concurrency; i++ {
		summaryInfo.failCount += workers[i].failCount
		summaryInfo.successCount += workers[i].successCount
	}

	cost := endTime.Sub(beginTime).Seconds()

	buffer := bytes.NewBuffer(make([]byte, 2048))

	report, err := template.New("Report").Parse(ReportTemplate)

	if err != nil {
		seelog.Debugf("create template error")
		return
	}

	report.Execute(buffer,
		struct {
			Concurrency, Count      int
			FailCount, SuccessCount uint64
			Cost, Tps               float64
			SuccessTps, FailedTps   float64
		}{
			Concurrency: concurrency, Count: count,
			FailCount:    summaryInfo.failCount,
			SuccessCount: summaryInfo.successCount,
			Cost:         endTime.Sub(beginTime).Seconds(),
			Tps:          float64(count*concurrency) / cost,
			SuccessTps:   float64(summaryInfo.successCount) / cost,
			FailedTps:    float64(summaryInfo.failCount) / cost,
		})
	output := buffer.String()
	fmt.Println(output)
	seelog.Debugf("%s", output)

	return

}
