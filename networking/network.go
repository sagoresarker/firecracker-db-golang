package networking

import (
	"fmt"

	"github.com/sagoresarker/firecracker-db-golang/database"
)

var (
	tapName1, tapName2       string
	bridge_ip_address_global = "Blank"
	bridge_gateway_ip_global = "Blank"
)

func SetUpNetwork() {
	fmt.Println("Setup full networking")
	bridgeName, userID, bridge_ip_address, _, err := SetupBridgeNetwork()
	bridge_ip_address_global = bridge_ip_address

	if err != nil {
		fmt.Println("(From SetUp network) - Error setting up bridge network:", err)
	}
	fmt.Println("(From SetUp network) - Bridge IP Address:", bridge_ip_address)
	tapName1, tapName2, err = SetupTapNetwork(bridgeName)

	if err != nil {
		fmt.Println("(From SetUp network) - Error setting up tap network:", err)
	}

	fmt.Println("Bridge Name:", bridgeName)
	fmt.Println("Tap Names:", tapName1, tapName2)

	database.SaveNetworkDetails(bridgeName, tapName1, tapName2, userID, bridge_ip_address)
}

func GetTapNames() (string, string) {

	fmt.Println("The Tap Names are (from Get Method):", tapName1, tapName2)

	return tapName1, tapName2
}

func GetBridgeIPAddress() (string, string) {
	fmt.Println("The Bridge IP Address is (from Get Method):", bridge_ip_address_global)
	return bridge_ip_address_global, bridge_gateway_ip_global
}
