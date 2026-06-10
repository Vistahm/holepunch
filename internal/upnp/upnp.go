package upnp

import (
	"fmt"
	"net"
	"time"

	"github.com/huin/goupnp/dcps/internetgateway1"
	"github.com/huin/goupnp/dcps/internetgateway2"
)

var (
	cleanupFunc func()
	usedPort    int
)

// ForwardPort opens a port via UPnP. Returns the external IP.
// It automatically handles stale mappings from previous runs.
func ForwardPort(port int) (string, error) {
	usedPort = port

	localIP, err := getLocalIP()
	if err != nil {
		return "", fmt.Errorf("get local IP: %w", err)
	}

	// Try IGDv1 first
	clients1, _, err1 := internetgateway1.NewWANIPConnection1Clients()
	if err1 == nil && len(clients1) > 0 {
		client := clients1[0]
		externalIP, _ := client.GetExternalIPAddress()

		// Aggressively clean up: try multiple times
		for i := 0; i < 3; i++ {
			client.DeletePortMapping("", uint16(port), "TCP")
			time.Sleep(300 * time.Millisecond)
		}

		// Now add the mapping
		err := client.AddPortMapping("", uint16(port), "TCP", uint16(port), localIP, true, "HolePunch", 0)
		if err != nil {
			// If still failing, wait longer and retry once more
			time.Sleep(2 * time.Second)
			client.DeletePortMapping("", uint16(port), "TCP")
			time.Sleep(500 * time.Millisecond)
			err = client.AddPortMapping("", uint16(port), "TCP", uint16(port), localIP, true, "HolePunch", 0)
			if err != nil {
				return "", fmt.Errorf("add port mapping (IGDv1): %w", err)
			}
		}

		cleanupFunc = func() {
			fmt.Println("\nRemoving port mapping...")
			client.DeletePortMapping("", uint16(port), "TCP")
			fmt.Println("Port mapping removed.")
		}

		return externalIP, nil
	}

	// Try IGDv2
	clients2, _, err2 := internetgateway2.NewWANIPConnection2Clients()
	if err2 == nil && len(clients2) > 0 {
		client := clients2[0]
		externalIP, _ := client.GetExternalIPAddress()

		// Aggressive cleanup
		for i := 0; i < 3; i++ {
			client.DeletePortMapping("", uint16(port), "TCP")
			time.Sleep(300 * time.Millisecond)
		}

		err := client.AddPortMapping("", uint16(port), "TCP", uint16(port), localIP, true, "HolePunch", 0)
		if err != nil {
			time.Sleep(2 * time.Second)
			client.DeletePortMapping("", uint16(port), "TCP")
			time.Sleep(500 * time.Millisecond)
			err = client.AddPortMapping("", uint16(port), "TCP", uint16(port), localIP, true, "HolePunch", 0)
			if err != nil {
				return "", fmt.Errorf("add port mapping (IGDv2): %w", err)
			}
		}

		cleanupFunc = func() {
			fmt.Println("\nRemoving port mapping...")
			client.DeletePortMapping("", uint16(port), "TCP")
			fmt.Println("Port mapping removed.")
		}

		return externalIP, nil
	}

	return "", fmt.Errorf("no UPnP gateway found")
}

// Cleanup removes the port mapping.
func Cleanup(port int) {
	if cleanupFunc != nil {
		cleanupFunc()
		cleanupFunc = nil
	}
}

func getLocalIP() (string, error) {
	conn, err := net.Dial("udp", "192.0.2.0:1")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String(), nil
}
