import { STATUS_LABELS, PRIORITY_LABELS } from "@/lib/types";
import type { Task } from "@/lib/types";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from "@/components/ui/select";

interface Props {
  task: Task;
  canDelete: boolean;
  onStatusChange: (taskId: string, status: string) => void;
  onDelete: (taskId: string) => void;
  onEdit: (task: Task) => void;
}

const priorityColors: Record<string, string> = {
  high: "bg-red-100 text-red-800",
  medium: "bg-yellow-100 text-yellow-800",
  low: "bg-green-100 text-green-800",
};

export default function TaskCard({ task, canDelete, onStatusChange, onDelete, onEdit }: Props) {
  return (
    <div className="border rounded-lg p-4 bg-card">
      <div className="flex items-start justify-between gap-2">
        <div className="flex-1 min-w-0">
          <h3
            className="font-medium text-foreground cursor-pointer hover:underline truncate"
            onClick={() => onEdit(task)}
          >
            {task.title}
          </h3>
          {task.description && (
            <p className="text-sm text-muted-foreground mt-1 line-clamp-2">
              {task.description}
            </p>
          )}
        </div>
        {canDelete && (
          <Button
            variant="ghost"
            size="sm"
            className="text-destructive hover:text-destructive shrink-0"
            onClick={() => onDelete(task.id)}
          >
            Delete
          </Button>
        )}
      </div>

      <div className="flex flex-wrap items-center gap-2 mt-3">
        <Select
          value={task.status}
          onValueChange={(val) => val && onStatusChange(task.id, val)}
        >
          <SelectTrigger className="w-[140px] h-8 text-xs">
            <span>{STATUS_LABELS[task.status] || task.status}</span>
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="todo">Todo</SelectItem>
            <SelectItem value="in_progress">In Progress</SelectItem>
            <SelectItem value="done">Done</SelectItem>
          </SelectContent>
        </Select>

        <Badge variant="outline" className={priorityColors[task.priority] || ""}>
          {PRIORITY_LABELS[task.priority] || task.priority}
        </Badge>

        {task.assignee_id ? (
          <Badge variant="secondary" className="text-xs">
            {task.assignee_name || "Assigned"}
          </Badge>
        ) : (
          <Badge variant="outline" className="text-xs text-muted-foreground">Unassigned</Badge>
        )}

        {task.due_date && (
          <span className="text-xs text-muted-foreground">
            Due: {new Date(task.due_date).toLocaleDateString()}
          </span>
        )}
      </div>
    </div>
  );
}
