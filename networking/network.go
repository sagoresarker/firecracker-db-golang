package networking

import (
	"fmt"

	"github.com/sagoresarker/firecracker-db-golang/database"
)

var tapName1, tapName2 string

func SetUpNetwork() {
	fmt.Println("Setup full networking")
	bridgeName, userID, bridge_ip_address, _ := SetupBridgeNetwork()
	tapName1, tapName2, _ := SetupTapNetwork(bridgeName)

	fmt.Println("Bridge Name:", bridgeName)
	fmt.Println("Tap Names:", tapName1, tapName2)

	database.SaveNetworkDetails(bridgeName, tapName1, tapName2, userID, bridge_ip_address)
}

func GetTapNames() (string, string) {

	fmt.Println("The Tap Names are (from Get Method):", tapName1, tapName2)

	return tapName1, tapName2
}

func GetBridgeIPAddress() (string, string) {
	_, _, bridge_ip_address, bridge_gateway_ip := SetupBridgeNetwork()
	return bridge_ip_address, bridge_gateway_ip
}
