package main

import (
	"fmt"

	"github.com/TangGaoDou/giostat"
	"github.com/gosuri/uilive"
)

func main() {
	// fmt.Println("device:           rrqm/s  wrqm/s      r/s      w/s    rMB/s    wMB/s avgrq-sz avgqu-sz     await   r_await   w_await   %%util")
	writer := uilive.New()
	writer.Start()
	for {
		s, err := giostat.Get()
		if err != nil {
			panic(err)
		}
		for _, v := range s {
			fmt.Fprintf(writer, "device:           rrqm/s  wrqm/s      r/s      w/s    rMB/s    wMB/s avgrq-sz avgqu-sz     await   r_await   w_await   %%util\n%s        %s  %s      %s      %s    %s    %s %s %s     %s   %s   %s   %s\n", v.Device, v.RRQM, v.WRQM, v.R, v.W, v.RMB, v.WMB, v.AvgrqSz, v.AvgquSz, v.Await, v.RWait, v.WWait, v.Util)
		}
	}

}
