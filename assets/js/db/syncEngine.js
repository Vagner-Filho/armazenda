/**
 * Armazenda Sync Engine
 * Handles synchronization between IndexedDB and server
 * @module syncEngine
 */

import { db, STORES } from './database.js';

/**
 * Manages data synchronization between client and server
 * @class
 */
class SyncEngine {
  constructor() {
    /** @type {boolean} Whether sync is currently in progress */
    this.isSyncing = false;
    /** @type {number|null} Sync interval timer ID */
    this.syncInterval = null;
    /** @type {number} Retry delay in milliseconds (1 minute) */
    this.retryDelay = 60000;
    /** @type {Array<Function>} Event listeners */
    this.listeners = [];
  }

  /**
   * Initialize the sync engine
   * Sets up event listeners and performs initial sync if online
   * @returns {Promise<void>}
   */
  async init() {
    await db.init();

    // Listen for online/offline events
    window.addEventListener('online', () => this.handleOnline());
    window.addEventListener('offline', () => this.handleOffline());

    // Listen for service worker messages
    if ('serviceWorker' in navigator) {
      navigator.serviceWorker.addEventListener('message', (event) => {
        this.handleServiceWorkerMessage(event.data);
      });
    }
    
    // If online, sync on init
    if (navigator.onLine) {
      this.sync();
    }
  }

  /**
   * Add a listener for sync events
   * @param {Function} callback - Event handler callback
   */
  addListener(callback) {
    this.listeners.push(callback);
  }

  /**
   * Remove a listener
   * @param {Function} callback - Callback to remove
   */
  removeListener(callback) {
    this.listeners = this.listeners.filter(cb => cb !== callback);
  }

  /**
   * Notify all listeners of an event
   * @param {Object} event - Event object to broadcast
   */
  notify(event) {
    this.listeners.forEach(callback => {
      try {
        callback(event);
      } catch (err) {
        console.error('Sync listener error:', err);
      }
    });
  }

  /**
   * Handle browser coming online
   * Triggers sync operation
   */
  handleOnline() {
    console.log('[Sync] Browser is online');
    this.notify({ type: 'ONLINE' });
    this.sync();
  }

  /**
   * Handle browser going offline
   * Notifies listeners of offline status
   */
  handleOffline() {
    console.log('[Sync] Browser is offline');
    this.notify({ type: 'OFFLINE' });
  }

  /**
   * Handle messages from service worker
   * @param {Object} data - Message data from service worker
   */
  handleServiceWorkerMessage(data) {
    switch (data.type) {
      case 'SYNC_QUEUED':
        this.notify({ type: 'SYNC_QUEUED', item: data.item });
        break;
      case 'SYNC_COMPLETE':
        this.notify({ type: 'SYNC_COMPLETE', data: data });
        break;
      case 'SYNC_FAILED':
        this.notify({ type: 'SYNC_FAILED', item: data.item, error: data.error });
        break;
    }
  }

  /**
   * Check if currently syncing
   * @returns {boolean} True if sync is in progress
   */
  isCurrentlySyncing() {
    return this.isSyncing;
  }

  /**
   * Main sync function - uploads pending changes and downloads updates
   * @returns {Promise<void>}
   */
  async sync() {
    if (this.isSyncing || !navigator.onLine) {
      return;
    }

    this.isSyncing = true;
    this.notify({ type: 'SYNC_START' });

    try {
      // 1. Upload pending changes
      await this.uploadPendingChanges();

      // 2. Download updates from server
      await this.downloadUpdates();

      this.notify({ type: 'SYNC_SUCCESS' });
    } catch (error) {
      console.error('[Sync] Sync failed:', error);
      this.notify({ type: 'SYNC_ERROR', error: error.message });
      
      // Retry after delay
      setTimeout(() => this.sync(), this.retryDelay);
    } finally {
      this.isSyncing = false;
    }
  }

