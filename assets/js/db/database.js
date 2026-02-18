/**
 * Armazenda IndexedDB Database
 * Handles all offline data storage and retrieval
 * @module database
 */

/** @constant {string} */
const DB_NAME = 'ArmazendaDB';

/** @constant {number} */
const DB_VERSION = 1;

/**
 * Database store names
 * @constant {Object}
 * @property {string} ENTRIES - Entry data store
 * @property {string} DEPARTURES - Departure data store
 * @property {string} PEOPLE - Person data store
 * @property {string} CROPS - Crop reference data store
 * @property {string} FIELDS - Field reference data store
 * @property {string} VEHICLES - Vehicle reference data store
 * @property {string} PENDING_CHANGES - Pending sync operations store
 * @property {string} SYNC_METADATA - Sync metadata store
 * @property {string} TEMPLATES - Cached HTML templates store
 */
const STORES = {
  ENTRIES: 'entries',
  DEPARTURES: 'departures',
  PEOPLE: 'people',
  CROPS: 'crops',
  FIELDS: 'fields',
  VEHICLES: 'vehicles',
  PENDING_CHANGES: 'pendingChanges',
  SYNC_METADATA: 'syncMetadata',
  TEMPLATES: 'templates'
};

/**
 * Manages IndexedDB database operations for offline storage
 * @class
 */
class ArmazendaDB {
  constructor() {
    /** @type {IDBDatabase|null} */
    this.db = null;
    /** @type {Promise|null} */
    this.initPromise = null;
  }

  /**
   * Initialize the database connection
   * Creates object stores if they don't exist
   * @returns {Promise<IDBDatabase>} The database instance
   */
  async init() {
    if (this.initPromise) {
      return this.initPromise;
    }

    this.initPromise = new Promise((resolve, reject) => {
      const request = indexedDB.open(DB_NAME, DB_VERSION);

      request.onerror = () => reject(request.error);
      request.onsuccess = () => {
        this.db = request.result;
        resolve(this.db);
      };

      request.onupgradeneeded = (event) => {
        const db = event.target.result;

        // Entries store
        if (!db.objectStoreNames.contains(STORES.ENTRIES)) {
          const entryStore = db.createObjectStore(STORES.ENTRIES, { keyPath: 'id' });
          entryStore.createIndex('farm', 'farm', { unique: false });
          entryStore.createIndex('synced', 'synced', { unique: false });
          entryStore.createIndex('modifiedAt', 'modifiedAt', { unique: false });
        }

        // Departures store
        if (!db.objectStoreNames.contains(STORES.DEPARTURES)) {
          const departureStore = db.createObjectStore(STORES.DEPARTURES, { keyPath: 'id' });
          departureStore.createIndex('farm', 'farm', { unique: false });
          departureStore.createIndex('synced', 'synced', { unique: false });
          departureStore.createIndex('modifiedAt', 'modifiedAt', { unique: false });
        }

        // People store
        if (!db.objectStoreNames.contains(STORES.PEOPLE)) {
          const peopleStore = db.createObjectStore(STORES.PEOPLE, { keyPath: 'id' });
          peopleStore.createIndex('farm', 'farm', { unique: false });
          peopleStore.createIndex('synced', 'synced', { unique: false });
        }

        // Crops store
        if (!db.objectStoreNames.contains(STORES.CROPS)) {
          const cropStore = db.createObjectStore(STORES.CROPS, { keyPath: 'id' });
          cropStore.createIndex('farm', 'farm', { unique: false });
        }

        // Fields store
        if (!db.objectStoreNames.contains(STORES.FIELDS)) {
          const fieldStore = db.createObjectStore(STORES.FIELDS, { keyPath: 'id' });
          fieldStore.createIndex('farm', 'farm', { unique: false });
        }

        // Vehicles store
        if (!db.objectStoreNames.contains(STORES.VEHICLES)) {
          const vehicleStore = db.createObjectStore(STORES.VEHICLES, { keyPath: 'id' });
          vehicleStore.createIndex('farm', 'farm', { unique: false });
        }

        // Pending changes queue
        if (!db.objectStoreNames.contains(STORES.PENDING_CHANGES)) {
          const pendingStore = db.createObjectStore(STORES.PENDING_CHANGES, { 
            keyPath: 'id',
            autoIncrement: true 
          });
          pendingStore.createIndex('entity', 'entity', { unique: false });
          pendingStore.createIndex('operation', 'operation', { unique: false });
          pendingStore.createIndex('timestamp', 'timestamp', { unique: false });
        }

        // Sync metadata
        if (!db.objectStoreNames.contains(STORES.SYNC_METADATA)) {
          db.createObjectStore(STORES.SYNC_METADATA, { keyPath: 'key' });
        }

        // Cached templates
        if (!db.objectStoreNames.contains(STORES.TEMPLATES)) {
          db.createObjectStore(STORES.TEMPLATES, { keyPath: 'name' });
        }
      };
    });

    return this.initPromise;
  }

