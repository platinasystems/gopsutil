// +build darwin

package load

import (
	"context"
	"os/exec"
	"strconv"
	"strings"

	"github.com/platinasystems/gopsutil/internal/common"
)

func Avg() (*AvgStat, error) {
	return AvgWithContext(context.Background())
}

func AvgWithContext(ctx context.Context) (*AvgStat, error) {
	values, err := common.DoSysctrlWithContext(ctx, "vm.loadavg")
	if err != nil {
		return nil, err
	}

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
		Load1:  float64(load1),
		Load5:  float64(load5),
		Load15: float64(load15),
	}

	return ret, nil
}

// Misc returnes miscellaneous host-wide statistics.
// darwin use ps command to get process running/blocked count.
// Almost same as FreeBSD implementation, but state is different.
// U means 'Uninterruptible Sleep'.
func Misc() (*MiscStat, error) {
	return MiscWithContext(context.Background())
}

//From "ps" man page
//
//state     The state is given by a sequence of characters, for example, ``RWNA''.  The first character indicates the run state of the process:
//
//I       Marks a process that is idle (sleeping for longer than about 20 seconds).
//R       Marks a runnable process.
//S       Marks a process that is sleeping for less than about 20 seconds.
//T       Marks a stopped process.
//U       Marks a process in uninterruptible wait.
//Z       Marks a dead process (a ``zombie'').
//
//Additional characters after these, if any, indicate additional state information:
//
//+       The process is in the foreground process group of its control terminal.
//<       The process has raised CPU scheduling priority.
//>       The process has specified a soft limit on memory requirements and is currently exceeding that limit; such a process is (necessarily) not swapped.
//A       the process has asked for random page replacement (VA_ANOM, from vadvise(2), for example, lisp(1) in a garbage collect).
//E       The process is trying to exit.
//L       The process has pages locked in core (for example, for raw I/O).
//N       The process has reduced CPU scheduling priority (see setpriority(2)).
//S       The process has asked for FIFO page replacement (VA_SEQL, from vadvise(2), for example, a large image processing program using virtual memory to sequentially
//address voluminous data).
//s       The process is a session leader.
//V       The process is suspended during a vfork(2).
//W       The process is swapped out.
//X       The process is being traced or debugged.
//
func MiscWithContext(ctx context.Context) (*MiscStat, error) {
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
		if strings.HasPrefix(l, "R") {
			ret.ProcsRunning++
		} else if strings.HasPrefix(l, "U") {
			// uninterruptible sleep == blocked
			ret.ProcsBlocked++
			ret.ProcsStopped++
		} else if strings.HasPrefix(l, "Z") {
			ret.ProcsZombie++
		} else if strings.HasPrefix(l, "T") {
			ret.ProcsStopped++
		} else if strings.HasPrefix(l, "S") {
			ret.ProcsSleeping++
		}
	}

	return &ret, nil
}
