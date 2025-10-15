# MyceliumCloud

MyceliumCloud is a CLI tool that helps you deploy and manage Kubernetes clusters on the decentralized TFGrid.

## Configuration

MyceliumCloud supports configuration through environment variables, CLI flags, and configuration files.

### Configuration File

By default, MyceliumCloud looks for a `config.json` file in the current directory. You can specify a custom configuration file path using the `--config` or `-c` flag:

```bash
myceliumcloud --config /path/to/config.json
```

The configuration file should be in JSON format. Example:

check the config [example](./config-example.json)

### Notification Configuration

MyceliumCloud supports a separate notification configuration file to define how different types of notifications are handled. This allows you to customize which channels (UI, email) and severity levels are used for different notification types.

#### Notification Config File

By default, MyceliumCloud looks for a `notification-config.json` file in the current directory. You can specify a custom notification configuration file path using the `--notification_config_path` flag:

```bash
myceliumcloud --notification_config_path /path/to/notification-config.json
```

You can also set it via environment variable:

```bash
export MYCELIUMCLOUD_NOTIFICATION_CONFIG_PATH=/path/to/notification-config.json
```

Or include it in your main configuration file:

```json
{
  "notification_config_path": "./notification-config.json"
  // ... other config
}
```

#### Default Behavior

If no notification configuration file is provided, MyceliumCloud will use default settings:

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
