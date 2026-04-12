import { useState, useEffect, useCallback } from "react";
import api from "@/lib/axios";
import type { Project, ListResponse } from "@/lib/types";

export function useProjects() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const fetchProjects = useCallback(async () => {
    try {
      setLoading(true);
      const res = await api.get<ListResponse<Project>>("/projects");
      setProjects(res.data.data || []);
      setError("");
    } catch {
      setError("Failed to load projects");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchProjects();
  }, [fetchProjects]);

  const createProject = async (name: string, description?: string) => {
    const res = await api.post<Project>(
      "/projects",
      { name, description },
      { headers: { "X-Idempotency-Key": crypto.randomUUID() } }
    );
    setProjects((prev) => [res.data, ...prev]);
    return res.data;
  };

  const deleteProject = async (id: string) => {
    await api.delete(`/projects/${id}`);
    setProjects((prev) => prev.filter((p) => p.id !== id));
  };

  return { projects, loading, error, createProject, deleteProject, refetch: fetchProjects };
}
