package networking

import (
	"fmt"
	"math/rand"
	"net"
	"os/exec"
	"time"

	"github.com/sagoresarker/firecracker-db-golang/database"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateUserID() string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 5)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func generateIPAddress(startRange, endRange string) (string, error) {
	// Parse start and end IP addresses
	startIP := net.ParseIP(startRange).To4()
	endIP := net.ParseIP(endRange).To4()

	if startIP == nil || endIP == nil {
		return "", fmt.Errorf("invalid IP address range")
	}

	// Convert IP addresses to integers
	start := int(startIP[0])<<24 | int(startIP[1])<<16 | int(startIP[2])<<8 | int(startIP[3])
	end := int(endIP[0])<<24 | int(endIP[1])<<16 | int(endIP[2])<<8 | int(endIP[3])

	// Generate a random IP address within the range
	randomIP := make(net.IP, 4)
	ipInt := rand.Intn(end-start+1) + start
	randomIP[0] = byte(ipInt >> 24 & 0xFF)
	randomIP[1] = byte(ipInt >> 16 & 0xFF)
	randomIP[2] = byte(ipInt >> 8 & 0xFF)
	randomIP[3] = byte(ipInt & 0xFF)

	return randomIP.String(), nil
}

func generateValue() (bridgeName string, userID string, ipAddress string) {
	fmt.Println("Generate a value for bridge-name, user-id and ip-address")

	startRange := "10.0.0.0"
	endRange := "10.255.255.255"

	userID = generateUserID()
	bridgeName = "br-" + userID

	ip, err := generateIPAddress(startRange, endRange)

	if err != nil {
		fmt.Println("Error Generating IP adress:", err)
		return
	}

	ipAddress = ip + "/24"

	return bridgeName, userID, ipAddress

}

func createBridge(bridgeName string, ipAddress string) error {

	cmd := exec.Command("sudo", "ip", "link", "add", "name", bridgeName, "type", "bridge")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create bridge: %v", err)
	}

	cmd = exec.Command("sudo", "ip", "addr", "add", ipAddress, "dev", bridgeName)
	if err := cmd.Run(); err != nil {
		// If assigning IP address fails, we need to delete the bridge
		cmd := exec.Command("sudo", "ip", "link", "delete", bridgeName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to delete bridge after IP assignment failure: %v", err)
		}
		return fmt.Errorf("failed to assign IP address to bridge: %v", err)
	}

	fmt.Printf("Bridge %s created and assigned IP Address %s\n", bridgeName, ipAddress)

	return nil
}

func SetupBridgeNetwork() {
	fmt.Println("Setting up bridge")

	bridgeName, userID, ipAddress := generateValue()

	fmt.Println("Bridge Name:", bridgeName)
	fmt.Println("User ID:", userID)
	fmt.Println("IP Address:", ipAddress)

	if err := createBridge(bridgeName, ipAddress); err != nil {
		fmt.Println("Error creating bridge:", err)
		return
	}
	database.SaveBridgeDetails(bridgeName, userID, ipAddress)
}
