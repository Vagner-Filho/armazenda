/**
 * Detached NF-e emission form (templates/pages/nfe-emit.html): source-mode
 * control for the Destinatário and Transporte blocks.
 *
 * Each block has two mutually exclusive sources — a farm contact picked from a
 * select ("existing") or inline fields ("new"). Only the active source is
 * submitted: its section is visible and its controls are enabled, while the
 * inactive one is hidden and its controls are disabled (disabled controls are
 * not submitted and are not focusable).
 *
 * The functions are DOM-agnostic and do not touch the DOM when imported, so
 * the pure state helper is unit-testable without a DOM library.
 */

/**
 * Resolves the active source flags for a mode.
 * @param {'existing'|'new'} mode the selected source.
 * @returns {{existing: boolean, new: boolean}} the per-source active flags.
 * @throws {Error} when mode is not a known source, so a typo cannot disable
 * both sources.
 */
export function modeState(mode) {
    if (mode !== 'existing' && mode !== 'new') {
        throw new Error(`modo de origem inválido: ${mode}`);
    }
    return { existing: mode === 'existing', new: mode === 'new' };
}

/**
 * Applies a source mode to a pair of sections.
 * @param {'existing'|'new'} mode the selected source.
 * @param {{existing: {section: Element, controls: Iterable<Element>}, new: {section: Element, controls: Iterable<Element>}}} sources
 *   the section element and the controls of each source.
 */
export function applySourceMode(mode, sources) {
    const state = modeState(mode);
    for (const source of ['existing', 'new']) {
        const entry = sources[source];
        if (!entry) continue;
        const active = state[source];
        if (entry.section && entry.section.classList) {
            entry.section.classList.toggle('hidden', !active);
        }
        for (const control of entry.controls || []) {
            if (control) control.disabled = !active;
        }
    }
}
