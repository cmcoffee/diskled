package main

import (
	"bytes"
	"bufio"
	"io/ioutil"
	"fmt"
	"strings"
	"time"
	"strconv"
	"github.com/cmcoffee/go-snuglib/eflag"
	"os"
)

const sys_path = "/sys/class/gpio"

type gpio struct {
	target string
	enabled int
}

func GPIO(n int) gpio {
	target := strconv.Itoa(n)
	ioutil.WriteFile(fmt.Sprintf("%s/export", sys_path), []byte(target), 0666)
	ioutil.WriteFile(fmt.Sprintf("%s/gpio%s/direction", sys_path, target), []byte("out"), 0666)
	return gpio{fmt.Sprintf("%s/gpio%s/value", sys_path, target), 0}
}

func (g gpio) Get() bool {
	out, _ := ioutil.ReadFile(g.target)
	if x, err := strconv.Atoi(string(out)); err == nil && x == 1 {
		return true
	} else {
		return false
	}
}

func (g *gpio) Enable() {
	ioutil.WriteFile(g.target, []byte("1"), 0666)
	g.enabled = 1
}

func (g *gpio) Disable() {
	if g.enabled == 1 {
		ioutil.WriteFile(g.target, []byte("0"), 0666)
		g.enabled = 0
	}
}

var stat_map = make(map[string]string)
var access_map = make(map[string]bool)

func get_disk_stats() {
	stats, err := ioutil.ReadFile("/proc/diskstats")
	if err != nil {
		return
	}
	scanner := bufio.NewScanner(bytes.NewReader(stats))
	for scanner.Scan() {
		txt := scanner.Text()
		txt_len := len(txt)
		split := strings.Split(txt, " ")
		split_len := len(split)
		if split_len > 15 {
			name := split[split_len - 16]
			access_map[name] = false

			if val, ok := stat_map[name]; ok {
				if txt_len == len(val) {
					for i := 0; i < len(txt); i++ {
						if val[i] != txt[i] {
							access_map[name] = true
							break
						}
					}
				}

			}
			stat_map[name] = txt
		}
	}
	return
}

func main() {
	fatal := func(input string) {
		fmt.Printf("fatal: %s\n", input)
		os.Exit(1)
	}

	freq := eflag.Int("interval", 300, "Repeat every x milliseconds.")
	input := eflag.Multi("disks", "<disk:gpio>", "Specify which disk to blink which gpo, example: --disks=sda:438")
	eflag.Footer(fmt.Sprintf("\nExample Syntax: %s --interval=300 --disks=sda:436,sdb:438", os.Args[0]))
	err := eflag.Parse()
	if err != nil {
		fatal(err.Error())
	}
	var (
		gpios []gpio
		disks []string
	)

	for _, v := range *input {
		split := strings.Split(v, ":")
		if len(split) < 2 {
			fatal(fmt.Sprint("Incorrect value for --disks, syntax is --disks=<disk>:<gpio>", v))
		}
		disks = append(disks, split[0])
		num, err := strconv.Atoi(split[1])
		if err != nil {
			fatal(fmt.Sprintf("gpio must be numeric value, got: --disks=%s, syntax is --disks=<disk>:<gpio>", v))
		}
		gpios = append(gpios, GPIO(num))
	}

	if len(gpios) == 0 {
		eflag.Usage()
		os.Exit(1)
	}

	for {
		get_disk_stats()

		for i, v := range disks {
			if access_map[v] {
				gpios[i].Enable()
			} else {
				gpios[i].Disable()
			}
		}

		time.Sleep(time.Millisecond * time.Duration(*freq/2))
		for _, v := range gpios {
			v.Disable()
		}
		time.Sleep(time.Millisecond * time.Duration(*freq/2))
	}
}