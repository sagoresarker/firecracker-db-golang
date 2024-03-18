package main

import (
	"fmt"

	"github.com/sagoresarker/firecracker-db-golang/database"
	"github.com/sagoresarker/firecracker-db-golang/networking"
)

func main() {
	fmt.Println("Hello World")
	database.InitMongoDB()

	networking.SetUpNetwork()

	bridgeDetails, err := database.GetBridgeDetails()
	if err != nil {
		fmt.Println("Error getting bridge details:", err)
		return
	}

	fmt.Println("Bridge Details:", bridgeDetails)
}
