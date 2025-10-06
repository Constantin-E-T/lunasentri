'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useSession } from '@/lib/useSession';
import { listUsers, createUser, deleteUser, type User, type CreateUserResponse } from '@/lib/api';

export default function UsersPage() {
  const router = useRouter();
  const { status, user: currentUser } = useSession();
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isCreating, setIsCreating] = useState(false);
  const [isDeletingId, setIsDeletingId] = useState<number | null>(null);

  // Form state
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [formError, setFormError] = useState<string | null>(null);
  const [tempPassword, setTempPassword] = useState<string | null>(null);

  // Redirect to login if unauthenticated
  useEffect(() => {
    if (status === 'unauthenticated') {
      router.push('/login');
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
      setError(err instanceof Error ? err.message : 'Failed to load users');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (status === 'authenticated') {
      fetchUsers();
    }
  }, [status]);

  const handleCreateUser = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!email) {
      setFormError('Email is required');
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
      setEmail('');
      setPassword('');

      // Refresh user list
      await fetchUsers();
    } catch (err) {
      setFormError(err instanceof Error ? err.message : 'Failed to create user');
    } finally {
      setIsCreating(false);
    }
  };

  const handleDeleteUser = async (userId: number, userEmail: string) => {
    if (!confirm(`Are you sure you want to delete user ${userEmail}?`)) {
      return;
    }

    try {
      setIsDeletingId(userId);
      await deleteUser(userId);
      await fetchUsers();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to delete user');
    } finally {
      setIsDeletingId(null);
    }
  };

  // Show loading state while checking authentication
  if (status === 'loading') {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-900 to-slate-800">
        <div className="text-slate-400">Loading...</div>
      </div>
    );
  }

  // Don't render if not authenticated (will redirect)
  if (status !== 'authenticated') {
    return null;
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 to-slate-800">
      {/* Header */}
      <div className="border-b border-slate-700/50 bg-slate-800/30 backdrop-blur-sm">
        <div className="max-w-7xl mx-auto px-4 py-4 flex justify-between items-center">
          <div className="flex items-center gap-3">
            <Link href="/" className="flex items-center gap-3 hover:opacity-80 transition-opacity">
              <span className="text-2xl">🌙</span>
              <span className="text-white font-semibold">LunaSentri</span>
            </Link>
          </div>
          <div className="flex items-center gap-4">
            <span className="text-sm text-slate-400">
              {currentUser?.email}
            </span>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="max-w-7xl mx-auto px-4 py-8">
        <div className="mb-8">
          <h1 className="text-4xl font-bold text-white mb-2">Manage Users</h1>
          <p className="text-slate-400">Create and manage user accounts</p>
        </div>

        {/* Add User Form */}
        <div className="bg-slate-800/30 backdrop-blur-sm border border-slate-700/50 rounded-lg p-6 mb-8">
          <h2 className="text-2xl font-semibold text-white mb-4">Add User</h2>

          {tempPassword && (
            <div className="mb-4 p-4 bg-blue-900/30 border border-blue-700/50 rounded-lg">
              <div className="flex justify-between items-start">
                <div>
                  <p className="text-blue-200 font-semibold mb-2">Temporary Password Generated</p>
                  <p className="text-sm text-blue-300 mb-2">
                    Share this password with the new user. They should change it after first login.
                  </p>
                  <code className="block bg-slate-900/50 px-3 py-2 rounded text-blue-200 text-sm break-all">
                    {tempPassword}
                  </code>
                </div>
                <button
                  onClick={() => setTempPassword(null)}
                  className="text-blue-300 hover:text-blue-100 text-xl leading-none"
                  aria-label="Dismiss"
                >
                  ×
                </button>
              </div>
            </div>
          )}

          <form onSubmit={handleCreateUser} className="space-y-4">
            <div>
              <label htmlFor="email" className="block text-sm font-medium text-slate-300 mb-2">
                Email
              </label>
              <input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full px-4 py-2 bg-slate-900/50 border border-slate-700 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="user@example.com"
                disabled={isCreating}
              />
            </div>

            <div>
              <label htmlFor="password" className="block text-sm font-medium text-slate-300 mb-2">
                Password <span className="text-slate-500">(optional - will be auto-generated if empty)</span>
              </label>
              <input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full px-4 py-2 bg-slate-900/50 border border-slate-700 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="Leave empty to auto-generate"
                disabled={isCreating}
              />
            </div>

            {formError && (
              <div className="p-3 bg-red-900/30 border border-red-700/50 rounded-lg text-red-200 text-sm">
                {formError}
              </div>
            )}

            <button
              type="submit"
              disabled={isCreating}
              className="px-6 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-slate-700 disabled:cursor-not-allowed text-white rounded-lg transition-colors"
            >
              {isCreating ? 'Adding...' : 'Add User'}
            </button>
          </form>
        </div>

        {/* Users Table */}
        <div className="bg-slate-800/30 backdrop-blur-sm border border-slate-700/50 rounded-lg overflow-hidden">
          <div className="px-6 py-4 border-b border-slate-700/50">
            <h2 className="text-2xl font-semibold text-white">Users</h2>
          </div>

          {loading ? (
            <div className="px-6 py-8 text-center text-slate-400">
              Loading users...
            </div>
          ) : error ? (
            <div className="px-6 py-8 text-center text-red-400">
              {error}
            </div>
          ) : users.length === 0 ? (
            <div className="px-6 py-8 text-center text-slate-400">
              No users found
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-slate-700/50 text-left">
                    <th className="px-6 py-3 text-sm font-semibold text-slate-300">Email</th>
                    <th className="px-6 py-3 text-sm font-semibold text-slate-300">Created</th>
                    <th className="px-6 py-3 text-sm font-semibold text-slate-300">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {users.map((user) => (
                    <tr key={user.id} className="border-b border-slate-700/50 hover:bg-slate-700/20">
                      <td className="px-6 py-4 text-white">{user.email}</td>
                      <td className="px-6 py-4 text-slate-400">
                        {user.created_at ? new Date(user.created_at).toLocaleDateString() : 'N/A'}
                      </td>
                      <td className="px-6 py-4">
                        {currentUser?.id === user.id ? (
                          <span className="text-sm text-slate-500">Current User</span>
                        ) : (
                          <button
                            onClick={() => handleDeleteUser(user.id, user.email)}
                            disabled={isDeletingId === user.id}
                            className="text-sm text-red-400 hover:text-red-300 disabled:text-slate-600 disabled:cursor-not-allowed transition-colors"
                          >
                            {isDeletingId === user.id ? 'Deleting...' : 'Delete'}
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
    </div>
  );
}