  /**
   * Get a transaction for a specific store
   * @param {string} storeName - Name of the object store
   * @param {string} [mode='readonly'] - Transaction mode ('readonly' or 'readwrite')
   * @returns {Promise<IDBTransaction>} The transaction object
   */
  async getTransaction(storeName, mode = 'readonly') {
    await this.init();
    return this.db.transaction(storeName, mode);
  }

  /**
   * Generic get method to retrieve a single record
   * @param {string} storeName - Name of the object store
   * @param {number|string} id - Primary key of the record
   * @returns {Promise<Object|null>} The retrieved record or null
   */
  async get(storeName, id) {
    const transaction = await this.getTransaction(storeName);
    const store = transaction.objectStore(storeName);
    
    return new Promise((resolve, reject) => {
      const request = store.get(id);
      request.onsuccess = () => resolve(request.result);
      request.onerror = () => reject(request.error);
    });
  }

  /**
   * Generic get all method to retrieve multiple records
   * @param {string} storeName - Name of the object store
   * @param {string} [indexName=null] - Optional index name to query
   * @param {IDBKeyRange} [query=null] - Optional key range query
   * @returns {Promise<Array>} Array of records
   */
  async getAll(storeName, indexName = null, query = null) {
    const transaction = await this.getTransaction(storeName);
    const store = transaction.objectStore(storeName);
    
    let target = store;
    if (indexName) {
      target = store.index(indexName);
    }
    
    return new Promise((resolve, reject) => {
      const request = target.getAll(query);
      request.onsuccess = () => resolve(request.result);
      request.onerror = () => reject(request.error);
    });
  }

  /**
   * Generic put method to save or update a record
   * @param {string} storeName - Name of the object store
   * @param {Object} data - Data to store
   * @returns {Promise<number|string>} The key of the stored record
   */
  async put(storeName, data) {
    const transaction = await this.getTransaction(storeName, 'readwrite');
    const store = transaction.objectStore(storeName);
    
    return new Promise((resolve, reject) => {
      const request = store.put(data);
      request.onsuccess = () => resolve(request.result);
      request.onerror = () => reject(request.error);
    });
  }

  /**
   * Generic delete method to remove a record
   * @param {string} storeName - Name of the object store
   * @param {number|string} id - Primary key of the record to delete
   * @returns {Promise<void>}
   */
  async delete(storeName, id) {
    const transaction = await this.getTransaction(storeName, 'readwrite');
    const store = transaction.objectStore(storeName);
    
    return new Promise((resolve, reject) => {
      const request = store.delete(id);
      request.onsuccess = () => resolve();
      request.onerror = () => reject(request.error);
    });
  }

  /**
   * Clear all records from a store
   * @param {string} storeName - Name of the object store to clear
   * @returns {Promise<void>}
   */
  async clear(storeName) {
    const transaction = await this.getTransaction(storeName, 'readwrite');
    const store = transaction.objectStore(storeName);
    
    return new Promise((resolve, reject) => {
      const request = store.clear();
      request.onsuccess = () => resolve();
      request.onerror = () => reject(request.error);
    });
  }

  // ==================== ENTRY OPERATIONS ====================

  /**
   * Get a single entry by ID
   * @param {number} id - Entry ID
   * @returns {Promise<Object|null>} The entry record
   */
  async getEntry(id) {
    return this.get(STORES.ENTRIES, id);
  }

  /**
   * Get all entries, optionally filtered by farm
   * @param {number} [farmId=null] - Optional farm ID filter
   * @returns {Promise<Array>} Array of entry records
   */
  async getAllEntries(farmId = null) {
    if (farmId) {
      return this.getAll(STORES.ENTRIES, 'farm', IDBKeyRange.only(farmId));
    }
    return this.getAll(STORES.ENTRIES);
  }

  /**
   * Save or update an entry
   * @param {Object} entry - Entry data to save
   * @param {boolean} [synced=false] - Whether the entry is synced with server
   * @returns {Promise<number|string>} The entry ID
   */
  async saveEntry(entry, synced = false) {
    const data = {
      ...entry,
      synced,
      modifiedAt: new Date().toISOString()
    };
    return this.put(STORES.ENTRIES, data);
  }

  /**
   * Delete an entry
   * @param {number} id - Entry ID to delete
   * @returns {Promise<void>}
   */
  async deleteEntry(id) {
    return this.delete(STORES.ENTRIES, id);
  }

