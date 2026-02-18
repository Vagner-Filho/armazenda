/**
 * Offline Template Renderer
 * Renders Go HTML templates client-side using cached templates and IndexedDB data
 * @module templateRenderer
 */

import { db } from './db/database.js';

/**
 * Renders Go HTML templates using cached templates and data
 * Supports basic Go template syntax including variables, ranges, conditionals, and includes
 * @class
 */
class TemplateRenderer {
  constructor() {
    /** @type {Map<string, string>} In-memory template cache */
    this.templates = new Map();
    /** @type {Map<string, Function>} Compiled template cache (reserved for future use) */
    this.compiledTemplates = new Map();
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
   * Supports Go template syntax: {{ .Field }}, {{ range }}, {{ if }}, {{ template }}
   * @param {string} template - Template string
   * @param {Object} data - Data object
   * @returns {Promise<string>} Rendered HTML
   */
  async renderTemplateString(template, data) {
    let result = template;

    // Handle {{ .Field }} - simple field access
    result = result.replace(/\{\{\s*\.([\w]+)\s*\}\}/g, (match, field) => {
      const value = this.getFieldValue(data, field);
      return value !== undefined ? this.escapeHtml(value) : '';
    });

    // Handle {{ .Field.Nested }} - nested field access
    result = result.replace(/\{\{\s*\.([\w.]+)\s*\}\}/g, (match, path) => {
      const value = this.getFieldValue(data, path);
      return value !== undefined ? this.escapeHtml(value) : '';
    });

    // Handle {{ range .Items }} ... {{ end }}
    result = this.renderRangeLoops(result, data);

    // Handle {{ if .Condition }} ... {{ else }} ... {{ end }}
    result = this.renderConditionals(result, data);

    // Handle {{ template "name" . }}
    result = await this.renderTemplateCalls(result, data);

    // Handle built-in functions
    result = this.renderBuiltins(result, data);

    return result;
  }

  /**
   * Get a field value from data object using dot notation
   * @param {Object} data - Data object
   * @param {string} path - Field path (e.g., "user.name")
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
   * Render Go template range loops
   * @param {string} template - Template string
   * @param {Object} data - Data object containing arrays to iterate
   * @returns {string} Template with rendered loops
   */
  renderRangeLoops(template, data) {
    const rangeRegex = /\{\{\s*range\s+\.(\w+)\s*\}\}([\s\S]*?)\{\{\s*end\s*\}\}/g;

    return template.replace(rangeRegex, (match, arrayName, innerTemplate) => {
      const array = data[arrayName];

      if (!Array.isArray(array) || array.length === 0) {
        return '';
      }

      return array.map((item, index) => {
        // Create context with . as current item and $ as root data
        const context = {
          ...item,
          '$': data,
          '@index': index
        };
        return this.renderTemplateStringSync(innerTemplate, context);
      }).join('');
    });
  }

  /**
   * Synchronous version of renderTemplateString for use within loops
   * Does not support nested template calls
   * @param {string} template - Template string
   * @param {Object} data - Data object
   * @returns {string} Rendered HTML
   * @private
   */
  renderTemplateStringSync(template, data) {
    let result = template;

    // Handle field access
    result = result.replace(/\{\{\s*\.([\w.]+)\s*\}\}/g, (match, path) => {
      const value = this.getFieldValue(data, path);
      return value !== undefined ? this.escapeHtml(value) : '';
    });

    // Handle conditionals (simplified)
    result = this.renderConditionals(result, data);

    return result;
  }

  /**
   * Render Go template conditionals
   * @param {string} template - Template string
   * @param {Object} data - Data object
   * @returns {string} Template with rendered conditionals
   */
  renderConditionals(template, data) {
    // {{ if .Condition }} ... {{ end }}
    const ifRegex = /\{\{\s*if\s+\.(\w+)\s*\}\}([\s\S]*?)\{\{\s*end\s*\}\}/g;

    template = template.replace(ifRegex, (match, condition, innerContent) => {
      const value = this.getFieldValue(data, condition);
      const isTruthy = this.isTruthy(value);

      // Check for else
      const elseMatch = innerContent.match(/^(.*?)\{\{\s*else\s*\}\}([\s\S]*)$/);

      if (elseMatch) {
        const ifContent = elseMatch[1];
        const elseContent = elseMatch[2];
        return isTruthy ? ifContent : elseContent;
      }

      return isTruthy ? innerContent : '';
    });

    // {{ if not .Condition }} ... {{ end }}
    const ifNotRegex = /\{\{\s*if\s+not\s+\.(\w+)\s*\}\}([\s\S]*?)\{\{\s*end\s*\}\}/g;

    template = template.replace(ifNotRegex, (match, condition, innerContent) => {
      const value = this.getFieldValue(data, condition);
      const isFalsy = !this.isTruthy(value);
      return isFalsy ? innerContent : '';
    });

    return template;
  }

  /**
   * Check if a value is truthy (Go template semantics)
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
   * Render Go template calls ({{ template "name" . }})
   * @param {string} template - Template string
   * @param {Object} data - Data object
   * @returns {Promise<string>} Template with rendered sub-templates
   */
  async renderTemplateCalls(template, data) {
    const templateCallRegex = /\{\{\s*template\s+"(\w+)"\s*(\.)?\s*\}\}/g;

    let result = template;
    const calls = [];

    let match;
    while ((match = templateCallRegex.exec(template)) !== null) {
      calls.push({
        fullMatch: match[0],
        templateName: match[1],
        useContext: !!match[2]
      });
    }

    for (const call of calls) {
      try {
        const subTemplate = await this.getTemplate(call.templateName);
        if (subTemplate) {
          const subData = call.useContext ? data : {};
          const rendered = await this.renderTemplateString(subTemplate, subData);
          result = result.replace(call.fullMatch, rendered);
        }
      } catch (error) {
        console.warn(`[Template] Failed to render sub-template ${call.templateName}:`, error);
      }
    }

    return result;
  }

  /**
   * Render built-in Go template functions
   * @param {string} template - Template string
   * @param {Object} data - Data object
   * @returns {string} Template with rendered functions
   */
  renderBuiltins(template, data) {
    // len function
    template = template.replace(/\{\{\s*len\s+\.(\w+)\s*\}\}/g, (match, field) => {
      const value = this.getFieldValue(data, field);
      if (Array.isArray(value)) {
        return value.length;
      }
      if (typeof value === 'string') {
        return value.length;
      }
      return 0;
    });

    // gt (greater than)
    template = template.replace(/\{\{\s*gt\s+\.(\w+)\s+(\d+)\s*\}\}/g, (match, field, num) => {
      const value = this.getFieldValue(data, field);
      return value > parseInt(num) ? 'true' : 'false';
    });

    return template;
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
      { name: 'entry-content', url: '/entry/list' },
      { name: 'entry-form', url: '/entry/form' },
      { name: 'entry-list-item', url: '/entry/form' },
      { name: 'entry-draft-form', url: '/entry/draft/form' },
      { name: 'entry-draft-list-item', url: '/entry/draft/list' },
      { name: 'departure-content', url: '/departure/list' },
      { name: 'departure-form', url: '/departure/form' },
      { name: 'departure-list-item', url: '/departure/form' },
      { name: 'person-form', url: '/person/form' },
      { name: 'person-list-item', url: '/pessoa' }
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
