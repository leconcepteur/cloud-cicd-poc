import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { Board } from './Board'

describe('Board', () => {
  it('renders 9 cells', () => {
    const board = Array(9).fill('')
    render(<Board board={board} onCellClick={() => {}} disabled={false} yourSymbol="X" />)

    const buttons = screen.getAllByRole('button')
    expect(buttons).toHaveLength(9)
  })

  it('displays X and O symbols correctly', () => {
    const board = ['X', 'O', '', '', 'X', '', '', '', 'O']
    render(<Board board={board} onCellClick={() => {}} disabled={false} yourSymbol="X" />)

    const buttons = screen.getAllByRole('button')
    expect(buttons[0]).toHaveTextContent('X')
    expect(buttons[1]).toHaveTextContent('O')
    expect(buttons[4]).toHaveTextContent('X')
    expect(buttons[8]).toHaveTextContent('O')
  })

  it('calls onCellClick when empty cell is clicked', () => {
    const board = Array(9).fill('')
    const handleClick = vi.fn()

    render(<Board board={board} onCellClick={handleClick} disabled={false} yourSymbol="X" />)

    const buttons = screen.getAllByRole('button')
    fireEvent.click(buttons[4])

    expect(handleClick).toHaveBeenCalledWith(4)
  })

  it('does not call onCellClick when occupied cell is clicked', () => {
    const board = ['X', '', '', '', '', '', '', '', '']
    const handleClick = vi.fn()

    render(<Board board={board} onCellClick={handleClick} disabled={false} yourSymbol="O" />)

    const buttons = screen.getAllByRole('button')
    fireEvent.click(buttons[0]) // Click on occupied cell

    expect(handleClick).not.toHaveBeenCalled()
  })

  it('disables all cells when disabled prop is true', () => {
    const board = Array(9).fill('')
    render(<Board board={board} onCellClick={() => {}} disabled={true} yourSymbol="X" />)

    const buttons = screen.getAllByRole('button')
    buttons.forEach((button) => {
      expect(button).toBeDisabled()
    })
  })

  it('disables occupied cells', () => {
    const board = ['X', 'O', '', '', '', '', '', '', '']
    render(<Board board={board} onCellClick={() => {}} disabled={false} yourSymbol="X" />)

    const buttons = screen.getAllByRole('button')
    expect(buttons[0]).toBeDisabled() // X cell
    expect(buttons[1]).toBeDisabled() // O cell
    expect(buttons[2]).not.toBeDisabled() // Empty cell
  })
})
