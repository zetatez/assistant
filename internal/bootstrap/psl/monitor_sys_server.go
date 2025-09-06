package psl

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

func getLocalIPv4() string {
	conn, err := net.DialTimeout("udp", "8.8.8.8:80", 1*time.Second)
	if err == nil {
		_ = conn.Close()
		if ua, ok := conn.LocalAddr().(*net.UDPAddr); ok {
			if ip := ua.IP.To4(); ip != nil {
				return ip.String()
			}
		}
	}

	ifAddrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, a := range ifAddrs {
			ipnet, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			if ipnet.IP == nil || ipnet.IP.IsLoopback() {
				continue
			}
			if ip := ipnet.IP.To4(); ip != nil {
				return ip.String()
			}
		}
	}

	return "127.0.0.1"
}

const registerSysServer = `
INSERT IGNORE INTO sys_server (idc, svr_ip, svr_status, cpu_usage, mem_usage)
VALUES ('', ?, ?, 0, 0)
`

func EnsureLocalSysServerRegistered(ctx context.Context) (string, error) {
	svrIP := getLocalIPv4()
	if _, err := GetDB().ExecContext(ctx, registerSysServer, svrIP, "ONLINE"); err != nil {
		return "", fmt.Errorf("register local sys_server '%s': %w", svrIP, err)
	}
	return svrIP, nil
}

const updateSysServerMetricsBySvrIP = `
UPDATE sys_server
SET cpu_usage = ?,
    mem_usage = ?,
    svr_status = ?
WHERE svr_ip = ?
LIMIT 1
`

const updateSysServerStatusBySvrIP = `
UPDATE sys_server
SET svr_status = ?
WHERE svr_ip = ?
LIMIT 1
`

func StartSysServerMonitor(ctx context.Context, svrIP string, interval time.Duration) context.CancelFunc {
	logger := GetLogger()
	if interval <= 0 {
		interval = 10 * time.Second
	}

	monitorCtx, cancel := context.WithCancel(ctx)
	registerCleanup(cancel)

	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-monitorCtx.Done():
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if GetDB() != nil {
					_, _ = GetDB().ExecContext(shutdownCtx, updateSysServerStatusBySvrIP, svrIP, "OFFLINE")
				}
				logger.Infof("[sys_server_monitor] stopped for %s", svrIP)
				return
			case <-ticker.C:
				if GetDB() == nil {
					continue
				}
				cpuUsage := float64(0)
				if v, err := cpu.PercentWithContext(monitorCtx, 0, false); err == nil && len(v) > 0 {
					cpuUsage = v[0]
				} else if err != nil {
					logger.Errorf("[sys_server_monitor] cpu percent failed: %v", err)
				}

				memUsage := float64(0)
				if vm, err := mem.VirtualMemoryWithContext(monitorCtx); err == nil {
					memUsage = vm.UsedPercent
				} else {
					logger.Errorf("[sys_server_monitor] mem percent failed: %v", err)
				}

				status := "ONLINE"
				_, err := GetDB().ExecContext(monitorCtx, updateSysServerMetricsBySvrIP, cpuUsage, memUsage, status, svrIP)
				if err != nil {
					logger.Errorf("[sys_server_monitor] update failed (svr_ip=%s): %v", svrIP, err)
				}
			}
		}
	}()

	logger.Infof("[sys_server_monitor] started for %s (interval=%v)", svrIP, interval)
	return cancel
}
