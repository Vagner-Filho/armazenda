/**
 * Searchable CFOP selector for the detached NF-e surfaces.
 *
 * The native select remains the successful form control. Its accessible
 * combobox/listbox UI is rendered from the select options so form submission,
 * server ordering, default selection, and farm-option refresh stay in sync.
 */

let initialized = false
let pendingOriginSelect = null
let nextOptionId = 0

export function normalizeCfopSearchText(value) {
  return String(value || '')
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
    .toLowerCase()
    .trim()
}

/**
 * Derives Origem/Destino from the first digit of a CFOP code
 * (1/5 same state, 2/6 other state, 3/7 foreign). Empty for invalid families.
 */
export function deriveOriginDestination(code) {
  const normalizedCode = String(code || '').trim()
  if (!/^\d{4}$/.test(normalizedCode)) return ''

  const first = normalizedCode[0]
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

function optionContainsQuery(option, needle) {
  const originDestination =
    option?.dataset?.originDestination || deriveOriginDestination(option?.value)
  return [option?.value, option?.textContent, originDestination].some((value) =>
    normalizeCfopSearchText(value).includes(needle)
  )
}

/** Matches code, description, or Origem/Destino with accent-insensitive text. */
export function cfopOptionMatches(option, query) {
  const needle = normalizeCfopSearchText(query)
  return !needle || Boolean(option?.selected) || optionContainsQuery(option, needle)
}

function selectorOf(element) {
  if (!element || !element.closest) return null
  if (element.matches && element.matches('[data-cfop-selector]')) return element
  return element.closest('[data-cfop-selector]')
}

function selectOf(element) {
  const selector = selectorOf(element)
  return selector ? selector.querySelector('select.cfop-select[data-cfop-select]') : null
}

function triggerOf(selector) {
  return selector ? selector.querySelector('[data-cfop-trigger]') : null
}

function menuOf(selector) {
  return selector ? selector.querySelector('[data-cfop-menu]') : null
}

function searchOf(selector) {
  return selector ? selector.querySelector('.cfop-search') : null
}

function optionsListOf(selector) {
  return selector ? selector.querySelector('[data-cfop-options-list]') : null
}

function isMenuOpen(menu) {
  return Boolean(menu && menu.dataset.cfopOpen === 'true')
}

function assignControlIds(selector) {
  const trigger = triggerOf(selector)
  const menu = menuOf(selector)
  const search = searchOf(selector)
  const optionsList = optionsListOf(selector)
  const label = selector.querySelector('[data-cfop-label]')

  if (!trigger || !menu || !search || !optionsList) return

  // Always refresh IDs: cloned item rows copy IDs from their source row.
  optionsList.id = `cfop-options-${++nextOptionId}`
  if (label) label.id = `cfop-label-${++nextOptionId}`
  trigger.setAttribute('aria-controls', optionsList.id)
  if (label) trigger.setAttribute('aria-labelledby', label.id)
  search.setAttribute('aria-controls', optionsList.id)
}

function appendOption(option, parent) {
  const item = document.createElement('div')
  item.className = 'cfop-option'
  item.id = `cfop-option-${++nextOptionId}`
  item.dataset.cfopOption = ''
  item.dataset.cfopValue = option.value
  item.dataset.cfopOptionIndex = String(option.index)
  item.dataset.cfopOriginDestination =
    option.dataset.originDestination || deriveOriginDestination(option.value)
  if (option.dataset.cfopFarm === 'true') item.dataset.cfopFarm = 'true'
  item.setAttribute('role', 'option')
  item.setAttribute('aria-selected', String(option.selected))
  item.tabIndex = -1
  item.textContent = option.textContent.trim()
  parent.appendChild(item)
  return item
}

function renderOptionList(selector) {
  const select = selectOf(selector)
  const list = optionsListOf(selector)
  if (!select || !list) return

  list.replaceChildren()
  Array.from(select.children).forEach((child) => {
    if (child.tagName === 'OPTGROUP') {
      const group = document.createElement('div')
      group.className = 'cfop-option-group'
      group.dataset.cfopOptionGroup = ''
      group.setAttribute('role', 'group')
      group.setAttribute('aria-label', child.label)

      const label = document.createElement('div')
      label.className = 'cfop-option-group-label'
      label.textContent = child.label
      group.appendChild(label)

      Array.from(child.children).forEach((option) => {
        if (option.tagName === 'OPTION') appendOption(option, group)
      })
      list.appendChild(group)
      return
    }

    if (child.tagName === 'OPTION') appendOption(child, list)
  })

  const emptyState = selector.querySelector('[data-cfop-empty]')
  if (emptyState) emptyState.hidden = select.options.length > 0
}

function syncSelectedOption(selector) {
  const select = selectOf(selector)
  const trigger = triggerOf(selector)
  const label = selector.querySelector('[data-cfop-selected-label]')
  const list = optionsListOf(selector)
  if (!select) return

  const selectedOption = select.selectedOptions[0]
  if (label) {
    label.textContent = selectedOption
      ? selectedOption.textContent.trim()
      : 'Selecione um CFOP'
  }

  if (list) {
    list.querySelectorAll('[data-cfop-option]').forEach((item) => {
      const option = select.options[Number(item.dataset.cfopOptionIndex)]
      item.setAttribute('aria-selected', String(Boolean(option && option.selected)))
    })
  }

  if (trigger) trigger.setAttribute('aria-expanded', String(isMenuOpen(menuOf(selector))))
}

function setActiveOption(selector, item) {
  const list = optionsListOf(selector)
  const search = searchOf(selector)
  if (!list || !search) return

  list.querySelectorAll('[data-cfop-option]').forEach((option) => {
    if (option === item) option.dataset.active = 'true'
    else delete option.dataset.active
  })

  if (item) search.setAttribute('aria-activedescendant', item.id)
  else search.removeAttribute('aria-activedescendant')
}

function visibleOptions(selector) {
  const list = optionsListOf(selector)
  if (!list) return []
  return Array.from(list.querySelectorAll('[data-cfop-option]')).filter(
    (option) => !option.hidden && !option.closest('[hidden]')
  )
}

function applyFilter(selector, query) {
  const select = selectOf(selector)
  const list = optionsListOf(selector)
  if (!select || !list) return

  list.querySelectorAll('[data-cfop-option]').forEach((item) => {
    const option = select.options[Number(item.dataset.cfopOptionIndex)]
    item.hidden = !cfopOptionMatches(option, query)
  })

  list.querySelectorAll('[data-cfop-option-group]').forEach((group) => {
    group.hidden = group.querySelectorAll('[data-cfop-option]:not([hidden])').length === 0
  })

  const emptyState = selector.querySelector('[data-cfop-empty]')
  if (emptyState) emptyState.hidden = visibleOptions(selector).length > 0

  const search = searchOf(selector)
  const active = search && search.getAttribute('aria-activedescendant')
  const activeItem = active ? document.getElementById(active) : null
  if (!activeItem || activeItem.hidden || activeItem.closest('[hidden]')) {
    setActiveOption(selector, visibleOptions(selector)[0] || null)
  }
}

function syncSelector(selector) {
  if (!selector) return
  assignControlIds(selector)
  renderOptionList(selector)
  syncSelectedOption(selector)
  applyFilter(selector, searchOf(selector)?.value || '')
}

function positionMenu(selector) {
  const trigger = triggerOf(selector)
  const menu = menuOf(selector)
  const list = optionsListOf(selector)
  if (!trigger || !menu || !list) return

  const rect = trigger.getBoundingClientRect()
  const margin = 8
  const gap = 4
  const width = Math.max(0, Math.min(rect.width, window.innerWidth - margin * 2))
  const left = Math.max(margin, Math.min(rect.left, window.innerWidth - width - margin))
  const spaceBelow = Math.max(0, window.innerHeight - rect.bottom - margin - gap)
  const spaceAbove = Math.max(0, rect.top - margin - gap)
  const openAbove = spaceBelow < 240 && spaceAbove > spaceBelow
  const available = openAbove ? spaceAbove : spaceBelow
  const panelMaxHeight = Math.min(360, available)

  menu.style.width = `${width}px`
  menu.style.left = `${left}px`
  menu.style.maxHeight = `${panelMaxHeight}px`
  menu.style.top = openAbove
    ? `${Math.max(margin, rect.top - gap - panelMaxHeight)}px`
    : `${rect.bottom + gap}px`
  list.style.maxHeight = `${Math.max(48, panelMaxHeight - 58)}px`
}

function closeMenu(selector, restoreFocus = false) {
  const menu = menuOf(selector)
  if (!menu || !isMenuOpen(menu)) return

  menu.dataset.cfopOpen = 'false'
  if (typeof menu.hidePopover === 'function') {
    try {
      menu.hidePopover()
    } catch (_) {
      // Older browsers fall back to the data attribute's CSS visibility.
    }
  }

  const search = searchOf(selector)
  if (search) {
    search.value = ''
    search.removeAttribute('aria-activedescendant')
  }
  applyFilter(selector, '')
  syncSelectedOption(selector)
  if (restoreFocus) triggerOf(selector)?.focus()
}

function closeOtherMenus(exceptSelector = null) {
  document.querySelectorAll('[data-cfop-selector]').forEach((selector) => {
    if (selector !== exceptSelector) closeMenu(selector)
  })
}

function openMenu(selector, initialDirection = 0) {
  const menu = menuOf(selector)
  const trigger = triggerOf(selector)
  const search = searchOf(selector)
  if (!menu || !trigger || !search) return

  closeOtherMenus(selector)
  menu.dataset.cfopOpen = 'true'
  if (typeof menu.showPopover === 'function') {
    try {
      menu.showPopover()
    } catch (_) {
      // The data attribute also supports browsers without the Popover API.
    }
  }

  trigger.setAttribute('aria-expanded', 'true')
  applyFilter(selector, search.value)
  positionMenu(selector)
  search.focus()

  const options = visibleOptions(selector)
  const selected = options.find((option) => option.getAttribute('aria-selected') === 'true')
  if (initialDirection) {
    setActiveOption(selector, initialDirection > 0 ? options[0] : options[options.length - 1])
  } else {
    setActiveOption(selector, selected || options[0] || null)
  }
}

function chooseOption(selector, item) {
  const select = selectOf(selector)
  if (!select || !item) return

  const value = item.dataset.cfopValue
  if (!Array.from(select.options).some((option) => option.value === value)) return
  select.value = value
  select.dispatchEvent(new Event('change', { bubbles: true }))
  closeMenu(selector, true)
}

function selectExactCode(selector, query) {
  const code = String(query || '').trim()
  if (!/^\d{4}$/.test(code)) return

  const select = selectOf(selector)
  if (!select) return
  const option = Array.from(select.options).find((candidate) => candidate.value === code)
  if (!option) return

  select.value = code
  syncSelectedOption(selector)
  const item = optionsListOf(selector)?.querySelector(
    `[data-cfop-option-index="${option.index}"]`
  )
  if (item && !item.hidden) setActiveOption(selector, item)
}

function handleClick(event) {
  const target = event.target
  if (!target || !target.closest) return

  const trigger = target.closest('[data-cfop-trigger]')
  if (trigger) {
    event.preventDefault()
    const selector = selectorOf(trigger)
    const menu = menuOf(selector)
    if (isMenuOpen(menu)) closeMenu(selector)
    else openMenu(selector)
    return
  }

  const option = target.closest('[data-cfop-option]')
  if (option) {
    event.preventDefault()
    chooseOption(selectorOf(option), option)
    return
  }

  if (!target.closest('[data-cfop-menu]')) closeOtherMenus()
}

function handleSearch(event) {
  const input = event.target && event.target.closest
    ? event.target.closest('.cfop-search')
    : null
  if (!input) return

  const selector = selectorOf(input)
  if (!selector) return
  applyFilter(selector, input.value)
  const needle = normalizeCfopSearchText(input.value)
  const select = selectOf(selector)
  const options = visibleOptions(selector)
  const matchingOption = needle && select
    ? options.find((item) => optionContainsQuery(
      select.options[Number(item.dataset.cfopOptionIndex)],
      needle,
    ))
    : options.find((item) => item.getAttribute('aria-selected') === 'true') || options[0]
  setActiveOption(selector, matchingOption || options[0] || null)
  selectExactCode(selector, input.value)
}

function handleKeydown(event) {
  const target = event.target
  if (!target || !target.closest) return

  const trigger = target.closest('[data-cfop-trigger]')
  if (trigger && (event.key === 'ArrowDown' || event.key === 'ArrowUp')) {
    event.preventDefault()
    openMenu(selectorOf(trigger), event.key === 'ArrowDown' ? 1 : -1)
    return
  }

  const search = target.closest('.cfop-search')
  if (!search) return
  const selector = selectorOf(search)

  if (event.key === 'Escape') {
    event.preventDefault()
    closeMenu(selector, true)
    return
  }

  if (event.key === 'Tab') {
    closeMenu(selector)
    return
  }

  if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
    const options = visibleOptions(selector)
    if (!options.length) return
    event.preventDefault()

    const activeId = search.getAttribute('aria-activedescendant')
    const current = options.findIndex((option) => option.id === activeId)
    const direction = event.key === 'ArrowDown' ? 1 : -1
    const next = current === -1
      ? (direction > 0 ? 0 : options.length - 1)
      : Math.max(0, Math.min(options.length - 1, current + direction))
    setActiveOption(selector, options[next])
    return
  }

  if (event.key === 'Enter') {
    const activeId = search.getAttribute('aria-activedescendant')
    const activeOption = activeId ? document.getElementById(activeId) : null
    const code = String(search.value || '').trim()
    const select = selectOf(selector)
    const exactOption = /^\d{4}$/.test(code) && select
      ? Array.from(select.options).find((option) => option.value === code)
      : null
    const exactItem = exactOption
      ? optionsListOf(selector)?.querySelector(`[data-cfop-option-index="${exactOption.index}"]`)
      : null
    const firstVisible = visibleOptions(selector)[0]
    const chosen = exactItem && !exactItem.hidden ? exactItem : activeOption || firstVisible
    if (chosen && !chosen.hidden && !chosen.closest('[hidden]')) {
      event.preventDefault()
      chooseOption(selector, chosen)
    }
  }
}

