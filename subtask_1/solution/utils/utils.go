package utils

import "load-balancer/models"

func GCD(serverList []*models.Server) int {
	gcd := serverList[0].Weight
	for _, server := range serverList {
		gcd = gcdHelper(gcd, server.Weight)
	}
	return gcd
}

func gcdHelper(a, b int) int {
	if b == 0 {
		return a
	}
	return gcdHelper(b, a%b)
}

func MaxWeight(serverList []*models.Server) int {
	maxWeight := 0
	for _, server := range serverList {
		if server.Weight > maxWeight {
			maxWeight = server.Weight
		}
	}
	return maxWeight
}
