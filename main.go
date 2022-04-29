package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-ping/ping"
)

const (
	CommandGitPull   = "git pull"
	CommandGitPush   = "git push"
	CommandGitStatus = "git status -s"
	CommandGitAdd    = "git add ."
	CommandGitCommit = "git commit -m '%s'"
)

var loopInterval int

// TODO
var netHealthy bool

func main() {
	flag.IntVar(&loopInterval, "i", 5, "执行间隔时间，单位：秒")
	flag.Parse()

	go gitLoop(loopInterval)

	waitSignal()
}

func execCommand(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	out, err := cmd.CombinedOutput()
	outString := string(out)
	if err != nil {
		log.Printf("exec command: %s, error: %s, msg: %s", cmd.String(), err, outString)
	} else {
		log.Printf("exec command: %s, msg: %s", cmd.String(), outString)
	}
	return outString, err
}

func gitLoop(loopInterval int) error {
	for {
		func() {
			if _, err := execCommand(CommandGitPull); err != nil {
				return
			}

			out, err := execCommand(CommandGitStatus)
			if err != nil {
				return
			}
			if len(out) == 0 {
				return
			}

			if _, err := execCommand(CommandGitAdd); err != nil {
				return
			}

			commitMsg := fmt.Sprintf("changed files:\n%s", out)
			commitCmd := fmt.Sprintf(CommandGitCommit, commitMsg)
			_, err = execCommand(commitCmd)
			if err != nil {
				return
			}

			if _, err := execCommand(CommandGitPush); err != nil {
				return
			}
		}()
		time.Sleep(time.Duration(loopInterval) * time.Second)
	}
}

func waitSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	si := <-c
	log.Printf("program exit %s!", si)
}

func checkNetHealth() {
	pinger, err := ping.NewPinger("114.114.114.114")
	if err != nil {
		log.Fatalln(err)
	}

	pinger.Interval = 1 * time.Second

	pinger.OnRecv = func(pkt *ping.Packet) {
		fmt.Printf("%d bytes from %s: icmp_seq=%d time=%v\n",
			pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt)
	}

	if err := pinger.Run(); err != nil {
		log.Fatalln(err)
	}
}
