package main

import (
	"container/heap"
	"flag"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"io/ioutil"
	"time"
	"github.com/rthornton128/goncurses"
	"github.com/dustin/go-humanize"
	"strings"
	"os"
	"os/signal"
	"syscall"
	"bufio"
//	"runtime/pprof"
	"strconv"
	"regexp"
)

// TODO: Написать тесты
// TODO: Написать README.md

// Size of typical Jumbo frames
// Hmm... What is the max size of 'mgets' may be?
const CAPTURE_SIZE = 9000
var Config_ Config
const topLimit = 1000
var startTs = time.Now()

func formatCommas(num uint64) string {
	str := strconv.FormatUint(num, 10)
	re := regexp.MustCompile("(\\d+)(\\d{3})")
	for i := 0; i < (len(str) - 1) / 3; i++ {
		str = re.ReplaceAllString(str, "$1,$2")
	}
	return str
}

func showStat(config Config, stat *Stat) {
	sleep_duration := time.Duration(config.Interval) * time.Second
	time.Sleep(sleep_duration)

	// Do ncurses cleanup on SIGINT/SIGTERM and redraw screen on SIGWINCH
	sigs := make(chan os.Signal, 1)
	wchg := make(chan os.Signal, 1)
	wchgscr := false
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	signal.Notify(wchg, syscall.SIGWINCH)

	stdscr, _ := goncurses.Init()
	defer goncurses.End()

	go func() {
		_ = <-sigs
		goncurses.End()
//		pprof.StopCPUProfile()
		os.Exit(0)
	}()

	go func() {
		for {
			_ = <-wchg
			wchgscr = true
		}
	}()

	var outHandle *os.File

	if config.OutputFile != "" {
		outfile, err := os.OpenFile(config.OutputFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0x644)
		if err != nil {
			panic(err)
		}
		defer outfile.Close()
		outHandle = outfile
	}

	for {
		st := time.Now()
		if wchgscr {
			goncurses.End()
			stdscr, _ = goncurses.Init()
			stdscr.Erase()
			wchgscr = false
		}

		rows, cols := stdscr.MaxYX()

		clear := false
		if config.OutputFile != "" {
			clear = true
		}
		rotated_stat := stat.Rotate(clear)
		top := rotated_stat.GetTopKeys()

		output := ""
		output  = "       Reads       Writes          Read bytes       Written bytes   Key\n"
		output += strings.Repeat("-", cols)

		foutput := ""
		i := 0
		var totalRC, totalWC, totalRB, totalWB uint64
		for {
			if top.Len() == 0 || i > topLimit {
				break
			}

			key := heap.Pop(top)

			if i <= (rows - 6) {
				output += fmt.Sprintf("%12s %12s %19s %19s   %s\n",
					formatCommas(key.(*KeyStat).RCount),
					formatCommas(key.(*KeyStat).WCount),
					humanize.Bytes(key.(*KeyStat).RBytes),
					humanize.Bytes(key.(*KeyStat).WBytes),
					key.(*KeyStat).Name)
			}

			totalRC += key.(*KeyStat).RCount
			totalWC += key.(*KeyStat).WCount
			totalRB += key.(*KeyStat).RBytes
			totalWB += key.(*KeyStat).WBytes

			if config.OutputFile != "" {
				foutput += fmt.Sprintf("%12d %12d %19d %19d   %s\n",
					key.(*KeyStat).RCount,
					key.(*KeyStat).WCount,
					key.(*KeyStat).RBytes,
					key.(*KeyStat).WBytes,
					key.(*KeyStat).Name)
			}
			i += 1
		}

		tsDiff := time.Now().Sub(startTs)
		output += strings.Repeat("-", cols)
		output += fmt.Sprintf("%12s %12s %19s %19s",
			formatCommas(totalRC/uint64(tsDiff.Seconds())) + "/sec",
			formatCommas(totalWC/uint64(tsDiff.Seconds())) + "/sec",
			humanize.Bytes(totalRB/uint64(tsDiff.Seconds())) + "/sec",
			humanize.Bytes(totalWB/uint64(tsDiff.Seconds())) + "/sec")

		stdscr.MovePrint(0, 0, output)
		stdscr.Refresh()
		goncurses.Cursor(0)

		if config.OutputFile != "" {
			writer := bufio.NewWriter(outHandle)
			fmt.Fprint(writer, foutput)
			writer.Flush()
		}

		elapsed := time.Now().Sub(st)
		time.Sleep(sleep_duration - elapsed)
	}
}

func main() {

	var err error
/*	f, err := os.Create("mcache-ktop.pprof")
	if err != nil {
		panic(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()
*/
	interval := flag.Int("d", 3, "update interval (seconds, default 3)")
	net_interface := flag.String("i", "any", "capture interface (default any)")
	ip := flag.String("h", "", "capture ip address (i.e. for bond with multiple IPs)")
	port := flag.Int("p", 11211, "capture port")
	output_file := flag.String("o", "", "file to write output to")
	config_file := flag.String("c", "", "config file")
	sortby := flag.String("s", "rcount", "sort by (rcount|wcount|rbytes|wbytes)")

	flag.Parse()

	if *config_file != "" {
		config_data, _ := ioutil.ReadFile(*config_file)
		Config_, err = NewConfig(config_data)
	} else {
		Config_, err = NewConfig([]byte{})
	}
	if err != nil {
		panic(err)
	}

	if *interval != 0 {
		Config_.Interval = *interval
	}
	if *net_interface != "" {
		Config_.Interface = *net_interface
	}
	if *ip != "" {
		Config_.IpAddress = *ip
	}
	if *port != 0 {
		Config_.Port = *port
	}
	if *output_file != "" {
		Config_.OutputFile = *output_file
	}
	if *sortby != "" {
		Config_.SortBy = *sortby
	}
	re_keys := NewRegexpKeys()
	for _, re := range Config_.Regexps {
		regexp_key, err := NewRegexpKey(re.Re, re.Name)
		if err != nil {
			panic(err)
		}
		re_keys.Add(regexp_key)
	}

	kstat := NewStat()

	handle, err := pcap.OpenLive(Config_.Interface, CAPTURE_SIZE, true, pcap.BlockForever)
	if err != nil {
		panic(err)
	}
	var filter string
	if *ip != "" {
		filter = fmt.Sprintf("tcp and host %s && port %d", Config_.IpAddress, Config_.Port)
	} else {
		filter = fmt.Sprintf("tcp and port %d", Config_.Port)
	}

	err = handle.SetBPFFilter(filter)
	if err != nil {
		panic(err)
	}
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	go showStat(Config_, kstat)

	var (
		payload  []byte
		keys     []KeyStat
		cmd_err  int
	)
	for packet := range packetSource.Packets() {
		app_data := packet.ApplicationLayer()
		if app_data == nil {
			continue
		}
		payload = app_data.Payload()
//		payload2 := payload
//		fmt.Printf("%d, %s\n", cmd_err, payload2)

		if len(payload) < 1 {
			continue // nothing to parse
		}

		keys, cmd_err = parse(payload)



		if cmd_err == ERR_NO_CMD {
			continue
		}

		if cmd_err == ERR_NONE {

			if len(Config_.Regexps) == 0 {
				kstat.Add(keys)
			} else {
				matches := []KeyStat{}
				for _, key := range keys {
					key.Name, err = re_keys.Match(key.Name)
					matches = append(matches, key)
				}
				kstat.Add(matches)
			}
		}
/*		if cmd_err != ERR_NONE_SKIP && cmd_err != ERR_NONE {
				fmt.Printf("%d, %s\n", cmd_err, payload2)
		}*/
	}
}
