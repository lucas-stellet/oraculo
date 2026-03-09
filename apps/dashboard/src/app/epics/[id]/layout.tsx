"use client";

import { use } from "react";
import { Sidebar } from "@/components/sidebar";
import { getEpic } from "@/lib/mock-data";

export default function EpicLayout({
  children,
  params,
}: {
  children: React.ReactNode;
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);
  const epic = getEpic();

  return (
    <div className="flex h-screen overflow-hidden bg-[#020617]">
      <Sidebar epicId={Number(id)} epicName={epic.name} />
      <div className="flex-1 overflow-y-auto">{children}</div>
    </div>
  );
}
