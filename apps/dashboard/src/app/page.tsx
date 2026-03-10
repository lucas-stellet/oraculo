"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { EpicCard } from "@/components/epic-card";
import { CreateEpicDialog } from "@/components/create-epic-dialog";
import type { EpicSummary } from "@/lib/types";

export default function LandingPage() {
  const router = useRouter();
  const [epics, setEpics] = useState<EpicSummary[]>([]);

  useEffect(() => {
    fetch("/api/epics")
      .then((res) => res.json())
      .then((data) => setEpics(data ?? []))
      .catch(() => setEpics([]));
  }, []);

  async function handleCreate(name: string, description: string) {
    const res = await fetch("/api/epics", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name, description }),
    });
    if (!res.ok) return;
    // Refetch all epics to get accurate summary data
    const updated = await fetch("/api/epics").then((r) => r.json());
    setEpics(updated ?? []);
  }

  function handleOpen(epic: EpicSummary) {
    router.push(`/epics/${encodeURIComponent(epic.name)}`);
  }

  return (
    <div className="mx-auto max-w-6xl px-8 py-12">
      {/* Hero */}
      <section className="mb-12">
        <h1 className="text-4xl font-bold tracking-tight font-[family-name:var(--font-display)]">
          Oraculo
        </h1>
        <p className="mt-3 max-w-2xl text-muted-foreground">
          Mission Control for AI-agent-driven development. Monitor your agents,
          track task progress, review artifacts, and approve decisions — all from
          one place.
        </p>
      </section>

      {/* Epic Grid */}
      <section>
        <h2 className="mb-6 text-2xl font-semibold font-[family-name:var(--font-display)]">
          Epics
        </h2>

        <div className="grid gap-6 sm:grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
          {epics.map((epic) => (
            <EpicCard key={epic.id} epic={epic} onOpen={handleOpen} />
          ))}
          <CreateEpicDialog onCreate={handleCreate} />
        </div>
      </section>
    </div>
  );
}
