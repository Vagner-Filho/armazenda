/**
 * Unit tests for TemplateRenderer
 * Tests the _parseTemplateFunctions method with real list-item templates
 */

import { describe, it, expect, beforeAll } from 'bun:test';
import { readFileSync } from 'fs';
import { join } from 'path';
import { fileURLToPath } from 'url';

// Import setup for DOM mocking
import './setup.js';

// Import the TemplateRenderer class
import { TemplateRenderer } from '../../assets/js/templateRenderer.js';

const __dirname = fileURLToPath(new URL('.', import.meta.url));

describe('TemplateRenderer._parseTemplateFunctions', () => {
  const renderer = new TemplateRenderer();
  let templates = {};

  // Load templates before running tests
  beforeAll(() => {
    templates = {
      entry: readFileSync(join(__dirname, '../../templates/entry/entry-list-item.html'), 'utf-8'),
      person: readFileSync(join(__dirname, '../../templates/person/person-list-item.html'), 'utf-8'),
      departure: readFileSync(join(__dirname, '../../templates/departure/departure-list-item.html'), 'utf-8'),
      entryDraft: readFileSync(join(__dirname, '../../templates/entry/entry-draft-list-item.html'), 'utf-8'),
    };
  });

  describe('entry-list-item template', () => {
    it('renders Milho with yellow color class', () => {
      const html = renderer._parseTemplateFunctions(templates.entry, {
        Id: 1,
        Product: 'Milho',
        Origin: 'Farm A',
        Field: 'Field 1',
        Vehicle: 'ABC1234',
        NetWeight: 15000.50,
        ArrivalDate: '2024-01-15'
      });

      // The if/eq conditional should be processed
      expect(html).toInclude('text-yellow-300');
      expect(html).not.toInclude('text-emerald-300');
      // Simple placeholders {{ .Product }} are NOT processed by _parseTemplateFunctions
      // They remain as-is
      expect(html).toInclude('{{ .Product }}');
    });

    it('renders Soja with emerald color class', () => {
      const html = renderer._parseTemplateFunctions(templates.entry, {
        Id: 2,
        Product: 'Soja',
        Origin: 'Farm B',
        Field: 'Field 2',
        Vehicle: 'DEF5678',
        NetWeight: 20000.75,
        ArrivalDate: '2024-01-16'
      });

      expect(html).toInclude('text-emerald-300');
      expect(html).not.toInclude('text-yellow-300');
    });

    it('removes if/eq template markers after processing', () => {
      const html = renderer._parseTemplateFunctions(templates.entry, {
        Id: 1,
        Product: 'Milho',
        Origin: 'Farm',
        Field: 'A1',
        Vehicle: 'ABC123',
        NetWeight: 1000,
        ArrivalDate: '2024-01-01'
      });

      // The if/eq/else/end structure should be removed
      expect(html).not.toInclude('{{ if eq .Product');
      expect(html).not.toInclude('{{ else }}');
      expect(html).not.toInclude('text-yellow-300{{ else }}');
    });

    it('preserves block wrapper and simple placeholders', () => {
      const html = renderer._parseTemplateFunctions(templates.entry, {
        Id: 1,
        Product: 'Milho',
        Origin: 'Farm',
        Field: 'A1',
        Vehicle: 'ABC123',
        NetWeight: 1000,
        ArrivalDate: '2024-01-01'
      });

      // Block wrapper is not processed
      expect(html).toInclude('{{ block "entry-list-item" . }}');
      expect(html).toInclude('{{ end }}');
      // Simple placeholders remain
      expect(html).toInclude('{{ .Id }}');
      expect(html).toInclude('{{ .Origin }}');
    });
  });

  describe('person-list-item template', () => {
    it('renders Pessoa Física for Type 0', () => {
      const html = renderer._parseTemplateFunctions(templates.person, {
        Id: 1,
        Type: 0,
        Name: 'João Silva',
        Document: '12345678901',
        IE: ''
      });

      expect(html).toInclude('Pessoa Física');
      expect(html).not.toInclude('Pessoa Jurídica');
      // The URL part with {{ .Id }} is not replaced but the if/eq is processed
      expect(html).toInclude('/person/natural/form/');
    });

    it('renders Pessoa Jurídica for Type 1', () => {
      const html = renderer._parseTemplateFunctions(templates.person, {
        Id: 2,
        Type: 1,
        Name: 'Empresa ABC Ltda',
        Document: '12345678000195',
        IE: '123456789'
      });

      expect(html).toInclude('Pessoa Jurídica');
      expect(html).not.toInclude('Pessoa Física');
      expect(html).toInclude('/person/legal/form/');
    });

    it('removes if/eq/else/end for Type display', () => {
      const html = renderer._parseTemplateFunctions(templates.person, {
        Id: 1,
        Type: 0,
        Name: 'João',
        Document: '123',
        IE: ''
      });

      // The if/eq structure for Type should be removed
      expect(html).not.toInclude('{{ if eq .Type');
      expect(html).not.toInclude('{{ else }}');
    });

    it('processes both Type conditionals in template', () => {
      const html = renderer._parseTemplateFunctions(templates.person, {
        Id: 1,
        Type: 0,
        Name: 'Test',
        Document: '123',
        IE: ''
      });

      // Both if/eq conditionals should be processed
      expect(html).toInclude('Pessoa Física');
      expect(html).toInclude('/person/natural/form/');
      // Simple placeholders remain
      expect(html).toInclude('{{ .Name }}');
    });
  });

  describe('custom template functions', () => {
    it('supports custom greater-than function', () => {
      const customFuncs = {
        gt: (a, b) => a > b
      };

      const template = '{{ if gt .Value 100 }}High{{ else }}Low{{ end }}';
      const highResult = renderer._parseTemplateFunctions(template, { Value: 150 }, customFuncs);
      const lowResult = renderer._parseTemplateFunctions(template, { Value: 50 }, customFuncs);

      expect(highResult).toInclude('High');
      expect(lowResult).toInclude('Low');
    });

    it('supports custom less-than function', () => {
      const customFuncs = {
        lt: (a, b) => a < b
      };

      const template = '{{ if lt .Value 50 }}Small{{ else }}Large{{ end }}';
      const smallResult = renderer._parseTemplateFunctions(template, { Value: 30 }, customFuncs);
      const largeResult = renderer._parseTemplateFunctions(template, { Value: 100 }, customFuncs);

      expect(smallResult).toInclude('Small');
      expect(largeResult).toInclude('Large');
    });

    it('custom functions override defaults when names match', () => {
      const customFuncs = {
        eq: (a, b) => a === b  // Strict equality instead of loose
      };

      const template = '{{ if eq .Value "5" }}Match{{ else }}No Match{{ end }}';
      // With strict equality, number 5 !== string "5"
      const result = renderer._parseTemplateFunctions(template, { Value: 5 }, customFuncs);
      
      expect(result).toInclude('No Match');
    });
  });

  describe('recursive template parsing', () => {
    it('processes multiple template functions in one pass', () => {
      // Template with multiple if/eq blocks
      const multiTemplate = `
        {{ if eq .Color "red" }}Red{{ else }}Not Red{{ end }}
        {{ if eq .Size "large" }}Large{{ else }}Small{{ end }}
      `;

      const html = renderer._parseTemplateFunctions(multiTemplate, {
        Color: 'red',
        Size: 'large'
      });

      expect(html).toInclude('Red');
      expect(html).toInclude('Large');
      expect(html).not.toInclude('{{ if');
    });

    it('handles complex nested-like templates', () => {
      // Simulating multiple independent if blocks
      const template = '{{ if eq .A "1" }}A1{{ end }}{{ if eq .B "2" }}B2{{ end }}';
      const html = renderer._parseTemplateFunctions(template, { A: '1', B: '2' });

      expect(html).toInclude('A1');
      expect(html).toInclude('B2');
      expect(html).not.toInclude('{{');
    });

    it('does not hang on unclosed templates', () => {
      // This should not cause infinite loop
      const template = '{{ .Id }} {{ .Name }}';
      const result = renderer._parseTemplateFunctions(template, { Id: 1, Name: 'Test' });
      
      // These are not processed, just skipped
      expect(result).toBeDefined();
    });

    it('processes templates without else', () => {
      const template = '{{ if eq .X "a" }}Yes{{ end }}';
      const result = renderer._parseTemplateFunctions(template, { X: 'a' });
      
      expect(result).toInclude('Yes');
    });
  });

  describe('edge cases', () => {
    it('handles empty data gracefully', () => {
      const html = renderer._parseTemplateFunctions(templates.entry, {});
      
      // Should not hang, just process what it can
      expect(html).toBeDefined();
      // Block wrapper and simple placeholders remain
      expect(html).toInclude('{{ block');
    });

    it('handles missing data fields', () => {
      const html = renderer._parseTemplateFunctions(templates.person, {
        Id: 1,
        Type: 0,
        Name: 'Test',
        Document: '123'
        // IE is missing
      });

      expect(html).toInclude('Pessoa Física');
      expect(html).toBeDefined();
    });

    it('preserves HTML structure', () => {
      const html = renderer._parseTemplateFunctions(templates.entry, {
        Id: 1,
        Product: 'Milho',
        Origin: 'Farm',
        Field: 'A1',
        Vehicle: 'ABC123',
        NetWeight: 1000,
        ArrivalDate: '2024-01-01'
      });

      // Should preserve tr element structure
      expect(html).toInclude('<tr');
      expect(html).toInclude('</tr>');
      // Should preserve td elements
      expect(html).toInclude('<td');
      expect(html).toInclude('</td>');
    });

    it('handles templates with no conditionals', () => {
      const template = '<div>{{ .Name }}</div>';
      const result = renderer._parseTemplateFunctions(template, { Name: 'Test' });
      
      // No if/eq to process, should just skip the {{ .Name }}
      expect(result).toBeDefined();
      expect(result).toInclude('<div>');
    });
  });
});