  /**
   * Upload pending changes to server
   * @returns {Promise<void>}
   * @throws {Error} If some changes fail to sync after retries
   */
  async uploadPendingChanges() {
    const pendingChanges = await db.getPendingChanges();
    
    if (pendingChanges.length === 0) {
      return;
    }

    console.log(`[Sync] Uploading ${pendingChanges.length} pending changes`);

    const failed = [];

    for (const change of pendingChanges) {
      try {
        await this.uploadChange(change);
        await db.removePendingChange(change.id);
      } catch (error) {
        console.error(`[Sync] Failed to upload change ${change.id}:`, error);
        
        // Increment retry count
        change.retries++;
        
        if (change.retries >= 3) {
          // Max retries reached, remove from queue and notify
          await db.removePendingChange(change.id);
          this.notify({
            type: 'SYNC_ITEM_FAILED',
            change: change,
            error: error.message
          });
        } else {
          failed.push(change);
        }
      }
    }

    if (failed.length > 0) {
      throw new Error(`${failed.length} changes failed to sync`);
    }
  }

  /**
   * Upload a single change to server
   * @param {Object} change - Change object with operation, entity, and data
   * @returns {Promise<Object|string>} Server response
   * @throws {Error} If upload fails
   */
  async uploadChange(change) {
    const { operation, entity, data } = change;

    let url;
    let method;
    let body = null;

    switch (entity) {
      case 'entry':
        url = operation === 'CREATE' ? '/entry' : `/entry/${data.id}`;
        method = operation === 'CREATE' ? 'POST' : (operation === 'UPDATE' ? 'PUT' : 'DELETE');
        if (operation !== 'DELETE') body = data;
        break;
      
      case 'entryDraft':
        url = operation === 'CREATE' ? '/entry/draft' : `/entry/draft/${data.id}`;
        method = operation === 'CREATE' ? 'POST' : (operation === 'UPDATE' ? 'PUT' : 'DELETE');
        if (operation !== 'DELETE') body = data;
        break;
      
      case 'departure':
        url = operation === 'CREATE' ? '/departure' : `/departure/${data.id}`;
        method = operation === 'CREATE' ? 'POST' : (operation === 'UPDATE' ? 'PUT' : 'DELETE');
        if (operation !== 'DELETE') body = data;
        break;
      
      case 'departureDraft':
        url = operation === 'CREATE' ? '/departure/draft' : `/departure/draft/${data.id}`;
        method = operation === 'CREATE' ? 'POST' : (operation === 'UPDATE' ? 'PUT' : 'DELETE');
        if (operation !== 'DELETE') body = data;
        break;
      
      case 'person':
        url = operation === 'CREATE' ? '/person' : `/person/${data.id}`;
        method = operation === 'CREATE' ? 'POST' : (operation === 'UPDATE' ? 'PUT' : 'DELETE');
        if (operation !== 'DELETE') body = data;
        break;
      
      default:
        throw new Error(`Unknown entity: ${entity}`);
    }

    const options = {
      method,
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'text/html, application/json'
      },
      credentials: 'same-origin'
    };

    if (body) {
      options.body = JSON.stringify(body);
    }

