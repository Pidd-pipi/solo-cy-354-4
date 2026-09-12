export const USER_ROLES = [
  { value: 'student', label: '学生' },
  { value: 'admin', label: '管理员' },
] as const

export function roleLabel(value: string): string {
  return USER_ROLES.find((r) => r.value === value)?.label ?? value
}
