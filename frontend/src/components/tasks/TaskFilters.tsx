import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

interface Props {
  statusFilter: string;
  onStatusChange: (val: string) => void;
  onClear: () => void;
}

export default function TaskFilters({ statusFilter, onStatusChange, onClear }: Props) {
  return (
    <div className="flex items-center gap-3">
      <Select value={statusFilter || "all"} onValueChange={(val) => val !== null && onStatusChange(val === "all" ? "" : val)}>
        <SelectTrigger className="w-[160px]">
          <SelectValue placeholder="Filter by status" />
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
