"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useSession } from "@/lib/useSession";
import { useMachines } from "@/lib/useMachines";
import { useAlertsWithNotifications } from "@/lib/alerts";
import { Navbar } from "@/components/Navbar";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { CheckIcon, CopyIcon } from "lucide-react";

interface RegisterFormData {
  name: string;
  hostname?: string;
  description?: string;
}

export default function MachinesPage() {
  const router = useRouter();
  const { status, user, logout } = useSession();
  const { machines, loading, error, refresh, register } = useMachines();
  const { events, newAlertsCount } = useAlertsWithNotifications(10);

  // Count unacknowledged events from live data
  const unacknowledgedCount =
    events?.filter((e) => !e.acknowledged).length || 0;

  const [showRegisterModal, setShowRegisterModal] = useState(false);
  const [registering, setRegistering] = useState(false);
  const [machineName, setMachineName] = useState("");
  const [machineHostname, setMachineHostname] = useState("");
  const [machineDescription, setMachineDescription] = useState("");
  const [apiKey, setApiKey] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const [registerError, setRegisterError] = useState<string | null>(null);

  // Edit modal state
  const [showEditModal, setShowEditModal] = useState(false);
  const [editingMachine, setEditingMachine] = useState<any>(null);
  const [editName, setEditName] = useState("");
  const [editHostname, setEditHostname] = useState("");
  const [editDescription, setEditDescription] = useState("");
  const [updating, setUpdating] = useState(false);
  const [updateError, setUpdateError] = useState<string | null>(null);

  // Delete modal state
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [deletingMachine, setDeletingMachine] = useState<{
    id: number;
    name: string;
  } | null>(null);
  const [deleting, setDeleting] = useState(false);

  // Redirect to login if unauthenticated
  useEffect(() => {
    if (status === "unauthenticated") {
      router.push("/login");
    }
  }, [status, router]);

  // Show loading state while checking authentication
  if (status === "loading" || loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-muted-foreground animate-pulse">Loading...</div>
      </div>
    );
  }

  // Don't render if not authenticated (will redirect)
  if (status !== "authenticated") {
    return null;
  }

  const handleRegister = async () => {
    if (!machineName.trim()) {
      setRegisterError("Machine name is required");
      return;
    }

    setRegistering(true);
    setRegisterError(null);

    try {
      const formData: RegisterFormData = {
        name: machineName.trim(),
      };
      if (machineHostname.trim()) {
        formData.hostname = machineHostname.trim();
      }
      if (machineDescription.trim()) {
        formData.description = machineDescription.trim();
      }

      const response = await register(formData);

      // Store API key to show to user
      setApiKey(response.api_key);
      setMachineName("");
      setMachineHostname("");
      setMachineDescription("");
    } catch (err) {
      setRegisterError(
        err instanceof Error ? err.message : "Failed to register machine"
      );
    } finally {
      setRegistering(false);
    }
  };

  const handleCopyApiKey = () => {
    if (apiKey) {
      navigator.clipboard.writeText(apiKey);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  const handleCloseModal = () => {
    setShowRegisterModal(false);
    setApiKey(null);
    setMachineName("");
    setMachineHostname("");
    setRegisterError(null);
    setCopied(false);
  };

  const handleEditMachine = (machine: any) => {
    setEditingMachine(machine);
    setEditName(machine.name);
    setEditHostname(machine.hostname || "");
    setEditDescription(machine.description || "");
    setShowEditModal(true);
  };

  const handleUpdateMachine = async () => {
    if (!editingMachine || !editName.trim()) {
      return;
    }

    setUpdating(true);
    setUpdateError(null);
    try {
      const { updateMachine } = await import("@/lib/api");
      await updateMachine(editingMachine.id, {
        name: editName.trim(),
        hostname: editHostname.trim() || undefined,
        description: editDescription.trim() || undefined,
      });

      // Refresh machines list
      await refresh();
      setShowEditModal(false);
      setEditingMachine(null);
    } catch (err) {
      console.error("Failed to update machine:", err);
      setUpdateError(
        err instanceof Error ? err.message : "Failed to update machine"
      );
    } finally {
      setUpdating(false);
    }
  };

  const handleDeleteMachine = (id: number, name: string) => {
    setDeletingMachine({ id, name });
    setShowDeleteModal(true);
  };

  const handleConfirmDelete = async () => {
    if (!deletingMachine) return;

    setDeleting(true);
    try {
      const { deleteMachine } = await import("@/lib/api");
      await deleteMachine(deletingMachine.id);

      // Refresh machines list
      await refresh();
      setShowDeleteModal(false);
      setDeletingMachine(null);
    } catch (err) {
      console.error("Failed to delete machine:", err);
      // Keep modal open to show error
    } finally {
      setDeleting(false);
    }
  };

  const getStatusColor = (status: string) => {
    return status === "online"
      ? "bg-green-500/20 text-green-500 border-green-500/30"
      : "bg-gray-500/20 text-gray-500 border-gray-500/30";
  };

  const formatLastSeen = (lastSeen: string) => {
    if (!lastSeen) return "Never";
    const date = new Date(lastSeen);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);

    if (diffMins < 1) return "Just now";
    if (diffMins < 60) return `${diffMins}m ago`;
    const diffHours = Math.floor(diffMins / 60);
    if (diffHours < 24) return `${diffHours}h ago`;
    const diffDays = Math.floor(diffHours / 24);
    return `${diffDays}d ago`;
  };

  return (
    <div className="min-h-screen">
      <Navbar
        user={user}
        unacknowledgedCount={unacknowledgedCount}
        newAlertsCount={newAlertsCount}
        onLogout={logout}
      />

      {/* Main Content */}
      <div className="px-4 py-8">
        <div className="max-w-7xl mx-auto">
          <div className="space-y-6">
            {/* Page Header */}
            <div className="flex justify-between items-center">
              <div>
                <h1 className="text-4xl font-semibold tracking-wide text-primary">
                  Machines
                </h1>
                <p className="text-muted-foreground mt-2">
                  Manage your monitored machines
                </p>
              </div>
              <Button
                onClick={() => setShowRegisterModal(true)}
                className="bg-primary text-primary-foreground hover:bg-primary/90"
              >
                Register Machine
              </Button>
            </div>

            {/* Error State */}
            {error && (
              <div className="rounded-lg bg-destructive/20 border border-destructive/30 p-4 text-destructive">
                {error}
              </div>
            )}

            {/* Machines List */}
            {!machines || machines.length === 0 ? (
              <div className="rounded-lg border border-border/40 bg-card/40 backdrop-blur-sm p-12 text-center">
                <div className="text-6xl mb-4">🖥️</div>
                <h3 className="text-xl font-semibold mb-2">No machines yet</h3>
                <p className="text-muted-foreground mb-6">
                  Register your first machine to start monitoring
                </p>
                <Button
                  onClick={() => setShowRegisterModal(true)}
                  className="bg-primary text-primary-foreground hover:bg-primary/90"
                >
                  Register Machine
                </Button>
              </div>
            ) : (
              <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
                {machines.map((machine) => (
                  <div
                    key={machine.id}
                    className="rounded-lg border border-border/40 bg-card/40 backdrop-blur-sm p-6 hover:border-border/60 transition-all"
                  >
                    <div className="flex justify-between items-start mb-3">
                      <h3 className="font-semibold text-lg">{machine.name}</h3>
                      <Badge className={getStatusColor(machine.status)}>
                        {machine.status}
                      </Badge>
                    </div>
                    <div className="space-y-2 text-sm text-muted-foreground">
                      {machine.description && (
                        <div>
                          <span className="font-medium">Description:</span>{" "}
                          {machine.description}
                        </div>
                      )}
                      <div>
                        <span className="font-medium">Hostname:</span>{" "}
                        {machine.hostname || "—"}
                      </div>
                      <div>
                        <span className="font-medium">Last seen:</span>{" "}
                        {formatLastSeen(machine.last_seen)}
                      </div>
                      <div>
                        <span className="font-medium">Registered:</span>{" "}
                        {new Date(machine.created_at).toLocaleDateString()}
                      </div>
                    </div>

                    {/* Action Buttons */}
                    <div className="mt-4 pt-4 border-t border-border/20 flex gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        className="flex-1"
                        onClick={() => handleEditMachine(machine)}
                      >
                        Edit
                      </Button>
                      <Button
                        variant="destructive"
                        size="sm"
                        className="flex-1"
                        onClick={() =>
                          handleDeleteMachine(machine.id, machine.name)
                        }
                      >
                        Delete
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Registration Dialog */}
      <Dialog open={showRegisterModal} onOpenChange={handleCloseModal}>
        <DialogContent className="sm:max-w-[600px]">
          <DialogHeader>
            <DialogTitle>
              {apiKey
                ? "Machine Registered Successfully!"
                : "Register New Machine"}
            </DialogTitle>
            <DialogDescription>
              {apiKey
                ? "Save this API key - it will only be shown once"
                : "Provide a name and optional hostname for your machine"}
            </DialogDescription>
          </DialogHeader>

          {apiKey ? (
            /* Show API Key */
            <div className="space-y-4">
              <div className="rounded-lg bg-yellow-500/10 border border-yellow-500/30 p-4 text-yellow-500">
                <div className="flex items-start gap-2">
                  <span className="text-xl">⚠️</span>
                  <div className="flex-1 text-sm">
                    <strong>Important:</strong> This API key will only be shown
                    once. Save it securely - you'll need it to configure the
                    agent.
                  </div>
                </div>
              </div>

              <div className="space-y-2">
                <Label>API Key</Label>
                <div className="flex gap-2">
                  <Input
                    value={apiKey}
                    readOnly
                    className="font-mono text-sm"
                  />
                  <Button onClick={handleCopyApiKey} variant="outline">
                    {copied ? (
                      <CheckIcon className="h-4 w-4" />
                    ) : (
                      <CopyIcon className="h-4 w-4" />
                    )}
                  </Button>
                </div>
              </div>

              <div className="rounded-lg bg-muted/50 p-4 text-sm space-y-2">
                <div className="font-medium">Next steps:</div>
                <ol className="list-decimal list-inside space-y-1 text-muted-foreground">
                  <li>Install the LunaSentri agent on your machine</li>
                  <li>Configure it with this API key</li>
                  <li>Start the agent to begin sending metrics</li>
                </ol>
              </div>
            </div>
          ) : (
            /* Registration Form */
            <div className="space-y-4">
              {registerError && (
                <div className="rounded-lg bg-destructive/20 border border-destructive/30 p-3 text-destructive text-sm">
                  {registerError}
                </div>
              )}

              <div className="space-y-2">
                <Label htmlFor="name">Machine Name *</Label>
                <Input
                  id="name"
                  placeholder="e.g., production-server-1"
                  value={machineName}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                    setMachineName(e.target.value)
                  }
                  disabled={registering}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="hostname">Hostname (optional)</Label>
                <Input
                  id="hostname"
                  placeholder="e.g., web-1.example.com"
                  value={machineHostname}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                    setMachineHostname(e.target.value)
                  }
                  disabled={registering}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="description">Description (optional)</Label>
                <Input
                  id="description"
                  placeholder="e.g., Main production server"
                  value={machineDescription}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                    setMachineDescription(e.target.value)
                  }
                  disabled={registering}
                />
              </div>
            </div>
          )}

          <DialogFooter>
            {apiKey ? (
              <Button onClick={handleCloseModal} className="w-full">
                Done
              </Button>
            ) : (
              <>
                <Button
                  variant="outline"
                  onClick={handleCloseModal}
                  disabled={registering}
                >
                  Cancel
                </Button>
                <Button onClick={handleRegister} disabled={registering}>
                  {registering ? "Registering..." : "Register"}
                </Button>
              </>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Edit Machine Dialog */}
      <Dialog open={showEditModal} onOpenChange={setShowEditModal}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>Edit Machine</DialogTitle>
            <DialogDescription>Update machine details</DialogDescription>
          </DialogHeader>

          {updateError && (
            <div className="rounded-lg bg-destructive/20 border border-destructive/30 p-3 text-destructive text-sm">
              {updateError}
            </div>
          )}

          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="edit-name">Name *</Label>
              <Input
                id="edit-name"
                placeholder="e.g., Production Server"
                value={editName}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                  setEditName(e.target.value)
                }
                disabled={updating}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="edit-hostname">Hostname</Label>
              <Input
                id="edit-hostname"
                placeholder="e.g., web-1.example.com"
                value={editHostname}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                  setEditHostname(e.target.value)
                }
                disabled={updating}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="edit-description">Description</Label>
              <Input
                id="edit-description"
                placeholder="Optional description"
                value={editDescription}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                  setEditDescription(e.target.value)
                }
                disabled={updating}
              />
            </div>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowEditModal(false)}
              disabled={updating}
            >
              Cancel
            </Button>
            <Button
              onClick={handleUpdateMachine}
              disabled={updating || !editName.trim()}
            >
              {updating ? "Updating..." : "Save Changes"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog open={showDeleteModal} onOpenChange={setShowDeleteModal}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>Delete Machine</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete &quot;{deletingMachine?.name}
              &quot;? This will remove all associated metrics and cannot be
              undone.
            </DialogDescription>
          </DialogHeader>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowDeleteModal(false)}
              disabled={deleting}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleConfirmDelete}
              disabled={deleting}
            >
              {deleting ? "Deleting..." : "Delete"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
