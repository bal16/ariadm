const BYTE_UNITS = ["B", "KB", "MB", "GB", "TB"] as const;
const RATE_UNITS = ["B/s", "KB/s", "MB/s", "GB/s"] as const;

function scale(value: number, units: readonly string[]): string {
  const base = 1024;
  const index = Math.min(
    Math.max(Math.floor(Math.log(value) / Math.log(base)), 0),
    units.length - 1,
  );
  const scaled = parseFloat((value / Math.pow(base, index)).toFixed(2));
  return `${scaled} ${units[index]}`;
}

export function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  return scale(bytes, BYTE_UNITS);
}

export function formatSpeed(bytesPerSecond: number): string {
  if (bytesPerSecond === 0) return "0.00 B/s";
  return scale(bytesPerSecond, RATE_UNITS);
}
