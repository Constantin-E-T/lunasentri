import type { Metadata } from "next";
import "./globals.css";
import { Toaster } from "@/components/ui/toaster";

export const metadata: Metadata = {
  title: "LunaSentri - Lightweight Server Monitoring",
  description: "Lightweight server monitoring dashboard for solo developers",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body
        className="min-h-screen bg-background text-foreground antialiased relative overflow-x-hidden"
        suppressHydrationWarning={true}
      >
        <div className="pointer-events-none fixed inset-0 -z-10 bg-[radial-gradient(circle_at_top,_rgba(104,154,255,0.35),_transparent_60%),radial-gradient(circle_at_bottom,_rgba(24,72,181,0.25),_transparent_65%)]" />
        <div className="pointer-events-none fixed inset-0 -z-20 opacity-20 mix-blend-screen bg-[url('https://www.transparenttextures.com/patterns/asfalt-dark.png')]" />
        <div className="relative z-0">{children}</div>
        <Toaster />
      </body>
    </html>
  );
}
