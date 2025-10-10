# Machine Credential Management Implementation Summary

**Date**: October 10, 2025  
**Status**: ✅ Complete  
**PR**: #[TBD]

## Overview

Implemented full machine credential lifecycle management, giving users complete control over machine API keys through disable/enable and rotation capabilities.

## What Was Implemented

### 1. Database Schema (Migration 015)

**New `machines` table fields:**
- `is_enabled` (BOOLEAN, default true): Controls machine authentication

**New `machine_api_keys` table:**
- `id`: Primary key
- `machine_id`: Foreign key to machines
- `api_key_hash`: SHA-256 hash of the API key
- `created_at`: Key generation timestamp
- `revoked_at`: Revocation timestamp (NULL if active)

**Indexes created:**
- `idx_machine_api_keys_machine_id`
- `idx_machine_api_keys_hash`
- `idx_machine_api_keys_revoked`
- `idx_machines_is_enabled`

**Migration features:**
- Automatically migrates existing API keys from `machines.api_key` to `machine_api_keys` table
- Backward compatible - existing agents continue working
- All machines default to `is_enabled = true`

### 2. Storage Layer (`internal/storage/machines.go`)

**New types:**
```go
type MachineAPIKey struct {
    ID         int
    MachineID  int
    APIKeyHash string
    CreatedAt  time.Time
    RevokedAt  *time.Time
}
```

**New methods:**
- `SetMachineEnabled(ctx, machineID, enabled)` - Enable/disable a machine
- `CreateMachineAPIKey(ctx, machineID, apiKeyHash)` - Create new API key entry
- `RevokeMachineAPIKey(ctx, keyID)` - Revoke a specific key
- `RevokeAllMachineAPIKeys(ctx, machineID)` - Revoke all keys for a machine
- `GetActiveAPIKeyForMachine(ctx, machineID)` - Get current active key
- `GetMachineAPIKeyByHash(ctx, apiKeyHash)` - Lookup key by hash
- `ListMachineAPIKeys(ctx, machineID)` - List all keys (active + revoked)

**Updated methods:**
- `CreateMachine()` - Now creates entry in `machine_api_keys` table
- `GetMachineByID()` - Now includes `is_enabled` field
- `GetMachineByAPIKey()` - Now queries `machine_api_keys` table with JOIN, only returns if key is not revoked

### 3. Service Layer (`internal/machines/service.go`)

**New methods:**
- `DisableMachine(ctx, machineID, userID)` - Disable a machine (checks ownership)
- `EnableMachine(ctx, machineID, userID)` - Re-enable a machine
- `RotateMachineAPIKey(ctx, machineID, userID)` - Generate new key, revoke old ones
- `GetMachineAPIKeyInfo(ctx, machineID, userID)` - Get key history for audit

**Updated methods:**
- `AuthenticateMachine()` - Now checks `is_enabled` flag and returns specific error messages:
  - `"machine disabled"` - Machine exists but is disabled
  - `"invalid API key"` - Key not found or revoked

### 4. HTTP Layer (`internal/http/agent_handlers.go`)

**New endpoints:**

1. **POST /machines/:id/disable**
   - Requires session auth
   - Disables machine (sets `is_enabled = false`)
   - Returns JSON success message

2. **POST /machines/:id/enable**
   - Requires session auth
   - Re-enables machine (sets `is_enabled = true`)
   - Returns JSON success message

