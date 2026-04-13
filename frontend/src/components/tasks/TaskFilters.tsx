import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from "@/components/ui/select";
import { STATUS_LABELS } from "@/lib/types";

interface Props {
  statusFilter: string;
  onStatusChange: (val: string) => void;
  onClear: () => void;
}

const filterLabels: Record<string, string> = {
  all: "All Statuses",
  ...STATUS_LABELS,
};

export default function TaskFilters({ statusFilter, onStatusChange, onClear }: Props) {
  const current = statusFilter || "all";

  return (
    <div className="flex items-center gap-3">
      <Select value={current} onValueChange={(val) => val !== null && onStatusChange(val === "all" ? "" : val)}>
        <SelectTrigger className="w-[160px]">
          <span>{filterLabels[current] || current}</span>
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Statuses</SelectItem>
          <SelectItem value="todo">Todo</SelectItem>
          <SelectItem value="in_progress">In Progress</SelectItem>
          <SelectItem value="done">Done</SelectItem>
        </SelectContent>
      </Select>

      {statusFilter && (
        <Button variant="ghost" size="sm" onClick={onClear}>
          Clear Filters
        </Button>
      )}
    </div>
  );
}
