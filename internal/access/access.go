package access

import (
	"net"
	"sync"

	"lanfiletransfertool/pkg/errors"
)

type Control struct {
	whitelist []string
	blacklist []string
	mu        sync.RWMutex
}

func NewControl(whitelist, blacklist []string) *Control {
	return &Control{
		whitelist: whitelist,
		blacklist: blacklist,
	}
}

func (c *Control) AllowAccess(clientIP string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.blacklist) > 0 {
		for _, blockedIP := range c.blacklist {
			if c.ipMatch(clientIP, blockedIP) {
				return false
			}
		}
	}

	if len(c.whitelist) > 0 {
		for _, allowedIP := range c.whitelist {
			if c.ipMatch(clientIP, allowedIP) {
				return true
			}
		}
		return false
	}

	return true
}

func (c *Control) AddToWhitelist(ip string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, existingIP := range c.whitelist {
		if existingIP == ip {
			return errors.ErrInvalidParameter
		}
	}

	c.whitelist = append(c.whitelist, ip)
	return nil
}

func (c *Control) RemoveFromWhitelist(ip string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, existingIP := range c.whitelist {
		if existingIP == ip {
			c.whitelist = append(c.whitelist[:i], c.whitelist[i+1:]...)
			return nil
		}
	}

	return errors.ErrFileNotFound
}

func (c *Control) AddToBlacklist(ip string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, existingIP := range c.blacklist {
		if existingIP == ip {
			return errors.ErrInvalidParameter
		}
	}

	c.blacklist = append(c.blacklist, ip)
	return nil
}

func (c *Control) RemoveFromBlacklist(ip string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, existingIP := range c.blacklist {
		if existingIP == ip {
			c.blacklist = append(c.blacklist[:i], c.blacklist[i+1:]...)
			return nil
		}
	}

	return errors.ErrFileNotFound
}

func (c *Control) GetWhitelist() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.whitelist
}

func (c *Control) GetBlacklist() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.blacklist
}

func (c *Control) SetWhitelist(ips []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.whitelist = ips
}

func (c *Control) SetBlacklist(ips []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.blacklist = ips
}

func (c *Control) ipMatch(clientIP, pattern string) bool {
	clientAddr := net.ParseIP(clientIP)
	patternAddr := net.ParseIP(pattern)

	if clientAddr != nil && patternAddr != nil {
		return clientAddr.Equal(patternAddr)
	}

	_, ipNet, err := net.ParseCIDR(pattern)
	if err == nil && ipNet != nil {
		return ipNet.Contains(clientAddr)
	}

	return clientIP == pattern
}

func (c *Control) IsIPBlocked(ip string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, blockedIP := range c.blacklist {
		if c.ipMatch(ip, blockedIP) {
			return true
		}
	}
	return false
}

func (c *Control) IsIPAllowed(ip string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.whitelist) == 0 {
		return !c.IsIPBlocked(ip)
	}

	for _, allowedIP := range c.whitelist {
		if c.ipMatch(ip, allowedIP) {
			return true
		}
	}
	return false
}
