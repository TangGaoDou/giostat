package giostat

/*
#include <unistd.h>
#include <sys/types.h>
#include <pwd.h>
#include <stdlib.h>
*/
import "C"
import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

var scClkTck C.long

func getHz() {
	ticks := C.sysconf(C._SC_CLK_TCK)
	if ticks == -1 {
		panic("unsupported platform")
	}
	scClkTck = ticks
}

func sValue(m uint64, n uint64, p float64) float64 {
	return float64(n-m) / p * float64(scClkTck)
}

func readUptime() (float64, error) {
	file, err := os.Open("/proc/uptime")
	if err != nil {
		return 0, err
	}
	defer file.Close()
	var upSec, upCent float64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		upSec, _ = strconv.ParseFloat(fields[0], 64)
		upCent, _ = strconv.ParseFloat(fields[1], 64)

	}
	return upSec*float64(scClkTck) + upCent*float64(scClkTck)/100, nil
}

func Get() (out []OutDiskStats, err error) {
	getHz()
	return collectDiskStats()
}

func collectDiskStats() (out []OutDiskStats, err error) {
	var (
		starTime, endTime float64
	)

	starTime, err = readUptime()
	if err != nil {
		return out, err
	}
	prevDiskStats, devNumS, err := getProcDiskStats()
	if err != nil {
		return out, err
	}
	time.Sleep(1 * time.Second)
	endTime, err = readUptime()
	if err != nil {
		return out, err
	}
	currDiskStats, devNumE, err := getProcDiskStats()
	if err != nil {
		return out, err
	}

	if devNumS != devNumE {
		return out, errors.New("device number has been changed")
	}

	itv := endTime - starTime
	out = make([]OutDiskStats, devNumE)
	for i := 0; i < devNumE; i++ {
		util := sValue(prevDiskStats[i].TimeSpentDoingIOs, currDiskStats[i].TimeSpentDoingIOs, itv)
		await := 0.0
		arqsz := 0.0

		if ((currDiskStats[i].ReadsCompleted + currDiskStats[i].WritesCompleted) - (prevDiskStats[i].ReadsCompleted + prevDiskStats[i].WritesCompleted)) != 0 {
			await = (float64(currDiskStats[i].TimeSpentReading-prevDiskStats[i].TimeSpentReading) + float64(currDiskStats[i].TimeSpentWriting-prevDiskStats[i].TimeSpentWriting)) / (float64((currDiskStats[i].ReadsCompleted + currDiskStats[i].WritesCompleted) - (prevDiskStats[i].ReadsCompleted + prevDiskStats[i].WritesCompleted)))

			arqsz = (float64(currDiskStats[i].SectorsRead-prevDiskStats[i].SectorsRead) + float64(currDiskStats[i].SectorsWritten-prevDiskStats[i].SectorsWritten)) / (float64((currDiskStats[i].ReadsCompleted + currDiskStats[i].WritesCompleted) - (prevDiskStats[i].ReadsCompleted + prevDiskStats[i].WritesCompleted)))
		}
		r_await := 0.0
		w_await := 0.0
		if currDiskStats[i].ReadsCompleted-prevDiskStats[i].ReadsCompleted != 0 {
			r_await = float64(currDiskStats[i].TimeSpentReading-prevDiskStats[i].TimeSpentReading) / float64(currDiskStats[i].ReadsCompleted-prevDiskStats[i].ReadsCompleted)
		}
		if currDiskStats[i].WritesCompleted-prevDiskStats[i].WritesCompleted != 0 {
			w_await = float64(currDiskStats[i].TimeSpentWriting-prevDiskStats[i].TimeSpentWriting) / float64(currDiskStats[i].WritesCompleted-prevDiskStats[i].WritesCompleted)
		}

		out[i].Device = currDiskStats[i].DeviceName
		out[i].RRQM = fmt.Sprintf("%8.2f", sValue(prevDiskStats[i].ReadsMerged, currDiskStats[i].ReadsMerged, itv))
		out[i].WRQM = fmt.Sprintf("%8.2f", sValue(prevDiskStats[i].WritesMerged, currDiskStats[i].WritesMerged, itv))
		out[i].R = fmt.Sprintf("%9.2f", sValue(prevDiskStats[i].ReadsCompleted, currDiskStats[i].ReadsCompleted, itv))
		out[i].W = fmt.Sprintf("%9.2f", sValue(prevDiskStats[i].WritesCompleted, currDiskStats[i].WritesCompleted, itv))
		out[i].RMB = fmt.Sprintf("%9.2f", sValue(prevDiskStats[i].SectorsRead, currDiskStats[i].SectorsRead, itv)/float64(2048))
		out[i].WMB = fmt.Sprintf("%9.2f", sValue(prevDiskStats[i].SectorsWritten, currDiskStats[i].SectorsWritten, itv)/float64(2048))
		out[i].AvgrqSz = fmt.Sprintf("%9.2f", arqsz)
		out[i].AvgquSz = fmt.Sprintf("%9.2f", sValue(prevDiskStats[i].WeightedTimeSpentDoingIOs, currDiskStats[i].WeightedTimeSpentDoingIOs, itv)/float64(1000))
		out[i].Await = fmt.Sprintf("%10.2f", await)
		out[i].RWait = fmt.Sprintf("%10.2f", r_await)
		out[i].WWait = fmt.Sprintf("%10.2f", w_await)
		out[i].Util = fmt.Sprintf("%10.2f", util/float64(10))
	}
	return out, nil
}

