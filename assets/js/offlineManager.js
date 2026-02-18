/**
 * Armazenda Offline Manager
 * Main entry point for offline functionality
 * Coordinates WASM, IndexedDB, Service Worker, and Sync Engine
 * @module offlineManager
 */

import { db } from './db/database.js';
import { syncEngine } from './db/syncEngine.js';
import { wasmCalculator } from './wasmCalculator.js';
import { templateRenderer } from './templateRenderer.js';

/**
 * Main coordinator for all offline functionality
 * Manages service worker, sync engine, WASM calculator, and UI indicators
 * @class
 */
class OfflineManager {
  constructor() {
    /** @type {boolean} Whether offline manager has been initialized */
    this.initialized = false;
    /** @type {boolean} Current online status */
    this.online = navigator.onLine;
    /** @type {Object} Current sync status */
    this.syncStatus = { lastSync: null, pendingCount: 0 };
    /** @type {Array<Function>} Event listeners */
    this.listeners = [];
  }

  /**
   * Initialize the offline manager
   * Sets up all offline components
   * @returns {Promise<void>}
   */
  async init() {
    if (this.initialized) {
      return;
    }

    console.log('[Offline] Initializing...');

    const cookieFarmId = await cookieStore.get('farmId');
    await this.initialLoad(Number(cookieFarmId.value));

    // Initialize IndexedDB
    await db.init();

    // Initialize sync engine
    await syncEngine.init();
    syncEngine.addListener(this.handleSyncEvent.bind(this));

    // Load WASM calculator
    try {
      await wasmCalculator.load();
    } catch (error) {
      console.warn('[Offline] WASM calculator failed to load:', error);
    }

    // Register service worker
    await this.registerServiceWorker();

    // Listen for online/offline events
    window.addEventListener('online', () => this.handleOnline());
    window.addEventListener('offline', () => this.handleOffline());

    // Pre-cache templates
    await templateRenderer.preCacheCommonTemplates();

    // Update sync status
    this.syncStatus = await syncEngine.getSyncStatus();

    this.initialized = true;
    console.log('[Offline] Initialization complete');

    // Show offline indicator if needed
    this.updateOfflineIndicator();
  }

