"use client";

import Navbar from "@/components/Navbar";
import { syncAuthCookie } from "@/lib/auth";
import { useEffect, useRef } from "react";

export default function ProtectedShell({
  children,
}: {
  children: React.ReactNode;
}) {
  const syncedRef = useRef(false);

  useEffect(() => {
    if (syncedRef.current) return;
    syncedRef.current = true;
    syncAuthCookie();
  }, []);

  return (
    <div className="flex min-h-full flex-col bg-grid">
      <Navbar />
      <main className="mx-auto w-full max-w-7xl flex-1 px-6 py-8">
        {children}
      </main>
    </div>
  );
}
