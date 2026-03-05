/**
 * Unit tests for TemplateRenderer
 * Tests the _parseTemplateFunctions method with real list-item templates
 */

import { describe, it, expect } from 'bun:test';
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

  // Load actual templates from the project
  const templates = {
    entry: readFileSync(join(__dirname, '../../templates/entry/entry-list-item.html'), 'utf-8'),
    person: readFileSync(join(__dirname, '../../templates/person/person-list-item.html'), 'utf-8'),
    departure: readFileSync(join(__dirname, '../../templates/departure/departure-list-item.html'), 'utf-8'),
    entryDraft: readFileSync(join(__dirname, '../../templates/entry/entry-draft-list-item.html'), 'utf-8'),
  };

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

      expect(html).toInclude('text-yellow-300');
      expect(html).not.toInclude('text-emerald-300');
      expect(html).toInclude('Milho');
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
      expect(html).toInclude('Soja');
    });

    it('replaces all placeholder values', () => {
      const html = renderer._parseTemplateFunctions(templates.entry, {
        Id: 123,
        Product: 'Milho',
        Origin: 'Test Farm',
        Field: 'North Field',
        Vehicle: 'XYZ9999',
        NetWeight: 25000.00,
        ArrivalDate: '2024-03-01'
      });

      expect(html).toInclude('123');
      expect(html).toInclude('Test Farm');
      expect(html).toInclude('North Field');
      expect(html).toInclude('XYZ9999');
    });

    it('removes template markers from output', () => {
      const html = renderer._parseTemplateFunctions(templates.entry, {
        Id: 1,
        Product: 'Milho',
        Origin: 'Farm',
        Field: 'A1',
        Vehicle: 'ABC123',
        NetWeight: 1000,
        ArrivalDate: '2024-01-01'
      });

      expect(html).not.toInclude('{{');
      expect(html).not.toInclude('{{ end }}');
      expect(html).not.toInclude('{{end}}');
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
      expect(html).toInclude('/person/natural/form/1');
      expect(html).not.toInclude('/person/legal/form/1');
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
      expect(html).toInclude('/person/legal/form/2');
      expect(html).not.toInclude('/person/natural/form/2');
    });

    it('replaces person data placeholders', () => {
      const html = renderer._parseTemplateFunctions(templates.person, {
        Id: 42,
        Type: 0,
        Name: 'Maria Souza',
        Document: '98765432100',
        IE: ''
      });

      expect(html).toInclude('Maria Souza');
      expect(html).toInclude('98765432100');
      expect(html).toInclude('person-42');
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

    it('combines multiple custom functions', () => {
      const customFuncs = {
        gt: (a, b) => a > b,
        lt: (a, b) => a < b,
        between: (a, b, c) => a > b && a < c
      };

      const template = '{{ if between .Value 10 20 }}In Range{{ else }}Out of Range{{ end }}';
      const inRange = renderer._parseTemplateFunctions(template, { Value: 15 }, customFuncs);
      const outRange = renderer._parseTemplateFunctions(template, { Value: 25 }, customFuncs);

      expect(inRange).toInclude('In Range');
      expect(outRange).toInclude('Out of Range');
    });
  });

  describe('recursive template parsing', () => {
    it('processes multiple template functions in one pass', () => {
      // Template with multiple if/eq blocks
      const multiTemplate = `
        {{ if eq .Color "red" }}Red{{ else }}Not Red{{ end }}
        {{ if eq .Size "large" }}Large{{ else }}Small{{ end }}
        {{ if eq .Active true }}Active{{ else }}Inactive{{ end }}
      `;

      const html = renderer._parseTemplateFunctions(multiTemplate, {
        Color: 'red',
        Size: 'large',
        Active: true
      });

      expect(html).toInclude('Red');
      expect(html).toInclude('Large');
      expect(html).toInclude('Active');
      expect(html).not.toInclude('{{ if');
    });

    it('handles nested conditionals', () => {
      // Simulating nested logic with multiple templates
      const html = renderer._parseTemplateFunctions(templates.person, {
        Id: 1,
        Type: 0,
        Name: 'Test Person',
        Document: '123',
        IE: 'IE123'
      });

      // Should have both Type display and correct edit URL
      expect(html).toInclude('Pessoa Física');
      expect(html).toInclude('/person/natural/form/1');
      // No template markers should remain
      expect(html).not.toInclude('{{');
    });
  });

  describe('edge cases', () => {
    it('handles empty data gracefully', () => {
      const html = renderer._parseTemplateFunctions(templates.entry, {});
      
      // Should still process template structure even with empty data
      expect(html).not.toInclude('{{');
    });

    it('handles missing optional fields', () => {
      const html = renderer._parseTemplateFunctions(templates.person, {
        Id: 1,
        Type: 0,
        Name: 'Test',
        Document: '123',
        // IE is missing but optional
      });

      expect(html).toInclude('Test');
      expect(html).not.toInclude('{{');
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
  });
});