  /**
   * Register service worker for PWA functionality
   * @returns {Promise<void>}
   */
  async registerServiceWorker() {
    if (!('serviceWorker' in navigator)) {
      console.warn('[Offline] Service Worker not supported');
      return;
    }

    try {
      const registration = await navigator.serviceWorker.register('/public/assets/sw.js');
      console.log('[Offline] Service Worker registered:', registration.scope);

      // Listen for updates
      registration.addEventListener('updatefound', () => {
        const newWorker = registration.installing;
        newWorker.addEventListener('statechange', () => {
          if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
            // New version available
            this.notify({ type: 'SW_UPDATE_AVAILABLE' });
          }
        });
      });

    } catch (error) {
      console.error('[Offline] Service Worker registration failed:', error);
    }
  }

  /**
   * Handle sync events from sync engine
   * @param {Object} event - Sync event object
   */
  handleSyncEvent(event) {
    this.syncStatus = {
      ...this.syncStatus,
      ...event
    };

    this.notify(event);
    this.updateOfflineIndicator();

    // Show toast for important events
    switch (event.type) {
      case 'SYNC_SUCCESS':
        this.showToast('Sincronização concluída', 'success');
        break;
      case 'SYNC_ERROR':
        this.showToast(`Erro na sincronização: ${event.error}`, 'error');
        break;
      case 'SYNC_ITEM_FAILED':
        this.showToast('Algumas alterações não puderam ser sincronizadas', 'warning');
        break;
      case 'OFFLINE':
        this.showToast('Você está offline. As alterações serão sincronizadas quando a conexão for restabelecida.', 'info');
        break;
      case 'ONLINE':
        this.showToast('Conexão restabelecida. Sincronizando...', 'success');
        break;
    }
  }

  /**
   * Handle browser coming online
   */
  handleOnline() {
    this.online = true;
    this.updateOfflineIndicator();
    this.notify({ type: 'ONLINE' });
  }

  /**
   * Handle browser going offline
   */
  handleOffline() {
    this.online = false;
    this.updateOfflineIndicator();
    this.notify({ type: 'OFFLINE' });
  }

  /**
   * Update offline/sync indicator in UI
   */
  updateOfflineIndicator() {
    let indicator = document.getElementById('offline-indicator');

    if (!indicator) {
      indicator = document.createElement('div');
      indicator.id = 'offline-indicator';
      indicator.style.cssText = `
        position: fixed;
        top: 0;
        left: 0;
        right: 0;
        background: #f59e0b;
        color: white;
        text-align: center;
        padding: 8px;
        font-size: 14px;
        font-weight: 500;
        z-index: 9999;
        transition: transform 0.3s ease;
      `;
      document.body.prepend(indicator);
    }

    if (!this.online) {
      const pendingText = this.syncStatus.pendingCount > 0
        ? ` (${this.syncStatus.pendingCount} pendente${this.syncStatus.pendingCount > 1 ? 's' : ''})`
        : '';
      indicator.textContent = `⚠️ Sem conexão${pendingText}`;
      indicator.style.transform = 'translateY(0)';
    } else if (this.syncStatus.pendingCount > 0) {
      indicator.textContent = `⏳ Sincronizando ${this.syncStatus.pendingCount} alteração${this.syncStatus.pendingCount > 1 ? 'es' : ''}...`;
      indicator.style.background = '#3b82f6';
      indicator.style.transform = 'translateY(0)';
    } else {
      indicator.style.transform = 'translateY(-100%)';
    }
  }

  /**
   * Show toast notification
   * @param {string} message - Toast message
   * @param {string} [type='info'] - Toast type ('info', 'success', 'warning', 'error')
   */
  showToast(message, type = 'info') {
    // Dispatch custom event that toast.js can listen to
    const event = new CustomEvent('toast', {
      bubbles: true,
      detail: {
        Message: message,
        Type: type === 'error' ? 3 : type === 'warning' ? 2 : type === 'success' ? 0 : 1,
        Hint: ''
      }
    });
    document.body.dispatchEvent(event);
  }

  /**
   * Add event listener
   * @param {Function} callback - Event handler
   */
  addListener(callback) {
    this.listeners.push(callback);
  }

  /**
   * Remove event listener
   * @param {Function} callback - Callback to remove
   */
  removeListener(callback) {
    this.listeners = this.listeners.filter(cb => cb !== callback);
  }

  /**
   * Notify all listeners
   * @param {Object} event - Event to broadcast
   */
  notify(event) {
    this.listeners.forEach(callback => {
      try {
        callback(event);
      } catch (err) {
        console.error('Offline manager listener error:', err);
      }
    });
  }

  /**
   * Check if currently online
   * @returns {boolean} True if online
   */
  isOnline() {
    return this.online;
  }

  /**
   * Get current sync status
   * @returns {Promise<Object>} Sync status object
   */
  async getSyncStatus() {
    return syncEngine.getSyncStatus();
  }

  /**
   * Force a sync operation
   * @returns {Promise<void>}
   */
  async forceSync() {
    if (!this.online) {
      this.showToast('Não é possível sincronizar offline', 'warning');
      return;
    }

    await syncEngine.sync();
  }

  /**
   * Initial data load after login
   * @param {number} farmId - Farm ID
   * @returns {Promise<void>}
   */
  async initialLoad(farmId) {
    this.showToast('Carregando dados...', 'info');

    try {
      await syncEngine.initialLoad(farmId);
      this.showToast('Dados carregados com sucesso', 'success');
    } catch (error) {
      console.error('[Offline] Initial load failed:', error);
      this.showToast('Erro ao carregar dados. Tente novamente.', 'error');
      throw error;
    }
  }

  /**
   * Handle HTMX request when offline
   * @param {string|Element} target - HTMX target element or selector
   * @param {string} url - Request URL
   * @param {string} method - HTTP method
   * @param {string|null} [body=null] - Request body
   * @returns {Promise<Object|null>} Queued status or null
   */
  async handleHtmxRequest(target, url, method, body = null) {
    if (this.online) {
      // Let HTMX handle it normally
      return null;
    }

    console.log(`[Offline] Handling HTMX request offline: ${method} ${url}`);

    try {
      // Queue the change
      const entity = this.getEntityFromUrl(url);
      const operation = this.getOperationFromMethod(method, url);

      if (entity && operation) {
        const data = body ? Object.fromEntries(new URLSearchParams(body)) : {};
        await syncEngine.queueChange(operation, entity, data);

        // Update UI optimistically
        await this.updateUIOffline(target, entity, operation, data);

        return { queued: true };
      }
    } catch (error) {
      console.error('[Offline] Failed to handle HTMX request:', error);
    }

    return null;
  }

  /**
   * Get entity type from URL
   * @param {string} url - Request URL
   * @returns {string|null} Entity type ('entry', 'departure', 'person', etc.)
   */
  getEntityFromUrl(url) {
    if (url.includes('/entry/draft')) return 'entryDraft';
    if (url.includes('/entry')) return 'entry';
    if (url.includes('/departure/draft')) return 'departureDraft';
    if (url.includes('/departure')) return 'departure';
    if (url.includes('/person')) return 'person';
    return null;
  }

  /**
   * Get operation from HTTP method
   * @param {string} method - HTTP method
   * @param {string} url - Request URL
   * @returns {string|null} Operation type ('CREATE', 'UPDATE', 'DELETE')
   */
  getOperationFromMethod(method, url) {
    if (method === 'POST') return 'CREATE';
    if (method === 'PUT' || method === 'PATCH') return 'UPDATE';
    if (method === 'DELETE') return 'DELETE';
    return null;
  }

  /**
   * Update UI optimistically when offline
   * @param {string|Element} target - Target element or selector
   * @param {string} entity - Entity type
   * @param {string} operation - Operation type
   * @param {Object} data - Operation data
   */
  async updateUIOffline(target, entity, operation, data) {
    // This would render the appropriate template with the new data
    // For now, just reload the relevant section from IndexedDB

    const targetElement = typeof target === 'string'
      ? document.querySelector(target)
      : target;

    if (!targetElement) return;

    // Show offline indicator on the element
    targetElement.style.opacity = '0.7';
    targetElement.dataset.offline = 'true';
  }

  /**
   * Calculate entry using WASM
   * @param {Object} entry - Entry data
   * @param {Object} personConfig - Person configuration
   * @returns {Promise<Object>} Calculation result
   */
  async calculateEntry(entry, personConfig) {
    if (!wasmCalculator.ready) {
      console.warn('[Offline] WASM not ready, falling back to JS calculation');
      return this.calculateEntryJS(entry, personConfig);
    }

    return wasmCalculator.calculateEntry(entry, personConfig);
  }

  /**
   * Fallback JS calculation (from discount.js)
   * @param {Object} entry - Entry data
   * @param {Object} personConfig - Person configuration
   * @returns {Promise<Object>} Calculation result
   */
  calculateEntryJS(entry, personConfig) {
    // Use the existing discount.js logic
    // This is a fallback when WASM is not available
    const event = new CustomEvent('calculate-entry', {
      detail: { entry, personConfig }
    });
    document.dispatchEvent(event);

    // Return a promise that resolves when calculation is done
    return new Promise((resolve) => {
      const handler = (e) => {
        document.removeEventListener('entry-calculated', handler);
        resolve(e.detail);
      };
      document.addEventListener('entry-calculated', handler);
    });
  }
}

/** @type {OfflineManager} Singleton offline manager instance */
export const offlineManager = new OfflineManager();

// Auto-initialize when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
  offlineManager.init().catch(console.error);
});

// Also initialize on htmx:afterSettle (for dynamically loaded content)
/*
document.addEventListener('htmx:afterSettle', () => {
  if (!offlineManager.initialized) {
    offlineManager.init().catch(console.error);
  }
});*/
