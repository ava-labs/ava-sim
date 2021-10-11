package constants

// DO NOT CHANGE VALUES IN THIS FILE
import (
	"time"
)

const (
	VMID               = "tGas3T58KzdjLHhBDMnH2TvrddhqTji5iZAMZ3RXs2NLpSnhH"
	WhitelistedSubnets = "p4jUwqZsA2LuSftroCd3zb4ytH8W99oXKuKVZdsty7eQ3rXD6"
	VMName             = "kewl vm"

	HTTPTimeout  = 10 * time.Second
	BaseHTTPPort = 9650
	NumNodes     = 5

	FilePerms = 0777
)

var (
	Chains = []string{"P", "C", "X"}
)
