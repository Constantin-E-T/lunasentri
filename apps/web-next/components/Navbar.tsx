"use client";

import Link from "next/link";
import { useState } from "react";
import { Badge } from "@/components/ui/badge";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Bell,
  Menu,
  Settings,
  User,
  LogOut,
  Users,
  Server,
  AlertTriangle,
  MessageSquare,
} from "lucide-react";

interface NavbarProps {
  user: {
    email: string;
    is_admin: boolean;
  } | null;
  unacknowledgedCount: number;
  newAlertsCount: number;
  onLogout: () => void;
}

export function Navbar({
  user,
  unacknowledgedCount,
  newAlertsCount,
  onLogout,
}: NavbarProps) {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const hasAlerts = unacknowledgedCount > 0 || newAlertsCount > 0;

  return (
    <nav className="sticky top-0 z-50 border-b border-border/40 bg-card/60 backdrop-blur-xl">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          {/* Logo */}
          <Link href="/" className="flex items-center gap-3 group">
            <span className="text-3xl group-hover:scale-110 transition-transform duration-200">
              🌙
            </span>
            <span className="font-bold text-xl tracking-tight bg-gradient-to-r from-primary to-blue-400 bg-clip-text text-transparent">
              LunaSentri
            </span>
          </Link>

          {/* Desktop Navigation */}
          <div className="hidden md:flex items-center gap-2">
            {/* Alerts - Prominent if active */}
            {hasAlerts ? (
              <Link
                href="/alerts"
                className="flex items-center gap-2 px-4 py-2 rounded-lg bg-destructive/10 border border-destructive/30 text-destructive hover:bg-destructive/20 transition-all duration-200"
              >
                <Bell className="w-4 h-4" />
                <span className="font-medium">Alerts</span>
                <div className="flex items-center gap-1">
                  {unacknowledgedCount > 0 && (
                    <Badge
                      variant="destructive"
                      className="text-xs px-2 py-0.5"
                    >
                      {unacknowledgedCount}
                    </Badge>
                  )}
                  {newAlertsCount > 0 && (
                    <Badge className="text-xs px-2 py-0.5 bg-blue-500">
                      {newAlertsCount}
                    </Badge>
                  )}
                </div>
              </Link>
            ) : null}

            {/* Machines */}
            <Link
              href="/machines"
              className="flex items-center gap-2 px-4 py-2 rounded-lg text-muted-foreground hover:text-foreground hover:bg-accent/10 transition-all duration-200"
            >
              <Server className="w-4 h-4" />
              <span>Machines</span>
            </Link>

            {/* Admin Menu */}
            {user?.is_admin && (
              <DropdownMenu>
                <DropdownMenuTrigger className="flex items-center gap-2 px-4 py-2 rounded-lg text-muted-foreground hover:text-foreground hover:bg-accent/10 transition-all duration-200 outline-none">
                  <Menu className="w-4 h-4" />
                  <span>Admin</span>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="w-48">
                  <DropdownMenuItem asChild>
                    <Link
                      href="/alerts"
                      className="flex items-center gap-2 cursor-pointer"
                    >
                      <AlertTriangle className="w-4 h-4" />
                      <span>Alert Rules</span>
                    </Link>
                  </DropdownMenuItem>
                  <DropdownMenuItem asChild>
                    <Link
                      href="/users"
                      className="flex items-center gap-2 cursor-pointer"
                    >
                      <Users className="w-4 h-4" />
                      <span>Manage Users</span>
                    </Link>
                  </DropdownMenuItem>
                  <DropdownMenuItem asChild>
                    <Link
                      href="/notifications/telegram"
                      className="flex items-center gap-2 cursor-pointer"
                    >
                      <MessageSquare className="w-4 h-4" />
                      <span>Telegram Alerts</span>
                    </Link>
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            )}
          </div>

          {/* User Menu */}
          <div className="hidden md:flex items-center gap-3">
            <DropdownMenu>
              <DropdownMenuTrigger className="flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-accent/10 transition-all duration-200 outline-none">
                <div className="w-8 h-8 rounded-full bg-gradient-to-br from-primary/20 to-blue-500/20 border border-primary/30 flex items-center justify-center">
                  <User className="w-4 h-4 text-primary" />
                </div>
                <span className="text-sm text-muted-foreground max-w-[150px] truncate">
                  {user?.email}
                </span>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-56">
                <div className="px-2 py-1.5 text-sm">
                  <p className="font-medium">{user?.email}</p>
                  <p className="text-xs text-muted-foreground mt-0.5">
                    {user?.is_admin ? "Administrator" : "User"}
                  </p>
                </div>
                <DropdownMenuSeparator />
                <DropdownMenuItem asChild>
                  <Link
                    href="/settings"
                    className="flex items-center gap-2 cursor-pointer"
                  >
                    <Settings className="w-4 h-4" />
                    <span>Settings</span>
                  </Link>
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  onClick={onLogout}
                  className="flex items-center gap-2 text-destructive cursor-pointer"
                >
                  <LogOut className="w-4 h-4" />
                  <span>Logout</span>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>

          {/* Mobile Menu Button */}
          <button
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            className="md:hidden p-2 rounded-lg hover:bg-accent/10 transition-all duration-200"
          >
            <Menu className="w-6 h-6" />
          </button>
        </div>

        {/* Mobile Menu */}
        {mobileMenuOpen && (
          <div className="md:hidden border-t border-border/40 py-4 space-y-2">
            {hasAlerts && (
              <Link
                href="/alerts"
                onClick={() => setMobileMenuOpen(false)}
                className="flex items-center justify-between px-4 py-3 rounded-lg bg-destructive/10 border border-destructive/30 text-destructive"
              >
                <div className="flex items-center gap-2">
                  <Bell className="w-4 h-4" />
                  <span className="font-medium">Alerts</span>
                </div>
                <div className="flex items-center gap-1">
                  {unacknowledgedCount > 0 && (
                    <Badge variant="destructive" className="text-xs">
                      {unacknowledgedCount}
                    </Badge>
                  )}
                  {newAlertsCount > 0 && (
                    <Badge className="text-xs bg-blue-500">
                      {newAlertsCount}
                    </Badge>
                  )}
                </div>
              </Link>
            )}

            <Link
              href="/machines"
              onClick={() => setMobileMenuOpen(false)}
              className="flex items-center gap-2 px-4 py-3 rounded-lg text-muted-foreground hover:text-foreground hover:bg-accent/10"
            >
              <Server className="w-4 h-4" />
              <span>Machines</span>
            </Link>

            {user?.is_admin && (
              <>
                <div className="px-4 py-2 text-xs font-semibold text-muted-foreground">
                  ADMIN
                </div>
                <Link
                  href="/alerts"
                  onClick={() => setMobileMenuOpen(false)}
                  className="flex items-center gap-2 px-4 py-3 rounded-lg text-muted-foreground hover:text-foreground hover:bg-accent/10"
                >
                  <AlertTriangle className="w-4 h-4" />
                  <span>Alert Rules</span>
                </Link>
                <Link
                  href="/users"
                  onClick={() => setMobileMenuOpen(false)}
                  className="flex items-center gap-2 px-4 py-3 rounded-lg text-muted-foreground hover:text-foreground hover:bg-accent/10"
                >
                  <Users className="w-4 h-4" />
                  <span>Manage Users</span>
                </Link>
                <Link
                  href="/notifications/telegram"
                  onClick={() => setMobileMenuOpen(false)}
                  className="flex items-center gap-2 px-4 py-3 rounded-lg text-muted-foreground hover:text-foreground hover:bg-accent/10"
                >
                  <MessageSquare className="w-4 h-4" />
                  <span>Telegram Alerts</span>
                </Link>
              </>
            )}

            <div className="border-t border-border/40 mt-2 pt-2">
              <Link
                href="/settings"
                onClick={() => setMobileMenuOpen(false)}
                className="flex items-center gap-2 px-4 py-3 rounded-lg text-muted-foreground hover:text-foreground hover:bg-accent/10"
              >
                <Settings className="w-4 h-4" />
                <span>Settings</span>
              </Link>
              <div className="px-4 py-2 text-sm text-muted-foreground">
                {user?.email}
              </div>
              <button
                onClick={() => {
                  setMobileMenuOpen(false);
                  onLogout();
                }}
                className="flex items-center gap-2 w-full px-4 py-3 rounded-lg text-destructive hover:bg-destructive/10"
              >
                <LogOut className="w-4 h-4" />
                <span>Logout</span>
              </button>
            </div>
          </div>
        )}
      </div>
    </nav>
  );
}
