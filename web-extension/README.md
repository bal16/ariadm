# AriaDM Browser Extension

Manifest V3 extension that sends downloads from the browser to the AriaDM desktop app.

## Development

Dependencies are managed from the repository root with Bun workspaces:

```bash
bun install
```

Run the extension dev build:

```bash
bun --filter @ariadm/web-extension dev
```

The extension discovers the local AriaDM API port, intercepts browser downloads, and forwards them to the desktop app.