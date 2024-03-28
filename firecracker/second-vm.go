package vm

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	models "github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	"github.com/sagoresarker/firecracker-db-golang/networking"
	"github.com/sirupsen/logrus"
)

func LaunchSecondVM(tapName2 string) {
	fmt.Println("Launching second VM")
	bridge_ip_address, bridge_gateway_ip := networking.GetBridgeIPAddress()

	bridge_ip_without_mask, _, err := net.ParseCIDR(bridge_ip_address)
	if err != nil {
		fmt.Println("Error parsing bridge IP address:", err)
		return
	}

	fmt.Println("(Launch VM) - Bridge IP without mask:", bridge_ip_without_mask)

	_, vm2_eth0_ip, err := networking.GetVMIPs(bridge_ip_without_mask.String())

	if err != nil {
		fmt.Println("Error getting VM IPs:", err)
		return
	}

	fmt.Printf("VM2 IP: %s\n", vm2_eth0_ip)

	vm2_eth0_ip_ipv4 := net.ParseIP(vm2_eth0_ip)
	if vm2_eth0_ip_ipv4 == nil {
		fmt.Println("Error parsing VM2 IP address")
		return
	}

	bridge_gateway_ip_ipv4 := net.ParseIP(bridge_gateway_ip)
	fmt.Printf("Bridge Gateway IP: %s and Type %s\n", bridge_gateway_ip_ipv4, reflect.TypeOf(bridge_gateway_ip_ipv4).String())

	if bridge_gateway_ip_ipv4 == nil {
		fmt.Println("Error parsing bridge gateway IP address")
		return
	}

	fmt.Println("tapName2 in LaunchSecondVM:", tapName2)

	cfg2 := firecracker.Config{
		SocketPath:      "/tmp/firecracker2.sock",
		LogFifo:         "/tmp/firecracker2.log",
		MetricsFifo:     "/tmp/firecracker2-metrics",
		LogLevel:        "Debug",
		KernelImagePath: "files/vmlinux",
		KernelArgs:      "ro console=ttyS0 reboot=k panic=1 pci=off",

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
				PathOnHost:   firecracker.String("files/build/rootfs.ext4"),
			},
		},
		NetworkInterfaces: []firecracker.NetworkInterface{
			{
				StaticConfiguration: &firecracker.StaticNetworkConfiguration{
					MacAddress:  "10:5b:ad:50:5c:13",
					HostDevName: tapName2,
					IPConfiguration: &firecracker.IPConfiguration{
						IPAddr: net.IPNet{
							IP:   vm2_eth0_ip_ipv4,
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

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	entry := logrus.NewEntry(logger)

	ctx := context.Background()
	m2, err := firecracker.NewMachine(ctx, cfg2, firecracker.WithLogger(entry))
	if err != nil {
		fmt.Printf("Failed to create VM2: %v\n", err)
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
		if err := m2.Start(vmmCtx); err != nil {
			fmt.Printf("Failed to start VM2: %v\n", err)
			return
		}
		if err := m2.Wait(vmmCtx); err != nil {
			fmt.Printf("VM2 exited with error: %v\n", err)
		} else {
			fmt.Println("VM2 exited successfully")
		}
	}()

	wg.Wait()
}
