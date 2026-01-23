import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useGame } from '../context/GameContext'
import { Board } from '../components/Board'
import { Button } from '../components/Button'
import { Spinner } from '../components/Spinner'

export function GamePage() {
  const navigate = useNavigate()
  const {
    gameState,
    game,
    yourSymbol,
    isYourTurn,
    opponent,
    gameResult,
    setReady,
    makeMove,
    forfeit,
    returnToLobby,
  } = useGame()

  useEffect(() => {
    if (gameState === 'idle') {
      navigate('/lobby')
    }
  }, [gameState, navigate])

  const handleReturnToLobby = () => {
    returnToLobby()
    navigate('/lobby')
  }

  if (gameState === 'queuing') {
    navigate('/lobby')
    return null
  }

  if (gameState === 'ready_check') {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="bg-white rounded-xl shadow-lg p-8 text-center max-w-md">
          <h2 className="text-2xl font-bold mb-4">Match Found!</h2>
          <p className="text-gray-600 mb-6">
            Opponent: <span className="font-semibold">{opponent || 'Unknown'}</span>
          </p>

          {game?.player_x_ready && game?.player_o_ready ? (
            <div className="flex flex-col items-center">
              <Spinner />
              <p className="mt-4">Starting game...</p>
            </div>
          ) : (
            <>
              <p className="text-sm text-gray-500 mb-4">Both players must be ready to start</p>
              <Button onClick={setReady} size="lg" className="w-full">
                Ready!
              </Button>
            </>
          )}
        </div>
      </div>
    )
  }

  if (gameState === 'finished' && gameResult) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="bg-white rounded-xl shadow-lg p-8 text-center max-w-md">
          <h2 className="text-3xl font-bold mb-4">
            {gameResult.isDraw ? (
              <span className="text-gray-600">Draw!</span>
            ) : gameResult.youWon ? (
              <span className="text-green-600">You Won!</span>
            ) : (
              <span className="text-red-600">You Lost</span>
            )}
          </h2>

          {game && (
            <div className="mb-6 flex justify-center">
              <Board
                board={game.board}
                onCellClick={() => {}}
                disabled={true}
                yourSymbol={yourSymbol}
              />
            </div>
          )}

          <Button onClick={handleReturnToLobby} size="lg" className="w-full">
            Return to Lobby
          </Button>
        </div>
      </div>
    )
  }

  if (gameState === 'playing' && game) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="bg-white rounded-xl shadow-lg p-8 text-center">
          <div className="mb-4">
            <p className="text-lg">
              You are{' '}
              <span
                className={`font-bold text-2xl ${yourSymbol === 'X' ? 'text-blue-600' : 'text-red-600'}`}
              >
                {yourSymbol}
              </span>
            </p>
            <p
              className={`text-xl font-semibold ${isYourTurn ? 'text-green-600' : 'text-gray-400'}`}
            >
              {isYourTurn ? 'Your turn!' : "Opponent's turn..."}
            </p>
          </div>

          <div className="flex justify-center mb-6">
            <Board
              board={game.board}
              onCellClick={makeMove}
              disabled={!isYourTurn}
              yourSymbol={yourSymbol}
            />
          </div>

          <Button variant="danger" onClick={forfeit}>
            Forfeit
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center">
      <Spinner size="lg" />
    </div>
  )
}
