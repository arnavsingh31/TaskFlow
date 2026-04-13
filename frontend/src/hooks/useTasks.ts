import { useState, useEffect, useCallback } from "react";
import api from "@/lib/axios";
import type { Task, PaginatedResponse } from "@/lib/types";

export function useTasks(projectId: string, pageLimit = 10) {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [statusFilter, setStatusFilter] = useState("");
  const [assigneeFilter, setAssigneeFilter] = useState("");
  const [page, setPage] = useState(1);
  const limit = pageLimit;
  const [total, setTotal] = useState(0);

  const fetchTasks = useCallback(async () => {
    try {
      setLoading(true);
      const params = new URLSearchParams();
      params.set("page", String(page));
      params.set("limit", String(limit));
      if (statusFilter) params.set("status", statusFilter);
      if (assigneeFilter) params.set("assignee", assigneeFilter);
      const res = await api.get<PaginatedResponse<Task>>(
        `/projects/${projectId}/tasks?${params.toString()}`
      );
      setTasks(res.data.data || []);
      setTotal(res.data.total);
      setError("");
    } catch {
      setError("Failed to load tasks");
    } finally {
      setLoading(false);
    }
  }, [projectId, statusFilter, assigneeFilter, page, limit]);

  useEffect(() => {
    fetchTasks();
  }, [fetchTasks]);

  // Reset to page 1 when filters change
  useEffect(() => {
    setPage(1);
  }, [statusFilter, assigneeFilter]);

  const createTask = async (data: {
    title: string;
    description?: string;
    status?: string;
    priority?: string;
    assignee_id?: string;
    due_date?: string;
  }) => {
    const res = await api.post<Task>(
      `/projects/${projectId}/tasks`,
      data,
      { headers: { "X-Idempotency-Key": crypto.randomUUID() } }
    );
    setPage(1);
    setTotal((prev) => prev + 1);
    return res.data;
  };

  const updateTask = async (
    taskId: string,
    data: Partial<Pick<Task, "title" | "description" | "status" | "priority" | "assignee_id" | "due_date">>
  ) => {
    const prev = tasks;
    setTasks((t) => t.map((task) => (task.id === taskId ? { ...task, ...data } : task)));
    try {
      const res = await api.patch<Task>(`/tasks/${taskId}`, data);
      setTasks((t) => t.map((task) => (task.id === taskId ? res.data : task)));
      return res.data;
    } catch {
      setTasks(prev);
      throw new Error("Failed to update task");
    }
  };

  const deleteTask = async (taskId: string) => {
    await api.delete(`/tasks/${taskId}`);
    const newTotal = total - 1;
    const newTotalPages = Math.ceil(newTotal / limit);
    if (page > newTotalPages && newTotalPages > 0) {
      setPage(newTotalPages);
    } else {
      setTasks((prev) => prev.filter((t) => t.id !== taskId));
    }
    setTotal(newTotal);
  };

  const totalPages = Math.ceil(total / limit);

  return {
    tasks,
    loading,
    error,
    createTask,
    updateTask,
    deleteTask,
    statusFilter,
    setStatusFilter,
    assigneeFilter,
    setAssigneeFilter,
    refetch: fetchTasks,
    page,
    setPage,
    total,
    totalPages,
  };
}
