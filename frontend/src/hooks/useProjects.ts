import { useState, useEffect, useCallback } from "react";
import api from "@/lib/axios";
import type { Project, PaginatedResponse } from "@/lib/types";

export function useProjects(pageLimit = 9) {
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [page, setPage] = useState(1);
  const limit = pageLimit;
  const [total, setTotal] = useState(0);

  const fetchProjects = useCallback(async () => {
    try {
      setLoading(true);
      const res = await api.get<PaginatedResponse<Project>>(
        `/projects?page=${page}&limit=${limit}`
      );
      setProjects(res.data.data || []);
      setTotal(res.data.total);
      setError("");
    } catch {
      setError("Failed to load projects");
    } finally {
      setLoading(false);
    }
  }, [page, limit]);

  useEffect(() => {
    fetchProjects();
  }, [fetchProjects]);

  const createProject = async (name: string, description?: string) => {
    const res = await api.post<Project>(
      "/projects",
      { name, description },
      { headers: { "X-Idempotency-Key": crypto.randomUUID() } }
    );
    setPage(1);
    setTotal((prev) => prev + 1);
    return res.data;
  };

  const deleteProject = async (id: string) => {
    await api.delete(`/projects/${id}`);
    const newTotal = total - 1;
    const newTotalPages = Math.ceil(newTotal / limit);
    if (page > newTotalPages && newTotalPages > 0) {
      setPage(newTotalPages);
    } else {
      setProjects((prev) => prev.filter((p) => p.id !== id));
    }
    setTotal(newTotal);
  };

  const totalPages = Math.ceil(total / limit);

  return {
    projects,
    loading,
    error,
    createProject,
    deleteProject,
    refetch: fetchProjects,
    page,
    setPage,
    total,
    totalPages,
  };
}
