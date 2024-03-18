package networking

import (
	"fmt"
	"os/exec"
)

func createTap(tapName string, bridgeName string) error {

	cmd := exec.Command("sudo", "ip", "tuntap", "add", "dev", tapName, "mode", "tap")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create tap: %v", err)
	}

	cmd = exec.Command("sudo", "ip", "link", "set", "dev", tapName, "up")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to bring up tap: %v", err)
	}

	cmd = exec.Command("sudo", "ip", "link", "set", "dev", tapName, "master", bridgeName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to assign tap to bridge: %v", err)
	}

	fmt.Printf("Tap %s assigned to Bridge Address %s\n", tapName, bridgeName)

	return nil
}

func SetupTapNetwork(bridgeName string) (string, error) {
	fmt.Println("Setting up tap")

	tapName := "tap-" + bridgeName

	if err := createTap(tapName, bridgeName); err != nil {
		fmt.Println("Error creating tap:", err)
		return "", err
	}
	return tapName, nil
}
