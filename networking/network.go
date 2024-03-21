package networking

import (
	"fmt"

	"github.com/sagoresarker/firecracker-db-golang/database"
)

var tapName1, tapName2 string

func SetUpNetwork() {
	fmt.Println("Setup full networking")
	bridgeName, userID, ipAddress := SetupBridgeNetwork()
	tapName1, tapName2, _ := SetupTapNetwork(bridgeName)

	fmt.Println("Bridge Name:", bridgeName)
	fmt.Println("Tap Names:", tapName1, tapName2)

	database.SaveNetworkDetails(bridgeName, tapName1, tapName2, userID, ipAddress)
}

func GetTapNames() (string, string) {

	fmt.Println("The Tap Names are (from Get Method):", tapName1, tapName2)

	return tapName1, tapName2
}

func BridgeIPAddress() string {
	_, _, ipAddress := SetupBridgeNetwork()
	return ipAddress
}
