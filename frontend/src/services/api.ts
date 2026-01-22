import type {
  User,
  UserStats,
  LoginRequest,
  RegisterRequest,
  LoginResponse,
  GameStateResponse,
} from '../types'

const API_BASE = import.meta.env.VITE_API_URL || '/api/v1'

class ApiError extends Error {
  constructor(
    public status: number,
    message: string
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

async function request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const url = `${API_BASE}${endpoint}`

  const response = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    credentials: 'include',
  })

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({ error: 'Unknown error' }))
    throw new ApiError(response.status, errorData.error || 'Request failed')
  }

  return response.json()
}

export const authApi = {
  register: (data: RegisterRequest): Promise<User> => request('/auth/register', {
    method: 'POST',
    body: JSON.stringify(data),
  }),

  login: (data: LoginRequest): Promise<LoginResponse> => request('/auth/login', {
    method: 'POST',
    body: JSON.stringify(data),
  }),

  logout: (): Promise<void> => request('/auth/logout', {
    method: 'POST',
  }),

  me: (): Promise<User> => request('/auth/me'),
}

export const gameApi = {
  getState: (): Promise<GameStateResponse> => request('/game/state'),

  ready: (): Promise<GameStateResponse> => request('/game/ready', {
    method: 'POST',
  }),

  move: (position: number): Promise<GameStateResponse> => request('/game/move', {
    method: 'POST',
    body: JSON.stringify({ position }),
  }),

  forfeit: (): Promise<GameStateResponse> => request('/game/forfeit', {
    method: 'POST',
  }),
}

export const matchmakingApi = {
  join: (): Promise<{ message: string }> => request('/matchmaking/join', {
    method: 'POST',
  }),

  leave: (): Promise<{ message: string }> => request('/matchmaking/leave', {
    method: 'POST',
  }),
}

export const statsApi = {
  getStats: (): Promise<UserStats> => request('/stats'),
}

export { ApiError }
