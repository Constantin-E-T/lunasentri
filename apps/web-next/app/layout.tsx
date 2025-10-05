import type { Metadata } from "next";
import "./globals.css";

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
      <body className="antialiased" suppressHydrationWarning={true}>
        {children}
      </body>
    </html>
  );
}