  /**
   * Get all unsynced entries
   * @returns {Promise<Array>} Array of unsynced entries
   */
  async getUnsyncedEntries() {
    return this.getAll(STORES.ENTRIES, 'synced', IDBKeyRange.only(false));
  }

  // ==================== DEPARTURE OPERATIONS ====================

  /**
   * Get a single departure by ID
   * @param {number} id - Departure ID
   * @returns {Promise<Object|null>} The departure record
   */
  async getDeparture(id) {
    return this.get(STORES.DEPARTURES, id);
  }

  /**
   * Get all departures, optionally filtered by farm
   * @param {number} [farmId=null] - Optional farm ID filter
   * @returns {Promise<Array>} Array of departure records
   */
  async getAllDepartures(farmId = null) {
    if (farmId) {
      return this.getAll(STORES.DEPARTURES, 'farm', IDBKeyRange.only(farmId));
    }
    return this.getAll(STORES.DEPARTURES);
  }

  /**
   * Save or update a departure
   * @param {Object} departure - Departure data to save
   * @param {boolean} [synced=false] - Whether the departure is synced with server
   * @returns {Promise<number|string>} The departure ID
   */
  async saveDeparture(departure, synced = false) {
    const data = {
      ...departure,
      synced,
      modifiedAt: new Date().toISOString()
    };
    return this.put(STORES.DEPARTURES, data);
  }

  /**
   * Delete a departure
   * @param {number} id - Departure ID to delete
   * @returns {Promise<void>}
   */
  async deleteDeparture(id) {
    return this.delete(STORES.DEPARTURES, id);
  }

  /**
   * Get all unsynced departures
   * @returns {Promise<Array>} Array of unsynced departures
   */
  async getUnsyncedDepartures() {
    return this.getAll(STORES.DEPARTURES, 'synced', IDBKeyRange.only(false));
  }

  // ==================== PERSON OPERATIONS ====================

  /**
   * Get a single person by ID
   * @param {number} id - Person ID
   * @returns {Promise<Object|null>} The person record
   */
  async getPerson(id) {
    return this.get(STORES.PEOPLE, id);
  }

  /**
   * Get all people, optionally filtered by farm
   * @param {number} [farmId=null] - Optional farm ID filter
   * @returns {Promise<Array>} Array of person records
   */
  async getAllPeople(farmId = null) {
    if (farmId) {
      return this.getAll(STORES.PEOPLE, 'farm', IDBKeyRange.only(farmId));
    }
    return this.getAll(STORES.PEOPLE);
  }

  /**
   * Save or update a person
   * @param {Object} person - Person data to save
   * @param {boolean} [synced=false] - Whether the person is synced with server
   * @returns {Promise<number|string>} The person ID
   */
  async savePerson(person, synced = false) {
    const data = {
      ...person,
      synced,
      modifiedAt: new Date().toISOString()
    };
    return this.put(STORES.PEOPLE, data);
  }

  /**
   * Delete a person
   * @param {number} id - Person ID to delete
   * @returns {Promise<void>}
   */
  async deletePerson(id) {
    return this.delete(STORES.PEOPLE, id);
  }

  // ==================== REFERENCE DATA ====================

  /**
   * Save multiple crops
   * @param {Array<Object>} crops - Array of crop objects
   * @returns {Promise<void>}
   */
  async saveCrops(crops) {
    for (const crop of crops) {
      await this.put(STORES.CROPS, { ...crop, synced: true });
    }
  }

  /**
   * Get all crops
   * @returns {Promise<Array>} Array of crop records
   */
  async getAllCrops() {
    return this.getAll(STORES.CROPS);
  }

  /**
   * Save multiple fields
   * @param {Array<Object>} fields - Array of field objects
   * @returns {Promise<void>}
   */
  async saveFields(fields) {
    for (const field of fields) {
      await this.put(STORES.FIELDS, { ...field, synced: true });
    }
  }

  /**
   * Get all fields
   * @returns {Promise<Array>} Array of field records
   */
  async getAllFields() {
    return this.getAll(STORES.FIELDS);
  }

  /**
   * Save multiple vehicles
   * @param {Array<Object>} vehicles - Array of vehicle objects
   * @returns {Promise<void>}
   */
  async saveVehicles(vehicles) {
    for (const vehicle of vehicles) {
      await this.put(STORES.VEHICLES, { ...vehicle, synced: true });
    }
  }

  /**
   * Get all vehicles
   * @returns {Promise<Array>} Array of vehicle records
   */
  async getAllVehicles() {
    return this.getAll(STORES.VEHICLES);
  }

