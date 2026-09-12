import type { TradeOrder } from '../types'

// Trade status values, mirrored with backend internal/constants/trade.go.
export const TRADE_STATUS = {
  PENDING: 'pending',
  CONFIRMED: 'confirmed',
  COMPLETED: 'completed',
  CANCELLED: 'cancelled',
} as const

export const TRADE_STATUSES = [
  { value: TRADE_STATUS.PENDING, label: '待确认', type: 'warning' },
  { value: TRADE_STATUS.CONFIRMED, label: '已确认', type: 'primary' },
  { value: TRADE_STATUS.COMPLETED, label: '已完成', type: 'success' },
  { value: TRADE_STATUS.CANCELLED, label: '已取消', type: 'info' },
] as const

export const REVIEW_RATINGS = [
  { value: 'good', label: '好评' },
  { value: 'medium', label: '中评' },
  { value: 'bad', label: '差评' },
] as const

// Order lifecycle actions. The button rules live here as ONE table so the UI
// shows an action exactly when the backend state machine accepts it:
//   pending   --buyer confirm-->  confirmed
//   confirmed --seller confirm--> completed
//   pending   --buyer/seller cancel--> cancelled
export type TradeAction = 'buyer_confirm' | 'seller_confirm' | 'cancel'

type TradeActor = 'buyer' | 'seller' | 'participant'

interface TradeActionRule {
  // Statuses from which the action is allowed.
  from: readonly string[]
  // Who may perform the action.
  actor: TradeActor
}

export const TRADE_ACTION_RULES: Record<TradeAction, TradeActionRule> = {
  buyer_confirm: { from: [TRADE_STATUS.PENDING], actor: 'buyer' },
  seller_confirm: { from: [TRADE_STATUS.CONFIRMED], actor: 'seller' },
  cancel: { from: [TRADE_STATUS.PENDING], actor: 'participant' },
}

type OrderLike = Pick<TradeOrder, 'status' | 'buyer_id' | 'seller_id'>

// canTrade is the single gate every order action button must pass. It applies
// the same role + status rules as the backend tradestate Guard.
export function canTrade(order: OrderLike, action: TradeAction, userId?: number): boolean {
  if (userId === undefined) {
    return false
  }
  const rule = TRADE_ACTION_RULES[action]
  if (!rule.from.includes(order.status)) {
    return false
  }
  switch (rule.actor) {
    case 'buyer':
      return order.buyer_id === userId
    case 'seller':
      return order.seller_id === userId
    case 'participant':
      return order.buyer_id === userId || order.seller_id === userId
    default:
      return false
  }
}

export function tradeStatusLabel(value: string): string {
  return TRADE_STATUSES.find((t) => t.value === value)?.label ?? value
}

export function tradeStatusType(value: string): string {
  return TRADE_STATUSES.find((t) => t.value === value)?.type ?? 'info'
}

export function ratingLabel(value: string): string {
  return REVIEW_RATINGS.find((r) => r.value === value)?.label ?? value
}
