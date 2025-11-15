package persistence

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	infrapers "kubecloud/internal/infrastructure/persistence"
	"github.com/stretchr/testify/require"
	"github.com/xmonader/ewf"
	"gorm.io/driver/sqlite"
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

	sqlDB, err := infrapers.NewGormStorage(sqlite.Open(dbFile))
	require.NoError(t, err)

	t.Cleanup(func() {
		sqlDB.Close()
	})

	store := NewGormEWFRepository(sqlDB)
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

	sqlDB, err := infrapers.NewGormStorage(sqlite.Open(dbFile))
	require.NoError(t, err)

	t.Cleanup(func() {
		sqlDB.Close()
	})

	store := NewGormEWFRepository(sqlDB)
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

// TestGormStore_SaveAndLoadQueues tests saving and loading of queues in GormStore.
func TestGormStore_SaveAndLoadQueues(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "ewf_gorm_store_test_*.db")
	require.NoError(t, err)
	dbFile := tmpFile.Name()
	require.NoError(t, tmpFile.Close())

	t.Cleanup(func() {
		err := os.Remove(dbFile)
		require.NoError(t, err)
	})

	sqlDB, err := infrapers.NewGormStorage(sqlite.Open(dbFile))
	require.NoError(t, err)

	t.Cleanup(func() {
		sqlDB.Close()
	})

	store := NewGormEWFRepository(sqlDB)
	if err != nil {
		t.Fatalf("NewGormStore() error = %v", err)
	}

	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Fatalf("failed to close store: %v", err)
		}
	})

	err = store.Setup()
	require.NoError(t, err)

	for i := 0; i < 4; i++ {
		name := fmt.Sprintf("store_test_queue_%d", i)
		workersDef := ewf.WorkersDefinition{Count: 2, PollInterval: 1 * time.Second}
		queueOpts := ewf.QueueOptions{AutoDelete: false, PopTimeout: 1 * time.Second}

		queueMetaData := &ewf.QueueMetadata{
			Name:         name,
			WorkersDef:   workersDef,
			QueueOptions: queueOpts,
		}

		err = store.SaveQueueMetadata(context.Background(), queueMetaData)
		if err != nil {
			t.Fatalf("Save() error = %v", err)
		}
	}

	queues, err := store.LoadAllQueueMetadata(context.Background())
	if err != nil {
		t.Fatalf("LoadAllQueueMetadata() error = %v", err)
	}

	if len(queues) != 4 {
		t.Errorf("Expected queues len to be 4, got %d", len(queues))
	}

	for i, q := range queues {
		expectedName := fmt.Sprintf("store_test_queue_%d", i)
		expectedWorkersDef := ewf.WorkersDefinition{Count: 2, PollInterval: 1 * time.Second}
		expectedQueueOpts := ewf.QueueOptions{AutoDelete: false, PopTimeout: 1 * time.Second}

		if q.Name != expectedName {
			t.Errorf("Expected queue name %s, got %s", q.Name, expectedName)
		}
		if !reflect.DeepEqual(q.WorkersDef, expectedWorkersDef) {
			t.Errorf("Expected workersDef to be %v, got %v", expectedWorkersDef, q.WorkersDef)
		}
		if !reflect.DeepEqual(q.QueueOptions, expectedQueueOpts) {
			t.Errorf("Expected queueOpts to be %v, got %v", expectedQueueOpts, q.QueueOptions)
		}
	}

}

// TestGormStore_DeleteQueue tests deleting a queue from GormStore.
func TestGormStore_DeleteQueue(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "ewf_gorm_store_test_*.db")
	require.NoError(t, err)
	dbFile := tmpFile.Name()
	require.NoError(t, tmpFile.Close())

	t.Cleanup(func() {
		err := os.Remove(dbFile)
		require.NoError(t, err)
	})

	sqlDB, err := infrapers.NewGormStorage(sqlite.Open(dbFile))
	require.NoError(t, err)

	t.Cleanup(func() {
		sqlDB.Close()
	})

	store := NewGormEWFRepository(sqlDB)
	if err != nil {
		t.Fatalf("NewGormStore() error = %v", err)
	}

	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Fatalf("failed to close store: %v", err)
		}
	})

	err = store.Setup()
	require.NoError(t, err)

	name := "store_test_queue"
	workersDef := ewf.WorkersDefinition{Count: 2, PollInterval: 1 * time.Second}
	queueOpts := ewf.QueueOptions{AutoDelete: false, PopTimeout: 1 * time.Second}

	queueMetaData := &ewf.QueueMetadata{
		Name:         name,
		WorkersDef:   workersDef,
		QueueOptions: queueOpts,
	}

	err = store.SaveQueueMetadata(context.Background(), queueMetaData)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	err = store.DeleteQueueMetadata(context.Background(), queueMetaData.Name)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	queues, err := store.LoadAllQueueMetadata(context.Background())
	if err != nil {
		t.Fatalf("loadAll() error = %v", err)
	}

	if len(queues) != 0 {
		t.Errorf("Expected queues len to be 0, got %d", len(queues))
	}
}