  // ==================== PENDING CHANGES QUEUE ====================

  /**
   * Queue a change operation
   * @param {string} operation - Operation type ('CREATE', 'UPDATE', 'DELETE')
   * @param {string} entity - Entity type ('entry', 'departure', 'person', etc.)
   * @param {Object} data - Change data
   * @returns {Promise<number|string>} The queued change ID
   */
  async queueChange(operation, entity, data) {
    const change = {
      operation,
      entity,
      data,
      timestamp: Date.now(),
      retries: 0
    };
    return this.put(STORES.PENDING_CHANGES, change);
  }

  /**
   * Get all pending changes
   * @returns {Promise<Array>} Array of pending changes
   */
  async getPendingChanges() {
    return this.getAll(STORES.PENDING_CHANGES);
  }

  /**
   * Remove a pending change
   * @param {number|string} id - Change ID to remove
   * @returns {Promise<void>}
   */
  async removePendingChange(id) {
    return this.delete(STORES.PENDING_CHANGES, id);
  }

  /**
   * Clear all pending changes
   * @returns {Promise<void>}
   */
  async clearPendingChanges() {
    return this.clear(STORES.PENDING_CHANGES);
  }

  // ==================== SYNC METADATA ====================

  /**
   * Set sync metadata value
   * @param {string} key - Metadata key
   * @param {*} value - Metadata value
   * @returns {Promise<number|string>} The metadata key
   */
  async setSyncMetadata(key, value) {
    return this.put(STORES.SYNC_METADATA, { key, value, updatedAt: new Date().toISOString() });
  }

  /**
   * Get sync metadata value
   * @param {string} key - Metadata key
   * @returns {Promise<*>} The metadata value or null
   */
  async getSyncMetadata(key) {
    const result = await this.get(STORES.SYNC_METADATA, key);
    return result ? result.value : null;
  }

  // ==================== TEMPLATE CACHE ====================

  /**
   * Save an HTML template
   * @param {string} name - Template name
   * @param {string} html - HTML content
   * @returns {Promise<string>} The template name
   */
  async saveTemplate(name, html) {
    return this.put(STORES.TEMPLATES, { name, html, cachedAt: new Date().toISOString() });
  }

  /**
   * Get an HTML template
   * @param {string} name - Template name
   * @returns {Promise<string|null>} The HTML content or null
   */
  async getTemplate(name) {
    const result = await this.get(STORES.TEMPLATES, name);
    return result ? result.html : null;
  }

  // ==================== BULK OPERATIONS ====================

  /**
   * Save multiple entries in bulk
   * @param {Array<Object>} entries - Array of entry objects
   * @returns {Promise<void>}
   */
  async bulkSaveEntries(entries) {
    const transaction = await this.getTransaction(STORES.ENTRIES, 'readwrite');
    const store = transaction.objectStore(STORES.ENTRIES);
    
    const promises = entries.map(entry => {
      return new Promise((resolve, reject) => {
        const request = store.put({
          ...entry,
          synced: true,
          modifiedAt: new Date().toISOString()
        });
        request.onsuccess = () => resolve();
        request.onerror = () => reject(request.error);
      });
    });
    
    await Promise.all(promises);
  }

  /**
   * Save multiple departures in bulk
   * @param {Array<Object>} departures - Array of departure objects
   * @returns {Promise<void>}
   */
  async bulkSaveDepartures(departures) {
    const transaction = await this.getTransaction(STORES.DEPARTURES, 'readwrite');
    const store = transaction.objectStore(STORES.DEPARTURES);
    
    const promises = departures.map(departure => {
      return new Promise((resolve, reject) => {
        const request = store.put({
          ...departure,
          synced: true,
          modifiedAt: new Date().toISOString()
        });
        request.onsuccess = () => resolve();
        request.onerror = () => reject(request.error);
      });
    });
    
    await Promise.all(promises);
  }

  /**
   * Save multiple people in bulk
   * @param {Array<Object>} people - Array of person objects
   * @returns {Promise<void>}
   */
  async bulkSavePeople(people) {
    const transaction = await this.getTransaction(STORES.PEOPLE, 'readwrite');
    const store = transaction.objectStore(STORES.PEOPLE);
    
    const promises = people.map(person => {
      return new Promise((resolve, reject) => {
        const request = store.put({
          ...person,
          synced: true,
          modifiedAt: new Date().toISOString()
        });
        request.onsuccess = () => resolve();
        request.onerror = () => reject(request.error);
      });
    });
    
    await Promise.all(promises);
  }
}

/** @type {ArmazendaDB} Singleton database instance */
export const db = new ArmazendaDB();

/** @type {Object} Store names constant */
export { STORES };
