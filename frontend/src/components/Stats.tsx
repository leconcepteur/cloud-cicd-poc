import { useEffect, useState } from 'react'
import type { UserStats } from '../types'
import { statsApi } from '../services/api'

export function Stats() {
  const [stats, setStats] = useState<UserStats | null>(null)

  useEffect(() => {
    statsApi.getStats().then(setStats).catch(console.error)
  }, [])

  if (!stats) {
    return null
  }

  return (
    <div className="flex gap-6 text-sm">
      <div className="text-center">
        <div className="text-2xl font-bold text-green-600">{stats.wins}</div>
        <div className="text-gray-500">Wins</div>
      </div>
      <div className="text-center">
        <div className="text-2xl font-bold text-gray-600">{stats.ties}</div>
        <div className="text-gray-500">Ties</div>
      </div>
      <div className="text-center">
        <div className="text-2xl font-bold text-red-600">{stats.losses}</div>
        <div className="text-gray-500">Losses</div>
      </div>
    </div>
  )
}
