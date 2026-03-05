/**
 * Offline Template Renderer
 * Renders pre-processed templates with simple placeholder replacement
 * @module templateRenderer
 */

import { db } from './db/database.js';

/**
 * Renders HTML templates using cached templates and data
 * Supports simple {Placeholder} syntax and data attributes for conditionals
 * @class
 */
class TemplateRenderer {
  constructor() {
    /** @type {Map<string, string>} In-memory template cache */
    this.templates = new Map();
  }

  /**
   * Cache a template from the server
   * @param {string} name - Template identifier
   * @param {string} url - URL to fetch template from
   * @returns {Promise<string>} The cached HTML content
   * @throws {Error} If fetching or caching fails
   */
  async cacheTemplate(name, url) {
    try {
      const response = await fetch(url);
      if (!response.ok) {
        throw new Error(`Failed to fetch template: ${response.status}`);
      }

      const html = await response.text();
      await db.saveTemplate(name, html);
      this.templates.set(name, html);

      return html;
    } catch (error) {
      console.error(`[Template] Failed to cache ${name}:`, error);
      throw error;
    }
  }

  /**
   * Get a template from cache or IndexedDB
   * @param {string} name - Template name
   * @returns {Promise<string|null>} The template HTML or null if not found
   */
  async getTemplate(name) {
    // Check memory cache
    if (this.templates.has(name)) {
      return this.templates.get(name);
    }

    // Check IndexedDB
    const html = await db.getTemplate(name);
    if (html) {
      this.templates.set(name, html);
      return html;
    }

    return null;
  }

  /**
   * Render a template with data
   * Replaces {FieldName} placeholders with values from data object
   * Handles data-show-if and data-hide-if attributes for conditionals
   * @param {string} templateName - Name of the template to render
   * @param {Object} data - Data object to use for rendering
   * @returns {Promise<string>} Rendered HTML string
   * @throws {Error} If template is not found
   */
  async render(templateName, data) {
    const template = await this.getTemplate(templateName);

    if (!template) {
      throw new Error(`Template not found: ${templateName}`);
    }

    return this.renderTemplateString(template, data);
  }

  /**
   * Render a template string with data
   * @param {string} template - Template string with {Placeholder} syntax
   * @param {Object} data - Data object
   * @returns {string} Rendered HTML
   */
  renderTemplateString(template, data) {
    let html = template;

    // Removes enclosing template syntax
    if (html.startsWith("{{")) {
      html = html.substring(html.indexOf("<"), html.lastIndexOf(">") + 1);
    }

    // Replace {FieldName} placeholders with data values
    // Matches {FieldName} or {Object.FieldName}
    // Uses negative lookbehind to skip ES6 import statements (e.g., import {setupEntryForm})
    html = html.replace(/(?<!import )\{(\w+(?:\.\w+)*)\}/g, (match, fieldPath) => {
      const value = this.getFieldValue(data, fieldPath);
      return value !== undefined && value !== null ? this.escapeHtml(value) : '';
    });

    const templateIdentifierIdx = html.indexOf("{{");
    if (templateIdentifierIdx > -1) {
      const assertionEnd = html.indexOf("}}", templateIdentifierIdx);
      let templateEnd = html.indexOf("{{ end }}", templateIdentifierIdx);
      if (templateEnd === -1) {
        templateEnd = html.indexOf("{{end}}", templateIdentifierIdx);
      }

      if (assertionEnd > -1 && templateEnd > -1) {
        const funcs = {
          if: (assertion, param1, param2) => assertion(param1, param2),
          eq: (a, b) => {
            if (typeof a !== typeof b) {
              return a == b;
            }
            const asserters = {
              string: (a, b) => a.match(b) || b.match(a),
              number: (a, b) => a === b,
            }

            return asserters[typeof a](a, b)
          }
        }
        const tokens = html.substring(templateIdentifierIdx, assertionEnd).split(' ').filter(token => !['', '{{', '}}'].includes(token.trim()));
        const [templateFunctionName, assertionName, param1Token, param2Token] = tokens;
        if (funcs[templateFunctionName] && funcs[assertionName]) {
          const templateFunction = funcs[templateFunctionName];
          const assertion = funcs[assertionName];

          const param1Parsed = param1Token.replace('.', '');
          const param1 = param1Parsed in data ? data[param1Parsed] : param1Parsed;
          const param2 = param2Token.replace('.', '') in data ? data[param2Token] : param2Token;
          if (templateFunction(assertion, param1, param2)) {
            // TODO: insert true condition value to html templase
            const nextTemplateStart = html.indexOf("{{", assertionEnd);
            const assertionValue = html.substring(assertionEnd + 2, nextTemplateStart);
            html = html.substring(0, templateIdentifierIdx) + assertionValue + html.substring(html.indexOf("}}", templateEnd) + 2);
          } else {
            // TODO: check if template has else, if it does, insert else value to html template
            const template = html.substring(templateIdentifierIdx, templateEnd)
            let elseMarker = template.indexOf("{{ else }}");
            if (elseMarker === -1) {
              elseMarker = template.indexOf("{{else}}");
            }
            if (elseMarker > -1) {
              const assertionValue = template.substring(template.indexOf("}}", elseMarker) + 2, templateEnd);
              html = html.substring(0, templateIdentifierIdx) + assertionValue + html.substring(html.indexOf("}}", templateEnd) + 2);
            }
          }
        }
      }
    }

    if (html.startsWith('<tr')) {
      return html;
    }

    // Parse HTML to handle data attributes
    const container = document.createElement('div');
    container.innerHTML = html;

    // Handle data-show-if attributes
    container.querySelectorAll('[data-show-if]').forEach(el => {
      const fieldPath = el.dataset.showIf;
      const value = this.getFieldValue(data, fieldPath);
      el.style.display = this.isTruthy(value) ? '' : 'none';
    });

    // Handle data-hide-if attributes
    container.querySelectorAll('[data-hide-if]').forEach(el => {
      const fieldPath = el.dataset.hideIf;
      const value = this.getFieldValue(data, fieldPath);
      el.style.display = this.isTruthy(value) ? 'none' : '';
    });

    return container.innerHTML;
  }

