# AI / Claude Code

Every togo project is born agent-ready. `togo new` scaffolds a `.claude/` tree
(skills, agents, rules) and a `.mcp.json` pre-wired to the togo MCP server.

## MCP server

[`mcp`](https://github.com/togo-framework/mcp) exposes the generators as
MCP tools: `make_resource`, `generate`, `list_resources`, `migrate`.

```bash
togo mcp:install --agent claude-code   # writes .mcp.json
togo mcp:serve                         # runs the stdio MCP server
```

With it wired, an agent can scaffold resources, run codegen, and migrate inside
the project — no shell glue required.

## `.claude/` tree

- `skills/` — slash commands wrapping the CLI (e.g. `/togo:resource`).
- `agents/` — specialist subagents (`togo-backend`, …) that know the stack.
- `rules/` — conventions agents follow (generator-first, ownership classes, codegen order).

## Agentic generators

`togo make:* --agent "<request>"` and `--ai` enable AI-assisted scaffolding.
