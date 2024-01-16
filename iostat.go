package giostat

/*
#include <unistd.h>
#include <sys/types.h>
#include <pwd.h>
#include <stdlib.h>
*/
import "C"

var ScClkTck C.long

func init() {
	ScClkTck = C.sysconf(C._SC_CLK_TCK)
}
