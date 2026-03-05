/**
 * Setup file for Bun unit tests
 * Provides minimal DOM mocking for templateRenderer.js
 */

// Minimal DOM mock for templateRenderer
global.document = {
  createElement: (tag) => ({
    innerHTML: '',
    querySelectorAll: () => [],
    querySelector: () => null,
  }),
};
