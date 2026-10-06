/**
 * CFOP selector behavior for the detached NF-e surfaces.
 *
 * - The per-row search input filters only its own row's select options
 *   (code + description, case/diacritics-insensitive). The currently selected
 *   option is never filtered away.
 * - Typing a full 4-digit code selects the matching option, enabling
 *   keyboard-only selection.
 * - The "+" button records the originating row's select so the register modal
 *   can auto-select the created code and refresh every row with the
 *   farm-ordered option list.
 *
 * The module is idempotent: initCfopSelectors() may be called from the page
 * and from each swapped-in dialog.
 */

let initialized = false
let pendingOriginSelect = null

function normalizeText(value) {
  return (value || '')
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
    .toLowerCase()
    .trim()
}

/**
 * Derives the Origem/Destino value from the first digit of a CFOP code
 * (1/5 same state, 2/6 other state, 3/7 foreign). Empty for invalid families.
 */
export function deriveOriginDestination(code) {
  const first = (code || '').trim()[0]
  switch (first) {
    case '1':
    case '5':
      return 'Mesmo estado'
    case '2':
    case '6':
      return 'Outro estado'
    case '3':
    case '7':
      return 'Exterior'
    default:
      return ''
  }
}

function rowOf(element) {
  return element && element.closest
    ? element.closest('.item-row, .rascunho-item-row')
    : null
}

function selectOf(row) {
  return row ? row.querySelector('select.cfop-select[data-cfop-select]') : null
}

function optionMatches(option, needle) {
  if (!needle || option.selected) return true
  return (
    normalizeText(option.value).includes(needle) ||
    normalizeText(option.textContent).includes(needle)
  )
}

function applyFilter(select, query) {
  const needle = normalizeText(query)

  select.querySelectorAll('optgroup').forEach((group) => {
    let visible = 0
    group.querySelectorAll('option').forEach((option) => {
      const matches = optionMatches(option, needle)
      option.hidden = !matches
      if (matches) visible += 1
    })
    group.hidden = visible === 0
  })

  // Options rendered outside an optgroup (fallback for a selected code that
  // is not in the catalog/farm lists).
  select.querySelectorAll(':scope > option').forEach((option) => {
    option.hidden = !optionMatches(option, needle)
  })
}

function highlightExactCode(select, query) {
  const code = (query || '').trim()
  if (!/^\d{4}$/.test(code)) return

  const match = Array.from(select.options).find(
    (option) => option.value === code && !option.hidden
  )
  if (match) {
    select.value = code
  }
}

function handleSearch(event) {
  const input = event.target && event.target.closest
    ? event.target.closest('.cfop-search')
    : null
  if (!input) return

  const select = selectOf(rowOf(input))
  if (!select) return

  applyFilter(select, input.value)
  highlightExactCode(select, input.value)
}

function handleAdd(event) {
  const button = event.target && event.target.closest
    ? event.target.closest('[data-cfop-add]')
    : null
  if (!button) return

  pendingOriginSelect = selectOf(rowOf(button))
}

/**
 * Installs the document-level listeners once. Safe to call from the page and
 * from every dialog that includes a selector.
 */
export function initCfopSelectors() {
  if (initialized) return
  initialized = true

  document.addEventListener('input', handleSearch)
  // Capture so the originating row is recorded before htmx issues the request.
  document.addEventListener('click', handleAdd, true)
}

/**
 * Refreshes every selector in the document with the farm-ordered option list,
 * preserving each row's current selection. The target select receives the
 * newly created code.
 */
export async function refreshCfopOptions(selectedCode, targetSelect) {
  const query = selectedCode ? `?selected=${encodeURIComponent(selectedCode)}` : ''
  const response = await fetch(`/nfe/cfop/options${query}`, {
    headers: { 'HX-Request': 'true' },
  })
  if (!response.ok) return

  const html = await response.text()
  document.querySelectorAll('select.cfop-select[data-cfop-select]').forEach((select) => {
    const previous = select.value
    select.innerHTML = html

    const preferred = select === targetSelect ? selectedCode : previous
    if (preferred) {
      select.value = preferred
    }

    const row = rowOf(select)
    const search = row ? row.querySelector('.cfop-search') : null
    if (search) {
      search.value = ''
    }
    applyFilter(select, '')
  })
}

/** Returns and clears the row select recorded by the last "+" click. */
export function takePendingOriginSelect() {
  const select = pendingOriginSelect
  pendingOriginSelect = null
  return select
}
