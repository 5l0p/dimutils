# Config Package TODO

## Template Management
- [ ] Move Go templates from embedded strings to external template files
- [ ] Create `templates/` directory structure for different generator types
- [ ] Add template loading mechanism with fallback to embedded defaults
- [ ] Support custom template directories via config

## Template Enhancement
- [ ] Flesh out Docker templates with multi-stage builds, health checks, security
- [ ] Enhance Kubernetes templates with ConfigMaps, Secrets, Ingress, HPA
- [ ] Improve Helm charts with proper values structure and conditional logic
- [ ] Expand Terraform templates with VPC, security groups, load balancers
- [ ] Add more comprehensive Makefile templates with testing, linting, deployment

## Script Generation
- [ ] Add shell/bash script generation for deployment pipelines
- [ ] Add PowerShell script generation for Windows environments
- [ ] Add Python script generation for data processing workflows
- [ ] Add GitHub Actions workflow generation
- [ ] Add GitLab CI/CD pipeline generation
- [ ] Add Jenkins pipeline generation

## Configuration Features
- [ ] Add environment-specific configurations (dev, staging, prod)
- [ ] Add configuration inheritance and composition
- [ ] Add configuration validation rules and schema
- [ ] Add configuration encryption/decryption for secrets
- [ ] Add configuration version management and migrations

## Advanced Generators
- [ ] Add Ansible playbook generation
- [ ] Add Pulumi configuration generation
- [ ] Add CDK (Cloud Development Kit) generation
- [ ] Add Skaffold configuration generation
- [ ] Add Tilt configuration generation

## Template Engine Improvements
- [ ] Add custom template functions for common operations
- [ ] Add conditional template rendering based on config
- [ ] Add template includes and partials
- [ ] Add template validation and linting