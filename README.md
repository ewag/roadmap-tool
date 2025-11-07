# IT Roadmap Visualizer

A Go-based web service for visualizing IT application and service line roadmaps. Upload roadmaps via YAML files and view them as interactive timelines.

## Features

- 📊 Interactive timeline visualization using vis-timeline
- 📝 YAML-based roadmap definitions
- 🎨 Color-coded status tracking (Planned, In Progress, Completed, Blocked)
- 🔗 Dependency visualization between roadmap items
- 🔍 Filter roadmaps by service line and status
- 🐳 Kubernetes-ready with Docker container
- 💾 File-based storage (easily upgradeable to database)

## Quick Start

### Local Development

1. **Install Go 1.21+**

2. **Clone and run:**
   ```bash
   go run cmd/server/main.go
   ```

3. **Access the application:**
   Open http://localhost:8080 in your browser

4. **Upload sample roadmaps:**
   Use the sample YAML files in the `samples/` directory to test the application.

### Running with Docker

1. **Build the image:**
   ```bash
   docker build -t roadmap-visualizer:latest .
   ```

2. **Run the container:**
   ```bash
   docker run -p 8080:8080 -v $(pwd)/data:/data roadmap-visualizer:latest
   ```

3. **Access the application:**
   Open http://localhost:8080

### Deploying to Kubernetes

1. **Build and push the Docker image:**
   ```bash
   docker build -t your-registry/roadmap-visualizer:latest .
   docker push your-registry/roadmap-visualizer:latest
   ```

2. **Update the image in the deployment:**
   Edit `deployments/kubernetes/deployment.yaml` and update the `image` field.

3. **Deploy to Kubernetes:**
   ```bash
   kubectl apply -f deployments/kubernetes/pvc.yaml
   kubectl apply -f deployments/kubernetes/configmap.yaml
   kubectl apply -f deployments/kubernetes/deployment.yaml
   kubectl apply -f deployments/kubernetes/service.yaml
   ```

4. **Access the service:**
   ```bash
   kubectl get service roadmap-visualizer
   ```

## YAML Format

Roadmaps are defined using YAML files with the following structure:

```yaml
roadmap:
  name: "Your Roadmap Name"
  service_line: "Service Line"
  owner: "Team Name"
  items:
    - id: "unique-id"
      name: "Item Name"
      start: "2025-Q1"  # or "2025-01-15"
      end: "2025-Q2"    # or "2025-03-31"
      status: "planned" # planned, in-progress, completed, blocked
      description: "Description of the item"
      dependencies: ["other-item-id"]
```

### Field Requirements

- `name`: Required - Name of the roadmap
- `service_line`: Required - Service line for grouping/filtering
- `owner`: Optional - Team or person responsible
- `notes`: Optional - Markdown-formatted notes for the roadmap
- `items`: Required - Array of roadmap items
  - `id`: Required - Unique identifier
  - `name`: Required - Display name
  - `start`: Required - Start date (YYYY-QN or YYYY-MM-DD)
  - `end`: Required - End date (YYYY-QN or YYYY-MM-DD)
  - `status`: Required - One of: planned, in-progress, completed, blocked
  - `description`: Optional - Detailed description
  - `notes`: Optional - Markdown-formatted notes for the item
  - `dependencies`: Optional - Array of item IDs this depends on

### Fiscal Year Quarter Format

**Quarters start on July 1st:**
- `2026-Q1` = July 1, 2025 - September 30, 2025
- `2026-Q2` = October 1, 2025 - December 31, 2025
- `2026-Q3` = January 1, 2026 - March 31, 2026
- `2026-Q4` = April 1, 2026 - June 30, 2026

You can also use standard date format: `2025-07-01` for specific dates.

## REST API

### Endpoints

- `POST /api/roadmaps` - Upload a new roadmap (accepts YAML in body)
- `GET /api/roadmaps` - List all roadmaps
- `GET /api/roadmaps/{id}` - Get a specific roadmap
- `DELETE /api/roadmaps/{id}` - Delete a roadmap
- `GET /health` - Health check endpoint
- `GET /ready` - Readiness check endpoint

### Example: Upload via cURL

```bash
curl -X POST http://localhost:8080/api/roadmaps \
  -H "Content-Type: application/x-yaml" \
  -H "X-File-Name: my-roadmap.yaml" \
  --data-binary @samples/authentication-services.yaml
```

## Configuration

Configuration is done via environment variables:

- `PORT` - HTTP port (default: 8080)
- `DATA_DIR` - Directory for storing roadmap files (default: ./data)

## Project Structure

```
roadmap-visualizer/
├── cmd/server/              # Application entry point
├── internal/
│   ├── handlers/           # HTTP request handlers
│   ├── models/             # Data models
│   ├── parser/             # YAML parsing
│   └── storage/            # File storage implementation
├── web/
│   ├── static/css/         # Stylesheets
│   └── templates/          # HTML templates
├── deployments/kubernetes/ # Kubernetes manifests
├── samples/                # Sample roadmap YAML files
├── Dockerfile             # Container definition
└── README.md
```

## Development

### Adding Features

The application is designed to be extended:

- **Database backend**: Replace `storage.FileStorage` with a database implementation
- **Authentication**: Add middleware to the HTTP handlers
- **Export functionality**: Add new endpoints and handlers
- **Advanced visualization**: Enhance the frontend JavaScript

### Testing

Test the application with the provided sample files:

1. Upload `samples/authentication-services.yaml`
2. Upload `samples/data-platform.yaml`
3. Upload `samples/customer-portal.yaml`
4. View the roadmap list and filter by service line
5. Click on a roadmap to view the interactive timeline

## Tech Stack

- **Backend**: Go 1.21+
- **Frontend**: HTML, CSS, JavaScript
- **Visualization**: vis-timeline library
- **Storage**: File-based (YAML + JSON metadata)
- **Container**: Docker
- **Orchestration**: Kubernetes

## License

MIT License - Feel free to use this for teaching, learning, or production use.

## Contributing

This is a teaching tool for DevOps courses. Contributions welcome!

---

**Generated with Claude Code** 🤖
