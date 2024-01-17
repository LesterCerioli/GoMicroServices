package internal

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/iox"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	cpuTicks  = 100
	cpuFields = 8
	cpuMax    = 1000
	statFile  = "/proc/stat"
)

var (
	preSystem uint64
	preTotal  uint64
	limit     float64
	cores     uint64
	initOnce  sync.Once
)

// if /proc not present, ignore the cpu calculation, like wsl linux
func initialize() {
	cpus, err := effectiveCpus()
	if err != nil {
		logx.Error(err)
		return
	}

	cores = uint64(cpus)
	limit = float64(cpus)
	quota, err := cpuQuota()
	if err == nil && quota > 0 {
		if quota < limit {
			limit = quota
		}
	}

	preSystem, err = systemCpuUsage()
	if err != nil {
		logx.Error(err)
		return
	}

	preTotal, err = cpuUsage()
	if err != nil {
		logx.Error(err)
		return
	}
}

// RefreshCpu refreshes cpu usage and returns.
func RefreshCpu() uint64 {
	initOnce.Do(initialize)

	total, err := cpuUsage()
	if err != nil {
		return 0
	}

	system, err := systemCpuUsage()
	if err != nil {
		return 0
	}

	var usage uint64
	cpuDelta := total - preTotal
	systemDelta := system - preSystem
	if cpuDelta > 0 && systemDelta > 0 {
		usage = uint64(float64(cpuDelta*cores*cpuMax) / (float64(systemDelta) * limit))
		if usage > cpuMax {
			usage = cpuMax
		}
	}
	preSystem = system
	preTotal = total

	return usage
}

func cpuQuota() (float64, error) {
	cg, err := currentCgroup()
	if err != nil {
		return 0, err
	}

	return cg.cpuQuota()
}

func cpuUsage() (usage uint64, err error) {
	var cg cgroup
	if cg, err = currentCgroup(); err != nil {
		return
	}

	return cg.cpuUsage()
}

func effectiveCpus() (int, error) {
	cg, err := currentCgroup()
	if err != nil {
		return 0, err
	}

	return cg.effectiveCpus()
}

func systemCpuUsage() (uint64, error) {
	lines, err := iox.ReadTextLines(statFile, iox.WithoutBlank())
	if err != nil {
		return 0, err
	}

	for _, line := range lines {
		fields := strings.Fields(line)
		if fields[0] == "cpu" {
			if len(fields) < cpuFields {
				return 0, fmt.Errorf("bad format of cpu stats")
			}

			var totalClockTicks uint64
			for _, i := range fields[1:cpuFields] {
				v, err := parseUint(i)
				if err != nil {
					return 0, err
				}

				totalClockTicks += v
			}

			return (totalClockTicks * uint64(time.Second)) / cpuTicks, nil
		}
	}

	return 0, errors.New("bad stats format")
}
