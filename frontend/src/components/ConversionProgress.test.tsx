import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import '@testing-library/jest-dom'
import { ConversionProgress } from './ConversionProgress'

describe('ConversionProgress', () => {
  it('returns null if progress is null, status is 0, or status is 3+', () => {
    const { container: c1 } = render(<ConversionProgress progress={null} />)
    expect(c1.firstChild).toBeNull()

    const { container: c2 } = render(<ConversionProgress progress={{ status: 0, message: 'Idle', converted: 0, total: 0 }} />)
    expect(c2.firstChild).toBeNull()

    const { container: c3 } = render(<ConversionProgress progress={{ status: 3, message: 'Done', converted: 5, total: 5 }} />)
    expect(c3.firstChild).toBeNull()
  })

  it('renders standard progress message and status percentage', () => {
    render(<ConversionProgress progress={{ status: 2, message: 'Converting tracks...', converted: 3, total: 10 }} />)
    expect(screen.getByText('Converting tracks...')).toBeInTheDocument()
    expect(screen.getByText('30%')).toBeInTheDocument()
  })
})