function handleChange(event) {
  const select = event.target
  if (!select || !select.matches || !select.matches('select.cfop-select[data-cfop-select]')) return
  const selector = selectorOf(select)
  if (selector) syncSelectedOption(selector)
}

function handleAdd(event) {
  const button = event.target && event.target.closest
    ? event.target.closest('[data-cfop-add]')
    : null
  if (!button) return
  pendingOriginSelect = selectOf(button)
}

function repositionOpenMenus() {
  document.querySelectorAll('[data-cfop-menu][data-cfop-open="true"]').forEach((menu) => {
    const selector = selectorOf(menu)
    if (selector) positionMenu(selector)
  })
}

/**
 * Installs delegated listeners once and synchronizes selectors in root. The
 * optional root supports the dynamically cloned detached item rows.
 */
export function initCfopSelectors(root = document) {
  if (!initialized) {
    initialized = true
    document.addEventListener('click', handleClick)
    document.addEventListener('input', handleSearch)
    document.addEventListener('keydown', handleKeydown)
    document.addEventListener('change', handleChange)
    // Capture scroll events from the rascunho editor's scroll container too.
    document.addEventListener('scroll', repositionOpenMenus, true)
    document.addEventListener('click', handleAdd, true)
    window.addEventListener('resize', repositionOpenMenus)
  }

  if (!root) return
  const selectors = []
  if (root.matches && root.matches('[data-cfop-selector]')) selectors.push(root)
  if (root.querySelectorAll) {
    selectors.push(...root.querySelectorAll('[data-cfop-selector]'))
  }
  selectors.forEach((selector) => {
    // A cloned row can inherit the open-state attribute from its source row.
    if (isMenuOpen(menuOf(selector))) closeMenu(selector)
    syncSelector(selector)
  })
}

/**
 * Refreshes every selector with the farm-ordered option list while preserving
 * row selections. The selector that initiated registration selects the new
 * CFOP after the refresh.
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
    if (preferred && Array.from(select.options).some((option) => option.value === preferred)) {
      select.value = preferred
    }

    const selector = selectorOf(select)
    const search = searchOf(selector)
    if (search) search.value = ''
    if (selector) syncSelector(selector)
  })
}

/** Returns and clears the select recorded by the last "+" click. */
export function takePendingOriginSelect() {
  const select = pendingOriginSelect
  pendingOriginSelect = null
  return select
}
