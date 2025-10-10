"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useSession } from "@/lib/useSession";
import { useAlertsWithNotifications } from "@/lib/alerts";
import { Navbar } from "@/components/Navbar";
import {
  listUsers,
  createUser,
  deleteUser,
  type User,
  type CreateUserResponse,
} from "@/lib/api";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";

export default function UsersPage() {
  const router = useRouter();
  const { status, user: currentUser, logout } = useSession();
  const { events, newAlertsCount } = useAlertsWithNotifications(10);
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isCreating, setIsCreating] = useState(false);
  const [isDeletingId, setIsDeletingId] = useState<number | null>(null);

  // Count unacknowledged events
  const unacknowledgedCount =
    events?.filter((e) => !e.acknowledged).length || 0;

  // Form state
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [formError, setFormError] = useState<string | null>(null);
  const [tempPassword, setTempPassword] = useState<string | null>(null);

  // Delete confirmation state
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [deletingUser, setDeletingUser] = useState<{
    id: number;
    email: string;
  } | null>(null);

  // Redirect to login if unauthenticated
  useEffect(() => {
    if (status === "unauthenticated") {
      router.push("/login");
    }
  }, [status, router]);

  // Fetch users
  const fetchUsers = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await listUsers();
      setUsers(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load users");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (status === "authenticated") {
      fetchUsers();
    }
  }, [status]);

  const handleCreateUser = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!email) {
      setFormError("Email is required");
      return;
    }

    try {
      setIsCreating(true);
      setFormError(null);
      setTempPassword(null);

      const response: CreateUserResponse = await createUser({
        email,
        password: password || undefined,
      });

      // If a temp password was generated, show it
      if (response.temp_password) {
        setTempPassword(response.temp_password);
      }

      // Reset form
      setEmail("");
      setPassword("");

      // Refresh user list
      await fetchUsers();
    } catch (err) {
      setFormError(
        err instanceof Error ? err.message : "Failed to create user"
      );
    } finally {
      setIsCreating(false);
    }
  };

  const handleDeleteUser = (userId: number, userEmail: string) => {
    setDeletingUser({ id: userId, email: userEmail });
    setShowDeleteModal(true);
  };

  const handleConfirmDelete = async () => {
    if (!deletingUser) return;

    try {
      setIsDeletingId(deletingUser.id);
      await deleteUser(deletingUser.id);
      await fetchUsers();
      setShowDeleteModal(false);
      setDeletingUser(null);
    } catch (err) {
      // Keep modal open on error so user can see what went wrong
      console.error("Failed to delete user:", err);
    } finally {
      setIsDeletingId(null);
    }
  };

  // Show loading state while checking authentication
  if (status === "loading") {
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

  return (
    <div className="min-h-screen">
      <Navbar
        user={currentUser}
        unacknowledgedCount={unacknowledgedCount}
        newAlertsCount={newAlertsCount}
        onLogout={logout}
      />

      {/* Main Content */}
      <div className="max-w-7xl mx-auto px-4 py-8">
        <div className="mb-8">
          <h1 className="text-4xl font-bold text-primary mb-2">Manage Users</h1>
          <p className="text-muted-foreground">
            Create and manage user accounts
          </p>
        </div>

        {/* Add User Form */}
        <div className="bg-card/70 backdrop-blur-xl border border-border/30 rounded-lg p-6 mb-8">
          <h2 className="text-2xl font-semibold text-card-foreground mb-4">
            Add User
          </h2>

          {tempPassword && (
            <div className="mb-4 p-4 bg-chart-1/10 border border-chart-1/30 rounded-lg">
              <div className="flex justify-between items-start">
                <div>
                  <p className="text-chart-1 font-semibold mb-2">
                    Temporary Password Generated
                  </p>
                  <p className="text-sm text-chart-1/80 mb-2">
                    Share this password with the new user. They should change it
                    after first login.
                  </p>
                  <code className="block bg-background/80 px-3 py-2 rounded text-chart-1 text-sm break-all">
                    {tempPassword}
                  </code>
                </div>
                <button
                  onClick={() => setTempPassword(null)}
                  className="text-chart-1 hover:text-chart-1/80 text-xl leading-none"
                  aria-label="Dismiss"
                >
                  ×
                </button>
              </div>
            </div>
          )}

          <form onSubmit={handleCreateUser} className="space-y-4">
            <div>
              <label
                htmlFor="email"
                className="block text-sm font-medium text-card-foreground mb-2"
              >
                Email
              </label>
              <input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full px-4 py-2 bg-background/50 border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                placeholder="user@example.com"
                disabled={isCreating}
              />
            </div>

            <div>
              <label
                htmlFor="password"
                className="block text-sm font-medium text-card-foreground mb-2"
              >
                Password{" "}
                <span className="text-muted-foreground">
                  (optional - will be auto-generated if empty)
                </span>
              </label>
              <input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full px-4 py-2 bg-background/50 border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                placeholder="Leave empty to auto-generate"
                disabled={isCreating}
              />
            </div>

            {formError && (
              <div className="p-3 bg-destructive/10 border border-destructive/30 rounded-lg text-destructive text-sm">
                {formError}
              </div>
            )}

            <button
              type="submit"
              disabled={isCreating}
              className="px-6 py-2 bg-primary hover:bg-primary/90 disabled:bg-muted disabled:cursor-not-allowed text-primary-foreground rounded-lg transition-all"
            >
              {isCreating ? "Adding..." : "Add User"}
            </button>
          </form>
        </div>

        {/* Users Table */}
        <div className="bg-card/70 backdrop-blur-xl border border-border/30 rounded-lg overflow-hidden">
          <div className="px-6 py-4 border-b border-border/30">
            <h2 className="text-2xl font-semibold text-card-foreground">
              Users
            </h2>
          </div>

          {loading ? (
            <div className="px-6 py-8 text-center text-muted-foreground">
              Loading users...
            </div>
          ) : error ? (
            <div className="px-6 py-8 text-center text-destructive">
              {error}
            </div>
          ) : users.length === 0 ? (
            <div className="px-6 py-8 text-center text-muted-foreground">
              No users found
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-border/30 text-left">
                    <th className="px-6 py-3 text-sm font-semibold text-card-foreground">
                      Email
                    </th>
                    <th className="px-6 py-3 text-sm font-semibold text-card-foreground">
                      Created
                    </th>
                    <th className="px-6 py-3 text-sm font-semibold text-card-foreground">
                      Actions
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {users.map((user) => (
                    <tr
                      key={user.id}
                      className="border-b border-border/30 hover:bg-background/20"
                    >
                      <td className="px-6 py-4 text-card-foreground">
                        {user.email}
                      </td>
                      <td className="px-6 py-4 text-muted-foreground">
                        {user.created_at
                          ? new Date(user.created_at).toLocaleDateString()
                          : "N/A"}
                      </td>
                      <td className="px-6 py-4">
                        {currentUser?.id === user.id ? (
                          <span className="text-sm text-muted-foreground">
                            Current User
                          </span>
                        ) : (
                          <button
                            onClick={() =>
                              handleDeleteUser(user.id, user.email)
                            }
                            disabled={isDeletingId === user.id}
                            className="text-sm text-destructive hover:text-destructive/80 disabled:text-muted-foreground disabled:cursor-not-allowed transition-colors"
                          >
                            {isDeletingId === user.id
                              ? "Deleting..."
                              : "Delete"}
                          </button>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>

      {/* Delete Confirmation Dialog */}
      <Dialog open={showDeleteModal} onOpenChange={setShowDeleteModal}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>Delete User</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete user &quot;{deletingUser?.email}
              &quot;? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowDeleteModal(false)}
              disabled={isDeletingId !== null}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleConfirmDelete}
              disabled={isDeletingId !== null}
            >
              {isDeletingId !== null ? "Deleting..." : "Delete User"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
