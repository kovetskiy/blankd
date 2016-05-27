package main

import "github.com/kovetskiy/lorg"

const (
	masterFormat = `master: ${level:%s\::left:true} ${time} %s`
	forkFormat   = `fork:   ${level:%s\::left:true} ${time} %s`
)

func getLogger() *lorg.Log {
	logger := lorg.NewLog()
	logger.SetLevel(lorg.LevelDebug)

	return logger
}
