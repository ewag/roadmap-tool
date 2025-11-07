# Cline Handoff Guide - Roadmap Visualizer

## What to Give Cline

Share the `roadmap-visualizer-spec.md` file with Cline and use this prompt:

---

**Prompt for Cline:**

"I need you to build an IT roadmap visualization tool based on the attached specification. This is a Go-based web service that will run in Kubernetes.

Please follow this approach:

1. **Start with project setup**: Initialize the Go module and create the directory structure as outlined in the spec
2. **Build backend first**: Implement the data models, YAML parser, file storage, and REST API
3. **Create frontend**: Build the HTML templates and visualization interface
4. **Containerize**: Create the Dockerfile and Kubernetes manifests
5. **Test**: Create at least 2-3 sample YAML roadmaps for testing

Key priorities:
- Keep it simple and focused on the MVP features
- Use standard library where possible
- Make sure YAML validation is robust
- Ensure the code is ready for K8s deployment

Please start by confirming you understand the requirements, then proceed with Phase 1 (Backend Core) from the specification."

---

## Tips for Working with Cline

1. **Let Cline drive the implementation details** - The spec provides the what, Cline will figure out the how

2. **Review at each phase** - After backend, frontend, and containerization, test each piece before moving on

3. **Ask Cline to explain choices** - If Cline picks a specific library or approach, ask why

4. **Iterate on the frontend** - The visualization is the most visible part, so don't hesitate to ask for adjustments

5. **Test the K8s deployment** - Make sure Cline includes instructions for deploying to your cluster

## What You'll Get

By the end, you should have:
- ✅ Complete Go application with REST API
- ✅ Web interface for uploading and viewing roadmaps
- ✅ Dockerfile ready to build
- ✅ Kubernetes manifests ready to deploy
- ✅ Sample YAML files for testing
- ✅ README with setup and deployment instructions

## After MVP

Once the MVP is working, you can ask Cline to add:
- Database backend (PostgreSQL)
- Authentication layer
- Export functionality
- More advanced dependency visualization
- CI/CD pipeline configuration

Good luck! This should be a fun project, and it'll be a great teaching tool for your DevOps courses.
