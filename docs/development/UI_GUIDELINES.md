# UI/UX Development Guidelines

## Core Principles

### 1. Never Use Native Browser Dialogs

**❌ NEVER DO THIS:**

```typescript
// Bad - breaks UI consistency
if (confirm("Are you sure?")) {
  deleteItem();
}

alert("Something went wrong!");
```

**✅ ALWAYS DO THIS:**

```typescript
// Good - use Dialog components
const [showConfirm, setShowConfirm] = useState(false);

<Dialog open={showConfirm} onOpenChange={setShowConfirm}>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>Confirm Action</DialogTitle>
      <DialogDescription>
        Are you sure you want to proceed?
      </DialogDescription>
    </DialogHeader>
    <DialogFooter>
      <Button variant="outline" onClick={() => setShowConfirm(false)}>
        Cancel
      </Button>
      <Button variant="destructive" onClick={handleConfirm}>
        Confirm
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
```

### 2. Destructive Action Confirmations

All destructive actions (delete, revoke, disable) MUST:

1. **Use a Dialog component** with clear messaging
2. **State what will be deleted/changed** (include the item name)
3. **Warn if action is irreversible** ("This cannot be undone")
4. **Show what data will be lost** (e.g., "This will remove all associated metrics")
5. **Use destructive variant** for the confirm button
6. **Disable actions during processing** with loading state

**Example Pattern:**

```typescript
const [showDeleteModal, setShowDeleteModal] = useState(false);
const [deletingItem, setDeletingItem] = useState<{id: number, name: string} | null>(null);
const [deleting, setDeleting] = useState(false);

const handleDelete = (id: number, name: string) => {
  setDeletingItem({ id, name });
  setShowDeleteModal(true);
};

const handleConfirmDelete = async () => {
  if (!deletingItem) return;
  
  setDeleting(true);
  try {
    await deleteItem(deletingItem.id);
    await refresh();
    setShowDeleteModal(false);
    setDeletingItem(null);
  } catch (err) {
    // Keep modal open on error
    console.error(err);
  } finally {
    setDeleting(false);
  }
};
```

### 3. Error Handling

**Display errors inline**, not as alerts:

```typescript
// State
const [error, setError] = useState<string | null>(null);

// In your form/modal
{error && (
  <div className="rounded-lg bg-destructive/20 border border-destructive/30 p-3 text-destructive text-sm">
    {error}
  </div>
)}

// Error handling
try {
  await submitForm();
} catch (err) {
  setError(err instanceof Error ? err.message : "Operation failed");
  // Don't close modal - let user see error and retry
}
```

### 4. Loading States

Always show loading states during async operations:

```typescript
<Button onClick={handleSubmit} disabled={loading}>
  {loading ? "Saving..." : "Save"}
</Button>
```

### 5. Data Field Consistency

**Rule:** If a field exists in the database, it MUST have UI for:

1. **Creating** the field (in registration/create forms)
2. **Viewing** the field (in list/detail views, conditionally if optional)
3. **Editing** the field (in update/edit forms)

**Example - Machine with description field:**

- ✅ Create: Registration form has description input
- ✅ View: Machine card displays description if present
- ✅ Edit: Edit modal has description input

## Component Usage

### Dialog Components

Import from: `@/components/ui/dialog`

```typescript
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
```

### Button Variants

- `default` - Primary actions
- `destructive` - Delete, revoke, dangerous actions
- `outline` - Cancel, secondary actions
- `ghost` - Tertiary actions
- `link` - Link-style buttons

### Form Validation

- Show validation errors inline, not in alerts
- Disable submit buttons when form is invalid
- Use `disabled={!fieldValue.trim()}` for required fields

## Examples

### Delete Confirmation (Complete Example)

See `apps/web-next/app/machines/page.tsx` for the complete implementation of:

- Delete confirmation dialog
- Loading states
- Error handling
- State management

### Edit Form (Complete Example)

See `apps/web-next/app/machines/page.tsx` for the complete implementation of:

- Edit modal with all fields
- Inline error display
- Loading states
- Form validation

## Quick Reference

| Action Type | Dialog Type | Button Variant | Loading Text |
|-------------|-------------|----------------|--------------|
| Delete | Confirmation | destructive | "Deleting..." |
| Save/Update | Form | default | "Saving..." |
| Cancel | - | outline | - |
| Create | Form | default | "Creating..." |

## Migration Checklist

When you find code using native dialogs:

- [ ] Replace `confirm()` with Dialog component
- [ ] Replace `alert()` with inline error display
- [ ] Add loading state
- [ ] Add error state
- [ ] Test error scenarios
- [ ] Ensure modal stays open on error
