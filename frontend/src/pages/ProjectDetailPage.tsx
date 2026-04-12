import { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import api from "@/lib/axios";
import type { ProjectDetail, Task } from "@/lib/types";
import { useTasks } from "@/hooks/useTasks";
import Navbar from "@/components/layout/Navbar";
import TaskCard from "@/components/tasks/TaskCard";
import TaskForm from "@/components/tasks/TaskForm";
import TaskFilters from "@/components/tasks/TaskFilters";
import { Button } from "@/components/ui/button";

export default function ProjectDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [project, setProject] = useState<ProjectDetail | null>(null);
  const [projectLoading, setProjectLoading] = useState(true);
  const [projectError, setProjectError] = useState("");

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

  useEffect(() => {
    const fetchProject = async () => {
      try {
        const res = await api.get<ProjectDetail>(`/projects/${id}`);
        setProject(res.data);
      } catch {
        setProjectError("Project not found");
      } finally {
        setProjectLoading(false);
      }
    };
    fetchProject();
  }, [id]);

  if (projectLoading) {
    return (
      <div className="min-h-screen bg-background">
        <Navbar />
        <div className="text-center py-12 text-muted-foreground">Loading...</div>
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

        <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-6">
          <TaskFilters
            statusFilter={statusFilter}
            onStatusChange={setStatusFilter}
            onClear={() => setStatusFilter("")}
          />
          <TaskForm onSubmit={createTask} />
        </div>

        {tasksLoading && (
          <div className="text-center py-8 text-muted-foreground">Loading tasks...</div>
        )}

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
                onStatusChange={(taskId, status) => updateTask(taskId, { status: status as Task["status"] })}
                onDelete={deleteTask}
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
            onSubmit={createTask}
            onUpdate={updateTask}
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
