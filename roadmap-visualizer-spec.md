# IT Roadmap Visualizer - Project Specification

## Project Overview
Build a Go-based web service for visualizing IT application and service line roadmaps, deployable to Kubernetes. Users submit roadmaps via YAML files, and the tool renders them as interactive timelines.

## Technology Stack
- **Backend**: Go (Golang)
- **Frontend**: HTML/CSS/JavaScript with a timeline visualization library
- **Storage**: File-based YAML (MVP - can migrate to database later)
- **Deployment**: Kubernetes-ready with Docker container
- **Visualization Library**: Recommend vis-timeline or similar lightweight option

## MVP Features (Priority Order)

### 1. Core Backend (Go)
- REST API with the following endpoints:
  - `POST /api/roadmaps` - Upload/create a new roadmap (accepts YAML)
  - `GET /api/roadmaps` - List all roadmaps
  - `GET /api/roadmaps/{id}` - Get specific roadmap details
  - `DELETE /api/roadmaps/{id}` - Delete a roadmap
- YAML parser to validate and load roadmap files
- File-based storage (store uploaded YAMLs in `/data` directory)
- CORS support for local development

### 2. YAML Schema
```yaml
roadmap:
  name: "Authentication Services"
  service_line: "Platform"
  owner: "Platform Team"
  items:
    - id: "oauth-impl"
      name: "OAuth2 Implementation"
      start: "2025-Q1"
      end: "2025-Q2"
      status: "in-progress"  # planned, in-progress, completed, blocked
      description: "Implement OAuth2 authentication flow"
      dependencies: []
    - id: "sso-integration"
      name: "SSO Integration"
      start: "2025-Q2"
      end: "2025-Q3"
      status: "planned"
      description: "Integrate with enterprise SSO providers"
      dependencies: ["oauth-impl"]
```

**Field Requirements:**
- `name`: Required, string
- `service_line`: Required, string (for grouping/filtering)
- `owner`: Optional, string
- `items`: Required, array of roadmap items
  - `id`: Required, unique identifier for dependency tracking
  - `name`: Required, string
  - `start`: Required, string (format: YYYY-QN or YYYY-MM-DD)
  - `end`: Required, string (same format as start)
  - `status`: Required, enum [planned, in-progress, completed, blocked]
  - `description`: Optional, string
  - `dependencies`: Optional, array of item IDs

### 3. Frontend Web Interface
- **Upload Page**: Simple form to upload YAML files
- **Roadmap List View**: Display all uploaded roadmaps with basic metadata
- **Roadmap Detail View**: Interactive timeline visualization showing:
  - Swimlane/Gantt-style timeline
  - Color coding by status (planned=blue, in-progress=yellow, completed=green, blocked=red)
  - Dependency lines/arrows between items
  - Hover tooltips showing item details
- **Basic Filtering**: Filter by service line or status
- Navigation between views

### 4. Containerization & K8s Deployment
- **Dockerfile**: Multi-stage build for Go application
- **Kubernetes Manifests**:
  - Deployment with 2 replicas
  - Service (ClusterIP or LoadBalancer)
  - PersistentVolumeClaim for `/data` directory
  - ConfigMap for any configuration
- **Health Checks**: `/health` and `/ready` endpoints

## Project Structure
```
roadmap-visualizer/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── handlers/
│   │   └── roadmap.go           # HTTP handlers
│   ├── models/
│   │   └── roadmap.go           # Data structures
│   ├── storage/
│   │   └── file.go              # File-based storage implementation
│   └── parser/
│       └── yaml.go              # YAML parsing and validation
├── web/
│   ├── static/
│   │   ├── css/
│   │   │   └── styles.css
│   │   └── js/
│   │       └── app.js           # Frontend JavaScript
│   └── templates/
│       ├── index.html           # Upload page
│       ├── list.html            # Roadmap list
│       └── view.html            # Roadmap visualization
├── deployments/
│   └── kubernetes/
│       ├── deployment.yaml
│       ├── service.yaml
│       └── pvc.yaml
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```

## Implementation Guidelines

### Phase 1: Backend Core
1. Initialize Go module
2. Create data models (structs) for roadmap schema
3. Implement YAML parser with validation
4. Build file storage layer (CRUD operations)
5. Create HTTP handlers and REST API
6. Add basic error handling and logging

### Phase 2: Frontend
1. Create HTML templates with clean, simple design
2. Implement upload form with file validation
3. Build roadmap list page fetching from API
4. Create timeline visualization using chosen library
5. Add dependency visualization
6. Implement filtering controls

### Phase 3: Containerization
1. Write multi-stage Dockerfile
2. Create Kubernetes manifests
3. Add health check endpoints
4. Test local deployment with kind or minikube
5. Document deployment process

## Features to Defer (Post-MVP)
- User authentication and authorization
- Database backend (PostgreSQL/MySQL)
- Version control/history of roadmaps
- Real-time collaboration
- Advanced dependency graph visualization
- Export to PDF/PNG
- Email notifications
- Integration with project management tools (Jira, etc.)

## Testing Approach
- Unit tests for YAML parser and validation
- Integration tests for API endpoints
- Manual testing of frontend interactions
- Smoke test in Kubernetes environment

## Success Criteria
- Can upload a YAML roadmap via web interface
- Roadmaps display as interactive timelines
- Dependencies are visually shown
- Can filter by service line or status
- Application runs in Kubernetes with persistent storage
- Basic error handling prevents invalid data

## Sample Data for Testing
Create 2-3 sample YAML files representing different service lines:
- Authentication Services roadmap
- Data Platform roadmap
- Customer Portal roadmap

Each with 4-6 items, some with dependencies, various statuses.

## Notes
- Keep it simple for MVP - focus on core visualization
- Use standard library where possible to minimize dependencies
- Ensure YAML validation prevents malformed submissions
- Design with future database migration in mind
- This could serve as a teaching example for cloud-native development

## Questions to Resolve During Development
1. Which timeline library to use? (vis-timeline, timeline.js, or custom CSS?)
2. Date format preference - quarters vs. specific dates vs. both?
3. Should the tool support multiple roadmaps per YAML file?
4. Naming convention for uploaded files (preserve original, generate UUID, use roadmap name)?
