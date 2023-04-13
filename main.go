package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-ping/ping"
)

const (
	CommandGitPull    = "git pull"
	CommandGitPush    = "git push"
	CommandGitStatus  = "git status"
	CommandGitStatuss = "git status -s"
	CommandGitAdd     = "git add ."
	CommandGitCommit  = "git commit -m '%s'"
)

const (
	GitStatusUpToDateMessage = "Your branch is up to date"
	GitStatusAheadMessage    = "Your branch is ahead of"
)

var (
	loopInterval int
	count        int
	pull         bool
)

// TODO
var netHealthy bool

func main() {
	flag.IntVar(&loopInterval, "i", 5, "执行间隔时间，单位：秒")
	flag.IntVar(&count, "c", 0, "执行次数（值 <=0 不限制次数，默认不限制次数）")
	flag.BoolVar(&pull, "pull", false, "指定执行 pull 操作")
	flag.Parse()

	quitChan := make(chan struct{})
	go gitLoop(loopInterval, quitChan)
	waitSignal(quitChan)
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

func gitLoop(loopInterval int, quitChan chan struct{}) error {
	c := 0
	for {
		gitExec()
		c++

		if count > 0 && c >= count {
			close(quitChan)
			return nil
		}

		// 执行间隔时间
		time.Sleep(time.Duration(loopInterval) * time.Second)
	}
}

func gitExec() {
	var (
		out string
		err error
	)

	defer func() {
		if err != nil {
			log.Printf("gitExec cause err: %s\n", err)
		}
	}()
	if _, err := execCommand(CommandGitPull); err != nil {
		return
	}

	// 如果指定了 pull 操作，直接返回
	if pull {
		return
	}

	out, err = execCommand(CommandGitStatuss)
	if err != nil {
		return
	}

	if len(out) == 0 {
		out, err = execCommand(CommandGitStatus)
		if err != nil {
			return
		}
		if strings.Contains(out, GitStatusUpToDateMessage) {
			// 本地和远程已同步，不做任何操作
			return
		} else if strings.Contains(out, GitStatusAheadMessage) {
			// 本地领先于远程，执行 push
			if _, err = execCommand(CommandGitPush); err != nil {
				return
			}
			return
		}
	}

	if _, err = execCommand(CommandGitAdd); err != nil {
		return
	}

	commitMsg := fmt.Sprintf("changed files:\n%s", out)
	commitCmd := fmt.Sprintf(CommandGitCommit, commitMsg)
	_, err = execCommand(commitCmd)
	if err != nil {
		return
	}

	if _, err = execCommand(CommandGitPush); err != nil {
		return
	}
}

func waitSignal(quitChan chan struct{}) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-quitChan:
		log.Printf("program exit quit")
	case si := <-c:
		log.Printf("program exit %s!", si)
	}
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
