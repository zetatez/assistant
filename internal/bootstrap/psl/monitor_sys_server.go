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

func EnsureLocalSysServerRegistered(ctx context.Context) (string, error) {
	svrIP := getLocalIPv4()
	const dml = "INSERT IGNORE INTO sys_server (idc, svr_ip, ak, sk, svr_status, cpu_usage, mem_usage) VALUES ('', ?, '', '', 'ONLINE', 0, 0)"
	if _, err := GetDB().ExecContext(ctx, dml, svrIP); err != nil {
		return "", fmt.Errorf("register local sys_server '%s': %w", svrIP, err)
	}
	return svrIP, nil
}

func StartSysServerMonitor(ctx context.Context, svrIP string, interval time.Duration) {
	logger := GetLogger()
	if interval <= 0 {
		interval = 10 * time.Second
	}

	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				_, _ = GetDB().ExecContext(context.Background(), "UPDATE sys_server SET svr_status = 'OFFLINE' WHERE svr_ip = ? LIMIT 1", svrIP)
				return
			case <-ticker.C:
				cpuUsage := float64(0)
				if v, err := cpu.PercentWithContext(ctx, 0, false); err == nil && len(v) > 0 {
					cpuUsage = v[0]
				} else if err != nil {
					logger.Warnf("sys_server monitor: cpu percent failed: %v", err)
				}

				memUsage := float64(0)
				if vm, err := mem.VirtualMemoryWithContext(ctx); err == nil {
					memUsage = vm.UsedPercent
				} else {
					logger.Warnf("sys_server monitor: mem percent failed: %v", err)
				}

				_, err := GetDB().ExecContext(ctx, "UPDATE sys_server SET cpu_usage = ?, mem_usage = ?, svr_status = 'ONLINE' WHERE svr_ip = ? LIMIT 1", cpuUsage, memUsage, svrIP)
				if err != nil {
					logger.Warnf("sys_server monitor update failed (svr_ip=%s): %v", svrIP, err)
				}
			}
		}
	}()
}