  /**
   * Get a field value from data object using dot notation
   * @param {Object} data - Data object
   * @param {string} path - Field path (e.g., "Entry.Id" or "Id")
   * @returns {*} The field value or undefined
   */
  getFieldValue(data, path) {
    const keys = path.split('.');
    let value = data;

    for (const key of keys) {
      if (value === null || value === undefined) {
        return undefined;
      }
      value = value[key];
    }

    return value;
  }

  /**
   * Check if a value is truthy
   * @param {*} value - Value to check
   * @returns {boolean} True if value is truthy
   */
  isTruthy(value) {
    if (value === null || value === undefined) return false;
    if (typeof value === 'boolean') return value;
    if (typeof value === 'number') return value !== 0;
    if (typeof value === 'string') return value.length > 0;
    if (Array.isArray(value)) return value.length > 0;
    if (typeof value === 'object') return Object.keys(value).length > 0;
    return true;
  }

  /**
   * Escape HTML entities to prevent XSS
   * @param {*} text - Text to escape
   * @returns {string} Escaped HTML string
   */
  escapeHtml(text) {
    if (text === null || text === undefined) {
      return '';
    }

    const str = String(text);
    return str
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#039;');
  }

  /**
   * Pre-cache common templates for offline use
   * @returns {Promise<void>}
   */
  async preCacheCommonTemplates() {
    const templates = [
      { name: 'entry-form', url: '/api/templates/entry-form' },
      { name: 'entry-list-item', url: '/api/templates/entry-list-item' },
      { name: 'entry-draft-form', url: '/api/templates/entry-draft-form' },
      { name: 'entry-draft-list-item', url: '/api/templates/entry-draft-list-item' },
      { name: 'departure-form', url: '/api/templates/departure-form' },
      { name: 'departure-list-item', url: '/api/templates/departure-list-item' },
      { name: 'departure-draft-form', url: '/api/templates/departure-draft-form' },
      { name: 'departure-draft-list-item', url: '/api/templates/departure-draft-list-item' },
      { name: 'person-form', url: '/api/templates/person-form' },
      { name: 'person-list-item', url: '/api/templates/person-list-item' },
      { name: 'crop-form', url: '/api/templates/crop-form' },
      { name: 'crop-option', url: '/api/templates/crop-option' },
      { name: 'vehicle-form', url: '/api/templates/vehicle-form' },
      { name: 'vehicle-option', url: '/api/templates/vehicle-option' },
      { name: 'field-form', url: '/api/templates/field-form' },
      { name: 'field-option', url: '/api/templates/field-option' }
    ];

    for (const { name, url } of templates) {
      try {
        await this.cacheTemplate(name, url);
      } catch (error) {
        console.warn(`[Template] Failed to pre-cache ${name}:`, error);
      }
    }
  }
}

/** @type {TemplateRenderer} Singleton template renderer instance */
export const templateRenderer = new TemplateRenderer();
