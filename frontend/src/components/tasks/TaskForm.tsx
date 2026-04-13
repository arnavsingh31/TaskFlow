import { useState, useEffect } from "react";
import api from "@/lib/axios";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from "@/components/ui/select";
import { STATUS_LABELS, PRIORITY_LABELS } from "@/lib/types";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import type { Task } from "@/lib/types";

interface Props {
  onSubmit: (data: {
    title: string;
    description?: string;
    status?: string;
    priority?: string;
    assignee_id?: string;
    due_date?: string;
  }) => Promise<any>;
  editTask?: Task | null;
  onUpdate?: (taskId: string, data: any) => Promise<any>;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
}

export default function TaskForm({ onSubmit, editTask, onUpdate, open, onOpenChange }: Props) {
  const [internalOpen, setInternalOpen] = useState(false);
  const isOpen = open !== undefined ? open : internalOpen;
  const setIsOpen = onOpenChange || setInternalOpen;

  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [status, setStatus] = useState<string>("todo");
  const [priority, setPriority] = useState<string>("medium");
  const [dueDate, setDueDate] = useState("");
  const [assigneeEmail, setAssigneeEmail] = useState("");
  const [assigneeId, setAssigneeId] = useState("");
  const [assigneeName, setAssigneeName] = useState("");
  const [assigneeCleared, setAssigneeCleared] = useState(false);
  const [verifying, setVerifying] = useState(false);
  const [verifyError, setVerifyError] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  // Sync form state when editTask changes
  useEffect(() => {
    if (editTask) {
      setTitle(editTask.title || "");
      setDescription(editTask.description || "");
      setStatus(editTask.status || "todo");
      setPriority(editTask.priority || "medium");
      setDueDate(editTask.due_date || "");
      setAssigneeId(editTask.assignee_id || "");
      setAssigneeEmail(editTask.assignee_email || "");
      setAssigneeName(
        editTask.assignee_name && editTask.assignee_email
          ? `${editTask.assignee_name} (${editTask.assignee_email})`
          : ""
      );
      setAssigneeCleared(false);
    } else {
      resetForm();
    }
  }, [editTask]);

  const resetForm = () => {
    setTitle("");
    setDescription("");
    setStatus("todo");
    setPriority("medium");
    setDueDate("");
    setAssigneeEmail("");
    setAssigneeId("");
    setAssigneeName("");
    setAssigneeCleared(false);
    setVerifyError("");
    setError("");
  };

  const verifyAssignee = async () => {
    if (!assigneeEmail.trim()) return;
    setVerifying(true);
    setVerifyError("");
    setAssigneeName("");
    try {
      const res = await api.get(`/users/search?email=${encodeURIComponent(assigneeEmail)}`);
      setAssigneeId(res.data.id);
      setAssigneeName(`${res.data.name} (${res.data.email})`);
      setAssigneeCleared(false);
    } catch {
      setVerifyError("No user found with this email");
      setAssigneeId("");
    } finally {
      setVerifying(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) {
      setError("Title is required");
      return;
    }
    setLoading(true);
    setError("");
    try {
      // Auto-verify if email is typed but not verified
      let resolvedAssigneeId = assigneeId;
      if (assigneeEmail.trim() && !assigneeId && !assigneeCleared) {
        try {
          const res = await api.get(`/users/search?email=${encodeURIComponent(assigneeEmail)}`);
          resolvedAssigneeId = res.data.id;
        } catch {
          setError("Could not find user with email: " + assigneeEmail);
          setLoading(false);
          return;
        }
      }

      const data: any = { title, status, priority };
      if (description) data.description = description;
      if (dueDate) data.due_date = dueDate;

      if (resolvedAssigneeId) {
        data.assignee_id = resolvedAssigneeId;
      } else if (assigneeCleared && editTask) {
        data.assignee_id = "";
      }

      if (editTask && onUpdate) {
        await onUpdate(editTask.id, data);
      } else {
        await onSubmit(data);
      }
      resetForm();
      setIsOpen(false);
    } catch {
      setError("Failed to save task");
    } finally {
      setLoading(false);
    }
  };

  const content = (
    <form onSubmit={handleSubmit} className="space-y-4">
      {error && (
        <div className="bg-destructive/10 text-destructive p-2 rounded text-sm">{error}</div>
      )}
      <div className="space-y-2">
        <Label>Title</Label>
        <Input value={title} onChange={(e) => setTitle(e.target.value)} placeholder="Task title" />
      </div>
      <div className="space-y-2">
        <Label>Description (optional)</Label>
        <Textarea value={description} onChange={(e) => setDescription(e.target.value)} placeholder="Details..." />
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Status</Label>
          <Select value={status} onValueChange={(v) => v && setStatus(v)}>
            <SelectTrigger><span>{STATUS_LABELS[status] || status}</span></SelectTrigger>
            <SelectContent>
              <SelectItem value="todo">Todo</SelectItem>
              <SelectItem value="in_progress">In Progress</SelectItem>
              <SelectItem value="done">Done</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div className="space-y-2">
          <Label>Priority</Label>
          <Select value={priority} onValueChange={(v) => v && setPriority(v)}>
            <SelectTrigger><span>{PRIORITY_LABELS[priority] || priority}</span></SelectTrigger>
            <SelectContent>
              <SelectItem value="low">Low</SelectItem>
              <SelectItem value="medium">Medium</SelectItem>
              <SelectItem value="high">High</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>
      <div className="space-y-2">
        <Label>Due Date (optional)</Label>
        <Input type="date" value={dueDate} onChange={(e) => setDueDate(e.target.value)} />
      </div>
      <div className="space-y-2">
        <Label>Assign to (optional)</Label>
        <div className="flex gap-2">
          <Input
            value={assigneeEmail}
            onChange={(e) => {
              setAssigneeEmail(e.target.value);
              setAssigneeId("");
              setAssigneeName("");
              setVerifyError("");
            }}
            placeholder="User email"
          />
          <Button type="button" variant="outline" onClick={verifyAssignee} disabled={verifying}>
            {verifying ? "..." : "Verify"}
          </Button>
        </div>
        {assigneeName && <p className="text-sm text-green-600">{assigneeName}</p>}
        {verifyError && <p className="text-sm text-destructive">{verifyError}</p>}
        {(assigneeId || assigneeName) && (
          <Button
            type="button"
            variant="ghost"
            size="sm"
            className="text-destructive"
            onClick={() => {
              setAssigneeId("");
              setAssigneeName("");
              setAssigneeEmail("");
              setAssigneeCleared(true);
            }}
          >
            Clear assignment
          </Button>
        )}
        {assigneeCleared && !assigneeId && (
          <p className="text-sm text-muted-foreground">Assignment will be removed on save.</p>
        )}
      </div>
      <Button type="submit" className="w-full" disabled={loading}>
        {loading ? "Saving..." : editTask ? "Update Task" : "Create Task"}
      </Button>
    </form>
  );

  if (open !== undefined) {
    return (
      <Dialog open={isOpen} onOpenChange={setIsOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editTask ? "Edit Task" : "New Task"}</DialogTitle>
          </DialogHeader>
          {content}
        </DialogContent>
      </Dialog>
    );
  }

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogTrigger render={<Button>Add Task</Button>} />
      <DialogContent>
        <DialogHeader>
          <DialogTitle>New Task</DialogTitle>
        </DialogHeader>
        {content}
      </DialogContent>
    </Dialog>
  );
}
