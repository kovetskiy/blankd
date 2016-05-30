package main

/*
#include <unistd.h>
#include <sys/types.h>
#include <pwd.h>
#include <stdlib.h>
*/
import "C"
import "bytes"

type buffer struct{ *bytes.Buffer }

func newBuffer(data []byte) *buffer {
	return &buffer{bytes.NewBuffer(data)}
}

func (*buffer) Close() error { return nil }

func getMaxFD() int {
	sc_open_max := C.sysconf(C._SC_OPEN_MAX)
	return int(sc_open_max)
}
