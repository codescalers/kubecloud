package deployment

import "fmt"

func getDeployNodeStepName(nodeIndex int) string {
	return fmt.Sprintf("deploy-%d%s-node", nodeIndex, getOrdinalSuffix(nodeIndex))
}

func getOrdinalSuffix(n int) string {
	switch n % 100 {
	case 11, 12, 13:
		return "th"
	default:
		switch n % 10 {
		case 1:
			return "st"
		case 2:
			return "nd"
		case 3:
			return "rd"
		default:
			return "th"
		}
	}
}
