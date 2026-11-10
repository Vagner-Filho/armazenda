/**
 * Unit tests for the pt-BR decimal mask core (assets/js/decimalMask.js).
 *
 * DOM-free: only the pure formatting/parsing/cursor surface is covered here
 * (same pattern as detachedNFeModes.js). The DOM wiring and the htmx
 * parameter override are covered by the E2E layer
 * (test/e2e/tests/detached_decimals.spec.js).
 */

import { describe, it, expect } from 'bun:test';

import {
    formatDecimalInput,
    computeCursorPosition,
} from '../../assets/js/decimalMask.js';

describe('formatDecimalInput — raw is the dot-decimal contract value', () => {
    it('accepts the comma as the decimal separator (placeholder-promised input)', () => {
        expect(formatDecimalInput('1250,90')).toEqual({ formatted: '1.250,90', raw: '1250.90' });
        expect(formatDecimalInput('17,00')).toEqual({ formatted: '17,00', raw: '17.00' });
        expect(formatDecimalInput('0,9')).toEqual({ formatted: '0,9', raw: '0.9' });
        expect(formatDecimalInput('1,5')).toEqual({ formatted: '1,5', raw: '1.5' });
    });

    it('accepts the dot as the decimal separator (server-contract input)', () => {
        expect(formatDecimalInput('1250.90')).toEqual({ formatted: '1.250,90', raw: '1250.90' });
        expect(formatDecimalInput('17.00')).toEqual({ formatted: '17,00', raw: '17.00' });
        expect(formatDecimalInput('0.900')).toEqual({ formatted: '0,900', raw: '0.900' });
    });

    it('accepts already-grouped pt-BR display strings (round-trips through editing)', () => {
        expect(formatDecimalInput('1.250,90')).toEqual({ formatted: '1.250,90', raw: '1250.90' });
        expect(formatDecimalInput('1.234.567,89')).toEqual({
            formatted: '1.234.567,89',
            raw: '1234567.89',
        });
    });

    it('keeps the integer part untouched when there is no separator', () => {
        expect(formatDecimalInput('1250')).toEqual({ formatted: '1.250', raw: '1250' });
        expect(formatDecimalInput('12500')).toEqual({ formatted: '12.500', raw: '12500' });
        expect(formatDecimalInput('7')).toEqual({ formatted: '7', raw: '7' });
    });

    it('accepts numbers and null-ish inputs', () => {
        expect(formatDecimalInput(1250.9)).toEqual({ formatted: '1.250,9', raw: '1250.9' });
        expect(formatDecimalInput(null)).toEqual({ formatted: '', raw: '' });
        expect(formatDecimalInput(undefined)).toEqual({ formatted: '', raw: '' });
    });

    it('returns empty formatted and raw for empty or separator-only input', () => {
        expect(formatDecimalInput('')).toEqual({ formatted: '', raw: '' });
        expect(formatDecimalInput(',')).toEqual({ formatted: '', raw: '' });
        expect(formatDecimalInput('.')).toEqual({ formatted: '', raw: '' });
        expect(formatDecimalInput(',,')).toEqual({ formatted: '', raw: '' });
    });

    it('strips characters that are not digits or separators', () => {
        expect(formatDecimalInput('abc1,2x3')).toEqual({ formatted: '1,23', raw: '1.23' });
        expect(formatDecimalInput(' 1250,90 ')).toEqual({
            formatted: '1.250,90',
            raw: '1250.90',
        });
    });

    it('keeps a single zero for a decimal-only value', () => {
        expect(formatDecimalInput(',5')).toEqual({ formatted: '0,5', raw: '0.5' });
        expect(formatDecimalInput('.05')).toEqual({ formatted: '0,05', raw: '0.05' });
    });

    it('keeps typed leading zeros in the raw while normalizing the display', () => {
        expect(formatDecimalInput('0,05')).toEqual({ formatted: '0,05', raw: '0.05' });
        expect(formatDecimalInput('5')).toEqual({ formatted: '5', raw: '5' });
    });
});

describe('formatDecimalInput — grouping and decimals caps', () => {
    it('groups thousands on every scale', () => {
        expect(formatDecimalInput('1234567,89').formatted).toBe('1.234.567,89');
        expect(formatDecimalInput('999')).toEqual({ formatted: '999', raw: '999' });
        expect(formatDecimalInput('1000')).toEqual({ formatted: '1.000', raw: '1000' });
    });

    it('truncates the display at the decimals cap and keeps typed digits in the raw', () => {
        const at3 = formatDecimalInput('1250,9055', { decimals: 3 });
        expect(at3.formatted).toBe('1.250,905');
        expect(at3.raw).toBe('1250.9055');

        const at2 = formatDecimalInput('1250,9055', { decimals: 2 });
        expect(at2.formatted).toBe('1.250,90');
        expect(at2.raw).toBe('1250.9055');

        expect(formatDecimalInput('17,005', { decimals: 2 }).formatted).toBe('17,00');
    });

    it('does not cap the integer part digit count', () => {
        const long = '123456789012,34';
        expect(formatDecimalInput(long).raw).toBe('123456789012.34');
    });
});

describe('formatDecimalInput — trailing separator (live vs final)', () => {
    it('keeps the just-typed separator visible while typing', () => {
        expect(formatDecimalInput('1250,')).toEqual({ formatted: '1.250,', raw: '1250' });
        expect(formatDecimalInput('12500,')).toEqual({ formatted: '12.500,', raw: '12500' });
        expect(formatDecimalInput('1250.')).toEqual({ formatted: '1.250,', raw: '1250' });
    });

    it('drops the trailing separator on the final (blur) format', () => {
        expect(formatDecimalInput('1250,', { final: true })).toEqual({
            formatted: '1.250',
            raw: '1250',
        });
        expect(formatDecimalInput('17,00', { final: true })).toEqual({
            formatted: '17,00',
            raw: '17.00',
        });
    });
});

describe('computeCursorPosition — generalized from weight.js', () => {
    it('keeps the cursor at the end when typing triggers regrouping', () => {
        // Typed "1250" (4 digits) and display regroups to "1.250"
        expect(computeCursorPosition('1250', 4, '1.250')).toBe(5);
        // Typed the 5th digit and regrouped to "12.500"
        expect(computeCursorPosition('12500', 5, '12.500')).toBe(6);
    });

    it('keeps the cursor at the end when deleting triggers de-regrouping', () => {
        expect(computeCursorPosition('1.250', 3, '125')).toBe(3);
    });

    it('maps a mid-string cursor by significant character count', () => {
        // Cursor after two significant chars in "1250" → after the separator in "1.250"
        expect(computeCursorPosition('1250', 2, '1.250')).toBe(5);
    });

    it('clamps impossible positions to the end', () => {
        expect(computeCursorPosition('', 0, '1.250')).toBe(0);
        expect(computeCursorPosition('1', 2, '999')).toBe(3);
    });
});
