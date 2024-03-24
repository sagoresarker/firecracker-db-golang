package vm

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	models "github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	"github.com/sagoresarker/firecracker-db-golang/networking"
	"github.com/sirupsen/logrus"
)

func LaunchFirstVM(tapName1 string, tapName2 string) {

	// Read the startup script from a file
	// startupScriptPath := "startup-script/startup-script-vm1.sh"
	// vm1_startupScript, err := ioutil.ReadFile(startupScriptPath)
	// if err != nil {
	// 	fmt.Printf("Failed to read vm1 startup script: %v\n", err)
	// 	return
	// }

	bridge_ip_address, _ := networking.GetBridgeIPAddress()

	bridge_ip_without_mask, _, err := net.ParseCIDR(bridge_ip_address)
	if err != nil {
		fmt.Println("Error parsing bridge IP address:", err)
		return
	}
	fmt.Println("(Launch VM) - Bridge IP without mask:", bridge_ip_without_mask)

	vm1_eth0_ip, _, err := networking.GetVMIPs(bridge_ip_without_mask.String())

	if err != nil {
		fmt.Println("Error getting VM IPs:", err)
		return
	}

	fmt.Printf("VM1 IP: %s\n", vm1_eth0_ip)
	vm1_eth0_ip_ipv4 := net.ParseIP(vm1_eth0_ip)
	if vm1_eth0_ip_ipv4 == nil {
		fmt.Println("Error parsing VM1 IP address")
		return
	}

	// script := fmt.Sprintf(`#!/bin/bash
	// ip link set eth0 up
	// ip addr add %s/24 dev eth0
	// ip route add default via %s dev eth0
	// `, vm1_eth0_ip, bridge_ip_address)

	cfg1 := firecracker.Config{
		SocketPath:      "/tmp/firecracker1.sock",
		LogFifo:         "/tmp/firecracker1.log",
		MetricsFifo:     "/tmp/firecracker1-metrics",
		LogLevel:        "Debug",
		KernelImagePath: "files/vmlinux",
		//KernelArgs:      fmt.Sprintf("ro console=ttyS0 reboot=k panic=1 pci=off %s", script),
		KernelArgs: "ro console=tty0 console=ttyS0 reboot=k panic=1 pci=off",

		MachineCfg: models.MachineConfiguration{
			VcpuCount:  firecracker.Int64(2),
			MemSizeMib: firecracker.Int64(256),
			Smt:        firecracker.Bool(false),
		},
		Drives: []models.Drive{
			{
				DriveID:      firecracker.String("1"),
				IsRootDevice: firecracker.Bool(true),
				IsReadOnly:   firecracker.Bool(false),
				PathOnHost:   firecracker.String("files/root-drive-with-ssh.img"),
			},
		},
		NetworkInterfaces: []firecracker.NetworkInterface{
			{
				StaticConfiguration: &firecracker.StaticNetworkConfiguration{
					MacAddress:  "10:5b:ad:53:5c:17",
					HostDevName: "tapName1",
					IPConfiguration: &firecracker.IPConfiguration{
						IPAddr: net.IPNet{
							IP:   vm1_eth0_ip_ipv4,
							Mask: net.CIDRMask(24, 32),
						},
						Gateway: bridge_ip_without_mask,
						IfName:  "eth0",
						Nameservers: []string{
							"8.8.8.8",
							"8.8.4.4",
						},
					},
				},
			},
		},
	}

	// // Read the startup script from a file
	// startupScriptPath2 := "startup-script/startup-script-vm2.sh"
	// vm2_startupScript, err := ioutil.ReadFile(startupScriptPath2)
	// if err != nil {
	// 	fmt.Printf("Failed to read vm2 startup script: %v\n", err)
	// 	return
	// }

	// script2 := fmt.Sprintf(`#!/bin/bash
	// ip addr add %s/24 dev eth0
	// ip link set eth0 up
	// ip route add default via %s dev eth0
	// `, vm2_eth0_ip, bridge_ip_address)

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	entry := logrus.NewEntry(logger)

	ctx := context.Background()
	m1, err := firecracker.NewMachine(ctx, cfg1, firecracker.WithLogger(entry))
	if err != nil {
		fmt.Printf("Failed to create VM1: %v\n", err)
		return
	}

	vmmCtx, vmmCancel := context.WithCancel(ctx)
	defer vmmCancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		fmt.Printf("Received signal: %s\n", sig)
		vmmCancel()
	}()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		if err := m1.Start(vmmCtx); err != nil {
			fmt.Printf("Failed to start VM1: %v\n", err)
			return
		}
		if err := m1.Wait(vmmCtx); err != nil {
			fmt.Printf("VM1 exited with error: %v\n", err)
		} else {
			fmt.Println("VM1 exited successfully")
		}
	}()

	wg.Wait()
}
