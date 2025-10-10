# Phase 4: Dashboard Integration - Implementation Complete

## Overview

Phase 4 successfully integrates machine selection into the main dashboard, allowing users to view metrics, system information, and alerts scoped to a selected machine. The implementation includes persistent machine selection, visual indicators, and comprehensive testing.

## Completed Components

### 1. Core Hooks

#### `useMachineSelection` Hook

**Location**: `apps/web-next/lib/useMachineSelection.ts`

**Features:**

- **localStorage Persistence**: Selected machine ID persists across page reloads
- **Cross-tab Sync**: Changes to selection in one tab sync to others via storage events
- **SSR Safety**: Guards against server-side rendering issues with typeof window checks
- **Null-safe**: Handles "no selection" state gracefully
- **Hydration Protection**: Prevents SSR/client mismatch with hydration flag

**API:**

```typescript
interface UseMachineSelectionReturn {
  selectedMachineId: number | null;
  selectMachine: (machineId: number | null) => void;
  clearSelection: () => void;
}
```

**Storage Key**: `lunasentri_selected_machine_id`

#### Updated `useMetrics` Hook

**Location**: `apps/web-next/lib/useMetrics.ts`

**Changes:**

- Added `machineId` to `UseMetricsOptions` interface
- WebSocket URL construction includes `?machine_id=` query parameter
- Polling function passes `machineId` to `fetchMetrics()`
- useEffect dependency array includes `machineId` for auto-refetch

#### Updated `useSystemInfo` Hook

**Location**: `apps/web-next/lib/useSystemInfo.ts`

**Changes:**

- Added `UseSystemInfoOptions` interface with optional `machineId`
- Function signature updated to accept options parameter
- `fetchData` passes `machineId` to `fetchSystemInfo()`
- useEffect dependency array includes `machineId` for auto-refetch

### 2. UI Components

#### `MachineSelector` Component

**Location**: `apps/web-next/components/MachineSelector.tsx`

**Features:**

- Dropdown selector showing all registered machines
- Status badges: `online` (green), `offline` (red), `never_seen` (gray)
- Shows machine name, hostname, and last seen timestamp
- Empty state with "Register Machine" CTA linking to `/machines`
- "Manage Machines" link in dropdown footer
- Closes on outside clicks (click-away listener)
- Responsive design with fixed positioning

**Props:**

```typescript
interface MachineSelectorProps {
  machines: Machine[];
  selectedMachineId: number | null;
  onSelectMachine: (machineId: number) => void;
}
```

#### Updated `MetricsCard` Component

**Location**: `apps/web-next/components/MetricsCard.tsx`

**Changes:**

- Added `MetricsCardProps` interface with optional `machineId`
- Passes `machineId` to `useMetrics()` hook
- Shows "Dev Only" badge when `machineId` is undefined (localhost fallback)
- Badge styling: yellow border, small text, positioned next to "Orbital Feed" label

#### Updated `SystemInfoCard` Component

**Location**: `apps/web-next/components/dashboard/SystemInfoCard.tsx`

**Changes:**

- Added `SystemInfoCardProps` interface with optional `machineId`
- Passes `machineId` to `useSystemInfo()` hook
- Shows "Dev Only" badge when `machineId` is undefined (localhost fallback)
- Badge positioned in card header next to title

### 3. Dashboard Integration

**Location**: `apps/web-next/app/page.tsx`

**Changes:**

- Imports `MachineSelector`, `useMachines`, and `useMachineSelection` hooks
- Added `MachineSelector` component to header (next to LunaSentri logo)
- Passes `selectedMachineId` to both `MetricsCard` and `SystemInfoCard`
- Converts `selectedMachineId` from `number | null` to `number | undefined` for component props

**Layout:**

```
Header: [Logo] [MachineSelector] ... [Navigation] [User Menu]
Body: [MetricsCard] [SystemInfoCard] [ActiveAlertsCard]
```

### 4. API Updates

**Location**: `apps/web-next/lib/api.ts`

**Updated Functions:**

- `fetchMetrics(machineId?: number)` - Adds `?machine_id=` query param when provided
- `fetchSystemInfo(machineId?: number)` - Adds `?machine_id=` query param when provided

**Note:** Backend now requires `machine_id` for production requests; localhost fallback remains available only when `LOCAL_HOST_METRICS=true`.

## Testing

### Test Coverage

**Test File**: `apps/web-next/__tests__/useMachineSelection.test.ts`

**Test Suites:** 6 suites, 15 tests total

