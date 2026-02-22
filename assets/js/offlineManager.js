/**
 * Armazenda Offline Manager
 * Main entry point for offline functionality
 * Coordinates WASM, IndexedDB, Service Worker, and Sync Engine
 * @module offlineManager
 */

import { db, STORES } from './db/database.js';
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
    /** @type {BroadcastChannel|null} Channel for SW communication */
    this.syncChannel = null;
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

    // Setup BroadcastChannel for SW communication
    this.setupBroadcastChannel();

    // Setup HTMX offline handler
    this.setupHtmxOfflineHandler();

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

  setupBroadcastChannel() {
    if (!('BroadcastChannel' in window)) {
      console.warn('[Offline] BroadcastChannel not supported');
      return;
    }

    this.syncChannel = new BroadcastChannel('sw-sync');

    this.syncChannel.onmessage = async (event) => {
      const { type } = event.data;

      switch (type) {
        case 'GET_QUEUE':
          const queue = await db.getSyncMetadata('sw-queue') || [];
          this.syncChannel.postMessage({ type: 'QUEUE_DATA', queue });
          break;

        case 'SAVE_QUEUE':
          await db.setSyncMetadata('sw-queue', event.data.queue);
          this.syncChannel.postMessage({ type: 'QUEUE_SAVED' });
          break;
      }
    };

    console.log('[Offline] BroadcastChannel setup complete');
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
    this.updateEntryButtonStates();
    this.notify({ type: 'ONLINE' });
  }

  /**
   * Handle browser going offline
   */
  handleOffline() {
    this.online = false;
    this.updateOfflineIndicator();
    this.updateEntryButtonStates();
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

  /**
   * Setup HTMX offline handler for Entry operations
   */
  setupHtmxOfflineHandler() {
    document.body.addEventListener('htmx:beforeRequest', async (event) => {
      if (this.online) return;

      const { elt, target, requestConfig } = event.detail;
      const parameters = requestConfig.parameters;

      const method = requestConfig.verb.toUpperCase();
      const path = requestConfig.path;

      // Handle GET requests for forms
      if (method === 'GET' && this.isFormRequest(path)) {
        event.preventDefault();
        await this.handleOfflineFormRequest(path, target);
        return;
      }

      // Only handle Entry mutating requests (not drafts)
      if (!this.isEntryMutatingRequest(method, path)) return;

      event.preventDefault();

      const operation = this.getOperationFromMethod(method, path);

      switch (operation) {
        case 'CREATE':
          await this.handleOfflineEntryCreate(parameters, target, elt);
          break;
        case 'UPDATE':
          await this.handleOfflineEntryUpdate(path, parameters, target, elt);
          break;
        case 'DELETE':
          await this.handleOfflineEntryDelete(path, target, elt);
          break;
      }
    });
  }

  /**
   * Check if request is for a form endpoint
   */
  isFormRequest(path) {
    return path === '/entry/form' ||
      path.startsWith('/entry/form/') ||
      path === '/departure/form' ||
      path.startsWith('/departure/form/');
  }

  /**
   * Handle offline form request by rendering template with IndexedDB data
   */
  async handleOfflineFormRequest(path, target) {
    console.log(`[Offline] Handling form request: ${path}`);

    try {
      // Extract entry ID if editing
      const entryId = this.extractIdFromPath(path);

      // Get reference data from IndexedDB
      const [fields, crops, vehicles, people] = await Promise.all([
        db.getAllFields(),
        db.getAllCrops(),
        db.getAllVehicles(),
        db.getAllPeople()
      ]);

      // Prepare template data
      const templateData = {
        Fields: fields,
        Crops: crops,
        Vehicles: vehicles,
        People: people,
        Entry: entryId ? await db.getEntry(entryId) : null,
        IsOffline: true
      };

      // Determine which template to render
      let templateName;
      if (path.startsWith('/entry/form')) {
        templateName = 'entry-form';
      } else if (path.startsWith('/departure/form')) {
        templateName = 'departure-form';
      } else {
        throw new Error(`Unknown form path: ${path}`);
      }

      // Render the form
      const html = await templateRenderer.render(templateName, templateData);

      // Use HTMX swap to insert HTML and execute scripts properly
      if (window.htmx) {
        await window.htmx.swap(target, html, {
          swapStyle: 'beforeend',
          swapDelay: 0,
          settleDelay: 20
        });

        // Show the dialog after HTMX has processed and inserted the content
        const targetElement = typeof target === 'string'
          ? document.querySelector(target)
          : target;
        const dialog = targetElement.querySelector('dialog');
        if (dialog && dialog.showModal) {
          dialog.showModal();
        }
      } else {
        // Fallback if HTMX is not available
        const targetElement = typeof target === 'string'
          ? document.querySelector(target)
          : target;
        if (targetElement) {
          targetElement.insertAdjacentHTML('beforeend', html);
        }
      }

      this.showToast('Formulário carregado offline', 'info');

    } catch (error) {
      console.error('[Offline] Failed to render form:', error);
      this.showToast('Erro ao carregar formulário offline', 'error');
    }
  }

  /**
   * Check if request is an Entry mutating request
   */
  isEntryMutatingRequest(method, path) {
    return ['POST', 'PUT', 'DELETE'].includes(method) &&
      path.startsWith('/entry') &&
      !path.includes('/draft');
  }

  /**
   * Extract ID from URL path
   */
  extractIdFromPath(path) {
    const match = path.match(/\/(\d+|offline_[\d]+)$/);
    return match ? match[1] : null;
  }

  /**
   * Get selected option text from a select element
   */
  getSelectText(selectId) {
    const select = document.getElementById(selectId);
    return select?.selectedOptions[0]?.text || '';
  }

  /**
   * Calculate net weight from gross and tare
   */
  calculateNetWeight(params) {
    const gross = parseFloat(params.grossWeight) || 0;
    const tare = parseFloat(params.tare) || 0;
    return (gross - tare).toFixed(2);
  }

  /**
   * Close form dialog
   */
  closeFormDialog(element) {
    const dialog = element.closest('dialog');
    if (dialog) {
      dialog.close();
      dialog.remove();
    }
  }

  /**
   * Render Entry list item HTML
   */
  renderEntryListItem(data) {
    const isPending = data.id.toString().startsWith('offline_');
    const pendingBadge = isPending
      ? '<span class="ml-2 px-2 py-0.5 text-xs bg-yellow-500/30 text-yellow-200 rounded-full">Pendente</span>'
      : '';

    const productClass = data.product === 'Milho' ? 'text-yellow-300' : 'text-emerald-300';
    const netWeight = parseFloat(data.netWeight).toLocaleString('pt-BR', { minimumFractionDigits: 2 });
    const arrivalDate = new Date(data.arrivalDate).toLocaleString('pt-BR');

    return `
      <tr class="text-sm border-b border-sky-50/30 hover:bg-sky-50/10 whitespace-nowrap bg-yellow-500/5"
          id="entry-${data.id}"
          ${isPending ? 'data-offline-pending="true"' : ''}>
        <td class="py-3 px-4 text-left">${data.id}</td>
        <td class="py-3 px-4 text-left ${productClass}">${data.product}</td>
        <td class="py-3 px-4 text-left truncate max-w-64" title="${data.originName}">${data.originName}</td>
        <td class="py-3 px-4 text-left">${data.fieldName}</td>
        <td class="py-3 px-4 text-left font-mono">${data.vehiclePlate}</td>
        <td class="py-3 px-4 text-left font-bold">${netWeight} kg</td>
        <td class="py-3 px-4 text-left">${arrivalDate}</td>
        <td class="py-3 px-4 text-left">
          <div class="flex items-center gap-2">
            ${pendingBadge}
            <button type="button" class="icon-btn" onclick="getEntryPdf(this)" data-id="${data.id}">
              <iconify-icon icon="mdi:file-pdf-box" class="text-xl"></iconify-icon>
            </button>
            <button type="button" hx-get="/entry/form/${data.id}" hx-target="body" hx-swap="beforeend"
                    title="Editar" class="icon-btn">
              <iconify-icon icon="mdi:pencil" class="text-xl"></iconify-icon>
            </button>
            <button type="button" hx-delete="/entry/${data.id}" hx-target="#entry-${data.id}"
                    hx-swap="delete" title="Excluir" class="icon-btn">
              <iconify-icon icon="mdi:trash-can" class="text-xl"></iconify-icon>
            </button>
          </div>
        </td>
      </tr>
    `;
  }

  /**
   * Handle offline Entry creation
   */
  async handleOfflineEntryCreate(parameters, target, formElement) {
    const tempId = `offline_${Date.now()}`;

    // Data for UI rendering (display fields + calculated values)
    const displayData = {
      id: tempId,
      field: Number(parameters.field),
      crop: Number(parameters.crop),
      vehicle: parameters.vehicle,
      origin: Number(parameters.origin),
      grossWeight: parameters.grossWeight,
      tare: parameters.tare,
      netWeight: this.calculateNetWeight(parameters),
      humidity: parameters.humidity,
      damage: parameters.damage,
      impurity: parameters.impurity,
      arrivalDate: parameters.arrivalDate || new Date().toISOString(),
      farm: this.farmId,
      product: sessionStorage.getItem('product') || '',
      originName: this.getSelectText('origin-selector'),
      fieldName: this.getSelectText('field-selector'),
      vehiclePlate: this.getSelectText('vehicle-selector')
    };

    // Data for API/backend (matches Go Entry struct)
    const apiData = {
      field: Number(parameters.field),
      crop: Number(parameters.crop),
      vehiclePlate: Number(parameters.vehicle),
      origin: parameters.origin ? Number(parameters.origin) : undefined,
      grossWeight: parameters.grossWeight,
      tare: parameters.tare,
      humidity: parameters.humidity,
      damage: parameters.damage,
      impurity: parameters.impurity,
      arrivalDate: parameters.arrivalDate || new Date().toISOString()
    };

    // Queue the change
    await syncEngine.queueCreate('entry', apiData);

    // Render optimistic list item
    const html = this.renderEntryListItem(displayData);
    target.insertAdjacentHTML('afterbegin', html);

    // Cleanup
    this.closeFormDialog(formElement);
    this.showToast('Entrada salva localmente', 'success');
    this.syncStatus.pendingCount++;
    this.updateOfflineIndicator();
  }

  /**
   * Handle offline Entry update
   */
  async handleOfflineEntryUpdate(path, parameters, target, formElement) {
    const id = this.extractIdFromPath(path);

    // Only allow editing offline-created entries
    if (!id || !id.startsWith('offline_')) {
      this.showToast('Não é possível editar entradas do servidor offline', 'warning');
      return;
    }

    // Find and update pending change
    const pending = await db.getPendingChanges();
    const change = pending.find(c =>
      c.entity === 'entry' &&
      c.data.id === id &&
      c.operation === 'CREATE'
    );

    if (!change) {
      this.showToast('Entrada não encontrada', 'error');
      return;
    }

    // Data for UI rendering (display fields + calculated values)
    const displayData = {
      ...change.data,
      ...parameters,
      netWeight: this.calculateNetWeight(parameters),
      product: sessionStorage.getItem('product') || change.data.product,
      originName: this.getSelectText('origin-selector') || change.data.originName,
      fieldName: this.getSelectText('field-selector') || change.data.fieldName,
      vehiclePlate: this.getSelectText('vehicle-selector') || change.data.vehiclePlate
    };

    // Data for API/backend (matches Go Entry struct)
    const apiData = {
      field: Number(parameters.field),
      crop: Number(parameters.crop),
      vehiclePlate: Number(parameters.vehicle),
      origin: parameters.origin ? Number(parameters.origin) : undefined,
      grossWeight: parameters.grossWeight,
      tare: parameters.tare,
      humidity: parameters.humidity,
      damage: parameters.damage,
      impurity: parameters.impurity,
      arrivalDate: parameters.arrivalDate || new Date().toISOString()
    };

    // Update the pending change with API data
    change.data = apiData;
    await db.put('pendingChanges', change);

    // Update DOM with display data
    const html = this.renderEntryListItem(displayData);
    target.outerHTML = html;

    this.closeFormDialog(formElement);
    this.showToast('Entrada atualizada localmente', 'success');
  }

  /**
   * Handle offline Entry deletion
   */
  async handleOfflineEntryDelete(path, target, buttonElement) {
    const id = this.extractIdFromPath(path);

    // Only allow deleting offline-created entries
    if (!id || !id.startsWith('offline_')) {
      this.showToast('Não é possível excluir entradas do servidor offline', 'warning');
      return;
    }

    // Remove from queue
    const pending = await db.getPendingChanges();
    const change = pending.find(c =>
      c.entity === 'entry' &&
      c.data.id === id &&
      c.operation === 'CREATE'
    );

    if (change) {
      await db.removePendingChange(change.id);
      target.remove();
      this.syncStatus.pendingCount--;
      this.updateOfflineIndicator();
      this.showToast('Entrada removida', 'success');
    }
  }

  /**
   * Update Entry button states based on offline status
   */
  updateEntryButtonStates() {
    const editButtons = document.querySelectorAll('[hx-get^="/entry/form/"]');
    const deleteButtons = document.querySelectorAll('[hx-delete^="/entry/"]');

    const disableIfServer = (btn, pathAttr) => {
      const path = btn.getAttribute(pathAttr) || '';
      const id = this.extractIdFromPath(path);

      if (!this.online && id && !id.toString().startsWith('offline_')) {
        btn.disabled = true;
        btn.classList.add('opacity-50', 'cursor-not-allowed');
        btn.title = 'Indisponível offline';
      } else {
        btn.disabled = false;
        btn.classList.remove('opacity-50', 'cursor-not-allowed');
        btn.title = '';
      }
    };

    editButtons.forEach(btn => disableIfServer(btn, 'hx-get'));
    deleteButtons.forEach(btn => disableIfServer(btn, 'hx-delete'));
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
