package services

import (
	"context"
	"kubecloud/internal/core/models"
	"kubecloud/internal/infrastructure/substrate"

	proxy "github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/client"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/types"
)

type StatsService struct {
	userRepo        models.UserRepository
	clusterRepo     models.ClusterRepository
	gridProxyClient proxy.Client
	substrateClient substrate.Substrate
	systemMnemonic  string
}

func NewStatsService(
	userRepo models.UserRepository,
	clusterRepo models.ClusterRepository,
	gridProxyClient proxy.Client,
	substrateClient substrate.Substrate,
	systemMnemonic string,
) StatsService {
	return StatsService{
		userRepo:        userRepo,
		clusterRepo:     clusterRepo,
		gridProxyClient: gridProxyClient,
		substrateClient: substrateClient,
		systemMnemonic:  systemMnemonic,
	}
}

type Stats struct {
	TotalUsers           uint32  `json:"total_users"`
	TotalClusters        uint32  `json:"total_clusters"`
	UpNodes              uint32  `json:"up_nodes"`
	Countries            uint32  `json:"countries"`
	Cores                uint32  `json:"cores"`
	SSD                  float64 `json:"ssd"`
	SystemAccountBalance float64 `json:"system_account_balance"`
}

func (svc *StatsService) GetStats(ctx context.Context) (Stats, error) {
	totalUsers, err := svc.userRepo.CountAllUsers()
	if err != nil {
		return Stats{}, err
	}

	totalClusters, err := svc.clusterRepo.CountAllClusters()
	if err != nil {
		return Stats{}, err
	}

	stats, err := svc.gridProxyClient.Stats(ctx, types.StatsFilter{Status: []string{"up", "standby"}})
	if err != nil {
		return Stats{}, err
	}

	// Fetch system account balance
	var systemBalance float64
	if svc.substrateClient != nil && svc.systemMnemonic != "" {
		balance, err := svc.substrateClient.GetUserBalanceUSD(svc.systemMnemonic)
		if err == nil {
			systemBalance = balance
		}
		// Continue even if balance fetch fails
	}

	return Stats{
		TotalUsers:           uint32(totalUsers),
		TotalClusters:        uint32(totalClusters),
		UpNodes:              uint32(stats.Nodes),
		Countries:            uint32(stats.Countries),
		Cores:                uint32(stats.TotalCRU),
		SSD:                  float64(stats.TotalSRU) / (1024 * 1024 * 1024),
		SystemAccountBalance: systemBalance,
	}, nil
}