- ✅ Initialization (4 tests)
- ✅ selectMachine (3 tests)
- ✅ clearSelection (2 tests)
- ✅ Cross-tab synchronization (4 tests)
- ✅ SSR safety (1 test)
- ✅ Cleanup (1 test)

**All tests passing** ✓

**Coverage includes:**

- Loading persisted values from localStorage
- Handling invalid/malformed localStorage data
- Updating selection and persisting to localStorage
- Clearing selection
- Cross-tab synchronization via storage events
- Ignoring unrelated storage keys
- Handling invalid values in storage events
- SSR environment compatibility
- Proper cleanup of event listeners on unmount

## User Experience

### Empty State (No Machines)

- MachineSelector shows "No machines registered"
- Dropdown contains "Register Machine" link to `/machines` page
- MetricsCard and SystemInfoCard show "Dev Only" badge
- Backend falls back to LOCAL_HOST_METRICS (if enabled)

### Machine Selected

- MachineSelector shows selected machine name and status
- MetricsCard fetches data with `?machine_id=X` parameter
- SystemInfoCard fetches data with `?machine_id=X` parameter
- No "Dev Only" badge visible
- Selection persists across page reloads
- Selection syncs across browser tabs

### Visual Indicators

- **Status Badges**: Color-coded machine status in dropdown
- **Dev Only Badge**: Yellow badge when using localhost fallback
- **WebSocket Indicator**: Animated dot showing live connection
- **Last Seen**: Relative timestamps for machine status

## Backend Integration Status

## Backend Integration

- Go handlers (`/metrics`, `/ws`, `/system/info`) enforce `machine_id`, check ownership, and fall back to dev metrics only when `LOCAL_HOST_METRICS=true`.
- Agents may post `system_info` + `uptime_s`; values persist on `machines` and `metrics_history` for dashboard display.
- SQLite migration `012_machine_system_info` must run before deploying this release.
The following backend changes are required for full functionality:

1. **`/metrics` Endpoint** (`apps/api-go/internal/http/metrics_handler.go`)
   - Parse `machine_id` query parameter
   - Fetch metrics from `metrics_history` table for specified machine
   - Fall back to LOCAL_HOST_METRICS if machine_id not provided

2. **`/ws` WebSocket Endpoint** (`apps/api-go/internal/http/websocket_handler.go`)
   - Parse `machine_id` query parameter from upgrade request
   - Stream metrics for specified machine
   - Handle machine-specific subscriptions

3. **`/system/info` Endpoint** (`apps/api-go/internal/http/system_handler.go`)
   - Parse `machine_id` query parameter
   - Fetch system info for specified machine (if stored)
   - Return machine-specific OS, CPU, memory details

4. **Alert Evaluation** (Future Phase)
   - Filter alert events by machine_id
   - Scope alert rules to specific machines
   - Update notification context with machine info

### Current Behavior

- Frontend sends `?machine_id=X` parameters
- Backend ignores these parameters
- Backend returns localhost metrics (if LOCAL_HOST_METRICS enabled)
- "Dev Only" badge correctly indicates localhost fallback

## File Changes Summary

### New Files Created

- `apps/web-next/lib/useMachineSelection.ts` (92 lines)
- `apps/web-next/components/MachineSelector.tsx` (135 lines)
- `apps/web-next/__tests__/useMachineSelection.test.ts` (224 lines)

### Files Modified

- `apps/web-next/lib/api.ts` - Added machineId parameters to fetchMetrics/fetchSystemInfo
- `apps/web-next/lib/useMetrics.ts` - Added machineId support to hook
- `apps/web-next/lib/useSystemInfo.ts` - Added machineId support to hook
- `apps/web-next/components/MetricsCard.tsx` - Added machineId prop, Dev Only badge
- `apps/web-next/components/dashboard/SystemInfoCard.tsx` - Added machineId prop, Dev Only badge
- `apps/web-next/app/page.tsx` - Integrated MachineSelector component

## Architecture Decisions

### 1. localStorage Over URL State

**Decision**: Use localStorage instead of URL query parameters for machine selection.

**Rationale:**

- Better UX: Selection persists across navigation
- Simpler implementation: No URL management needed
- Cross-tab sync: Storage events enable tab synchronization
- Clean URLs: Dashboard URL remains `/` without query clutter

### 2. Optional machineId Parameter

**Decision**: Make `machineId` optional in hooks and API functions.

**Rationale:**

- Backward compatibility: Existing code works without changes
- Graceful degradation: Falls back to localhost metrics
- Clear intent: `undefined` means "localhost", explicit ID means "specific machine"
- Visual feedback: "Dev Only" badge indicates fallback mode

