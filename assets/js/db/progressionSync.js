/**
 * Humidity Progression Sync Module
 * Manages syncing of humidity progressions between IndexedDB and server
 * @module progressionSync
 */

import { db, STORES } from './database.js';

/**
 * Manages humidity progression synchronization
 * @class
 */
class ProgressionSync {
  constructor() {
    /** @type {boolean} Whether progressions are synced */
    this.isSynced = false;
  }

  /**
   * Initialize the progression sync
   * @returns {Promise<void>}
   */
  async init() {
    await db.init();
    // Check if we have any progressions
    const progressions = await db.getAllProgressions();
    this.isSynced = progressions.length > 0;
  }

  /**
   * Check if progressions are available (synced)
   * @returns {Promise<boolean>} True if progressions are available
   */
  async hasProgressions() {
    const progressions = await db.getAllProgressions();
    return progressions.length > 0;
  }

  /**
   * Sync progressions from server
   * @param {number} farmId - Farm ID
   * @param {string} lastSync - ISO timestamp of last sync
   * @returns {Promise<number>} Number of progressions synced
   */
  async syncProgressions(farmId, lastSync) {
    try {
      const response = await fetch(
        `/api/humidity-progressions/sync?since=${encodeURIComponent(lastSync)}`,
        { credentials: 'same-origin' }
      );

      if (!response.ok) {
        throw new Error(`Failed to sync progressions: ${response.status}`);
      }

      const progressions = await response.json();

      if (progressions.length > 0) {
        // Store progressions with farm reference
        // Handle inactive progressions (soft deleted)
        const progressionsWithFarm = progressions.map(p => ({
          ...p,
          farm: farmId
        }));

        // Remove inactive progressions from IndexedDB (they were soft deleted)
        const inactiveProgressions = progressions.filter(p => !p.isActive);
        for (const inactive of inactiveProgressions) {
          await db.deleteProgression(inactive.id);
          console.log(`[ProgressionSync] Removed inactive progression ${inactive.id}`);
        }

        // Save active progressions
        const activeProgressions = progressions.filter(p => p.isActive);
        await db.saveProgressions(activeProgressions);
        console.log(`[ProgressionSync] Downloaded ${activeProgressions.length} active progressions`);
        this.isSynced = true;
      }

      return progressions.length;
    } catch (error) {
      console.error('[ProgressionSync] Sync failed:', error);
      // Don't throw - progressions are reference data, app can work with defaults
      return 0;
    }
  }

  /**
   * Get a progression by ID
   * @param {number} id - Progression ID
   * @returns {Promise<Object|null>} The progression or null
   */
  async getProgression(id) {
    const progression = await db.getProgression(Number(id));
    // Only return active progressions
    if (progression && progression.isActive !== false) {
      return progression;
    }
    return null;
  }

  /**
   * Get the system default progression
   * @returns {Promise<Object|null>} System default progression
   */
  async getSystemDefault() {
    const all = await db.getAllProgressions();
    return all.find(p => p.isSystemDefault && p.isActive !== false) || null;
  }

  /**
   * Get all active progressions for a farm (including system default)
   * @param {number} [farmId=null] - Optional farm ID filter
   * @returns {Promise<Array>} Array of active progressions
   */
  async getAllProgressions(farmId = null) {
    const all = await db.getAllProgressions();

    // Filter active progressions only
    let activeProgressions = all.filter(p => p.isActive !== false);

    if (farmId) {
      // Return farm-specific + system default
      return activeProgressions.filter(p => p.farm === farmId || p.isSystemDefault);
    }

    return activeProgressions;
  }

  /**
   * Find the appropriate tier for a given humidity value
   * @param {Array} tiers - Array of tier objects
   * @param {number} humidity - Humidity value
   * @returns {Object|null} The matching tier or null
   */
  findTierForHumidity(tiers, humidity) {
    if (!tiers || tiers.length === 0) {
      return null;
    }

    // Sort by threshold descending
    const sorted = [...tiers].sort((a, b) => b.thresholdHumidity - a.thresholdHumidity);

    // Find first tier where threshold <= humidity
    return sorted.find(t => t.thresholdHumidity <= humidity) || null;
  }

  /**
   * Get the discount value for a humidity level from a progression
   * @param {Object} progression - Progression object with tiers
   * @param {number} humidity - Humidity value
   * @returns {number} Discount value (0 if no tier applies)
   */
  getDiscountForHumidity(progression, humidity) {
    if (!progression || !progression.tiers || progression.isActive === false) {
      return 0;
    }

    const tier = this.findTierForHumidity(progression.tiers, humidity);
    return tier ? tier.discountValue : 0;
  }

  /**
   * Get the current progression for a person/farm
   * Follows the fallback chain: Person -> Farm -> System Default
   * Falls back if progression is inactive
   * @param {Object} personConfig - Person configuration object
   * @param {Object} farmConfig - Farm configuration object
   * @returns {Promise<Object>} The applicable progression
   */
  async getCurrentProgression(personConfig, farmConfig) {
    // 1. Check person's progression
    if (personConfig?.humidityProgressionId) {
      const personProgression = await this.getProgression(personConfig.humidityProgressionId);
      if (personProgression && personProgression.isActive !== false) {
        return { ...personProgression, source: 'person' };
      }
    }

    // 2. Check farm's progression
    if (farmConfig?.humidityProgressionId) {
      const farmProgression = await this.getProgression(farmConfig.humidityProgressionId);
      if (farmProgression && farmProgression.isActive !== false) {
        return { ...farmProgression, source: 'farm' };
      }
    }

    // 3. Fall back to system default
    const systemDefault = await this.getSystemDefault();
    if (systemDefault) {
      return { ...systemDefault, source: 'system' };
    }

    // 4. Last resort - return null (calculations will use 0 discount)
    return null;
  }

  /**
   * Format a tier for display (e.g., "16% → 1.8")
   * @param {Object} tier - Tier object
   * @returns {string} Formatted string
   */
  formatTier(tier) {
    if (!tier) return '';
    return `${tier.thresholdHumidity}% → ${tier.discountValue}`;
  }

  /**
   * Get display info for the current humidity tier
   * @param {Object} progression - Progression object
   * @param {number} humidity - Current humidity value
   * @returns {Object} Display info object
   */
  getTierDisplayInfo(progression, humidity) {
    if (!progression || !progression.tiers || progression.isActive === false) {
      return {
        hasTier: false,
        tier: null,
        displayText: '',
        progressionName: ''
      };
    }

    const tier = this.findTierForHumidity(progression.tiers, humidity);

    return {
      hasTier: !!tier,
      tier: tier,
      displayText: tier ? this.formatTier(tier) : '',
      progressionName: progression.name,
      isDefault: progression.isSystemDefault,
      source: progression.source || 'unknown'
    };
  }
}

/** @type {ProgressionSync} Singleton instance */
export const progressionSync = new ProgressionSync();
