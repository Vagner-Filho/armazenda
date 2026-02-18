/**
 * WASM Calculator Loader
 * Loads and initializes the Go WASM calculator module
 * @module wasmCalculator
 */

/**
 * Manages loading and interaction with the Go WASM calculator
 * @class
 */
class WasmCalculator {
  constructor() {
    /** @type {boolean} Whether WASM is ready for use */
    this.ready = false;
    /** @type {Promise|null} Loading promise to prevent duplicate loads */
    this.loadingPromise = null;
  }

  /**
   * Load the WASM module
   * @returns {Promise<void>}
   */
  async load() {
    if (this.ready) {
      return;
    }

    if (this.loadingPromise) {
      return this.loadingPromise;
    }

    this.loadingPromise = this.doLoad();
    return this.loadingPromise;
  }

  /**
   * Internal method to perform WASM loading
   * @returns {Promise<void>}
   * @private
   */
  async doLoad() {
    try {
      // Load wasm_exec.js if not already loaded
      if (typeof Go === 'undefined') {
        await this.loadScript('/public/assets/wasm/wasm_exec.js');
      }

      const go = new Go();
      
      // Fetch and instantiate WASM
      const response = await fetch('/public/assets/wasm/calculator.wasm');
      if (!response.ok) {
        throw new Error(`Failed to load WASM: ${response.status}`);
      }

      const wasmBuffer = await response.arrayBuffer();
      const wasmModule = await WebAssembly.instantiate(wasmBuffer, go.importObject);
      
      // Start the Go runtime
      go.run(wasmModule.instance);
      
      this.ready = true;
      console.log('[WASM] Calculator loaded successfully');
    } catch (error) {
      console.error('[WASM] Failed to load calculator:', error);
      this.loadingPromise = null;
      throw error;
    }
  }

  /**
   * Load a script dynamically
   * @param {string} src - Script source URL
   * @returns {Promise<void>}
   * @private
   */
  loadScript(src) {
    return new Promise((resolve, reject) => {
      const script = document.createElement('script');
      script.src = src;
      script.onload = resolve;
      script.onerror = () => reject(new Error(`Failed to load ${src}`));
      document.head.appendChild(script);
    });
  }

  /**
   * Calculate entry weight and discounts
   * @param {Object} entry - Entry data with cargoWeight, analysis, etc.
   * @param {Object} [personConfig={}] - Person configuration with discounts
   * @returns {Object} Calculation result with netWeight, discounts, etc.
   * @throws {Error} If WASM is not loaded
   */
  calculateEntry(entry, personConfig) {
    if (!this.ready) {
      throw new Error('WASM calculator not loaded');
    }

    const input = JSON.stringify({
      entry: this.serializeEntry(entry),
      personConfig: personConfig || {}
    });

    const result = window.calculateEntry(input);
    return JSON.parse(result);
  }

  /**
   * Calculate departure weight
   * @param {Object} departure - Departure data
   * @returns {Object} Calculation result with netWeight
   * @throws {Error} If WASM is not loaded
   */
  calculateDeparture(departure) {
    if (!this.ready) {
      throw new Error('WASM calculator not loaded');
    }

    const input = JSON.stringify({
      departure: this.serializeDeparture(departure)
    });

    const result = window.calculateDeparture(input);
    return JSON.parse(result);
  }

  /**
   * Calculate individual discounts
   * @param {number|null} humidity - Humidity percentage
   * @param {number|null} damage - Damage percentage
   * @param {number|null} impurity - Impurity percentage
   * @param {number} grossWeight - Gross weight
   * @param {number} tare - Tare weight
   * @param {number|null} humidityModifier - Humidity discount modifier
   * @returns {Object} Discount breakdown
   * @throws {Error} If WASM is not loaded
   */
  calculateDiscounts(humidity, damage, impurity, grossWeight, tare, humidityModifier) {
    if (!this.ready) {
      throw new Error('WASM calculator not loaded');
    }

    const input = JSON.stringify({
      humidity: humidity || null,
      damage: damage || null,
      impurity: impurity || null,
      grossWeight: grossWeight.toString(),
      tare: tare.toString(),
      humidityModifier: humidityModifier || null
    });

    const result = window.calculateDiscounts(input);
    return JSON.parse(result);
  }

  /**
   * Validate entry data
   * @param {Object} entry - Entry data to validate
   * @returns {Object} Validation result with isValid and errors
   * @throws {Error} If WASM is not loaded
   */
  validateEntry(entry) {
    if (!this.ready) {
      throw new Error('WASM calculator not loaded');
    }

    const input = JSON.stringify(this.serializeEntry(entry));
    const result = window.validateEntry(input);
    return JSON.parse(result);
  }

  /**
   * Serialize entry for WASM
   * @param {Object} entry - Entry data
   * @returns {Object} Serialized entry for WASM consumption
   * @private
   */
  serializeEntry(entry) {
    return {
      id: entry.id || 0,
      field: entry.field || 0,
      crop: entry.crop || 0,
      vehicle: entry.vehicle || 0,
      cargoWeight: {
        grossWeight: (entry.grossWeight || entry.CargoWeight?.GrossWeight || 0).toString(),
        tare: (entry.tare || entry.CargoWeight?.Tare || 0).toString(),
        netWeight: (entry.netWeight || entry.CargoWeight?.NetWeight || 0).toString()
      },
      analysis: {
        humidity: entry.humidity || entry.Analysis?.Humidity || null,
        damage: entry.damage || entry.Analysis?.Damage || null,
        impurity: entry.impurity || entry.Analysis?.Impurity || null
      },
      arrivalDate: entry.arrivalDate || new Date().toISOString(),
      farm: entry.farm || 0,
      origin: entry.origin || null
    };
  }

  /**
   * Serialize departure for WASM
   * @param {Object} departure - Departure data
   * @returns {Object} Serialized departure for WASM consumption
   * @private
   */
  serializeDeparture(departure) {
    return {
      id: departure.id || 0,
      departureDate: departure.departureDate || new Date().toISOString(),
      vehicle: departure.vehicle || 0,
      crop: departure.crop || 0,
      cargoWeight: {
        grossWeight: (departure.grossWeight || departure.CargoWeight?.GrossWeight || 0).toString(),
        tare: (departure.tare || departure.CargoWeight?.Tare || 0).toString(),
        netWeight: (departure.netWeight || departure.CargoWeight?.NetWeight || 0).toString()
      },
      analysis: {
        humidity: departure.humidity || departure.Analysis?.Humidity || null,
        damage: departure.damage || departure.Analysis?.Damage || null,
        impurity: departure.impurity || departure.Analysis?.Impurity || null
      },
      farm: departure.farm || 0,
      recipient: departure.recipient || null,
      origin: departure.origin || null
    };
  }
}

/** @type {WasmCalculator} Singleton WASM calculator instance */
export const wasmCalculator = new WasmCalculator();
