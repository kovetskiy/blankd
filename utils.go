package main

/*
#include <unistd.h>
#include <sys/types.h>
#include <pwd.h>
#include <stdlib.h>
*/
import "C"

func getMaxFD() int {
	sc_open_max := C.sysconf(C._SC_OPEN_MAX)
	return int(sc_open_max)
}
