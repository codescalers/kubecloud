package internal

import "github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"

var (
	FireSquidURLs = map[string][]string{
		deployer.DevNetwork: {
			"https://firesquid.dev.grid.tf/graphql",
			"https://firesquid.dev.threefold.me/graphql",
		},
		deployer.TestNetwork: {
			"https://firesquid.test.grid.tf/graphql",
			"https://firesquid.test.threefold.me/graphql",
		},
		deployer.QaNetwork: {
			"https://firesquid.qa.grid.tf/graphql",
			"https://firesquid.qa.threefold.me/graphql",
		},
		deployer.MainNetwork: {
			"https://firesquid.grid.tf/graphql",
			"https://firesquid.be.grid.tf/graphql",
			"https://firesquid.grid.threefold.me/graphql",
			"https://firesquid.sg.grid.tf/graphql",
			"https://firesquid.us.grid.tf/graphql",
			"https://firesquid.grid.threefold.io/graphql",
		},
	}

	KYCURLs = map[string]string{
		deployer.DevNetwork:  "https://kyc.dev.grid.tf",
		deployer.TestNetwork: "https://kyc.test.grid.tf",
		deployer.QaNetwork:   "https://kyc.qa.grid.tf",
		deployer.MainNetwork: "https://kyc.grid.tf",
	}

	ActivationServiceURLs = map[string]string{
		deployer.DevNetwork:  "https://activation.dev.grid.tf/activation/activate",
		deployer.TestNetwork: "https://activation.test.grid.tf/activation/activate",
		deployer.QaNetwork:   "https://activation.qa.grid.tf/activation/activate",
		deployer.MainNetwork: "https://activation.grid.tf/activation/activate",
	}
)