### 3. SSR Safety with typeof window

**Decision**: Use `typeof window === "undefined"` checks instead of try-catch.

**Rationale:**

- More explicit: Clear intent of SSR vs browser environment
- No error suppression: Catch blocks hide unexpected errors
- Performance: Direct check faster than exception handling
- Standard pattern: Widely used in Next.js ecosystem

### 4. Hydration Flag

**Decision**: Don't expose `selectedMachineId` until after hydration.

**Rationale:**

- Prevents SSR mismatch errors
- Ensures consistent initial render between server/client
- localStorage only accessible on client, server always returns null
- Avoids React hydration warnings

### 5. "Dev Only" Badge Placement

**Decision**: Show badge inline with card title/label.

**Rationale:**

- Non-intrusive: Small, subtle indicator
- Contextual: Appears directly on affected components
- Clear meaning: Indicates localhost fallback mode
- Easy to remove: When backend implements machine_id support

## Next Steps

### Immediate

1. Update backend to process `machine_id` query parameters
2. Implement machine-scoped metrics retrieval
3. Add machine context to alert evaluations
4. Test end-to-end flow with real machine data

### Future Enhancements

1. Add machine filtering to alerts page
2. Show machine-specific alert history
3. Add "Compare Machines" view (multi-select)
4. Machine groups/tags for bulk selection
5. Default machine preference in user settings

## Migration Notes

### For Developers

**Breaking Changes:** None - all changes are backward compatible.

**New Dependencies:** None - uses existing React, UI components.

**Environment Variables:** None required.

### For Users

**Action Required:** None - feature works out of the box.

**New Workflow:**

1. Register machines via `/machines` page (Phase 3)
2. Select machine from header dropdown on dashboard
3. View machine-specific metrics and system info
4. Selection persists across sessions

**Fallback Behavior:** If no machines registered, dashboard shows localhost data (if available) with "Dev Only" badge.

## Performance Considerations

### localStorage Operations

- Read: Once on mount (cached in state)
- Write: Only on selection change (not frequent)
- Impact: Negligible - synchronous API, fast key-value lookup

### Cross-tab Sync

- Event listener: Lightweight, only fires on storage changes
- Handler: Quick parseInt, no heavy computation
- Impact: Minimal - event-driven, no polling

### Re-renders

- Selection change triggers: MetricsCard, SystemInfoCard refetch
- WebSocket reconnection: With new machine_id parameter
- Impact: Expected behavior, user-initiated action

### Bundle Size

- New hook: ~3KB uncompressed
- MachineSelector component: ~4KB uncompressed
- Total increase: ~7KB (minified + gzipped: ~2KB)

## Security Considerations

### localStorage Security

- **XSS Protection**: Next.js CSP headers prevent script injection
- **Data Type**: Only stores numeric machine ID (no sensitive data)
- **Validation**: parseInt with NaN check prevents malicious values
- **Scope**: localStorage scoped to origin, no cross-site leakage

### API Parameter Validation

- **Backend Responsibility**: Go server must validate machine_id parameter
- **Authorization**: Backend must verify user has access to requested machine
- **SQL Injection**: Use parameterized queries for machine_id lookups
- **Rate Limiting**: Apply per-user limits to prevent enumeration

## Documentation Updates

### User Documentation

- ✅ This file documents Phase 4 implementation
- ⏳ Update main README.md with machine selection feature
- ⏳ Add screenshots/GIFs to docs/features/

### Developer Documentation

- ✅ JSDoc comments in all new/modified functions
- ✅ TypeScript interfaces fully documented
- ✅ Test file demonstrates usage patterns
- ⏳ Update AGENT_GUIDELINES.md with machine selection UX

## Success Metrics

### Implementation Completeness

- ✅ All frontend components implemented
- ✅ All hooks updated with machineId support
- ✅ All tests passing (15/15)
- ✅ TypeScript compilation clean
- ⏳ Backend integration (Phase 2 follow-up)

### Code Quality

- ✅ SSR-safe implementation
- ✅ Comprehensive error handling
- ✅ Type-safe interfaces
- ✅ Well-documented code
- ✅ Consistent styling

### User Experience

- ✅ Persistent selection across reloads
- ✅ Cross-tab synchronization
- ✅ Clear empty state handling
- ✅ Visual status indicators
- ✅ Responsive dropdown UI

---

**Phase 4 Status**: ✅ **Frontend Complete** | ⏳ **Backend Pending**

**Next Phase**: Backend implementation of machine_id query parameter support (Phase 2 follow-up)
