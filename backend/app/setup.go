package app

// import (
// 	"fmt"
// 	"os"
// 	"path/filepath"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// )

// func SetUpTesting(t testing.T)*App{

// 	tempDir := t.TempDir()

// 	configPath := filepath.Join(tempDir, "test_config.json")
// 	dbPath := filepath.Join(tempDir, "testing.db")

// 	config := fmt.Sprintf(`
// {
//   "server": {
//     "host": "localhost",
//     "port": "3000"
//   },
//   "token": {
//     "secret": "secret",
//     "access_token_expiry_minutes": 60,
//     "refresh_token_expiry_hours": 24
//   },
//   "admins": [""],
//   "currency": "usd",
//   "stripe_secret": "secret-testing",
//   "tfchain_url": "wss://tfchain.grid.tf/wss",
//   "gridproxy_url": "https://gridproxy.dev.grid.tf/",
//   "voucher_name_length": 5,
//   "terms_and_conditions": {
//     "document_link": "https://manual.grid.tf/labs/knowledge_base/terms_conditions_all3",
//     "document_hash": "6f2b4109704ba2883d978a7b94e5f295"
//   },
//   "activation_service_url": "https://activation.dev.grid.tf/activation/activate",
//   "system_account": {
//     "mnemonics": "winner giant reward damage expose pulse recipe manual brand volcano dry avoid",
//     "network": "dev"
//   },
//   "graphql_url": "https://graphql.dev.grid.tf/graphql",
//   "firesquid_url": "https://firesquid.dev.grid.tf/graphql",
//   "redis": {
//     "host": "localhost",
//     "port": 6379,
//     "password": "",
//     "db": 0
//   },
//   "grid": {
//     "mne": "mom picnic deliver again rug night rabbit music motion hole lion where",
//     "net": "main"
//   },
//   "deployer_workers_num": 3,
//   "invoice": {
//     "name": "Codescalers Egypt",
//     "address": "9 Al Wardi street, El Hegaz St",
//     "governorate": "Cairo Governorate 11341"
//   }
// }
// 	`, dbPath)
// 	err := os.WriteFile(configPath, []byte(config), 0644)
// 	assert.NoError(t, err)

// 	configuration, err := ReadConfFile(configPath)

// }