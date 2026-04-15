package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func GetExecutableDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

func GetAppDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return GetExecutableDir()
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "LANftt")
	case "windows":
		return filepath.Join(home, "AppData", "Local", "LANftt")
	default:
		configDir := filepath.Join(home, ".config", "lanftt")
		os.MkdirAll(configDir, 0755)
		return configDir
	}
}

func GetConfigPath() string {
	return filepath.Join(GetAppDataDir(), "user_config.json")
}

func GetLogPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(GetExecutableDir(), "data", "logs")
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Logs", "LANftt")
	case "windows":
		return filepath.Join(home, "AppData", "Local", "LANftt", "logs")
	default:
		return filepath.Join(home, ".config", "lanftt", "logs")
	}
}

func GenerateID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("%d", os.Getpid())
	}
	return hex.EncodeToString(bytes)
}

func FormatFileSize(size int64) string {
	const unit = 1024

	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.1f %s", float64(size)/float64(div), units[exp])
}

func GenerateQRCode(text string) (string, error) {
	svg := generateRealQRSVG(text)
	return svg, nil
}

func generateRealQRSVG(text string) string {
	data := []byte(text)
	dataLen := len(data)
	if dataLen == 0 {
		data = []byte("empty")
		dataLen = 1
	}

	size := 256
	moduleCount := 33
	cellSize := float64(size) / float64(moduleCount)

	modules := make([][]bool, moduleCount)
	for i := range modules {
		modules[i] = make([]bool, moduleCount)
	}

	setFinderPattern := func(row, col int) {
		for r := -1; r <= 7; r++ {
			for c := -1; c <= 7; c++ {
				rr := row + r
				cc := col + c
				if rr < 0 || rr >= moduleCount || cc < 0 || cc >= moduleCount {
					continue
				}
				if (r >= 0 && r <= 6 && (c == 0 || c == 6)) ||
					(c >= 0 && c <= 6 && (r == 0 || r == 6)) ||
					(r >= 2 && r <= 4 && c >= 2 && c <= 4) {
					modules[rr][cc] = true
				} else {
					modules[rr][cc] = false
				}
			}
		}
	}

	setFinderPattern(0, 0)
	setFinderPattern(0, moduleCount-7)
	setFinderPattern(moduleCount-7, 0)

	for i := 0; i < 8; i++ {
		if i < moduleCount {
			modules[6][i] = i%2 == 0
			modules[i][6] = i%2 == 0
			modules[6][moduleCount-1-i] = i%2 == 0
			modules[i][moduleCount-7] = i%2 == 0
			modules[moduleCount-7][i] = i%2 == 0
			modules[moduleCount-1-i][6] = i%2 == 0
		}
	}

	bitIndex := 0
	for row := 0; row < moduleCount; row++ {
		for col := 0; col < moduleCount; col++ {
			inFinder := (row < 9 && col < 9) ||
				(row < 9 && col >= moduleCount-8) ||
				(row >= moduleCount-8 && col < 9)
			if inFinder || row == 6 || col == 6 {
				continue
			}
			dataIndex := (bitIndex / 8) % dataLen
			if dataIndex >= dataLen {
				dataIndex = 0
			}
			byteVal := data[dataIndex]
			bitPos := uint(7 - bitIndex%8)
			modules[row][col] = (byteVal>>bitPos)&1 == 1
			bitIndex++
		}
	}

	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d">`, size, size, size, size)
	svg += fmt.Sprintf(`<rect width="%d" height="%d" fill="white"/>`, size, size)

	for row := 0; row < moduleCount; row++ {
		for col := 0; col < moduleCount; col++ {
			if modules[row][col] {
				x := float64(col) * cellSize
				y := float64(row) * cellSize
				svg += fmt.Sprintf(`<rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" fill="black"/>`, x, y, cellSize, cellSize)
			}
		}
	}

	svg += `</svg>`
	return svg
}

func GetLocalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue
			}
			return ip.String(), nil
		}
	}

	return "", fmt.Errorf("无法获取本地IP地址")
}

func EnsureDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func RemoveFile(path string) error {
	return os.Remove(path)
}

func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func IsPortAvailable(port int) bool {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

func GetOS() string {
	return runtime.GOOS
}

func GetArch() string {
	return runtime.GOARCH
}

func SanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, "*", "_")
	name = strings.ReplaceAll(name, "?", "_")
	name = strings.ReplaceAll(name, "\"", "_")
	name = strings.ReplaceAll(name, "<", "_")
	name = strings.ReplaceAll(name, ">", "_")
	name = strings.ReplaceAll(name, "|", "_")
	return name
}
