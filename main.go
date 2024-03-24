package main

import (
	"fmt"

	"github.com/sagoresarker/firecracker-db-golang/database"
	vm "github.com/sagoresarker/firecracker-db-golang/firecracker"
	"github.com/sagoresarker/firecracker-db-golang/networking"
)

func main() {
	fmt.Println("Hello Poridhians!")
	database.InitMongoDB()

	networking.SetUpNetwork()

	// bridgeDetails, err := database.GetBridgeDetails()
	// if err != nil {
	// 	fmt.Println("Error getting bridge details:", err)
	// 	return
	// }

	// fmt.Println("Bridge Details:", bridgeDetails)

	tapName1, tapName2 := networking.GetTapNames()

	vm.LaunchFirstVM(tapName1, tapName2)
}
