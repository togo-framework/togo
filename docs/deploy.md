# Deploy

togo ships built-in Terraform. Scaffold infrastructure for a provider, then deploy:

```bash
togo infra:init gcp          # or: fly | docker
docker build -t myapp:latest .
togo infra:plan              # terraform plan
togo deploy                  # terraform apply
```

`infra:init <provider>` writes a `Dockerfile` and `infra/main.tf`:

- **gcp** — Cloud Run v2 service (public), region configurable.
- **fly** — Fly.io app + machine.
- **docker** — local/remote Docker container.

`togo infra:plan`, `togo infra:apply`, and `togo deploy` run `terraform -chdir=infra`
(auto-running `terraform init` the first time). Keep Terraform state out of the
app repo — point it at a separate infra repo or a remote backend.

Set provider credentials via environment variables / Terraform variables before
applying (e.g. `var.project` + `var.image` for GCP, `var.fly_api_token` for Fly).