    const response = await fetch(url, options);

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`Server returned ${response.status}: ${errorText}`);
    }

    // Parse response
    const contentType = response.headers.get('content-type');
    if (contentType && contentType.includes('application/json')) {
      return response.json();
    }
    
    return response.text();
  }

  /**
   * Download updates from server
   * @returns {Promise<void>}
   */
  async downloadUpdates() {
    const lastSync = await db.getSyncMetadata('lastSync') || '1970-01-01T00:00:00Z';
    const farmId = await db.getSyncMetadata('farmId');

    if (!farmId) {
      console.warn('[Sync] No farmId found, skipping download');
      return;
    }

    try {
      // Download entries
      await this.downloadEntries(lastSync, farmId);
      
      // Download departures
      await this.downloadDepartures(lastSync, farmId);
      
      // Download people
      await this.downloadPeople(lastSync, farmId);

      // Update last sync timestamp
      await db.setSyncMetadata('lastSync', new Date().toISOString());
      
    } catch (error) {
      console.error('[Sync] Download failed:', error);
      throw error;
    }
  }

  /**
   * Download entries from server
   * @param {string} lastSync - ISO timestamp of last sync
   * @param {number} farmId - Farm ID
   * @returns {Promise<void>}
   */
  async downloadEntries(lastSync, farmId) {
    try {
      const response = await fetch(`/api/entries?since=${encodeURIComponent(lastSync)}&farm=${farmId}`, {
        credentials: 'same-origin'
      });

      if (!response.ok) {
        throw new Error(`Failed to download entries: ${response.status}`);
      }

      const entries = await response.json();
      
      if (entries.length > 0) {
        await db.bulkSaveEntries(entries);
        console.log(`[Sync] Downloaded ${entries.length} entries`);
      }
    } catch (error) {
      console.error('[Sync] Failed to download entries:', error);
      // Don't throw - continue with other downloads
    }
  }

  /**
   * Download departures from server
   * @param {string} lastSync - ISO timestamp of last sync
   * @param {number} farmId - Farm ID
   * @returns {Promise<void>}
   */
  async downloadDepartures(lastSync, farmId) {
    try {
      const response = await fetch(`/api/departures?since=${encodeURIComponent(lastSync)}&farm=${farmId}`, {
        credentials: 'same-origin'
      });

      if (!response.ok) {
        throw new Error(`Failed to download departures: ${response.status}`);
      }

      const departures = await response.json();
      
      if (departures.length > 0) {
        await db.bulkSaveDepartures(departures);
        console.log(`[Sync] Downloaded ${departures.length} departures`);
      }
    } catch (error) {
      console.error('[Sync] Failed to download departures:', error);
    }
  }

  /**
   * Download people from server
   * @param {string} lastSync - ISO timestamp of last sync
   * @param {number} farmId - Farm ID
   * @returns {Promise<void>}
   */
  async downloadPeople(lastSync, farmId) {
    try {
      const response = await fetch(`/api/people?since=${encodeURIComponent(lastSync)}&farm=${farmId}`, {
        credentials: 'same-origin'
      });

      if (!response.ok) {
        throw new Error(`Failed to download people: ${response.status}`);
      }

      const people = await response.json();
      
      if (people.length > 0) {
        await db.bulkSavePeople(people);
        console.log(`[Sync] Downloaded ${people.length} people`);
      }
    } catch (error) {
      console.error('[Sync] Failed to download people:', error);
    }
  }

  /**
   * Queue a create operation
   * @param {string} entity - Entity type ('entry', 'departure', 'person', etc.)
   * @param {Object} data - Entity data
   * @returns {Promise<number|string>} The queued change ID
   */
  async queueCreate(entity, data) {
    return db.queueChange('CREATE', entity, data);
  }

  /**
   * Queue an update operation
   * @param {string} entity - Entity type
   * @param {Object} data - Entity data with ID
   * @returns {Promise<number|string>} The queued change ID
   */
  async queueUpdate(entity, data) {
    return db.queueChange('UPDATE', entity, data);
  }

  /**
   * Queue a delete operation
   * @param {string} entity - Entity type
   * @param {number} id - Entity ID to delete
   * @returns {Promise<number|string>} The queued change ID
   */
  async queueDelete(entity, id) {
    return db.queueChange('DELETE', entity, { id });
  }

  /**
   * Initial data load - call after login
   * Clears existing data and downloads all data from server
   * @param {number} farmId - Farm ID to load data for
   * @returns {Promise<void>}
   */
  async initialLoad(farmId) {
    await db.setSyncMetadata('farmId', farmId);
    
    // Clear existing data
    await db.clear(STORES.ENTRIES);
    await db.clear(STORES.DEPARTURES);
    await db.clear(STORES.PEOPLE);
    
    // Download all data
    await this.downloadUpdates();
    
    console.log('[Sync] Initial load complete');
  }

  /**
   * Get current sync status
   * @returns {Promise<Object>} Sync status object with lastSync, pendingCount, isOnline, isSyncing
   */
  async getSyncStatus() {
    const lastSync = await db.getSyncMetadata('lastSync');
    const pendingChanges = await db.getPendingChanges();
    
    return {
      lastSync,
      pendingCount: pendingChanges.length,
      isOnline: navigator.onLine,
      isSyncing: this.isSyncing
    };
  }
}

/** @type {SyncEngine} Singleton sync engine instance */
export const syncEngine = new SyncEngine();
