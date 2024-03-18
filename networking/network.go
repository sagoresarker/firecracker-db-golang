package networking

import (
	"fmt"

	"github.com/sagoresarker/firecracker-db-golang/database"
)

func SetUpNetwork() {
	fmt.Println("Setup full networking")
	bridgeName, userID, ipAddress := SetupBridgeNetwork()
	tapName, _ := SetupTapNetwork(bridgeName)

	fmt.Println("Bridge Name:", bridgeName)
	fmt.Println("Tap Name:", tapName)

	database.SaveNetworkDetails(bridgeName, tapName, userID, ipAddress)

}
