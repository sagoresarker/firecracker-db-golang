package vm

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"net"
// 	"net/http"
// 	"os"
// 	"os/exec"
// 	"path/filepath"
// 	"time"
// )

// const (
// 	apiSocket      = "/tmp/firecracker1.socket"
// 	logFile        = "./firecracker1.log"
// 	kernel         = "files/vmlinux-5.10.209"
// 	kernelBootArgs = "console=ttyS0 reboot=k panic=1 pci=off"
// 	rootFS         = "files/rootfs.ext4"
// 	fcMAC          = "06:00:AC:10:00:02"
// 	tapIP          = "172.16.0.2"
// )

// type logger struct {
// 	LogPath       string `json:"log_path"`
// 	Level         string `json:"level"`
// 	ShowLevel     bool   `json:"show_level"`
// 	ShowLogOrigin bool   `json:"show_log_origin"`
// }

// type machineConfig struct {
// 	MemSizeMiB uint `json:"mem_size_mib"`
// 	VCPUCount  uint `json:"vcpu_count"`
// }

// type bootSource struct {
// 	KernelImagePath string `json:"kernel_image_path"`
// 	BootArgs        string `json:"boot_args"`
// }

// type drive struct {
// 	DriveID      string `json:"drive_id"`
// 	PathOnHost   string `json:"path_on_host"`
// 	IsRootDevice bool   `json:"is_root_device"`
// 	IsReadOnly   bool   `json:"is_read_only"`
// }

// type networkInterface struct {
// 	IfaceID     string `json:"iface_id"`
// 	GuestMAC    string `json:"guest_mac"`
// 	HostDevName string `json:"host_dev_name"`
// }

// type action struct {
// 	ActionType string `json:"action_type"`
// }

// func createSocket() error {
// 	apiSocket := "/tmp/firecracker1.socket"

// 	// Remove existing socket file
// 	err := os.RemoveAll(apiSocket)
// 	if err != nil {
// 		return err
// 	}

// 	// Run the firecracker command
// 	cmd := exec.Command("sudo", "./firecracker", "--api-sock", apiSocket)
// 	err = cmd.Start()
// 	if err != nil {
// 		return err
// 	}

// 	// Wait for the firecracker process to start
// 	time.Sleep(1 * time.Second)

// 	return nil
// }

// func Vmlaunch() {
// 	// Create log file
// 	file, err := os.Create(logFile)
// 	if err != nil {
// 		fmt.Println("Error creating log file:", err)
// 		return
// 	}
// 	file.Close()

// 	// Set log file
// 	err = sendRequest("PUT", "http://localhost/logger", &logger{
// 		LogPath:       logFile,
// 		Level:         "Debug",
// 		ShowLevel:     true,
// 		ShowLogOrigin: true,
// 	})
// 	if err != nil {
// 		fmt.Println("Error setting log file:", err)
// 		return
// 	}

// 	// Machine config
// 	err = sendRequest("PUT", "http://localhost/machine-config", &machineConfig{
// 		MemSizeMiB: 2048,
// 		VCPUCount:  1,
// 	})
// 	if err != nil {
// 		fmt.Println("Error setting machine config:", err)
// 		return
// 	}

// 	// Set boot source
// 	err = sendRequest("PUT", "http://localhost/boot-source", &bootSource{
// 		KernelImagePath: kernel,
// 		BootArgs:        kernelBootArgs,
// 	})
// 	if err != nil {
// 		fmt.Println("Error setting boot source:", err)
// 		return
// 	}

// 	// Set rootfs
// 	err = sendRequest("PUT", "http://localhost/drives/rootfs", &drive{
// 		DriveID:      "rootfs",
// 		PathOnHost:   rootFS,
// 		IsRootDevice: true,
// 		IsReadOnly:   false,
// 	})
// 	if err != nil {
// 		fmt.Println("Error setting rootfs:", err)
// 		return
// 	}

// 	// Set network interface
// 	tapDev := os.Getenv("TAP_DEV")
// 	err = sendRequest("PUT", "http://localhost/network-interfaces/eth0", &networkInterface{
// 		IfaceID:     "eth0",
// 		GuestMAC:    fcMAC,
// 		HostDevName: tapDev,
// 	})
// 	if err != nil {
// 		fmt.Println("Error setting network interface:", err)
// 		return
// 	}

// 	// Sleep to allow configuration to be set
// 	time.Sleep(15 * time.Millisecond)

// 	// Start microVM
// 	err = sendRequest("PUT", "http://localhost/actions", &action{
// 		ActionType: "InstanceStart",
// 	})
// 	if err != nil {
// 		fmt.Println("Error starting microVM:", err)
// 		return
// 	}

// 	// Sleep to allow microVM to start
// 	time.Sleep(15 * time.Millisecond)

// 	// SSH into the microVM
// 	keyPath, _ := filepath.Abs("./key_pairs/id_rsa")
// 	fmt.Println("To ssh into the microVM run: ssh -i", keyPath, "root@"+tapIP)
// 	fmt.Println("Use 'root' for both the login and password. Run 'reboot' to exit.")
// }

// func sendRequest(method, url string, payload interface{}) error {
// 	data, err := json.Marshal(payload)
// 	if err != nil {
// 		return err
// 	}

// 	conn, err := net.Dial("unix", apiSocket)
// 	if err != nil {
// 		return err
// 	}
// 	defer conn.Close()

// 	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
// 	if err != nil {
// 		return err
// 	}

// 	resp, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		body, _ := ioutil.ReadAll(resp.Body)
// 		return fmt.Errorf("HTTP request failed with status code %d: %s", resp.StatusCode, body)
// 	}

// 	return nil
// }
