import { useState, useEffect, useCallback } from "react";
import api from "@/lib/axios";
import type { Task, ListResponse } from "@/lib/types";

export function useTasks(projectId: string) {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [statusFilter, setStatusFilter] = useState("");
  const [assigneeFilter, setAssigneeFilter] = useState("");

  const fetchTasks = useCallback(async () => {
    try {
      setLoading(true);
      const params = new URLSearchParams();
      if (statusFilter) params.set("status", statusFilter);
      if (assigneeFilter) params.set("assignee", assigneeFilter);
      const url = `/projects/${projectId}/tasks${params.toString() ? "?" + params.toString() : ""}`;
      const res = await api.get<ListResponse<Task>>(url);
      setTasks(res.data.data || []);
      setError("");
    } catch {
      setError("Failed to load tasks");
    } finally {
      setLoading(false);
    }
  }, [projectId, statusFilter, assigneeFilter]);

  useEffect(() => {
    fetchTasks();
  }, [fetchTasks]);

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
    setTasks((prev) => [res.data, ...prev]);
    return res.data;
  };

  const updateTask = async (
    taskId: string,
    data: Partial<Pick<Task, "title" | "description" | "status" | "priority" | "assignee_id" | "due_date">>
  ) => {
    // Optimistic update
    const prev = tasks;
    setTasks((t) => t.map((task) => (task.id === taskId ? { ...task, ...data } : task)));
    try {
      const res = await api.patch<Task>(`/tasks/${taskId}`, data);
      setTasks((t) => t.map((task) => (task.id === taskId ? res.data : task)));
      return res.data;
    } catch {
      setTasks(prev); // Revert
      throw new Error("Failed to update task");
    }
  };

  const deleteTask = async (taskId: string) => {
    await api.delete(`/tasks/${taskId}`);
    setTasks((prev) => prev.filter((t) => t.id !== taskId));
  };

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
  };
}
