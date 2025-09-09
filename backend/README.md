# KubeCloud

KubeCloud is a CLI tool that helps you deploy and manage Kubernetes clusters on the decentralized TFGrid.

## Configuration

KubeCloud supports configuration through environment variables, CLI flags, and configuration files.

### Configuration File

By default, KubeCloud looks for a `config.json` file in the current directory. You can specify a custom configuration file path using the `--config` or `-c` flag:

```bash
kubecloud --config /path/to/config.json
```

The configuration file should be in JSON format. Example:

```json
{
  "server": {
    "host": "localhost",
    "port": "8080"
  },
  "database": {
    "file": "/path/to/db/file"
  },
  "token": {
    "secret": "your-secret-key",
    "access_token_expiry_minutes": 60,
    "refresh_token_expiry_hours": 24
  },
  "currency": "USD",
  "stripe_secret": "your-stripe-secret",
  "tfchain_url": "wss://tfchain.dev.grid.tf/wss",
  "gridproxy_url": "https://gridproxy.dev.grid.tf/",
  "activation_service_url": "https://activation.grid.tf/activation/activate",
  "graphql_url": "https://graphql.grid.tf/graphql",
  "firesquid_url": "https://firesquid.grid.tf/graphql",
  "redis": {
    "host": "localhost",
    "port": 6379,
    "password": "",
    "db": 0
  },
  "grid": {
    "mne": "your-mnemonic",
    "net": "main"
  },
  "invoice": {
    "name": "KubeCloud Invoice",
    "address": "123 KubeCloud St, Cloud City, CC 12345",
    "governorate": "Cloud Governorate"
  },
  "logger": {
  "log_dir": "./app/logs",
  "max_size": 512,
  "max_backups": 12,
  "max_age": 30,
  "compress": true
  }
}
```

### Environment Variables

You can override configuration values using environment variables. Environment variables should be prefixed with `KUBECLOUD_`. For example:

```bash
export KUBECLOUD_SERVER_HOST=localhost
export KUBECLOUD_SERVER_PORT=8080
```

### CLI Flags

Some configuration options can be passed directly as CLI flags. For example:

```bash
kubecloud --config /path/to/config.json
```

### Priority Order

The priority order for configuration is:

1. CLI flags
2. Environment variables
3. Configuration file
4. Default values

This allows you to override specific settings without modifying the configuration file.
