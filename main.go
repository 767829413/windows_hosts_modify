package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

type Hosts struct {
	ip   string
	host string
	path string
	mod  string
}

var hosts Hosts
var wg = &sync.WaitGroup{}

func init() {
	flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
	flag.CommandLine.Usage = func() {
		if _, err := fmt.Fprintf(os.Stderr, "Usage of %s:\n", "modify hosts file"); err != nil {
			panic(err)
		}
		flag.PrintDefaults()
	}
	flag.StringVar(&hosts.ip, "ip", "127.0.0.1", "The ip.")
	flag.StringVar(&hosts.host, "host", "localhost", "The host address.")
	flag.StringVar(&hosts.path, "path", "C:\\Windows\\System32\\drivers\\etc\\", "The hosts file address.")
	flag.StringVar(&hosts.mod, "mod", "Unknown", "The hosts file operate type modify or rollback")
}

func main() {
	flag.Parse()
	switch hosts.mod {
	case "modify": //修改hosts,并备份
		if hosts.host == "127.0.0.1" && hosts.ip == "localhost" {
			panic(errors.New("no operate"))
		}
		wg.Add(1)
		go modify(&hosts)
	case "rollback": //回滚上一次操作
		wg.Add(1)
		go rollback(&hosts)
	default:
		panic(errors.New("Unknown operate"))
	}
	wg.Wait()
}

func modify(hosts *Hosts) {
	hosts.bakHosts()
	hosts.write()
	wg.Done()
}

func rollback(hosts *Hosts) {
	hosts.rollbackHosts()
	wg.Done()
}

func (h *Hosts) bakHosts() {
	hostsFile, err := os.Open(h.path + "hosts")
	checkErr(err)
	hostsBak, err := os.Create(h.path + "hosts_bak")
	if err != nil {
		hostsBak, err = os.Open(h.path + "hosts_bak")
		checkErr(err)
	}
	defer fileClose(hostsFile, hostsBak)
	buf := bufio.NewReader(hostsFile)
	for {
		line, err := readLine(buf)
		if err != nil {
			break
		}
		_, err = hostsBak.Write(*line)
		checkErr(err)
	}
}

func (h *Hosts) write() {
	f, err := os.OpenFile(h.path+"hosts", os.O_APPEND|os.O_WRONLY, 0600)
	checkErr(err)
	defer fileClose(f)
	_, err = f.WriteString("\n" + h.ip + " " + h.host)
	checkErr(err)

}

func (h *Hosts) rollbackHosts() {
	hostsBak, err := os.Open(h.path + "hosts_bak")
	if err != nil {
		panic(errors.New("hosts_bak not found"))
	}
	hostsFile, err := os.Open(h.path + "hosts")
	if err != nil {
		panic(errors.New("hosts not found"))
	}
	fileClose(hostsBak, hostsFile)
	err = os.Rename(h.path+"hosts", h.path+"hosts"+"_"+strconv.Itoa(int(time.Now().Unix())))
	checkErr(err)
	err = os.Rename(h.path+"hosts_bak", h.path+"hosts")
}

func fileClose(fileHandles ...*os.File) {
	for _, v := range fileHandles {
		err := v.Close()
		checkErr(err)
	}
}

func readLine(r *bufio.Reader) (*[]byte, error) {
	line, isprefix, err := r.ReadLine()
	for isprefix && err == nil {
		var bs []byte
		bs, isprefix, err = r.ReadLine()
		line = append(line, bs...)
	}
	line = append(line, []byte("\n")...)
	return &line, err
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
