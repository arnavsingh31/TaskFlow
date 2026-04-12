import { useProjects } from "@/hooks/useProjects";
import ProjectCard from "@/components/projects/ProjectCard";
import ProjectForm from "@/components/projects/ProjectForm";
import Navbar from "@/components/layout/Navbar";
import { Button } from "@/components/ui/button";

export default function ProjectsPage() {
  const { projects, loading, error, createProject, refetch } = useProjects();

  return (
    <div className="min-h-screen bg-background">
      <Navbar />
      <div className="max-w-6xl mx-auto px-4 py-8">
        <div className="flex justify-between items-center mb-8">
          <h1 className="text-2xl font-bold">Projects</h1>
          <ProjectForm
            onSubmit={async (name, desc) => {
              await createProject(name, desc);
            }}
          />
        </div>

        {loading && (
          <div className="text-center py-12 text-muted-foreground">
            Loading projects...
          </div>
        )}

        {error && (
          <div className="text-center py-12">
            <p className="text-destructive mb-4">{error}</p>
            <Button variant="outline" onClick={refetch}>
              Retry
            </Button>
          </div>
        )}

        {!loading && !error && projects.length === 0 && (
          <div className="text-center py-12">
            <p className="text-muted-foreground mb-2">No projects yet.</p>
            <p className="text-sm text-muted-foreground">
              Create your first project to get started!
            </p>
          </div>
        )}

        {!loading && !error && projects.length > 0 && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {projects.map((project) => (
              <ProjectCard key={project.id} project={project} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
