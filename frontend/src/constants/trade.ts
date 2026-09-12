export const TRADE_STATUSES = [
  { value: 'pending', label: '待确认', type: 'warning' },
  { value: 'confirmed', label: '已确认', type: 'primary' },
  { value: 'completed', label: '已完成', type: 'success' },
  { value: 'cancelled', label: '已取消', type: 'info' },
] as const

export const REVIEW_RATINGS = [
  { value: 'good', label: '好评' },
  { value: 'medium', label: '中评' },
  { value: 'bad', label: '差评' },
] as const

export function tradeStatusLabel(value: string): string {
  return TRADE_STATUSES.find((t) => t.value === value)?.label ?? value
}

export function tradeStatusType(value: string): string {
  return TRADE_STATUSES.find((t) => t.value === value)?.type ?? 'info'
}

export function ratingLabel(value: string): string {
  return REVIEW_RATINGS.find((r) => r.value === value)?.label ?? value
}
