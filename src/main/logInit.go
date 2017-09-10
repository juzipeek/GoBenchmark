package main

import (
	"fmt"
	"os"

	"github.com/cihub/seelog"
)

func LogInit() {

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
		fmt.Printf("LoggerFromConfigAsBytes Error:%s", err.Error())
		os.Exit(-1)
	}

	err = seelog.ReplaceLogger(logger)
	if err != nil {
		fmt.Printf("ReplaceLogger Error!%s\n", err.Error())
		os.Exit(-2)
	}
}
