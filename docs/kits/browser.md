# Browser Kit

Chromium browser via [Playwright](https://playwright.dev/) for browser automation.

**Activation: Opt-in** — only active if explicitly enabled in your config.

## What's Included

- **Chromium** browser with all system dependencies
- **playwright** CLI for browser automation

## Configuration

```yaml
kits:
  browser: {}
```

## Dependencies

Depends on the [Node.js](node.md) kit (Playwright is installed via npm).

## Cache

The Playwright browser cache is persisted at `/home/claude/.cache/ms-playwright` in a named Docker volume. This avoids re-downloading Chromium on every container start.

## Usage

```sh
# Take a screenshot
npx playwright screenshot https://example.com screenshot.png

# Run a Playwright script
npx playwright test

# Launch interactive codegen
npx playwright codegen https://example.com
```

Playwright can also be used programmatically from Node.js or Python scripts inside the container.
