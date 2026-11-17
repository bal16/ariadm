# AriaDM Frontend

SolidJS interface for the AriaDM desktop download manager.

## Development

Dependencies are managed from the repository root with Bun workspaces:

```bash
bun install
```

Run the frontend dev server:

```bash
bun --filter @ariadm/frontend dev
```

Type-check:

```bash
bun --filter @ariadm/frontend typecheck
```

The UI is built with SolidJS, Tailwind CSS v4, and shared design tokens from `packages/tokens`.