import browser from "webextension-polyfill";
import { healthUrl, PORT_RANGE_END, PORT_RANGE_START } from "./config";

const PORT_CACHE_KEY = "port";
const PROBE_TIMEOUT_MS = 800;

/**
 * Ask a port whether it belongs to ariadm. A different service can occupy any
 * candidate port, so the identity marker in /health is required.
 */
export async function probePort(port: number): Promise<boolean> {
  try {
    const response = await fetch(healthUrl(port), {
      signal: AbortSignal.timeout(PROBE_TIMEOUT_MS),
    });
    if (!response.ok) {
      return false;
    }

    const payload = (await response.json()) as { app?: string };
    return payload.app === "ariadm";
  } catch {
    return false;
  }
}

/**
 * Find the port the local API is listening on. The cached port is checked
 * first, then the whole range is scanned in parallel. The winner is cached so
 * later calls only pay for one probe.
 */
export async function resolvePort(): Promise<number | null> {
  const cached = await browser.storage.local.get({ [PORT_CACHE_KEY]: null });
  const cachedPort = typeof cached[PORT_CACHE_KEY] === "number" ? cached[PORT_CACHE_KEY] : null;

  if (cachedPort !== null) {
    if (await probePort(cachedPort)) {
      return cachedPort;
    }
    await browser.storage.local.remove(PORT_CACHE_KEY);
  }

  const ports: number[] = [];
  for (let port = PORT_RANGE_START; port <= PORT_RANGE_END; port++) {
    ports.push(port);
  }

  const results = await Promise.allSettled(ports.map((port) => probePort(port)));

  for (let index = 0; index < results.length; index++) {
    const result = results[index];
    if (result.status === "fulfilled" && result.value) {
      const port = ports[index];
      await browser.storage.local.set({ [PORT_CACHE_KEY]: port });
      return port;
    }
  }

  return null;
}

/** Drop the cached port so the next resolution performs a fresh scan. */
export async function clearCachedPort(): Promise<void> {
  await browser.storage.local.remove(PORT_CACHE_KEY);
}
