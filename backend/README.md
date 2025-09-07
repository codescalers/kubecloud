# Mycelium Cloud

Mycelium Cloud is a CLI tool that helps you deploy and manage Kubernetes clusters on the decentralized TFGrid.

## Configuration

Mycelium Cloud supports configuration through environment variables, CLI flags, and configuration files.

### Configuration File

By default, Mycelium Cloud looks for a `config.json` file in the current directory. You can specify a custom configuration file path using the `--config` or `-c` flag:

```bash
myceliumcloud --config /path/to/config.json
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
    "name": "Mycelium Cloud Invoice",
    "address": "123 Mycelium Cloud St, Cloud City, CC 12345",
    "governorate": "Cloud Governorate"
  }
}
```

### Environment Variables

You can override configuration values using environment variables. Environment variables should be prefixed with `MYCELIUMCLOUD_`. For example:

```bash
export MYCELIUMCLOUD_SERVER_HOST=localhost
export MYCELIUMCLOUD_SERVER_PORT=8080
```

### CLI Flags

Some configuration options can be passed directly as CLI flags. For example:

```bash
myceliumcloud --config /path/to/config.json
```

### Priority Order

The priority order for configuration is:

1. CLI flags
2. Environment variables
3. Configuration file
4. Default values

This allows you to override specific settings without modifying the configuration file.
