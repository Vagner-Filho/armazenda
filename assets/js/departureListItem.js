import { formatDateToDisplay } from "./date.js"
import { formatWeight } from "./weight.js"

/**
 * Formats a departure list item row
 * @param {string} selector - The CSS selector for the row element
 */
export function formatDepartureListItem(selector) {
    const itemRow = document.querySelector(selector)
    if (itemRow) {
        const dateEl = itemRow.querySelector("td#departureDate")
        if (dateEl && dateEl.hasAttribute("data-raw")) {
            const dateValue = dateEl.getAttribute("data-raw")
            dateEl.textContent = formatDateToDisplay(dateValue)
        }

        const weightEl = itemRow.querySelector("td#netWeight")
        if (weightEl) {
            weightEl.textContent = formatWeight(weightEl.textContent)
        }
    }
}

/**
 * Formats all departure list items on the page
 */
export function formatAllDepartureListItems() {
    const items = document.querySelectorAll("tr[id^='departure-']")
    items.forEach(row => {
        const dateEl = row.querySelector("td#departureDate")
        if (dateEl && dateEl.hasAttribute("data-raw")) {
            const dateValue = dateEl.getAttribute("data-raw")
            dateEl.textContent = formatDateToDisplay(dateValue)
        }

        const weightEl = row.querySelector("td#netWeight")
        if (weightEl) {
            weightEl.textContent = formatWeight(weightEl.textContent)
        }
    })
}

// Make it available globally for HTMX afterSwap calls
window.formatDepartureListItem = formatDepartureListItem
