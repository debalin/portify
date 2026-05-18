import { describe, it, expect } from 'vitest'
import { render } from '@testing-library/react'
import { DynamicProviderIcon } from './Icons'

describe('Icons', () => {
  it('renders spotify icon', () => {
    const { container } = render(<DynamicProviderIcon providerId="spotify" />)
    expect(container.querySelector('svg')).toBeDefined()
  })
  it('renders youtube icon', () => {
    const { container } = render(<DynamicProviderIcon providerId="youtube" />)
    expect(container.querySelector('svg')).toBeDefined()
  })
  it('renders tidal icon', () => {
    const { container } = render(<DynamicProviderIcon providerId="tidal" />)
    expect(container.querySelector('svg')).toBeDefined()
  })
  it('renders default icon for unknown provider', () => {
    const { container } = render(<DynamicProviderIcon providerId="unknown" />)
    expect(container.querySelector('svg')).toBeDefined()
  })
})
