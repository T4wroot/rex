package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

func collectSysinfo() map[string]interface{} {
	info := map[string]interface{}{
		"os":   runtime.GOOS,
		"arch": runtime.GOARCH,
		"cpus": runtime.NumCPU(),
	}

	if mem, err := readMeminfo(); err == nil {
		info["mem_total_kb"] = mem["MemTotal"]
		info["mem_available_kb"] = mem["MemAvailable"]
		info["mem_free_kb"] = mem["MemFree"]
	}

	if uptime, err := readUptime(); err == nil {
		info["uptime_seconds"] = uptime
	}

	if load, err := readLoadAvg(); err == nil {
		info["load_1m"] = load[0]
		info["load_5m"] = load[1]
		info["load_15m"] = load[2]
	}

	return info
}

func readMeminfo() (map[string]int64, error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	result := make(map[string]int64)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) < 2 {
			continue
		}
		key := strings.TrimSuffix(parts[0], ":")
		val, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			continue
		}
		result[key] = val
	}
	return result, scanner.Err()
}

func readUptime() (float64, error) {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0, err
	}
	parts := strings.Fields(string(data))
	if len(parts) == 0 {
		return 0, fmt.Errorf("empty uptime")
	}
	return strconv.ParseFloat(parts[0], 64)
}

func readLoadAvg() ([]float64, error) {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return nil, err
	}
	parts := strings.Fields(string(data))
	if len(parts) < 3 {
		return nil, fmt.Errorf("unexpected loadavg format")
	}
	loads := make([]float64, 3)
	for i := 0; i < 3; i++ {
		v, err := strconv.ParseFloat(parts[i], 64)
		if err != nil {
			return nil, err
		}
		loads[i] = v
	}
	return loads, nil
}
