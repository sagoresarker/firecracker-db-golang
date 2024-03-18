package vm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

const (
	apiSocket      = "/tmp/firecracker.socket"
	logFile        = "./firecracker1.log"
	kernel         = "vmlinux-5.10.209"
	kernelBootArgs = "console=ttyS0 reboot=k panic=1 pci=off"
	rootFS         = "rootfs.ext4"
	fcMAC          = "06:00:AC:10:00:02"
	tapIP          = "172.16.0.2"
)

type logger struct {
	LogPath       string `json:"log_path"`
	Level         string `json:"level"`
	ShowLevel     bool   `json:"show_level"`
	ShowLogOrigin bool   `json:"show_log_origin"`
}

type machineConfig struct {
	MemSizeMiB uint `json:"mem_size_mib"`
	VCPUCount  uint `json:"vcpu_count"`
}

type bootSource struct {
	KernelImagePath string `json:"kernel_image_path"`
	BootArgs        string `json:"boot_args"`
}

type drive struct {
	DriveID      string `json:"drive_id"`
	PathOnHost   string `json:"path_on_host"`
	IsRootDevice bool   `json:"is_root_device"`
	IsReadOnly   bool   `json:"is_read_only"`
}

type networkInterface struct {
	IfaceID     string `json:"iface_id"`
	GuestMAC    string `json:"guest_mac"`
	HostDevName string `json:"host_dev_name"`
}

type action struct {
	ActionType string `json:"action_type"`
}

type instanceInfo struct {
	State string `json:"state"`
}

func waitForSignal(cmd *exec.Cmd) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received
	sig := <-signalChan

	fmt.Printf("\nReceived signal: %s\n", sig)

	// Terminate the Firecracker process gracefully
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		fmt.Printf("Failed to terminate Firecracker process: %v\n", err)
	}

	// Wait for the Firecracker process to exit
	if _, err := cmd.Process.Wait(); err != nil {
		fmt.Printf("Firecracker process exited with error: %v\n", err)
	}
}

func checkInstanceState() error {
	unixTransport := &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", apiSocket)
		},
	}

	client := &http.Client{
		Transport: unixTransport,
	}

	resp, err := client.Get("http://unix/machine-config")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("HTTP request failed with status code %d: %s", resp.StatusCode, body)
	}

	var info instanceInfo
	err = json.NewDecoder(resp.Body).Decode(&info)
	if err != nil {
		return err
	}

	if info.State != "Running" {
		return fmt.Errorf("microVM instance is not running (state: %s)", info.State)
	}

	fmt.Println("microVM instance is running")
	return nil
}

func createSocket() error {
	apiSocket := "/tmp/firecracker.socket"

	// Remove existing socket file
	err := os.RemoveAll(apiSocket)
	if err != nil {
		return err
	}

	// Run the firecracker command
	cmd := exec.Command("sudo", "firecracker/firecracker", "--api-sock", apiSocket)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		return err
	}

	// Wait for the socket file to be created
	for {
		_, err = os.Stat(apiSocket)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Start a goroutine to handle signals
	go waitForSignal(cmd)

	return nil
}

func Vmlaunch() {

	err := createSocket()
	if err != nil {
		fmt.Println("Error creating socket and starting firecracker:", err)
		return
	}

	// Create log file
	file, err := os.Create(logFile)
	if err != nil {
		fmt.Println("Error creating log file:", err)
		return
	}
	file.Close()

	// Set log file
	err = sendRequest("PUT", "http://unix/logger", &logger{
		LogPath:       logFile,
		Level:         "Debug",
		ShowLevel:     true,
		ShowLogOrigin: true,
	})
	if err != nil {
		fmt.Println("Error setting log file:", err)
		return
	}

	// Machine config
	err = sendRequest("PUT", "http://unix/machine-config", &machineConfig{
		MemSizeMiB: 2048,
		VCPUCount:  1,
	})
	if err != nil {
		fmt.Println("Error setting machine config:", err)
		return
	}

	// Set boot source
	err = sendRequest("PUT", "http://unix/boot-source", &bootSource{
		KernelImagePath: kernel,
		BootArgs:        kernelBootArgs,
	})
	if err != nil {
		fmt.Println("Error setting boot source:", err)
		return
	}

	// Set rootfs
	err = sendRequest("PUT", "http://unix/drives/rootfs", &drive{
		DriveID:      "rootfs",
		PathOnHost:   rootFS,
		IsRootDevice: true,
		IsReadOnly:   false,
	})
	if err != nil {
		fmt.Println("Error setting rootfs:", err)
		return
	}

	// Set network interface
	tapDev := os.Getenv("TAP_DEV")
	err = sendRequest("PUT", "http://unix/network-interfaces/eth0", &networkInterface{
		IfaceID:     "eth0",
		GuestMAC:    fcMAC,
		HostDevName: tapDev,
	})
	if err != nil {
		fmt.Println("Error setting network interface:", err)
		return
	}

	// Sleep to allow configuration to be set
	time.Sleep(15 * time.Millisecond)

	// Start microVM
	err = sendRequest("PUT", "http://unix/actions", &action{
		ActionType: "InstanceStart",
	})
	if err != nil {
		fmt.Println("Error starting microVM:", err)
		return
	}

	// Sleep to allow microVM to start
	time.Sleep(15 * time.Millisecond)

	// Check if the instance is running
	err = checkInstanceState()
	if err != nil {
		fmt.Println(err)
		return
	}
	// SSH into the microVM
	keyPath, _ := filepath.Abs("./key_pairs/id_rsa")
	fmt.Println("To ssh into the microVM run: ssh -i", keyPath, "root@"+tapIP)
	fmt.Println("Use 'root' for both the login and password. Run 'reboot' to exit.")
}

func sendRequest(method, url string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Create a Unix domain socket transport
	unixTransport := &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", apiSocket)
		},
	}

	// Create a custom HTTP client using the Unix domain socket transport
	client := &http.Client{
		Transport: unixTransport,
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		// Treat 204 No Content as a success
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("HTTP request failed with status code %d: %s", resp.StatusCode, body)
	}

	return nil
}
