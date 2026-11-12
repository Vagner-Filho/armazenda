/**
 * Unit tests for the weight input mask (assets/js/weight.js).
 *
 * DOM-free: the display formatter (formatForDisplay) and full input
 * sequences through the shared state machine (applyInputToRaw, from
 * decimalMask.js) are covered here, the same pattern as decimalMask.test.js.
 * The keydown filter (comma is the only typable separator; the dot is
 * paste-only) and the DOM wiring are covered by the E2E layer
 * (test/e2e/tests/departure.spec.js).
 *
 * Contract: the mask tracks its own rendering and never re-parses it, so a
 * pasted dot-decimal ("47148.55"), a pasted grouped display ("47.148,55")
 * and a comma-typed value ("47148,55") all canonicalize to the same
 * dot-decimal raw — and a formatted display like "50.000 kg" round-trips to
 * 50000 instead of folding to 50.
 */

import { describe, it, expect } from 'bun:test';

import { formatForDisplay } from '../../assets/js/weight.js';
import { applyInputToRaw, parseToRaw } from '../../assets/js/decimalMask.js';

describe('formatForDisplay — renders the canonical raw', () => {
    it('formats dot-decimal raws with the pt-BR kg suffix', () => {
        expect(formatForDisplay('47148.55')).toEqual({ formatted: '47.148,55 kg', raw: '47148.55' });
        expect(formatForDisplay('9845.13')).toEqual({ formatted: '9.845,13 kg', raw: '9845.13' });
        expect(formatForDisplay('50000')).toEqual({ formatted: '50.000 kg', raw: '50000' });
        expect(formatForDisplay('7')).toEqual({ formatted: '7 kg', raw: '7' });
    });

    it('formats numbers like the server-prefilled init value', () => {
        expect(formatForDisplay(47148.55)).toEqual({ formatted: '47.148,55 kg', raw: '47148.55' });
    });

    it('normalizes a zero-prefixed integer part', () => {
        expect(formatForDisplay('0.5')).toEqual({ formatted: '0,5 kg', raw: '0.5' });
        expect(formatForDisplay('.5')).toEqual({ formatted: '0,5 kg', raw: '0.5' });
    });

    it('renders a trailing point as a live comma without the unit suffix', () => {
        expect(formatForDisplay('47148.')).toEqual({ formatted: '47.148,', raw: '47148' });
    });

    it('drops the trailing point on finalize (blur), keeping the committed integer', () => {
        expect(formatForDisplay('47148.', { final: true })).toEqual({ formatted: '47.148 kg', raw: '47148' });
    });

    it('returns empty formatted and raw for empty input', () => {
        expect(formatForDisplay('')).toEqual({ formatted: '', raw: '' });
        expect(formatForDisplay(null)).toEqual({ formatted: '', raw: '' });
        expect(formatForDisplay(undefined)).toEqual({ formatted: '', raw: '' });
    });
});

describe('parseToRaw — full parse of values that did not come from the mask', () => {
    it('digests a pasted dot-decimal without multiplying it (the 22003 regression)', () => {
        expect(parseToRaw('47148.55')).toBe('47148.55');
        expect(parseToRaw('9845.13')).toBe('9845.13');
    });

    it('digests pasted pt-BR display strings (last comma is the decimal)', () => {
        expect(parseToRaw('47.148,55')).toBe('47148.55');
        expect(parseToRaw('9.845,13')).toBe('9845.13');
        expect(parseToRaw('1.234.567,89')).toBe('1234567.89');
    });

    it('round-trips the mask display with the kg suffix and noise', () => {
        expect(parseToRaw('47.148,55 kg')).toBe('47148.55');
        expect(parseToRaw(' 47148,55 ')).toBe('47148.55');
    });

    it('reads strict pt-BR grouped integers as grouping, not decimals', () => {
        expect(parseToRaw('50.000 kg')).toBe('50000');
        expect(parseToRaw('47.148')).toBe('47148');
        expect(parseToRaw('1.234.567')).toBe('1234567');
    });

    it('keeps server dot-decimal values untouched', () => {
        expect(parseToRaw('17.00')).toBe('17.00');
        expect(parseToRaw('12500')).toBe('12500');
    });

    it('returns empty for empty or digit-less input', () => {
        expect(parseToRaw('')).toBe('');
        expect(parseToRaw('kg')).toBe('');
        expect(parseToRaw(null)).toBe('');
    });
});

describe('applyInputToRaw — tracked typing never re-parses the display', () => {
    // Simulates the DOM flow: each keystroke applies to the state the mask
    // rendered, and the next event diffs against the rendered value.
    function typeKeys(keys, { format = (raw) => formatForDisplay(raw).formatted } = {}) {
        let raw = '';
        let rendered = '';
        for (const key of keys) {
            const inserted = rendered.slice(0, rendered.length) // render placeholder
            // the keystroke inserts at the caret (end of the rendered value
            // in these sequences)
            const nextValue = rendered + key;
            void inserted;
            const result = applyInputToRaw(raw, rendered, nextValue);
            raw = result.raw;
            const display = formatForDisplay(raw);
            rendered = display.formatted === nextValue ? nextValue : display.formatted;
        }
        return { raw, rendered };
    }

    it('typing "47148,55" (keyboard case) canonicalizes to the dot-decimal', () => {
        const { raw, rendered } = typeKeys(['4', '7', '1', '4', '8', ',', '5', '5']);
        expect(raw).toBe('47148.55');
        expect(rendered).toBe('47.148,55 kg');
    });

    it('typing "50000" (integer case) never folds through its own grouping', () => {
        const { raw, rendered } = typeKeys(['5', '0', '0', '0', '0']);
        expect(raw).toBe('50000');
        expect(rendered).toBe('50.000 kg');
    });

    it('a pasted dot-decimal into an empty field parses through the full path', () => {
        const { raw } = applyInputToRaw('', '', '47148.55');
        expect(raw).toBe('47148.55');
    });

    it('a pasted grouped display into an empty field parses through the full path', () => {
        expect(applyInputToRaw('', '', '47.148,55').raw).toBe('47148.55');
        expect(applyInputToRaw('', '', '9.845,13').raw).toBe('9845.13');
    });

    it('clearing the field (fill("")) empties the raw', () => {
        const before = applyInputToRaw('', '', '47148.55');
        const { raw } = applyInputToRaw(before.raw, '47.148,55 kg', '');
        expect(raw).toBe('');
    });

    it('re-focusing shows the comma raw and keeps typing into the fraction', () => {
        // After blur the display is "47.148,55 kg"; focus shows the raw with
        // a comma; the next digit appends to the fraction.
        const typed = typeKeys(['4', '7', '1', '4', '8', ',', '5', '5']);
        const focused = typed.rendered === '' ? '' : '47148,55'; // handleFocus writes this
        const { raw } = applyInputToRaw(typed.raw, focused, '47148,555');
        expect(raw).toBe('47148.555');
    });

    it('blur finalizes from the tracked raw instead of re-parsing the display', () => {
        // The "50.000 kg" fold: formatting 50000 and re-parsing the display
        // used to read back 50.000 (= fifty).
        const { raw } = typeKeys(['5', '0', '0', '0', '0']);
        const blurred = formatForDisplay(raw, { final: true });
        expect(blurred.raw).toBe('50000');
        expect(blurred.formatted).toBe('50.000 kg');
    });
});
