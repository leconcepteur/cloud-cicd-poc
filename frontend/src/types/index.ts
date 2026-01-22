export interface User {
  id: string
  username: string
}

export interface UserStats {
  user_id: string
  wins: number
  losses: number
  ties: number
}

export type GameStatus = 'waiting' | 'in_progress' | 'finished'
export type GameResult = '' | 'win_x' | 'win_o' | 'draw'

export interface Game {
  id: string
  player_x: string
  player_o: string
  board: string[]
  current_turn: string
  status: GameStatus
  result: GameResult
  winner?: string
  player_x_ready: boolean
  player_o_ready: boolean
  created_at: string
  updated_at: string
}

export interface GameStateResponse {
  game: Game | null
  your_symbol: string
  is_your_turn: boolean
}

export interface LoginRequest {
  username: string
  password: string
}

export interface RegisterRequest {
  username: string
  password: string
}

export interface LoginResponse {
  user: User
  session_id: string
}

export type EventType =
  | 'match_found'
  | 'game_start'
  | 'game_update'
  | 'game_end'
  | 'opponent_ready'
  | 'opponent_left'
  | 'queue_update'
  | 'error'
  | 'ready_timeout'
  | 'disconnect_grace'

export interface SSEEvent<T = unknown> {
  type: EventType
  data: T
}

export interface MatchFoundData {
  game_id: string
  opponent: string
}

export interface GameUpdateData {
  game: Game
  your_symbol: string
  is_your_turn: boolean
}

export interface GameEndData {
  game: Game
  your_symbol: string
  you_won: boolean
  is_draw: boolean
}