func getProcDiskStats() (procDiskStats []DiskStats, deviceNumber int, err error) {
	file, err := os.Open("/proc/diskstats")
	if err != nil {
		return procDiskStats, 0, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 14 {
			continue
		}
		major, _ := strconv.Atoi(fields[0])
		minor, _ := strconv.Atoi(fields[1])
		devname := fields[2]
		r_completed, _ := strconv.ParseUint(fields[3], 10, 64)
		r_merged, _ := strconv.ParseUint(fields[4], 10, 64)
		r_sectors, _ := strconv.ParseUint(fields[5], 10, 64)
		r_spent, _ := strconv.ParseUint(fields[6], 10, 64)
		w_completed, _ := strconv.ParseUint(fields[7], 10, 64)
		w_merged, _ := strconv.ParseUint(fields[8], 10, 64)
		w_sectors, _ := strconv.ParseUint(fields[9], 10, 64)
		w_spent, _ := strconv.ParseUint(fields[10], 10, 64)
		io_in_progress, _ := strconv.ParseUint(fields[11], 10, 64)
		t_spent, _ := strconv.ParseUint(fields[12], 10, 64)
		t_weighted, _ := strconv.ParseUint(fields[13], 10, 64)
		procDiskStats = append(procDiskStats, DiskStats{
			Major:                     major,
			Minor:                     minor,
			DeviceName:                devname,
			ReadsCompleted:            r_completed,
			ReadsMerged:               r_merged,
			SectorsRead:               r_sectors,
			TimeSpentReading:          r_spent,
			WritesCompleted:           w_completed,
			WritesMerged:              w_merged,
			SectorsWritten:            w_sectors,
			TimeSpentWriting:          w_spent,
			IOsCurrentlyInProgress:    io_in_progress,
			TimeSpentDoingIOs:         t_spent,
			WeightedTimeSpentDoingIOs: t_weighted,
		})
		deviceNumber++
	}
	//fmt.Println(procDiskStats, deviceNumber)
	return procDiskStats, deviceNumber, nil
}

// https://www.kernel.org/doc/Documentation/ABI/testing/procfs-diskstats
type DiskStats struct {
	Major                     int    `json:"major"`
	Minor                     int    `json:"minor"`
	DeviceName                string `json:"devname"`
	ReadsCompleted            uint64 `json:"r_completed"`
	ReadsMerged               uint64 `json:"r_merged"`
	SectorsRead               uint64 `json:"r_sectors"`
	TimeSpentReading          uint64 `json:"r_spent"`
	WritesCompleted           uint64 `json:"w_completed"`
	WritesMerged              uint64 `json:"w_merged"`
	SectorsWritten            uint64 `json:"w_sectors"`
	TimeSpentWriting          uint64 `json:"w_spent"`
	IOsCurrentlyInProgress    uint64 `json:"io_in_progress"`
	TimeSpentDoingIOs         uint64 `json:"t_spent"`
	WeightedTimeSpentDoingIOs uint64 `json:"t_weighted"`
}

type OutDiskStats struct {
	Device  string `json:"devname"`
	RRQM    string `json:"rrqm"`
	WRQM    string `json:"wrqm"`
	R       string `json:"r"`
	W       string `json:"w"`
	RMB     string `json:"rMB"`
	WMB     string `json:"wMB"`
	AvgrqSz string `json:"avgrq-sz"`
	AvgquSz string `json:"avgqu-sz"`
	Await   string `json:"await"`
	RWait   string `json:"r_wait"`
	WWait   string `json:"w_wait"`
	Util    string `json:"util"`
}
