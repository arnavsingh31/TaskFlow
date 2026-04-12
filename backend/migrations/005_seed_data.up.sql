-- Test user: test@example.com / password123
INSERT INTO users (id, name, email, password) VALUES (
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'Test User',
    'test@example.com',
    '$2a$12$sN87zE2ZptgAGR1jk3BqueAqqHolxY8D7Dg2QLQtD.ASLNaXUyaRK'
) ON CONFLICT DO NOTHING;

-- Test project
INSERT INTO projects (id, name, description, owner_id) VALUES (
    'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22',
    'Sample Project',
    'A demo project with sample tasks',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'
) ON CONFLICT DO NOTHING;

-- Three tasks with different statuses
INSERT INTO tasks (title, status, priority, project_id, created_by) VALUES
    ('Design the homepage', 'todo', 'high', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'),
    ('Set up CI/CD pipeline', 'in_progress', 'medium', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'),
    ('Write API documentation', 'done', 'low', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11')
ON CONFLICT DO NOTHING;
