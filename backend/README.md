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
  },
  "logger": {
  "log_dir": "./app/logs",
  "max_size": 512,
  "max_backups": 12,
  "max_age": 30,
  "compress": true
  },
  "cluster_health_check_interval_hours":6
}
```

### Notification Configuration

KubeCloud supports a separate notification configuration file to define how different types of notifications are handled. This allows you to customize which channels (UI, email) and severity levels are used for different notification types.

#### Notification Config File

By default, KubeCloud looks for a `notification-config.json` file in the current directory. You can specify a custom notification configuration file path using the `--notification_config_path` flag:

```bash
kubecloud --notification_config_path /path/to/notification-config.json
```

You can also set it via environment variable:

```bash
export KUBECLOUD_NOTIFICATION_CONFIG_PATH=/path/to/notification-config.json
```

Or include it in your main configuration file:

```json
{
  "notification_config_path": "./notification-config.json"
  // ... other config
}
```

#### Default Behavior

If no notification configuration file is provided, KubeCloud will use default settings:

- **All channels**: `["ui"]` (UI notifications only)
- **All severity levels**: `"info"`
- **All notification types**: Use the default settings unless specifically overridden

#### Notification Config Structure

The notification configuration file should follow this structure:

```json
{
  "template_types": {
    "deployment": {
      "default": {
        "channels": ["ui"],
        "severity": "info"
      },
      "by_status": {
        "started": {
          "channels": ["ui"],
          "severity": "info"
        },
        "succeeded": {
          "channels": ["ui", "email"],
          "severity": "success"
        },
        "failed": {
          "channels": ["ui", "email"],
          "severity": "error"
        },
        "deleted": {
          "channels": ["ui"],
          "severity": "warning"
        }
      }
    },
    "billing": {
      "default": {
        "channels": ["ui"],
        "severity": "info"
      },
      "by_status": {
        "funds_succeeded": {
          "channels": ["ui", "email"],
          "severity": "success"
        },
        "funds_failed": {
          "channels": ["ui", "email"],
          "severity": "error"
        }
      }
    },
    "user": {
      "default": {
        "channels": ["ui"],
        "severity": "info"
      },
      "by_status": {
        "password_changed": {
          "channels": ["ui", "email"],
          "severity": "success"
        }
      }
    }
  }
}
```

#### Configuration Options

- **Channels**: Available channels are `["ui", "email"]`
- **Severity Levels**: Available severities are `"info"`, `"success"`, `"warning"`, `"error"`
- **Template Types**: Currently supported types are `deployment`, `billing`, and `user`
- **Status Overrides**: You can override the default behavior for specific statuses within each template type

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
