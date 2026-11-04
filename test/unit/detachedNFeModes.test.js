import { describe, expect, it } from 'bun:test'
import { modeState, applySourceMode } from '../../assets/js/detachedNFeModes.js'

/** Minimal element stub: tracks the hidden class and the disabled property. */
function fakeElement() {
  return {
    disabled: false,
    classes: new Set(),
    classList: {
      toggle(name, force) {
        if (force) this.owner.classes.add(name)
        else this.owner.classes.delete(name)
      },
    },
  }
}

function fakeSection() {
  const element = fakeElement()
  element.classList.owner = element
  return element
}

function fakeControl() {
  return fakeElement()
}

describe('mode helper', () => {
  it('returns the active source flags', () => {
    expect(modeState('existing')).toEqual({ existing: true, new: false })
    expect(modeState('new')).toEqual({ existing: false, new: true })
  })

  it('throws for an unknown mode so both sources cannot stay disabled', () => {
    expect(() => modeState('other')).toThrow()
    expect(() => modeState('')).toThrow()
    expect(() => modeState(undefined)).toThrow()
  })
})

describe('applySourceMode', () => {
  function sources() {
    const existing = { section: fakeSection(), controls: [fakeControl()] }
    const fresh = { section: fakeSection(), controls: [fakeControl(), fakeControl()] }
    return { existing, new: fresh }
  }

  it('shows and enables the existing source, hiding and disabling the new one', () => {
    const { existing, new: fresh } = sources()

    applySourceMode('existing', { existing, new: fresh })

    expect(existing.section.classes.has('hidden')).toBe(false)
    expect(existing.controls.every((control) => control.disabled === false)).toBe(true)
    expect(fresh.section.classes.has('hidden')).toBe(true)
    expect(fresh.controls.every((control) => control.disabled === true)).toBe(true)
  })

  it('reverses the state when the new source is selected', () => {
    const { existing, new: fresh } = sources()

    applySourceMode('new', { existing, new: fresh })

    expect(existing.section.classes.has('hidden')).toBe(true)
    expect(existing.controls.every((control) => control.disabled === true)).toBe(true)
    expect(fresh.section.classes.has('hidden')).toBe(false)
    expect(fresh.controls.every((control) => control.disabled === false)).toBe(true)
  })

  it('throws for an unknown mode', () => {
    const { existing, new: fresh } = sources()
    expect(() => applySourceMode('nope', { existing, new: fresh })).toThrow()
  })
})
