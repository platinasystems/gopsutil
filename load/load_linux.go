// +build linux

package load

import (
	"context"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/gregorycallea/gopsutil/internal/common"
)

func Avg() (*AvgStat, error) {
	return AvgWithContext(context.Background())
}

func AvgWithContext(ctx context.Context) (*AvgStat, error) {
	filename := common.HostProc("loadavg")
	line, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	values := strings.Fields(string(line))

	load1, err := strconv.ParseFloat(values[0], 64)
	if err != nil {
		return nil, err
	}
	load5, err := strconv.ParseFloat(values[1], 64)
	if err != nil {
		return nil, err
	}
	load15, err := strconv.ParseFloat(values[2], 64)
	if err != nil {
		return nil, err
	}

	ret := &AvgStat{
		Load1:  load1,
		Load5:  load5,
		Load15: load15,
	}

	return ret, nil
}

// Misc returnes miscellaneous host-wide statistics.
// Note: the name should be changed near future.
func Misc() (*MiscStat, error) {
	return MiscWithContext(context.Background())
}

//From "ps" man page
//PROCESS STATE CODES
//       Here are the different values that the s, stat and state output specifiers (header "STAT" or "S") will display to describe the state of a process:
//
//               D    uninterruptible sleep (usually IO)
//               R    running or runnable (on run queue)
//               S    interruptible sleep (waiting for an event to complete)
//               T    stopped by job control signal
//               t    stopped by debugger during the tracing
//               W    paging (not valid since the 2.6.xx kernel)
//               X    dead (should never be seen)
//               Z    defunct ("zombie") process, terminated but not reaped by its parent
//
//       For BSD formats and when the stat keyword is used, additional characters may be displayed:
//
//               <    high-priority (not nice to other users)
//               N    low-priority (nice to other users)
//               L    has pages locked into memory (for real-time and custom IO)
//               s    is a session leader
//               l    is multi-threaded (using CLONE_THREAD, like NPTL pthreads do)
//               +    is in the foreground process group
func MiscWithContext(ctx context.Context) (*MiscStat, error) {
	filename := common.HostProc("stat")
	out, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	ret := &MiscStat{}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		v, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			continue
		}
		switch fields[0] {
		case "procs_running":
			ret.ProcsRunning = int(v)
		case "procs_blocked":
			ret.ProcsBlocked = int(v)
		case "ctxt":
			ret.Ctxt = int(v)
		default:
			continue
		}

	}

	bin, err := exec.LookPath("ps")
	if err != nil {
		return nil, err
	}
	out, err := invoke.CommandWithContext(ctx, bin, "axo", "state")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(out), "\n")

	ret := MiscStat{}
	for _, l := range lines {
		if l == "" {
			continue
		}
		ret.ProcsTotal++
		if strings.HasPrefix(l, "Z") {
			ret.ProcsZombie++
		} else if strings.HasPrefix(l, "T") || strings.HasPrefix(l, "t") {
			ret.ProcsStopped++
		} else if strings.HasPrefix(l, "S") {
			ret.ProcsSleeping++
		}
	}

	return ret, nil
}
