export const PORT_RANGE_START = 9999;
export const PORT_RANGE_END = 10008;

export function healthUrl(port: number): string {
  return `http://localhost:${port}/health`;
}

export function downloadUrl(port: number): string {
  return `http://localhost:${port}/download`;
}
