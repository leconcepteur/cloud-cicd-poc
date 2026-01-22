interface BoardProps {
  board: string[]
  onCellClick: (index: number) => void
  disabled: boolean
  yourSymbol: string
}

export function Board({ board, onCellClick, disabled, yourSymbol }: BoardProps) {
  return (
    <div className="grid grid-cols-3 gap-2 w-72 h-72">
      {board.map((cell, index) => (
        <button
          key={index}
          onClick={() => onCellClick(index)}
          disabled={disabled || cell !== ''}
          className={`
            w-full h-full text-5xl font-bold rounded-lg
            transition-all duration-150
            ${cell === '' && !disabled ? 'bg-gray-100 hover:bg-gray-200 cursor-pointer' : 'bg-gray-100'}
            ${cell === 'X' ? 'text-blue-600' : ''}
            ${cell === 'O' ? 'text-red-600' : ''}
            ${cell === '' && !disabled ? 'hover:scale-105' : ''}
            disabled:cursor-not-allowed
          `}
        >
          {cell || (cell === '' && !disabled ? (
            <span className="text-gray-300 text-3xl opacity-0 hover:opacity-50">
              {yourSymbol}
            </span>
          ) : null)}
        </button>
      ))}
    </div>
  )
}
