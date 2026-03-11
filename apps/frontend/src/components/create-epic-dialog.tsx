"use client";

import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Plus } from "lucide-react";

interface CreateEpicDialogProps {
  onCreate: (name: string, description: string) => void;
}

export function CreateEpicDialog({ onCreate }: CreateEpicDialogProps) {
  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!name.trim()) return;
    onCreate(name.trim(), description.trim());
    setName("");
    setDescription("");
    setOpen(false);
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger
        render={
          <button className="group flex min-h-[220px] w-full flex-col items-center justify-center gap-3 rounded-xl border border-dashed border-border/60 bg-transparent transition-colors hover:border-primary/40 hover:bg-primary/5 cursor-pointer" />
        }
      >
        <div className="flex h-12 w-12 items-center justify-center rounded-full border border-dashed border-border/60 transition-colors group-hover:border-primary/40 group-hover:bg-primary/10">
          <Plus className="h-5 w-5 text-muted-foreground transition-colors group-hover:text-primary" />
        </div>
        <span className="text-sm font-medium text-muted-foreground transition-colors group-hover:text-foreground">
          New Epic
        </span>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create Epic</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <label htmlFor="epic-name" className="text-sm font-medium">
              Name
            </label>
            <Input
              id="epic-name"
              placeholder="e.g. user-authentication"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>
          <div className="space-y-2">
            <label htmlFor="epic-desc" className="text-sm font-medium">
              Description
            </label>
            <Textarea
              id="epic-desc"
              placeholder="Brief description of the epic..."
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={3}
            />
          </div>
          <div className="flex justify-end gap-3">
            <Button
              type="button"
              variant="ghost"
              onClick={() => setOpen(false)}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={!name.trim()}>
              Create
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
