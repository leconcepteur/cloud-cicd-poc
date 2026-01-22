import {
  createContext,
  useContext,
  useState,
  useEffect,
  useCallback,
  type ReactNode,
} from 'react'
import type { Game, SSEEvent, MatchFoundData, GameUpdateData, GameEndData } from '../types'
import { gameApi, matchmakingApi } from '../services/api'
import { sseService } from '../services/sse'
import { useAuth } from './AuthContext'

type GameState = 'idle' | 'queuing' | 'ready_check' | 'playing' | 'finished'

interface GameContextType {
  gameState: GameState
  game: Game | null
  yourSymbol: string
  isYourTurn: boolean
  opponent: string | null
  queueTime: number
  gameResult: { youWon: boolean; isDraw: boolean } | null
  joinQueue: () => Promise<void>
  leaveQueue: () => Promise<void>
  setReady: () => Promise<void>
  makeMove: (position: number) => Promise<void>
  forfeit: () => Promise<void>
  returnToLobby: () => void
}

const GameContext = createContext<GameContextType | null>(null)

export function GameProvider({ children }: { children: ReactNode }) {
  const { user } = useAuth()
  const [gameState, setGameState] = useState<GameState>('idle')
  const [game, setGame] = useState<Game | null>(null)
  const [yourSymbol, setYourSymbol] = useState('')
  const [isYourTurn, setIsYourTurn] = useState(false)
  const [opponent, setOpponent] = useState<string | null>(null)
  const [queueTime, setQueueTime] = useState(0)
  const [gameResult, setGameResult] = useState<{ youWon: boolean; isDraw: boolean } | null>(null)

  // Check for existing game on mount
  useEffect(() => {
    if (user) {
      gameApi.getState().then((response) => {
        if (response.game) {
          setGame(response.game)
          setYourSymbol(response.your_symbol)
          setIsYourTurn(response.is_your_turn)

          if (response.game.status === 'waiting') {
            setGameState('ready_check')
          } else if (response.game.status === 'in_progress') {
            setGameState('playing')
          } else if (response.game.status === 'finished') {
            setGameState('finished')
          }
        }
      })
    }
  }, [user])

  // SSE event handling
  useEffect(() => {
    if (!user) return

    sseService.connect()

    const unsubscribe = sseService.subscribe((event: SSEEvent) => {
      switch (event.type) {
        case 'match_found': {
          const data = event.data as MatchFoundData
          setOpponent(data.opponent)
          setGameState('ready_check')
          break
        }
        case 'opponent_ready': {
          // Opponent clicked ready
          break
        }
        case 'game_update': {
          const data = event.data as GameUpdateData
          setGame(data.game)
          setYourSymbol(data.your_symbol)
          setIsYourTurn(data.is_your_turn)
          if (data.game.status === 'in_progress') {
            setGameState('playing')
          }
          break
        }
        case 'game_end': {
          const data = event.data as GameEndData
          setGame(data.game)
          setYourSymbol(data.your_symbol)
          setGameResult({ youWon: data.you_won, isDraw: data.is_draw })
          setGameState('finished')
          break
        }
        case 'ready_timeout': {
          setGameState('idle')
          setGame(null)
          setOpponent(null)
          break
        }
        case 'opponent_left': {
          // Handle opponent disconnect
          break
        }
      }
    })

    return () => {
      unsubscribe()
      sseService.disconnect()
    }
  }, [user])

  // Queue timer
  useEffect(() => {
    let interval: number | undefined
    if (gameState === 'queuing') {
      setQueueTime(0)
      interval = window.setInterval(() => {
        setQueueTime((t) => t + 1)
      }, 1000)
    }
    return () => {
      if (interval) clearInterval(interval)
    }
  }, [gameState])

  const joinQueue = useCallback(async () => {
    await matchmakingApi.join()
    setGameState('queuing')
  }, [])

  const leaveQueue = useCallback(async () => {
    await matchmakingApi.leave()
    setGameState('idle')
    setQueueTime(0)
  }, [])

  const setReady = useCallback(async () => {
    const response = await gameApi.ready()
    setGame(response.game)
    setYourSymbol(response.your_symbol)
    setIsYourTurn(response.is_your_turn)
    if (response.game?.status === 'in_progress') {
      setGameState('playing')
    }
  }, [])

  const makeMove = useCallback(async (position: number) => {
    const response = await gameApi.move(position)
    setGame(response.game)
    setIsYourTurn(response.is_your_turn)
  }, [])

  const forfeit = useCallback(async () => {
    const response = await gameApi.forfeit()
    setGame(response.game)
    setGameResult({ youWon: false, isDraw: false })
    setGameState('finished')
  }, [])

  const returnToLobby = useCallback(() => {
    setGameState('idle')
    setGame(null)
    setOpponent(null)
    setGameResult(null)
    setYourSymbol('')
    setIsYourTurn(false)
  }, [])

  return (
    <GameContext.Provider
      value={{
        gameState,
        game,
        yourSymbol,
        isYourTurn,
        opponent,
        queueTime,
        gameResult,
        joinQueue,
        leaveQueue,
        setReady,
        makeMove,
        forfeit,
        returnToLobby,
      }}
    >
      {children}
    </GameContext.Provider>
  )
}

export function useGame() {
  const context = useContext(GameContext)
  if (!context) {
    throw new Error('useGame must be used within a GameProvider')
  }
  return context
}
