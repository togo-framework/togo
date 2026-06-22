# The core flow

The spine of every togo app — full-stack in four commands:

```bash
togo make:model Post title:string body:text:nullable   # data layer (interactive if no fields)
togo migrate                                            # apply schema
togo make:controller Post                               # REST + GraphQL + docs + hooks + transformer
togo make:view Post                                     # Next.js page + data hook
togo generate                                           # sqlc → gqlgen → atlas → OpenAPI
```

`togo make:resource Post ...` runs model + controller + view at once.
`togo.resources.yaml` is the source of truth; registries are regenerated from it.
Add behavior with **Actions** (`make:action`) wired to **hooks/events**, and
customize API output with **resources** (transformers).
