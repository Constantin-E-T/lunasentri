# Machine Credential Lifecycle Management

This document describes how to manage machine credentials (API keys) in LunaSentri, including revocation and rotation capabilities.

## Overview

LunaSentri provides full control over machine credentials through:

1. **Enable/Disable**: Temporarily revoke access without deleting the machine
2. **API Key Rotation**: Generate fresh API keys while invalidating old ones
3. **Audit Trail**: Track all API key versions and their status

## Architecture

### Database Schema

#### `machines` table
- `is_enabled` (BOOLEAN): Controls whether the machine can authenticate
- Default: `true` for new machines

#### `machine_api_keys` table
Stores all API key versions for audit and rotation:
- `id`: Primary key
- `machine_id`: Foreign key to machines table
- `api_key_hash`: SHA-256 hash of the API key
- `created_at`: When the key was generated
- `revoked_at`: When the key was revoked (NULL if active)

### Authentication Flow

When an agent attempts to authenticate:

1. Hash the provided API key
2. Look up the key in `machine_api_keys` table
3. Check that `revoked_at IS NULL` (key is active)
4. Check that the machine's `is_enabled = true`
5. If both conditions pass, authentication succeeds

Error responses:
- `"invalid API key"`: Key not found or revoked
- `"machine disabled"`: Machine exists but is disabled

## API Endpoints

All endpoints require session authentication.

### Disable Machine

**POST** `/machines/:id/disable`

Disables a machine, preventing all API key authentication.

**Example:**
```bash
curl -X POST https://api.lunasentri.example.com/machines/123/disable \
  -H "Cookie: session=..." \
  -H "Content-Type: application/json"
```

**Response:**
```json
{
  "message": "Machine disabled successfully"
}
```

**Effects:**
- Machine's `is_enabled` set to `false`
- All agent authentication attempts will fail with "machine disabled" error
- Machine remains in the system with all its data
- Can be re-enabled at any time

### Enable Machine

**POST** `/machines/:id/enable`

Re-enables a previously disabled machine.

**Example:**
```bash
curl -X POST https://api.lunasentri.example.com/machines/123/enable \
  -H "Cookie: session=..." \
  -H "Content-Type: application/json"
```

**Response:**
```json
{
  "message": "Machine enabled successfully"
}
```

**Effects:**
- Machine's `is_enabled` set to `true`
- Agent can authenticate with existing valid API key
- No new API key is generated

### Rotate API Key

**POST** `/machines/:id/rotate-key`

Generates a new API key and revokes all previous keys.

**Example:**
```bash
curl -X POST https://api.lunasentri.example.com/machines/123/rotate-key \
  -H "Cookie: session=..." \
  -H "Content-Type: application/json"
```

**Response:**
```json
{
  "message": "API key rotated successfully",
  "api_key": "NEW_PLAINTEXT_API_KEY_SHOWN_ONCE"
}
```

**⚠️ Important:**
- The new API key is returned **only once** in the response
- Save it immediately - it cannot be retrieved later
- Old keys are immediately revoked
- Update the agent configuration with the new key

