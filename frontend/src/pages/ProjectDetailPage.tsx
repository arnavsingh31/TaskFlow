import { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import api from "@/lib/axios";
import type { ProjectDetail, Task, ProjectStats } from "@/lib/types";
import { STATUS_LABELS } from "@/lib/types";
import { useTasks } from "@/hooks/useTasks";
import { useAuth } from "@/context/AuthContext";
import Navbar from "@/components/layout/Navbar";
import TaskCard from "@/components/tasks/TaskCard";
import TaskForm from "@/components/tasks/TaskForm";
import TaskFilters from "@/components/tasks/TaskFilters";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { FullPageSpinner } from "@/components/ui/spinner";
import { toast } from "sonner";

export default function ProjectDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();
  const [project, setProject] = useState<ProjectDetail | null>(null);
  const [projectLoading, setProjectLoading] = useState(true);
  const [projectError, setProjectError] = useState("");
  const [stats, setStats] = useState<ProjectStats | null>(null);

  // Edit project state
  const [editProjectOpen, setEditProjectOpen] = useState(false);
  const [editName, setEditName] = useState("");
  const [editDescription, setEditDescription] = useState("");
  const [editLoading, setEditLoading] = useState(false);

  // Delete project state
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [deleteLoading, setDeleteLoading] = useState(false);

  const {
    tasks,
    loading: tasksLoading,
    error: tasksError,
    createTask,
    updateTask,
    deleteTask,
    statusFilter,
    setStatusFilter,
    refetch,
    page,
    setPage,
    total,
    totalPages,
  } = useTasks(id!);

  const [editTask, setEditTask] = useState<Task | null>(null);
  const [editOpen, setEditOpen] = useState(false);

  const isOwner = user?.id === project?.owner_id;

  const fetchProjectAndStats = async () => {
    try {
      const [projectRes, statsRes] = await Promise.all([
        api.get<ProjectDetail>(`/projects/${id}`),
        api.get<ProjectStats>(`/projects/${id}/stats`),
      ]);
      setProject(projectRes.data);
      setStats(statsRes.data);
    } catch {
      setProjectError("Project not found");
    } finally {
      setProjectLoading(false);
    }
  };

  useEffect(() => {
    fetchProjectAndStats();
  }, [id]);

  const refreshStats = async () => {
    try {
      const res = await api.get<ProjectStats>(`/projects/${id}/stats`);
      setStats(res.data);
    } catch {
      // non-critical
    }
  };

  const handleCreateTask = async (data: Parameters<typeof createTask>[0]) => {
    const result = await createTask(data);
    refreshStats();
    toast.success("Task created");
    return result;
  };

  const handleUpdateTask = async (taskId: string, data: Parameters<typeof updateTask>[1]) => {
    const result = await updateTask(taskId, data);
    refreshStats();
    toast.success("Task updated");
    return result;
  };

  const handleDeleteTask = async (taskId: string) => {
    await deleteTask(taskId);
    refreshStats();
    toast.success("Task deleted");
  };

  const handleEditProject = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editName.trim()) return;
    setEditLoading(true);
    try {
      const res = await api.patch<ProjectDetail>(`/projects/${id}`, {
        name: editName,
        description: editDescription || null,
      });
      setProject(res.data);
      setEditProjectOpen(false);
      toast.success("Project updated");
    } catch {
      toast.error("Failed to update project");
    } finally {
      setEditLoading(false);
    }
  };

  const handleDeleteProject = async () => {
    setDeleteLoading(true);
    try {
      await api.delete(`/projects/${id}`);
      toast.success("Project deleted");
      navigate("/");
    } catch {
      toast.error("Failed to delete project");
    } finally {
      setDeleteLoading(false);
    }
  };

  const openEditProject = () => {
    if (project) {
      setEditName(project.name);
      setEditDescription(project.description || "");
      setEditProjectOpen(true);
    }
  };

  if (projectLoading) {
    return (
      <div className="min-h-screen bg-background">
        <Navbar />
        <FullPageSpinner message="Loading project..." />
      </div>
    );
  }

  if (projectError || !project) {
    return (
      <div className="min-h-screen bg-background">
        <Navbar />
        <div className="text-center py-12">
          <p className="text-destructive mb-4">{projectError || "Project not found"}</p>
          <Button variant="outline" onClick={() => navigate("/")}>
            Back to Projects
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background">
      <Navbar />
      <div className="max-w-4xl mx-auto px-4 py-8">
        <Button variant="ghost" size="sm" onClick={() => navigate("/")} className="mb-4">
          &larr; Back to Projects
        </Button>

        <div className="flex items-start justify-between mb-8">
          <div>
            <h1 className="text-2xl font-bold">{project.name}</h1>
            {project.description && (
              <p className="text-muted-foreground mt-1">{project.description}</p>
            )}
          </div>
          {isOwner && (
            <div className="flex gap-2 shrink-0">
              <Button variant="outline" size="sm" onClick={openEditProject}>
                Edit
              </Button>
              <Button
                variant="outline"
                size="sm"
                className="text-destructive hover:text-destructive"
                onClick={() => setDeleteConfirmOpen(true)}
              >
                Delete
              </Button>
            </div>
          )}
        </div>

        {stats && (
          <div className="mb-8">
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-4">
              {([
                { key: "todo" as const, border: "border-l-blue-500", text: "text-blue-600" },
                { key: "in_progress" as const, border: "border-l-amber-500", text: "text-amber-600" },
                { key: "done" as const, border: "border-l-green-500", text: "text-green-600" },
              ]).map(({ key, border, text }) => (
                <Card key={key} className={`border-l-4 ${border}`}>
                  <CardContent className="pt-4 pb-4 text-center">
                    <p className={`text-2xl font-bold ${text}`}>
                      {stats[key]}
                    </p>
                    <p className="text-sm text-muted-foreground">
                      {STATUS_LABELS[key]}
                    </p>
                  </CardContent>
                </Card>
              ))}
            </div>

            {stats.by_assignee.length > 0 && (
              <div className="flex flex-wrap gap-2">
                {stats.by_assignee.map((a, i) => (
                  <span
                    key={i}
                    className="text-xs bg-secondary text-secondary-foreground px-2 py-1 rounded"
                  >
                    {a.assignee_name}: {a.count} task{a.count !== 1 ? "s" : ""}
                  </span>
                ))}
              </div>
            )}
          </div>
        )}

        <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-6">
          <TaskFilters
            statusFilter={statusFilter}
            onStatusChange={setStatusFilter}
            onClear={() => setStatusFilter("")}
          />
          <TaskForm onSubmit={handleCreateTask} />
        </div>

        {tasksLoading && <FullPageSpinner message="Loading tasks..." />}

        {tasksError && (
          <div className="text-center py-8">
            <p className="text-destructive mb-4">{tasksError}</p>
            <Button variant="outline" onClick={refetch}>Retry</Button>
          </div>
        )}

        {!tasksLoading && !tasksError && tasks.length === 0 && (
          <div className="text-center py-8">
            {statusFilter ? (
              <>
                <p className="text-muted-foreground mb-2">No tasks match your filters</p>
                <Button variant="outline" size="sm" onClick={() => setStatusFilter("")}>
                  Clear Filters
                </Button>
              </>
            ) : (
              <>
                <p className="text-muted-foreground mb-2">No tasks in this project.</p>
                <p className="text-sm text-muted-foreground">Add your first task!</p>
              </>
            )}
          </div>
        )}

        {!tasksLoading && !tasksError && tasks.length > 0 && (
          <>
            <div className="space-y-3">
              {tasks.map((task) => (
                <TaskCard
                  key={task.id}
                  task={task}
                  canDelete={isOwner || task.created_by === user?.id}
                  onStatusChange={(taskId, status) => handleUpdateTask(taskId, { status: status as Task["status"] })}
                  onDelete={handleDeleteTask}
                  onEdit={(t) => {
                    setEditTask(t);
                    setEditOpen(true);
                  }}
                />
              ))}
            </div>

            <div className="flex items-center justify-center gap-4 mt-6">
              <Button
                variant="outline"
                size="sm"
                disabled={page <= 1}
                onClick={() => setPage(page - 1)}
              >
                Previous
              </Button>
              <span className="text-sm text-muted-foreground">
                Page {page} of {totalPages} ({total} task{total !== 1 ? "s" : ""})
              </span>
              <Button
                variant="outline"
                size="sm"
                disabled={page >= totalPages}
                onClick={() => setPage(page + 1)}
              >
                Next
              </Button>
            </div>
          </>
        )}

        {editTask && (
          <TaskForm
            editTask={editTask}
            onSubmit={handleCreateTask}
            onUpdate={handleUpdateTask}
            open={editOpen}
            onOpenChange={(open) => {
              setEditOpen(open);
              if (!open) setEditTask(null);
            }}
          />
        )}

        {/* Edit Project Dialog */}
        <Dialog open={editProjectOpen} onOpenChange={setEditProjectOpen}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Edit Project</DialogTitle>
            </DialogHeader>
            <form onSubmit={handleEditProject} className="space-y-4">
              <div className="space-y-2">
                <Label>Name</Label>
                <Input
                  value={editName}
                  onChange={(e) => setEditName(e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <Label>Description</Label>
                <Textarea
                  value={editDescription}
                  onChange={(e) => setEditDescription(e.target.value)}
                />
              </div>
              <Button type="submit" className="w-full" disabled={editLoading}>
                {editLoading ? "Saving..." : "Save Changes"}
              </Button>
            </form>
          </DialogContent>
        </Dialog>

        {/* Delete Project Confirmation Dialog */}
        <Dialog open={deleteConfirmOpen} onOpenChange={setDeleteConfirmOpen}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Delete Project</DialogTitle>
            </DialogHeader>
            <p className="text-sm text-muted-foreground">
              Are you sure you want to delete <strong>{project.name}</strong>? This will also delete all tasks in this project. This action cannot be undone.
            </p>
            <div className="flex gap-2 justify-end mt-4">
              <Button variant="outline" onClick={() => setDeleteConfirmOpen(false)}>
                Cancel
              </Button>
              <Button
                variant="destructive"
                onClick={handleDeleteProject}
                disabled={deleteLoading}
              >
                {deleteLoading ? "Deleting..." : "Delete Project"}
              </Button>
            </div>
          </DialogContent>
        </Dialog>
      </div>
    </div>
  );
}
