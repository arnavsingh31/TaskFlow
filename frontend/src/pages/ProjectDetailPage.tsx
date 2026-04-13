import { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import api from "@/lib/axios";
import type { ProjectDetail, Task, ProjectStats } from "@/lib/types";
import { STATUS_LABELS } from "@/lib/types";
import { useTasks } from "@/hooks/useTasks";
import Navbar from "@/components/layout/Navbar";
import TaskCard from "@/components/tasks/TaskCard";
import TaskForm from "@/components/tasks/TaskForm";
import TaskFilters from "@/components/tasks/TaskFilters";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { FullPageSpinner } from "@/components/ui/spinner";

export default function ProjectDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [project, setProject] = useState<ProjectDetail | null>(null);
  const [projectLoading, setProjectLoading] = useState(true);
  const [projectError, setProjectError] = useState("");
  const [stats, setStats] = useState<ProjectStats | null>(null);

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
  } = useTasks(id!);

  const [editTask, setEditTask] = useState<Task | null>(null);
  const [editOpen, setEditOpen] = useState(false);

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

  // Refresh stats when tasks change
  const refreshStats = async () => {
    try {
      const res = await api.get<ProjectStats>(`/projects/${id}/stats`);
      setStats(res.data);
    } catch {
      // Stats refresh failure is non-critical
    }
  };

  const handleCreateTask = async (data: Parameters<typeof createTask>[0]) => {
    const result = await createTask(data);
    refreshStats();
    return result;
  };

  const handleUpdateTask = async (taskId: string, data: Parameters<typeof updateTask>[1]) => {
    const result = await updateTask(taskId, data);
    refreshStats();
    return result;
  };

  const handleDeleteTask = async (taskId: string) => {
    await deleteTask(taskId);
    refreshStats();
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

        <div className="mb-8">
          <h1 className="text-2xl font-bold">{project.name}</h1>
          {project.description && (
            <p className="text-muted-foreground mt-1">{project.description}</p>
          )}
        </div>

        {stats && (
          <div className="mb-8">
            <div className="grid grid-cols-3 gap-4 mb-4">
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
          <div className="space-y-3">
            {tasks.map((task) => (
              <TaskCard
                key={task.id}
                task={task}
                onStatusChange={(taskId, status) => handleUpdateTask(taskId, { status: status as Task["status"] })}
                onDelete={handleDeleteTask}
                onEdit={(t) => {
                  setEditTask(t);
                  setEditOpen(true);
                }}
              />
            ))}
          </div>
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
      </div>
    </div>
  );
}
