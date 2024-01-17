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
			fmt.Fprintf(writer, "Device:%s\nrrqm/s:%s\nwrqm/s:%s\nr/s:%s\nw/s:%s\nrMB/s:%s\nwMB/s:%s\navgrq-sz:%s\navgqu-sz:%s\nawait:%s\nr_await:%s\nw_await:%s\n%%util:%s", v.Device, v.RRQM, v.WRQM, v.R, v.W, v.RMB, v.WMB, v.AvgrqSz, v.AvgquSz, v.Await, v.RWait, v.WWait, v.Util)
		}
	}

}
