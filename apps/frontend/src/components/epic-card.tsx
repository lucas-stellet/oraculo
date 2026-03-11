import {
  Card,
  CardAction,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
import type { EpicSummary, Phase } from "@/lib/types";

const phaseColors: Record<Phase, string> = {
  discover: "bg-amber-500/10 text-amber-500 border-amber-500/20",
  plan: "bg-blue-500/10 text-blue-500 border-blue-500/20",
  execute: "bg-purple-500/10 text-purple-500 border-purple-500/20",
  validate: "bg-green-500/10 text-green-500 border-green-500/20",
};

interface EpicCardProps {
  epic: EpicSummary;
  onOpen: (epic: EpicSummary) => void;
}

export function EpicCard({ epic, onOpen }: EpicCardProps) {
  const completionPct =
    epic.task_count > 0
      ? Math.round((epic.completed_task_count / epic.task_count) * 100)
      : 0;

  return (
    <Card className="flex flex-col">
      <CardHeader>
        <CardTitle className="text-lg font-semibold">{epic.name}</CardTitle>
        <CardAction>
          <Badge variant="outline" className={phaseColors[epic.phase]}>
            {epic.phase}
          </Badge>
        </CardAction>
      </CardHeader>
      <CardContent className="flex flex-1 flex-col gap-4">
        <p className="text-sm text-muted-foreground line-clamp-2">
          {epic.description}
        </p>

        <div className="space-y-2">
          <Progress value={completionPct} className="h-2" />
          <div className="flex items-center justify-between text-xs text-muted-foreground font-[family-name:var(--font-mono)]">
            <span>{epic.story_count} stories</span>
            <span>
              {epic.completed_task_count}/{epic.task_count} tasks
            </span>
            <span>{completionPct}%</span>
          </div>
        </div>

        <Button className="mt-auto w-full" onClick={() => onOpen(epic)}>
          Open
        </Button>
      </CardContent>
    </Card>
  );
}