**Effects:**
- All existing API keys for this machine are revoked (`revoked_at` timestamp set)
- New API key is generated and stored (hashed)
- New plaintext key returned in response (only time it's visible)
- Old keys will no longer authenticate

## Agent Configuration

After rotating a key, update the agent configuration file:

```yaml
# /etc/lunasentri/agent.yaml
server_url: https://api.lunasentri.example.com
api_key: "NEW_API_KEY_HERE"  # Update this value
machine_name: "production-server-01"
interval: 30
```

Restart the agent:
```bash
sudo systemctl restart lunasentri-agent
```

## Common Workflows

### Temporary Suspension

If you need to temporarily stop an agent from reporting:

1. Disable the machine via UI or API
2. Agent will receive 403 errors when attempting to post metrics
3. Re-enable when ready to resume monitoring

### Security Incident Response

If an API key is compromised:

1. **Immediately** disable the machine to stop all authentication
2. Rotate the API key to get a new one
3. Securely deliver the new key to the legitimate server
4. Update agent configuration
5. Re-enable the machine

### Scheduled Key Rotation

For enhanced security, rotate keys periodically:

```bash
#!/bin/bash
# Rotate keys for a machine and update agent config

MACHINE_ID=123
AGENT_CONFIG="/etc/lunasentri/agent.yaml"

# Rotate key via API
RESPONSE=$(curl -s -X POST https://api.lunasentri.example.com/machines/$MACHINE_ID/rotate-key \
  -H "Cookie: session=$SESSION_COOKIE")

# Extract new key
NEW_KEY=$(echo $RESPONSE | jq -r '.api_key')

# Update config (requires root)
sudo sed -i "s/^api_key:.*/api_key: \"$NEW_KEY\"/" $AGENT_CONFIG

# Restart agent
sudo systemctl restart lunasentri-agent

echo "✓ API key rotated successfully"
```

## Frontend Integration

The `/machines` page should display:
- Enable/disable toggle for each machine
- "Rotate Key" button with confirmation dialog
- Last key rotation timestamp
- Current enabled status

### Example UI Flow

```jsx
// TODO: Implement in apps/web-next/app/machines/page.tsx

function MachineRow({ machine }) {
  const handleRotateKey = async () => {
    const confirmed = await showDialog({
      title: "Rotate API Key?",
      message: "This will invalidate the current key. You'll need to update the agent configuration.",
    });
    
    if (confirmed) {
      const response = await fetch(`/machines/${machine.id}/rotate-key`, {
        method: 'POST',
      });
      const data = await response.json();
      
      // Show new key in a modal
      showKeyModal(data.api_key);
    }
  };

  return (
    <tr>
      <td>{machine.name}</td>
      <td>
        <Toggle 
          checked={machine.is_enabled}
          onChange={() => toggleMachineStatus(machine.id)}
        />
      </td>
      <td>
        <Button onClick={handleRotateKey}>
          Rotate Key
        </Button>
      </td>
    </tr>
  );
}
```

## Security Considerations

### API Key Storage

- **Server**: Only SHA-256 hashes are stored in the database
- **Agent**: Keys stored in configuration files with restricted permissions (0600)
- **Transit**: Keys transmitted only once during registration/rotation via HTTPS

### Audit Trail

All key rotations and enable/disable actions are:
- Logged to the server logs with timestamps and user IDs
- Trackable via `machine_api_keys` table history
- Available for security audits

### Best Practices

1. **Rotate keys regularly**: At least every 90 days for production systems
2. **Immediate rotation**: If a server is decommissioned or compromised
3. **Secure delivery**: Use secure channels to deliver new keys to servers
4. **Monitor logs**: Watch for authentication failures that might indicate key issues
5. **Document rotations**: Keep a record of when and why keys were rotated

## Troubleshooting

### Agent Returns 401 Unauthorized

**Possible causes:**
1. API key was rotated but agent config not updated
2. Machine is disabled
3. Network/firewall issues preventing authentication

**Resolution:**
1. Check machine status in UI (enabled/disabled)
2. Verify API key in agent config matches current key
3. Check server logs for specific error message
4. If needed, rotate key and update agent

### Machine Shows Offline After Key Rotation

**Cause:** Agent still using old (revoked) key

**Resolution:**
1. Update `/etc/lunasentri/agent.yaml` with new key
2. Restart agent: `sudo systemctl restart lunasentri-agent`
3. Check agent logs: `sudo journalctl -u lunasentri-agent -f`

### Cannot Disable Machine

**Cause:** Insufficient permissions

**Resolution:**
- Ensure you're logged in as the machine owner
- Only the user who created the machine can manage it

## Migration Notes

For existing deployments upgrading to this version:

1. Migration `015_machine_credential_management` runs automatically on startup
2. Existing API keys are migrated to `machine_api_keys` table
3. All machines default to `is_enabled = true`
4. No agent configuration changes required
5. New endpoints are immediately available

## Related Documentation

- [Agent Installation](./agent/INSTALLATION.md) - Initial agent setup
- [Heartbeat Implementation](../features/HEARTBEAT_IMPLEMENTATION.md) - Machine monitoring
- [Deployment Guide](./deployment/DEPLOYMENT.md) - Server configuration
