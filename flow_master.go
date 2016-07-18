package main

import (
	"fmt"
	"os"
	"os/signal"
)

func masterFlow() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, ListeningStartedSignal)

	logger.Debugf("starting fork process")

	process, err := os.StartProcess(
		os.Args[0], append(os.Args),
		&os.ProcAttr{
			Env: append(os.Environ(), "BLANKD_FORK=1"),
		},
	)
	if err != nil {
		logger.Fatalf("can't fork: %s", err)
	}

	logger.Debugf("fork pid = %d", process.Pid)

	go func() {
		state, err := process.Wait()
		if err != nil {
			logger.Fatal(err)
		}

		logger.Fatalf("fork process unexpectedly exited: %s", state)
	}()

	logger.Debugf("waiting for signal %d", ListeningStartedSignal)

	<-signals

	logger.Debugf("got signal %d", ListeningStartedSignal)

	fmt.Println(process.Pid)

	process.Release()
}
