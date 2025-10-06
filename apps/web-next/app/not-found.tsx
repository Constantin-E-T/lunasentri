"use client";

import Link from "next/link";

export default function NotFoundPage() {
  return (
    <div className="min-h-screen flex items-center justify-center">
      <div className="text-center space-y-6 max-w-lg">
        <div className="space-y-2">
          <h1 className="text-6xl font-bold text-primary">🌙</h1>
          <h2 className="text-4xl font-semibold text-foreground">404</h2>
          <p className="text-muted-foreground text-lg">
            This page couldn't be found in the lunar system.
          </p>
        </div>

        <Link
          href="/"
          className="inline-flex items-center gap-2 rounded-full bg-primary/20 border border-primary/30 px-6 py-3 text-primary-foreground transition-all duration-200 hover:bg-primary/30 hover:-translate-y-0.5"
        >
          <span>🏠</span>
          Return to Dashboard
        </Link>
      </div>
    </div>
  );
}
