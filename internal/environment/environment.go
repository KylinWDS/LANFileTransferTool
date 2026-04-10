package environment

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

type Checker struct {
	mu sync.RWMutex
}

type CheckResult struct {
	Firewall CheckItem
	Network  CheckItem
	Port     CheckItem
	Solutions []string
}

type CheckItem struct {
	Status  string
	Message string
	Details string
}

func NewChecker() *Checker {
	return &Checker{}
}

func (c *Checker) CheckAll() (*CheckResult, error) {
	result := &CheckResult{}

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		result.Firewall = c.checkFirewall()
	}()

	go func() {
		defer wg.Done()
		result.Network = c.checkNetwork()
	}()

	go func() {
		defer wg.Done()
		result.Port = c.checkPort(8080)
	}()

	wg.Wait()

	result.Solutions = c.generateSolutions(result)
	return result, nil
}

func (c *Checker) checkFirewall() CheckItem {
	switch runtime.GOOS {
	case "windows":
		return c.checkWindowsFirewall()
	case "darwin":
		return c.checkMacOSFirewall()
	case "linux":
		return c.checkLinuxFirewall()
	default:
		return CheckItem{
			Status:  "unknown",
			Message: "无法检测防火墙状态",
			Details: "不支持的操作系统",
		}
	}
}

func (c *Checker) checkWindowsFirewall() CheckItem {
	cmd := exec.Command("netsh", "advfirewall", "show", "allprofiles", "state")
	output, err := cmd.Output()
	if err != nil {
		return CheckItem{
			Status:  "error",
			Message: "无法获取防火墙状态",
			Details: err.Error(),
		}
	}

	status := strings.ToLower(string(output))
	if strings.Contains(status, "on") {
		return CheckItem{
			Status:  "warning",
			Message: "Windows防火墙已开启",
			Details: "可能需要添加入站规则以允许应用通信",
		}
	}

	return CheckItem{
		Status:  "ok",
		Message: "Windows防火墙已关闭",
		Details: "网络连接不受限制",
	}
}

func (c *Checker) checkMacOSFirewall() CheckItem {
	cmd := exec.Command("/usr/libexec/ApplicationFirewall/socketfilterfw", "--getglobalstate")
	output, err := cmd.Output()
	if err != nil {
		return CheckItem{
			Status:  "error",
			Message: "无法获取防火墙状态",
			Details: err.Error(),
		}
	}

	status := strings.TrimSpace(string(output))
	if status == "true" {
		return CheckItem{
			Status:  "warning",
			Message: "macOS防火墙已开启",
			Details: "可能需要在系统偏好设置中允许应用",
		}
	}

	return CheckItem{
		Status:  "ok",
		Message: "macOS防火墙已关闭",
		Details: "网络连接不受限制",
	}
}

func (c *Checker) checkLinuxFirewall() CheckItem {
	cmds := []struct {
		name string
		args []string
	}{
		{"ufw", []string{"status"}},
		{"firewall-cmd", []string{"--state"}},
		{"iptables", []string{"-L", "-n"}},
	}

	for _, cmd := range cmds {
		output, err := exec.Command(cmd.name, cmd.args...).Output()
		if err != nil {
			continue
		}

		status := strings.ToLower(string(output))
		if strings.Contains(status, "active") || strings.Contains(status, "enabled") {
			return CheckItem{
				Status:  "warning",
				Message: "Linux防火墙已开启",
				Details: fmt.Sprintf("使用 %s 管理", cmd.name),
			}
		}
	}

	return CheckItem{
		Status:  "ok",
		Message: "未检测到Linux防火墙",
		Details: "网络连接不受限制",
	}
}

func (c *Checker) checkNetwork() CheckItem {
	conns, err := net.Interfaces()
	if err != nil {
		return CheckItem{
			Status:  "error",
			Message: "无法检测网络接口",
			Details: err.Error(),
		}
	}

	hasActiveInterface := false
	var activeInterfaces []string

	for _, conn := range conns {
		if conn.Flags&net.FlagUp != 0 && conn.Flags&net.FlagLoopback == 0 {
			addrs, err := conn.Addrs()
			if err == nil && len(addrs) > 0 {
				hasActiveInterface = true
				activeInterfaces = append(activeInterfaces, conn.Name)
			}
		}
	}

	if hasActiveInterface {
		return CheckItem{
			Status:  "ok",
			Message: "网络连接正常",
			Details: fmt.Sprintf("活动接口: %s", strings.Join(activeInterfaces, ", ")),
		}
	}

	return CheckItem{
		Status:  "error",
		Message: "未检测到活动网络接口",
		Details: "请检查网络连接",
	}
}

func (c *Checker) checkPort(port int) CheckItem {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return CheckItem{
			Status:  "error",
			Message: fmt.Sprintf("端口 %d 已被占用", port),
			Details: "请更换端口或停止占用端口的进程",
		}
	}
	listener.Close()

	return CheckItem{
		Status:  "ok",
		Message: fmt.Sprintf("端口 %d 可用", port),
		Details: "可以正常使用该端口启动服务",
	}
}

func (c *Checker) generateSolutions(result *CheckResult) []string {
	var solutions []string

	if result.Firewall.Status == "warning" || result.Firewall.Status == "error" {
		solutions = append(solutions, "1. 配置防火墙规则，允许应用端口（默认8080）的入站连接")
	}

	if result.Network.Status == "error" {
		solutions = append(solutions, "2. 检查网络电缆或WiFi连接")
		solutions = append(solutions, "3. 确保网络适配器已启用")
	}

	if result.Port.Status == "error" {
		solutions = append(solutions, "4. 更换应用端口或在设置中修改端口号")
		solutions = append(solutions, "5. 查找并结束占用端口的进程")
	}

	if len(solutions) == 0 {
		solutions = append(solutions, "✓ 环境检测通过，所有项目正常")
	}

	return solutions
}
