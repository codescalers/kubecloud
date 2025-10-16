package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
)

func TestListAllContractsInPeriod(t *testing.T) {
	// Setup in-memory SQLite database for testing
	db, err := NewGormStorage(sqlite.Open("file::memory:?cache=shared"))
	assert.NoError(t, err)

	// Clean up after test
	defer func() {
		sqlDB, _ := db.GetDB().DB()
		sqlDB.Close()
	}()

	gormDB := db.(*GormDB)

	// Define test time periods
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	twoDaysAgo := now.Add(-48 * time.Hour)
	tomorrow := now.Add(24 * time.Hour)

	// Create test contracts with different creation and deletion times

	// Contract 1: Created two days ago, not deleted (should be included)
	contract1 := &UserContractData{
		UserID:     1,
		ContractID: 1001,
		NodeID:     1,
		Type:       ContractTypeRented,
		CreatedAt:  twoDaysAgo,
	}
	assert.NoError(t, gormDB.CreateUserContractData(contract1))

	// Contract 2: Created yesterday, deleted today (should be included)
	contract2 := &UserContractData{
		UserID:     1,
		ContractID: 1002,
		NodeID:     2,
		Type:       ContractTypeRented,
		CreatedAt:  yesterday,
	}
	// Manually update the DeletedAt field since CreateUserContractData doesn't set it
	assert.NoError(t, gormDB.CreateUserContractData(contract2))
	assert.NoError(t, gormDB.GetDB().Model(&UserContractData{}).
		Where("contract_id = ?", 1002).
		Update("deleted_at", now).Error)

	// Contract 3: Created two days ago, deleted yesterday (should be included)
	contract3 := &UserContractData{
		UserID:     2,
		ContractID: 1003,
		NodeID:     3,
		Type:       ContractTypeDeployed,
		CreatedAt:  twoDaysAgo,
	}
	assert.NoError(t, gormDB.CreateUserContractData(contract3))
	assert.NoError(t, gormDB.GetDB().Model(&UserContractData{}).
		Where("contract_id = ?", 1003).
		Update("deleted_at", yesterday).Error)

	// Contract 4: Will be created tomorrow (should NOT be included)
	contract4 := &UserContractData{
		UserID:     2,
		ContractID: 1004,
		NodeID:     4,
		Type:       ContractTypeDeployed,
		CreatedAt:  tomorrow,
	}
	assert.NoError(t, gormDB.CreateUserContractData(contract4))

	// Test period: from yesterday to today
	periodStart := yesterday
	periodEnd := now

	// Test with userID = 0 (should return all contracts in the period)
	contracts, err := gormDB.ListAllContractsInPeriod(0, periodStart, periodEnd)
	assert.NoError(t, err)

	// Should include contracts 1, 2, and 3, but not 4
	assert.Equal(t, 3, len(contracts), "Expected 3 contracts in the period")

	// Verify contract IDs
	contractIDs := make([]uint64, len(contracts))
	for i, c := range contracts {
		contractIDs[i] = c.ContractID
	}

	assert.Contains(t, contractIDs, uint64(1001), "Contract 1001 should be in the period")
	assert.Contains(t, contractIDs, uint64(1002), "Contract 1002 should be in the period")
	assert.Contains(t, contractIDs, uint64(1003), "Contract 1003 should be in the period")

	// Test with userID = 2 (should only return contracts for user 2)
	contractsUser2, err := gormDB.ListAllContractsInPeriod(2, periodStart, periodEnd)
	assert.NoError(t, err)

	// Should only include contract 3, not 1, 2, or 4
	assert.Equal(t, 1, len(contractsUser2), "Expected 1 contract for user 2 in the period")

	// Verify contract IDs for user 2
	contractIDsUser2 := make([]uint64, len(contractsUser2))
	for i, c := range contractsUser2 {
		contractIDsUser2[i] = c.ContractID
	}

	assert.Contains(t, contractIDsUser2, uint64(1003), "Contract 1003 should be in the period for user 2")
	assert.NotContains(t, contractIDs, uint64(1004), "Contract 1004 should not be in the period")
}
