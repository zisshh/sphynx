package loadbalancing

import (
	"errors"
	"sync"

	"load-balancer/models"
	"load-balancer/utils"

	"github.com/sirupsen/logrus"
)

var (
	rrIndex       int
	rrMutex       sync.Mutex
	currentIndex  int
	currentWeight int

	HealthState      = make(map[string]bool)
	HealthStateMutex sync.Mutex
)

// Thread-safe update of HealthState
func UpdateHealthState(url string, health bool) {
	HealthStateMutex.Lock()
	defer HealthStateMutex.Unlock()
	HealthState[url] = health
}

func getHealthState(url string) (bool, bool) {
	HealthStateMutex.Lock()
	defer HealthStateMutex.Unlock()
	health, exists := HealthState[url]
	return health, exists
}

func GetRoundRobinServer(serverList []*models.Server, logger *logrus.Logger) (*models.Server, error) {
	rrMutex.Lock()
	defer rrMutex.Unlock()

	if len(serverList) == 0 {
		return nil, errors.New("no servers available")
	}

	for {
		rrIndex = (rrIndex + 1) % len(serverList)
		server := serverList[rrIndex]
		if server.Health {
			logServerHealthChanges(serverList, logger)
			return server, nil
		}
	}
}

func GetWeightedRoundRobinServer(serverList []*models.Server, logger *logrus.Logger) (*models.Server, error) {
	totalWeight := 0
	for _, server := range serverList {
		if server.Health {
			totalWeight += server.Weight
		}
	}
	if totalWeight == 0 {
		return nil, errors.New("no healthy hosts")
	}

	for {
		currentIndex = (currentIndex + 1) % len(serverList)
		if currentIndex == 0 {
			currentWeight -= utils.GCD(serverList)
			if currentWeight <= 0 {
				currentWeight = utils.MaxWeight(serverList)
			}
		}
		if serverList[currentIndex].Health && serverList[currentIndex].Weight >= currentWeight {
			logServerHealthChanges(serverList, logger)
			return serverList[currentIndex], nil
		}
	}
}

func GetLeastConnectionsServer(serverList []*models.Server, logger *logrus.Logger) (*models.Server, error) {
	var leastConnServer *models.Server
	for _, server := range serverList {
		if server.Health {
			if leastConnServer == nil || server.Connections < leastConnServer.Connections {
				leastConnServer = server
			}
		}
	}
	if leastConnServer == nil {
		return nil, errors.New("no healthy hosts")
	}
	logServerHealthChanges(serverList, logger)
	return leastConnServer, nil
}

func GetWeightedLeastConnectionsServer(serverList []*models.Server, logger *logrus.Logger) (*models.Server, error) {
	var bestServer *models.Server
	var minRatio float64 = -1
	for _, server := range serverList {
		if server.Health {
			ratio := float64(server.Connections) / float64(server.Weight)
			if minRatio == -1 || ratio < minRatio {
				minRatio = ratio
				bestServer = server
			}
		}
	}
	if bestServer == nil {
		return nil, errors.New("no healthy hosts")
	}
	logServerHealthChanges(serverList, logger)
	return bestServer, nil
}

func logServerHealthChanges(serverList []*models.Server, logger *logrus.Logger) {
	for _, server := range serverList {
		previousHealth, exists := getHealthState(server.URL)
		if !exists || previousHealth != server.Health {
			if server.Health {
				logger.WithFields(logrus.Fields{
					"server": server.URL,
				}).Info("Server is back to healthy state")
			} else {
				logger.WithFields(logrus.Fields{
					"server": server.URL,
				}).Warn("Server is down")
			}
			UpdateHealthState(server.URL, server.Health)
		}
	}
}

func GetHealthyServer(vs *models.VirtualService) (*models.Server, error) {
	switch vs.Algorithm {
	case "round_robin":
		return GetRoundRobinServer(vs.ServerList, vs.Logger)
	case "weighted_round_robin":
		return GetWeightedRoundRobinServer(vs.ServerList, vs.Logger)
	case "least_connections":
		return GetLeastConnectionsServer(vs.ServerList, vs.Logger)
	case "weighted_least_connections":
		return GetWeightedLeastConnectionsServer(vs.ServerList, vs.Logger)
	default:
		return nil, errors.New("unknown load balancing algorithm")
	}
}
