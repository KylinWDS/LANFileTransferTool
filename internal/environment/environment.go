package environment

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strings"
)

type CheckResult struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Details string `json:"details"`
}

type AllResults struct {
	Firewall  CheckResult `json:"firewall"`
	Network   CheckResult `json:"network"`
	Port      CheckResult `json:"port"`
	Solutions []string    `json:"solutions"`
}

type Checker struct{}

func NewChecker() *Checker {
	return &Checker{}
}

func (c *Checker) CheckAll() (*AllResults, error) {
	results := &AllResults{
		Firewall:  c.checkFirewall(),
		Network:   c.checkNetwork(),
		Port:      c.checkPort(8080),
		Solutions: []string{},
	}

	if results.Firewall.Status == "error" || results.Firewall.Status == "blocked" {
		results.Solutions = append(results.Solutions, "请检查防火墙设置，确保允许应用通过防火墙")
	}
	if results.Network.Status == "error" {
		results.Solutions = append(results.Solutions, "请检查网络连接，确保设备在同一局域网内")
	}
	if results.Port.Status == "error" {
		results.Solutions = append(results.Solutions, "端口被占用，请在设置中更换端口")
	}

	return results, nil
}

func (c *Checker) checkFirewall() CheckResult {
	switch runtime.GOOS {
	case "darwin":
		out, err := exec.Command("/usr/libexec/ApplicationFirewall/socketfilterfw", "--getglobalstate").Output()
		if err != nil {
			return CheckResult{Status: "ok", Message: "防火墙状态未知", Details: err.Error()}
		}
		s := strings.TrimSpace(string(out))
		if strings.Contains(s, "enabled") {
			return CheckResult{Status: "warning", Message: "防火墙已启用", Details: s}
		}
		return CheckResult{Status: "ok", Message: "防火墙未启用", Details: s}

	case "linux":
		out, err := exec.Command("ufw", "status").Output()
		if err == nil {
			s := strings.TrimSpace(string(out))
			if strings.Contains(s, "active") {
				return CheckResult{Status: "warning", Message: "防火墙已启用 (UFW)", Details: s}
			}
			return CheckResult{Status: "ok", Message: "防火墙未启用", Details: s}
		}
		return CheckResult{Status: "ok", Message: "无法检测防火墙状态", Details: ""}

	case "windows":
		out, err := exec.Command("netsh", "advfirewall", "show", "currentprofile", "state").Output()
		if err == nil {
			s := strings.TrimSpace(string(out))
			if strings.Contains(s, "ON") {
				return CheckResult{Status: "warning", Message: "防火墙已启用", Details: s}
			}
			return CheckResult{Status: "ok", Message: "防火墙未启用", Details: s}
		}
		return CheckResult{Status: "ok", Message: "无法检测防火墙状态", Details: ""}

	default:
		return CheckResult{Status: "ok", Message: "不支持的平台", Details: runtime.GOOS}
	}
}

func (c *Checker) checkNetwork() CheckResult {
	ifaces, err := net.Interfaces()
	if err != nil {
		return CheckResult{Status: "error", Message: "无法获取网络接口", Details: err.Error()}
	}

	var ips []string
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
				ips = append(ips, fmt.Sprintf("%s: %s", iface.Name, ipnet.IP.String()))
			}
		}
	}

	if len(ips) > 0 {
		return CheckResult{Status: "ok", Message: "网络正常", Details: strings.Join(ips, ", ")}
	}
	return CheckResult{Status: "error", Message: "未检测到活动网络", Details: ""}
}

func (c *Checker) checkPort(port int) CheckResult {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return CheckResult{Status: "warning", Message: fmt.Sprintf("端口 %d 已被占用", port), Details: err.Error()}
	}
	listener.Close()
	return CheckResult{Status: "ok", Message: fmt.Sprintf("端口 %d 可用", port), Details: ""}
}
