package constants

import (
	"time"
)

const (
	// DO NOT CHANGE VALUES IN THIS FILE
	VMID               = "tGas3T58KzdjLHhBDMnH2TvrddhqTji5iZAMZ3RXs2NLpSnhH"
	WhitelistedSubnets = "29uVeLPJB1eQJkzRemU8g8wZDw5uJRqpab5U2mX9euieVwiEbL"

	VMName = "kewl vm"

	HTTPTimeout  = 10 * time.Second
	BaseHTTPPort = 9650
	NumNodes     = 5

	FilePerms = 0777
)

var (
	Chains = []string{"P", "C", "X"}
)