3. **POST /machines/:id/rotate-key**
   - Requires session auth
   - Revokes all existing keys
   - Generates new API key
   - Returns new key in response (only time it's visible)

**Route registration:**
Updated `/machines/` handler to route to new endpoints based on URL suffix.

### 5. Tests

**Storage tests** (`internal/storage/machine_credentials_test.go`):
- ✅ CreateMachineWithAPIKey
- ✅ DisableMachine
- ✅ EnableMachine
- ✅ RotateAPIKey
- ✅ GetMachineByAPIKey (with revoked keys)
- ✅ GetActiveAPIKeyForMachine
- ✅ MachineAPIKeyMigration

**Service tests** (`internal/machines/service_credentials_test.go`):
- ✅ DisableMachine
- ✅ EnableMachine
- ✅ RotateMachineAPIKey
- ✅ RotateMachineAPIKeyUnauthorized
- ✅ AuthenticateMachineWithRevokedKey
- ✅ GetMachineAPIKeyInfo

**HTTP tests** (`internal/http/agent_credentials_test.go`):
- ✅ HandleDisableMachine
- ✅ HandleEnableMachine
- ✅ HandleRotateMachineAPIKey
- ✅ HandleDisableMachineUnauthorized
- ✅ AgentAuthenticationWithDisabledMachine

**Mock updates:**
- Updated all mock stores to implement new interface methods
- Files updated:
  - `internal/auth/service_test.go`
  - `internal/notifications/http_test.go`
  - `internal/notifications/telegram_http_test.go`
  - `internal/notifications/webhooks_test.go`

### 6. Documentation

**New documentation** (`docs/features/MACHINE_CREDENTIAL_LIFECYCLE.md`):
- Architecture overview
- Database schema details
- API endpoint specifications with examples
- Authentication flow diagram
- Common workflows (suspension, incident response, scheduled rotation)
- Agent configuration update procedures
- Security best practices
- Troubleshooting guide
- Frontend integration examples

**Updated documentation** (`docs/agent/INSTALLATION.md`):
- Added API Key Management section
- References to credential lifecycle documentation
- Security considerations for key storage

## Test Results

```
$ go test ./...
ok      github.com/Constantin-E-T/lunasentri/apps/api-go
ok      github.com/Constantin-E-T/lunasentri/apps/api-go/internal/alerts
ok      github.com/Constantin-E-T/lunasentri/apps/api-go/internal/auth
ok      github.com/Constantin-E-T/lunasentri/apps/api-go/internal/http
ok      github.com/Constantin-E-T/lunasentri/apps/api-go/internal/machines
ok      github.com/Constantin-E-T/lunasentri/apps/api-go/internal/notifications
ok      github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage
ok      github.com/Constantin-E-T/lunasentri/apps/api-go/internal/system
```

All tests pass ✅

## Files Changed

### New Files
- `apps/api-go/internal/storage/machine_credentials_test.go`
- `apps/api-go/internal/machines/service_credentials_test.go`
- `apps/api-go/internal/http/agent_credentials_test.go`
- `docs/features/MACHINE_CREDENTIAL_LIFECYCLE.md`

### Modified Files
- `apps/api-go/internal/storage/sqlite.go` - Added migration 015
- `apps/api-go/internal/storage/machines.go` - Added new types and methods
- `apps/api-go/internal/storage/interface.go` - Updated Store interface
- `apps/api-go/internal/machines/service.go` - Added credential management methods
- `apps/api-go/internal/http/agent_handlers.go` - Added new HTTP handlers
- `apps/api-go/internal/http/handlers.go` - Updated route registration
- `apps/api-go/internal/auth/service_test.go` - Updated mock store
- `apps/api-go/internal/notifications/http_test.go` - Updated mock store
- `apps/api-go/internal/notifications/telegram_http_test.go` - Updated mock store
- `apps/api-go/internal/notifications/webhooks_test.go` - Updated mock store
- `docs/agent/INSTALLATION.md` - Added credential management section

## API Examples

### Disable a machine
```bash
curl -X POST https://api.lunasentri.com/machines/123/disable \
  -H "Cookie: session=..." \
  -H "Content-Type: application/json"
```

### Enable a machine
```bash
curl -X POST https://api.lunasentri.com/machines/123/enable \
  -H "Cookie: session=..." \
  -H "Content-Type: application/json"
```

### Rotate API key
```bash
curl -X POST https://api.lunasentri.com/machines/123/rotate-key \
  -H "Cookie: session=..." \
  -H "Content-Type: application/json"
```

Response:
```json
{
  "message": "API key rotated successfully",
  "api_key": "NEW_API_KEY_SHOWN_ONCE"
}
```

## Security Features

1. **Granular Control**: Disable machines without deleting them
2. **Audit Trail**: All API key versions tracked in database
3. **One-time Key Display**: Rotated keys shown only once
4. **Ownership Enforcement**: Only machine owner can manage credentials
5. **Specific Error Messages**: Clear distinction between revoked keys and disabled machines

## Backward Compatibility

✅ **Fully backward compatible**
- Migration automatically runs on server startup
- Existing API keys migrated to new table structure
- No agent configuration changes required
- All existing functionality preserved

## Next Steps (Frontend)

TODO items for frontend implementation:

1. **Machines List Page** (`apps/web-next/app/machines/page.tsx`):
   - Add enable/disable toggle for each machine
   - Add "Rotate Key" button
   - Display last key rotation timestamp
   - Show enabled/disabled status badge

2. **Rotate Key Modal**:
   - Confirmation dialog before rotation
   - Display new API key (one-time view)
   - Copy-to-clipboard functionality
   - Warning about updating agent config

3. **Machine Details Page**:
   - Show API key history (created dates, revoked status)
   - Enable/disable controls
   - Key rotation button

## Deployment Notes

1. **Database Migration**: Runs automatically on server startup
2. **No Downtime Required**: Changes are backward compatible
3. **Agent Updates**: Not required - existing agents continue working
4. **API Changes**: Additive only - no breaking changes

## Related Issues

Closes #[TBD] - Machine credential revocation and rotation

## Acceptance Criteria

✅ All tests pass (`go test ./...`)  
✅ Migrations run automatically on startup  
✅ New endpoints return JSON with clear error messages  
✅ Disabled machines cannot post metrics (401/403)  
✅ Rotation response contains new plaintext key once  
✅ Documentation includes rotation workflow  

## Author

Implementation by GitHub Copilot  
Date: October 10, 2025
