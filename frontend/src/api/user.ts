import request from '../utils/request'
import type { LoginResponse, User } from '../types'

export function register(data: { phone: string; password: string; nickname: string; campus: string }) {
  return request.post<never, { code: number; message: string; data: User }>('/users/register', data)
}

export function login(data: { phone: string; password: string }) {
  return request.post<never, { code: number; message: string; data: LoginResponse }>('/users/login', data)
}

export function getProfile() {
  return request.get<never, { code: number; message: string; data: User }>('/users/me')
}

export function updateProfile(data: Partial<Pick<User, 'nickname' | 'avatar' | 'campus'>>) {
  return request.put<never, { code: number; message: string; data: User }>('/users/me', data)
}
