import { useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { useGame } from '../context/GameContext'
import { Button } from '../components/Button'
import { Stats } from '../components/Stats'
import { Spinner } from '../components/Spinner'

export function LobbyPage() {
  const navigate = useNavigate()
  const { user, logout } = useAuth()
  const { gameState, queueTime, joinQueue, leaveQueue } = useGame()

  const handleLogout = async () => {
    await logout()
    navigate('/login')
  }

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60)
    const secs = seconds % 60
    return `${mins}:${secs.toString().padStart(2, '0')}`
  }

  return (
    <div className="min-h-screen bg-gray-50 p-8">
      <div className="max-w-md mx-auto">
        <div className="bg-white rounded-xl shadow-lg p-6">
          <div className="flex justify-between items-center mb-6">
            <h1 className="text-2xl font-bold">Welcome, {user?.username}!</h1>
            <Button variant="secondary" size="sm" onClick={handleLogout}>
              Logout
            </Button>
          </div>

          <div className="mb-8">
            <h2 className="text-lg font-semibold mb-3">Your Stats</h2>
            <Stats />
          </div>

          <div className="border-t pt-6">
            {gameState === 'idle' && (
              <Button onClick={joinQueue} className="w-full" size="lg">
                Find Match
              </Button>
            )}

            {gameState === 'queuing' && (
              <div className="text-center">
                <Spinner />
                <p className="mt-4 text-gray-600">Looking for opponent...</p>
                <p className="text-2xl font-mono mt-2">{formatTime(queueTime)}</p>
                <Button variant="secondary" onClick={leaveQueue} className="mt-4">
                  Cancel
                </Button>
              </div>
            )}

            {(gameState === 'ready_check' || gameState === 'playing') && (
              <Button onClick={() => navigate('/game')} className="w-full" size="lg">
                Continue Game
              </Button>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
