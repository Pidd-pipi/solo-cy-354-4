export const PRODUCT_CATEGORIES = [
  { value: 'books', label: '书籍' },
  { value: 'electronics', label: '电子产品' },
  { value: 'daily', label: '生活用品' },
  { value: 'clothing', label: '服饰' },
] as const

export const PRODUCT_STATUSES = [
  { value: 'on_sale', label: '在售', type: 'success' },
  { value: 'reserved', label: '已预订', type: 'warning' },
  { value: 'sold', label: '已售出', type: 'info' },
  { value: 'removed', label: '已下架', type: 'info' },
] as const

export function categoryLabel(value: string): string {
  return PRODUCT_CATEGORIES.find((c) => c.value === value)?.label ?? value
}

export function productStatusLabel(value: string): string {
  return PRODUCT_STATUSES.find((s) => s.value === value)?.label ?? value
}

export function productStatusType(value: string): string {
  return PRODUCT_STATUSES.find((s) => s.value === value)?.type ?? 'info'
}
