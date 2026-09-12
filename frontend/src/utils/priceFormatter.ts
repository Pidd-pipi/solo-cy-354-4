export function formatPrice(value: number): string {
  return `¥${Number(value).toFixed(2)}`
}

export function parsePrice(value: number): number {
  return Math.round(value * 100) / 100
}
