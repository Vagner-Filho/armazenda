/**
 * Armazenda Sync Engine
 * Handles synchronization between IndexedDB and server
 * @module syncEngine
 */

import { db, STORES } from './database.js';
import { progressionSync } from './progressionSync.js';

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

    // PASS 1: Upload reference entities (fields, crops, vehicles) and build ID mapping
    const referenceEntities = ['field', 'crop', 'vehicle'];
    const referenceChanges = pendingChanges.filter(c => referenceEntities.includes(c.entity));
    const idMapping = new Map(); // tempId -> serverId
    const failedReferences = [];

    for (const change of referenceChanges) {
      try {
        const serverResponse = await this.uploadChange(change);
        await db.removePendingChange(change.id);

        // Store the ID mapping when server returns new ID for offline-created entities
        if (change.data.id?.toString().startsWith('offline_')) {
          let serverId = null;

          // Handle different response formats
          if (serverResponse && serverResponse.id) {
            serverId = serverResponse.id;
          } else if (typeof serverResponse === 'object' && serverResponse.serverId) {
            serverId = serverResponse.serverId;
          }

          if (serverId) {
            idMapping.set(change.data.id.toString(), serverId.toString());
            console.log(`[Sync] ID mapping: ${change.data.id} -> ${serverId} (${change.entity})`);
          }
        }
      } catch (error) {
        console.error(`[Sync] Failed to upload reference ${change.entity} ${change.data.id}:`, error);
        change.retries++;
        if (change.retries >= 3) {
          await db.removePendingChange(change.id);
          this.notify({
            type: 'SYNC_ITEM_FAILED',
            change: change,
            error: error.message
          });
        } else {
          failedReferences.push(change);
        }
      }
    }

    // PASS 2: Update entries/departures in pendingChanges with new server IDs
    const dependentEntities = ['entry', 'entryDraft', 'departure', 'departureDraft'];
    const dependentChanges = pendingChanges.filter(c => dependentEntities.includes(c.entity));

    if (idMapping.size > 0 && dependentChanges.length > 0) {
      console.log(`[Sync] Updating ${dependentChanges.length} dependent changes with ID mappings`);

      for (const change of dependentChanges) {
        debugger
        const updatedData = this.replaceTempIds(change.data, idMapping);

        // Only update if IDs were actually replaced
        if (JSON.stringify(updatedData) !== JSON.stringify(change.data)) {
          change.data = updatedData;
          await db.put('pendingChanges', change);
          console.log(`[Sync] Updated ${change.entity} ${change.data.id} with server IDs`);
        }
      }
    }

    // PASS 3: Upload dependent entities (entries and departures) with corrected IDs
    const failed = [];

    for (const change of dependentChanges) {
      try {
        await this.uploadChange(change);
        await db.removePendingChange(change.id);
      } catch (error) {
        console.error(`[Sync] Failed to upload dependent ${change.entity} ${change.data.id}:`, error);
        change.retries++;
        if (change.retries >= 3) {
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

    // Re-add failed reference changes back to the queue
    for (const change of failedReferences) {
      if (change.retries < 3) {
        await db.put('pendingChanges', change);
      }
    }

    if (failed.length > 0 || failedReferences.length > 0) {
      throw new Error(`${failed.length + failedReferences.length} changes failed to sync`);
    }
  }

  /**
   * Replace temporary IDs with server IDs in entity data
   * @param {Object} data - Entity data
   * @param {Map} idMapping - Map of tempId -> serverId
   * @returns {Object} Updated data with server IDs
   */
  replaceTempIds(data, idMapping) {
    if (!data || idMapping.size === 0) {
      return data;
    }

    const updated = { ...data };

    // Map field ID
    if (data.field && idMapping.has(data.field.toString())) {
      updated.field = Number(idMapping.get(data.field.toString()));
    }

    // Map crop ID
    if (data.crop && idMapping.has(data.crop.toString())) {
      updated.crop = Number(idMapping.get(data.crop.toString()));
    }

    // Map vehicle ID (used in some contexts)
    if (data.vehicle && idMapping.has(data.vehicle.toString())) {
      updated.vehicle = Number(idMapping.get(data.vehicle.toString()));
    }

    // Map vehiclePlate for entries (this stores vehicle ID in entry data)
    if (data.vehiclePlate && idMapping.has(data.vehiclePlate.toString())) {
      updated.vehiclePlate = Number(idMapping.get(data.vehiclePlate.toString()));
    }

    return updated;
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
        if (operation !== 'DELETE') body = structuredClone(data);
        if (operation === 'CREATE') {
          delete body.id;
        }
        break;

      case 'entryDraft':
        url = operation === 'CREATE' ? '/entry/draft' : `/entry/draft/${data.id}`;
        method = operation === 'CREATE' ? 'POST' : (operation === 'UPDATE' ? 'PUT' : 'DELETE');
        if (operation !== 'DELETE') body = structuredClone(data);
        break;

      case 'departure':
        url = operation === 'CREATE' ? '/departure' : `/departure/${data.id}`;
        method = operation === 'CREATE' ? 'POST' : (operation === 'UPDATE' ? 'PUT' : 'DELETE');
        if (operation !== 'DELETE') body = structuredClone(data);
        break;

      case 'departureDraft':
        url = operation === 'CREATE' ? '/departure/draft' : `/departure/draft/${data.id}`;
        method = operation === 'CREATE' ? 'POST' : (operation === 'UPDATE' ? 'PUT' : 'DELETE');
        if (operation !== 'DELETE') body = structuredClone(data);
        break;

      case 'person':
        url = operation === 'CREATE' ? '/person' : `/person/${data.id}`;
        method = operation === 'CREATE' ? 'POST' : (operation === 'UPDATE' ? 'PUT' : 'DELETE');
        if (operation !== 'DELETE') body = structuredClone(data);
        break;

      case 'crop':
        url = '/crop';
        method = 'POST';
        body = {
          name: data.name,
          product: data.product,
          startDate: data.startDate
        };
        break;

      case 'vehicle':
        url = '/vehicle';
        method = 'POST';
        body = structuredClone(data);
        delete body.id;
        break;

      case 'field':
        url = '/field';
        method = 'POST';
        body = structuredClone(data);
        delete body.id;
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

    // Handle Entry CREATE response - update DOM with server ID
    if (operation === 'CREATE' && entity === 'entry' && data.id.toString().startsWith('offline_')) {
      const contentType = response.headers.get('content-type');
      if (contentType && contentType.includes('application/json')) {
        const serverResponse = await response.json();
        const serverId = serverResponse.id;

        // Update DOM: replace temp ID with server ID
        const row = document.querySelector(`tr[id="entry-${data.id}"]`);
        if (row) {
          row.id = `entry-${serverId}`;
          row.removeAttribute('data-offline-pending');
          row.classList.remove('bg-yellow-500/5');

          // Update ID cell
          const idCell = row.querySelector('td:first-child');
          if (idCell) idCell.textContent = serverId;

          // Remove pending badge
          const badge = row.querySelector('.pending-badge');
          if (badge) badge.remove();

          // Update action buttons with new ID
          const buttons = row.querySelectorAll('button');
          buttons.forEach(btn => {
            const hxGet = btn.getAttribute('hx-get');
            const hxDelete = btn.getAttribute('hx-delete');
            if (hxGet) btn.setAttribute('hx-get', hxGet.replace(data.id, serverId));
            if (hxDelete) {
              btn.setAttribute('hx-delete', hxDelete.replace(data.id, serverId));
              btn.setAttribute('hx-target', `#entry-${serverId}`);
            }
          });
        }

        // Notify user
        this.notify({
          type: 'SYNC_ENTRY_COMPLETE',
          tempId: data.id,
          serverId: serverId
        });

        return serverResponse;
      }
    }

    // Handle Crop CREATE response - server returns HTML option element
    if (operation === 'CREATE' && entity === 'crop' && data.id.toString().startsWith('offline_')) {
      const htmlResponse = await response.text();

      // Parse the HTML to extract the server ID
      const tempId = data.id;
      const parser = new DOMParser();
      const doc = parser.parseFromString(htmlResponse, 'text/html');
      const newOption = doc.querySelector('option');

      if (newOption) {
        const serverId = newOption.getAttribute('value');
        const serverName = newOption.textContent.trim();
        const serverProductId = newOption.dataset.productId;

        // Update DOM: find and replace the offline option
        const cropSelector = document.getElementById('crop-selector');
        if (cropSelector) {
          const offlineOption = cropSelector.querySelector(`option[value="${tempId}"]`);
          if (offlineOption) {
            // Update the option with server data
            offlineOption.value = serverId;
            offlineOption.textContent = serverName;
            offlineOption.dataset.productId = serverProductId;
            offlineOption.removeAttribute('data-offline-pending');
            offlineOption.style.fontStyle = '';
          }
        }

        // Update IndexedDB: replace temp crop with server crop
        try {
          const tempCrop = await db.get(STORES.CROPS, tempId);
          if (tempCrop) {
            await db.delete(STORES.CROPS, tempId);
            await db.put(STORES.CROPS, {
              id: Number(serverId),
              name: serverName,
              product: Number(serverProductId),
              startDate: tempCrop.startDate,
              farm: tempCrop.farm,
              synced: true,
              modifiedAt: new Date().toISOString()
            });
          }
        } catch (dbError) {
          console.error('[Sync] Failed to update crop in IndexedDB:', dbError);
        }

        // Notify user
        this.notify({
          type: 'SYNC_CROP_COMPLETE',
          tempId: tempId,
          serverId: serverId,
          name: serverName
        });

        return { id: serverId, name: serverName, product: serverProductId };
      }
    }

    // Handle Vehicle CREATE response - server returns HTML option element
    if (operation === 'CREATE' && entity === 'vehicle' && data.id.toString().startsWith('offline_')) {
      const htmlResponse = await response.text();

      // Parse the HTML to extract the server ID
      const tempId = data.id;
      const parser = new DOMParser();
      const doc = parser.parseFromString(htmlResponse, 'text/html');
      const newOption = doc.querySelector('option');

      if (newOption) {
        const serverId = newOption.getAttribute('value');
        const displayText = newOption.textContent.trim();

        // Parse display text to extract plate and name
        let serverPlate = displayText;
        let serverName = '';
        if (displayText.includes('|')) {
          const parts = displayText.split('|').map(p => p.trim());
          serverPlate = parts[0];
          serverName = parts[1];
        }

        // Update DOM: find and replace the offline option
        const vehicleSelector = document.getElementById('vehicle-selector');
        if (vehicleSelector) {
          const offlineOption = vehicleSelector.querySelector(`option[value="${tempId}"]`);
          if (offlineOption) {
            // Update the option with server data
            offlineOption.value = serverId;
            offlineOption.textContent = displayText;
            offlineOption.removeAttribute('data-offline-pending');
            offlineOption.style.fontStyle = '';
          }
        }

        // Update IndexedDB: replace temp vehicle with server vehicle
        try {
          const tempVehicle = await db.get(STORES.VEHICLES, tempId);
          if (tempVehicle) {
            await db.delete(STORES.VEHICLES, tempId);
            await db.put(STORES.VEHICLES, {
              id: Number(serverId),
              plate: serverPlate,
              name: serverName,
              farm: tempVehicle.farm,
              synced: true,
              modifiedAt: new Date().toISOString()
            });
          }
        } catch (dbError) {
          console.error('[Sync] Failed to update vehicle in IndexedDB:', dbError);
        }

        // Notify user
        this.notify({
          type: 'SYNC_VEHICLE_COMPLETE',
          tempId: tempId,
          serverId: serverId,
          plate: serverPlate
        });

        return { id: serverId, plate: serverPlate, name: serverName };
      }
    }

    // Handle Field CREATE response - server returns HTML option element
    if (operation === 'CREATE' && entity === 'field' && data.id.toString().startsWith('offline_')) {
      const htmlResponse = await response.text();

      // Parse the HTML to extract the server ID
      const tempId = data.id;
      const parser = new DOMParser();
      const doc = parser.parseFromString(htmlResponse, 'text/html');
      const newOption = doc.querySelector('option');

      if (newOption) {
        const serverId = newOption.getAttribute('value');
        const serverName = newOption.textContent.trim();

        // Update DOM: find and replace the offline option
        const fieldSelector = document.getElementById('field-selector');
        if (fieldSelector) {
          const offlineOption = fieldSelector.querySelector(`option[value="${tempId}"]`);
          if (offlineOption) {
            // Update the option with server data
            offlineOption.value = serverId;
            offlineOption.textContent = serverName;
            offlineOption.removeAttribute('data-offline-pending');
            offlineOption.style.fontStyle = '';
          }
        }

        // Update IndexedDB: replace temp field with server field
        try {
          const tempField = await db.get(STORES.FIELDS, tempId);
          if (tempField) {
            await db.delete(STORES.FIELDS, tempId);
            await db.put(STORES.FIELDS, {
              id: Number(serverId),
              name: serverName,
              hectares: tempField.hectares,
              farm: tempField.farm,
              synced: true,
              modifiedAt: new Date().toISOString()
            });
          }
        } catch (dbError) {
          console.error('[Sync] Failed to update field in IndexedDB:', dbError);
        }

        // Notify user
        this.notify({
          type: 'SYNC_FIELD_COMPLETE',
          tempId: tempId,
          serverId: serverId,
          name: serverName
        });

        return { id: serverId, name: serverName };
      }
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
      // Sync progressions first (reference data)
      await this.syncProgressions(lastSync, farmId);

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
   * Sync humidity progressions from server
   * @param {string} lastSync - ISO timestamp of last sync
   * @param {number} farmId - Farm ID
   * @returns {Promise<void>}
   */
  async syncProgressions(lastSync, farmId) {
    try {
      await progressionSync.init();
      await progressionSync.syncProgressions(farmId, lastSync);
    } catch (error) {
      console.error('[Sync] Failed to sync progressions:', error);
      // Don't throw - progressions are reference data
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
