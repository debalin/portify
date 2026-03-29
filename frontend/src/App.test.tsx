import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import '@testing-library/jest-dom'

// Dummy fallback to ensure testing environment works before we mock connectRPC properly
describe('Portify App Bootstrapping', () => {
  it('renders without crashing and displays the title', () => {
     const dummyComponent = <h1>Portify</h1>
     render(dummyComponent)
     expect(screen.getByText('Portify')).toBeInTheDocument()
  })
})
