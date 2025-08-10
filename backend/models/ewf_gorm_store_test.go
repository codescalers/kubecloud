package models

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xmonader/ewf"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestSQLiteStore_SaveAndLoad tests saving and loading a workflow in SQLiteStore.
func TestGormStore_SaveAndLoad(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "ewf_gorm_store_test_*.db")
	require.NoError(t, err)
	dbFile := tmpFile.Name()
	require.NoError(t, tmpFile.Close())

	t.Cleanup(func() {
		err := os.Remove(dbFile)
		require.NoError(t, err)
	})

	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)

	t.Cleanup(func() {
		sqlDB.Close()
	})

	store := NewGormStore(db)
	require.NoError(t, err)

	t.Cleanup(func() {
		err := store.Close()
		require.NoError(t, err)
	})

	err = store.Setup()
	require.NoError(t, err)
	wfName := "test-gorm-workflow"
	wf := ewf.NewWorkflow(wfName)
	wf.Steps = []ewf.Step{{Name: "dummy_activity"}}
	wf.State["key"] = "value"
	wf.CurrentStep = 2
	wf.Status = ewf.StatusCompleted

	err = store.SaveWorkflow(context.Background(), wf)
	require.NoError(t, err)

	loadedWf, err := store.LoadWorkflowByUUID(context.Background(), wf.UUID)
	require.NoError(t, err)

	// Also test loading by name
	loadedByName, err := store.LoadWorkflowByName(context.Background(), wf.Name)
	require.NoError(t, err)

	require.Equal(t, wf.UUID, loadedByName.UUID, "Expected workflow UUID to match")
	require.Equal(t, wfName, loadedWf.Name, "Expected workflow name to match")
	require.Equal(t, 2, loadedWf.CurrentStep, "Expected CurrentStep to be 2")
	require.Equal(t, ewf.StatusCompleted, loadedWf.Status, "Expected Status to be COMPLETED")
	require.Equal(t, "value", loadedWf.State["key"], "Expected state['key'] to be 'value'")
}

func TestGormStore_LoadNotFound(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "ewf_gorm_store_test_*.db")
	require.NoError(t, err)
	dbFile := tmpFile.Name()
	require.NoError(t, tmpFile.Close())

	t.Cleanup(func() {
		err := os.Remove(dbFile)
		require.NoError(t, err)
	})

	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)

	t.Cleanup(func() {
		sqlDB.Close()
	})

	store := NewGormStore(db)
	if err != nil {
		t.Fatalf("NewGormStore() error = %v", err)
	}

	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Fatalf("failed to close store: %v", err)
		}
	})

	// Test LoadWorkflowByUUID with non-existent UUID
	_, err = store.LoadWorkflowByUUID(context.Background(), "non-existent-id")
	require.Error(t, err, "Expected an error when loading a non-existent workflow by UUID")

	// Test LoadWorkflowByName with non-existent name
	_, err = store.LoadWorkflowByName(context.Background(), "non-existent-name")
	require.Error(t, err, "Expected an error when loading a non-existent workflow by name")
}
