package vm

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	models "github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	"github.com/sirupsen/logrus"
)

func LaunchVM(tapName1 string, tapName2 string) {

	// Read the startup script from a file
	startupScriptPath := "startup-script/startup-script-vm1.sh"
	vm1_startupScript, err := ioutil.ReadFile(startupScriptPath)
	if err != nil {
		fmt.Printf("Failed to read vm1 startup script: %v\n", err)
		return
	}
	cfg1 := firecracker.Config{
		SocketPath:      "/tmp/firecracker1.sock",
		LogFifo:         "/tmp/firecracker1.log",
		MetricsFifo:     "/tmp/firecracker1-metrics",
		LogLevel:        "Debug",
		KernelImagePath: "files/vmlinux",
		KernelArgs:      fmt.Sprintf("ro console=ttyS0 reboot=k panic=1 pci=off %s", vm1_startupScript),
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
				},
			},
		},
	}

	// Read the startup script from a file
	startupScriptPath2 := "startup-script/startup-script-vm2.sh"
	vm2_startupScript, err := ioutil.ReadFile(startupScriptPath2)
	if err != nil {
		fmt.Printf("Failed to read vm2 startup script: %v\n", err)
		return
	}

	cfg2 := firecracker.Config{
		SocketPath:      "/tmp/firecracker2.sock",
		LogFifo:         "/tmp/firecracker2.log",
		MetricsFifo:     "/tmp/firecracker2-metrics",
		LogLevel:        "Debug",
		KernelImagePath: "files/vmlinux",
		KernelArgs:      fmt.Sprintf("ro console=ttyS0 reboot=k panic=1 pci=off %s", vm2_startupScript),
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
					MacAddress:  "02:42:27:c3:1c:87",
					HostDevName: "tapName2",
				},
			},
		},
	}

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	entry := logrus.NewEntry(logger)

	ctx := context.Background()
	m1, err := firecracker.NewMachine(ctx, cfg1, firecracker.WithLogger(entry))
	if err != nil {
		fmt.Printf("Failed to create VM1: %v\n", err)
		return
	}

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
	wg.Add(2)

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
