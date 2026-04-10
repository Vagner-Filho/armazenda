/**
 * Armazenda Service Worker
 * Handles offline caching, request interception, and background sync
 * @module ServiceWorker
 */

/** @constant {string} Cache version name */
const CACHE_NAME = 'armazenda-v1';

/**
 * Static assets to cache on install
 * @constant {Array<string>}
 */
const STATIC_ASSETS = [
  '/',
  '/romaneio',
  '/pessoa',
  '/public/assets/static/css/output.css',
  '/public/assets/static/htmx.min.js.gz',
  '/public/assets/js/sweetalert2.min.js',
  '/public/assets/js/toast.js',
  '/public/assets/js/form.js',
  '/public/assets/js/selectOption.js',
  '/public/assets/js/weight.js',
  '/public/assets/js/date.js',
  '/public/assets/js/discount.js',
  '/public/assets/js/getEntryPdf.js',
  '/public/assets/js/getDeparturePdf.js',
  '/public/assets/wasm/calculator.wasm',
  '/public/assets/wasm/wasm_exec.js'
];

/**
 * HTML templates to cache for offline use
 * @constant {Array<string>}
 */
const HTML_TEMPLATES = [
  '/entry/list',
  '/entry/form',
  '/entry/filters',
  '/entry/draft/form',
  '/entry/draft/list',
  '/departure/list',
  '/departure/form',
  '/departure/filters',
  '/departure/draft/form',
  '/departure/draft/list',
  '/person/form'
];

/**
 * Service Worker Install Event
 * Caches static assets
 * @event install
 */
self.addEventListener('install', (event) => {
  console.log('[Service Worker] Installing...');
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then((cache) => {
        console.log('[Service Worker] Caching static assets');
        return cache.addAll(STATIC_ASSETS);
      })
      .catch((err) => {
        console.error('[Service Worker] Failed to cache:', err);
      })
  );
  self.skipWaiting();
});

/**
 * Service Worker Activate Event
 * Cleans up old caches
 * @event activate
 */
self.addEventListener('activate', (event) => {
  console.log('[Service Worker] Activating...');
  event.waitUntil(
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames
          .filter((name) => name !== CACHE_NAME)
          .map((name) => caches.delete(name))
      );
    })
  );
  self.clients.claim();
});

/**
 * Check if request is an API call
 * @param {URL} url - Parsed URL object
 * @returns {boolean} True if request is an API endpoint
 */
function isApiRequest(url) {
  return url.pathname.startsWith('/entry') ||
    url.pathname.startsWith('/departure') ||
    url.pathname.startsWith('/person') ||
    url.pathname.startsWith('/crop') ||
    url.pathname.startsWith('/field') ||
    url.pathname.startsWith('/vehicle');
}

/**
 * Check if request is a mutating operation (POST, PUT, DELETE)
 * @param {Request} request - Fetch request object
 * @returns {boolean} True if request modifies data
 */
function isMutatingRequest(request) {
  return request.method === 'POST' ||
    request.method === 'PUT' ||
    request.method === 'PATCH' ||
    request.method === 'DELETE';
}

/**
 * Service Worker Fetch Event
 * Intercepts all network requests
 * @event fetch
 */
self.addEventListener('fetch', (event) => {
  const { request } = event;
  const url = new URL(request.url);

  // Skip non-GET requests for API calls
  if (isApiRequest(url) && isMutatingRequest(request)) {
    event.respondWith(handleMutatingRequest(request));
    return;
  }

  // Handle GET requests
  if (request.method === 'GET') {
    event.respondWith(handleGetRequest(request, url));
    return;
  }

  // Default: pass through to network
  event.respondWith(fetch(request));
});

/**
 * Handle GET requests with cache-first strategy
 * @param {Request} request - Fetch request
 * @param {URL} url - Parsed URL
 * @returns {Promise<Response>} Response from cache or network
 */
async function handleGetRequest(request, url) {
  // Try cache first
  const cachedResponse = await caches.match(request);

  if (cachedResponse) {
    // Return cached response and update cache in background
    fetch(request)
      .then((networkResponse) => {
        if (networkResponse.ok) {
          caches.open(CACHE_NAME).then((cache) => {
            cache.put(request, networkResponse.clone());
          });
        }
      })
      .catch(() => {
        // Network failed, already serving from cache
      });

    return cachedResponse;
  }

  // Not in cache, fetch from network
  try {
    const networkResponse = await fetch(request);

    if (networkResponse.ok) {
      // Cache the response
      const cache = await caches.open(CACHE_NAME);
      cache.put(request, networkResponse.clone());
    }

    return networkResponse;
  } catch (error) {
    // Network failed, not in cache
    console.error('[Service Worker] Network request failed:', error);

    // For HTML templates, return offline page
    if (request.headers.get('accept')?.includes('text/html')) {
      return new Response(
        `<html><body style="font-family: sans-serif; padding: 20px; text-align: center;">
          <h1>Sem conexão</h1>
          <p>Você está offline. Algumas funcionalidades podem estar indisponíveis.</p>
          <button onclick="location.reload()" style="padding: 10px 20px; font-size: 16px;">Tentar novamente</button>
        </body></html>`,
        {
          status: 503,
          headers: { 'Content-Type': 'text/html' }
        }
      );
    }

    throw error;
  }
}

/**
 * Handle mutating requests (POST, PUT, DELETE)
 * Queues for sync when offline
 * @param {Request} request - Fetch request
 * @returns {Promise<Response>} Response from network or queued response
 */
async function handleMutatingRequest(request) {
  // Try network first
  try {
    const networkResponse = await fetch(request);

    if (networkResponse.ok) {
      // Success - notify clients to sync
      notifyClients({ type: 'SYNC_COMPLETE', url: request.url });
      return networkResponse;
    }

    throw new Error(`Server returned ${networkResponse.status}`);
  } catch (error) {
    // Network failed - queue for sync
    console.log('[Service Worker] Network failed, queuing for sync:', request.url);

    const queueItem = await queueForSync(request);

    // Return a synthetic success response
    // The client will handle the actual sync later
    return new Response(
      JSON.stringify({
        queued: true,
        id: queueItem.id,
        message: 'Operação salva localmente. Será sincronizada quando a conexão for restabelecida.'
      }),
      {
        status: 202,
        headers: {
          'Content-Type': 'application/json',
          'X-Offline-Queued': 'true'
        }
      }
    );
  }
}

/**
 * Queue a request for background sync
 * @param {Request} request - Request to queue
 * @returns {Promise<Object>} Queue item with generated ID
 */
async function queueForSync(request) {
  const queue = await getSyncQueue();
  const id = Date.now().toString(36) + Math.random().toString(36).substr(2);

  const queueItem = {
    id,
    url: request.url,
    method: request.method,
    headers: Array.from(request.headers.entries()),
    timestamp: Date.now(),
    retries: 0
  };

  // Clone and store the body if it's a POST/PUT
  if (request.method === 'POST' || request.method === 'PUT' || request.method === 'PATCH') {
    const body = await request.clone().text();
    queueItem.body = body;
  }

  queue.push(queueItem);
  await saveSyncQueue(queue);

  // Notify clients
  notifyClients({ type: 'SYNC_QUEUED', item: queueItem });

  return queueItem;
}

/**
 * Get sync queue from IndexedDB via BroadcastChannel
 * @returns {Promise<Array>} Array of queued changes
 */
async function getSyncQueue() {
  return new Promise((resolve) => {
    const channel = new BroadcastChannel('sw-sync');
    channel.postMessage({ type: 'GET_QUEUE' });
    channel.onmessage = (event) => {
      if (event.data.type === 'QUEUE_DATA') {
        channel.close();
        resolve(event.data.queue || []);
      }
    };

    // Timeout fallback
    setTimeout(() => {
      channel.close();
      resolve([]);
    }, 1000);
  });
}

/**
 * Save sync queue to IndexedDB via BroadcastChannel
 * @param {Array} queue - Queue array to save
 * @returns {Promise<void>}
 */
async function saveSyncQueue(queue) {
  return new Promise((resolve) => {
    const channel = new BroadcastChannel('sw-sync');
    channel.postMessage({ type: 'SAVE_QUEUE', queue });
    channel.onmessage = (event) => {
      if (event.data.type === 'QUEUE_SAVED') {
        channel.close();
        resolve();
      }
    };

    setTimeout(() => {
      channel.close();
      resolve();
    }, 1000);
  });
}

/**
 * Notify all clients of an event
 * @param {Object} message - Message to broadcast
 */
function notifyClients(message) {
  self.clients.matchAll().then((clients) => {
    clients.forEach((client) => {
      client.postMessage(message);
    });
  });
}

/**
 * Service Worker Message Event
 * Listen for messages from clients
 * @event message
 */
self.addEventListener('message', (event) => {
  if (event.data.type === 'SKIP_WAITING') {
    self.skipWaiting();
  }

  if (event.data.type === 'CACHE_TEMPLATES') {
    cacheTemplates();
  }
});

/**
 * Cache HTML templates
 * @returns {Promise<void>}
 */
async function cacheTemplates() {
  const cache = await caches.open(CACHE_NAME);

  for (const template of HTML_TEMPLATES) {
    try {
      const response = await fetch(template);
      if (response.ok) {
        cache.put(template, response.clone());
        console.log('[Service Worker] Cached template:', template);
      }
    } catch (err) {
      console.error('[Service Worker] Failed to cache template:', template, err);
    }
  }
}

/**
 * Background Sync Event
 * Syncs pending changes when connection is restored
 * @event sync
 */
self.addEventListener('sync', (event) => {
  if (event.tag === 'sync-pending-changes') {
    event.waitUntil(syncPendingChanges());
  }
});

/**
 * Sync pending changes with server
 * @returns {Promise<void>}
 */
async function syncPendingChanges() {
  const queue = await getSyncQueue();

  if (queue.length === 0) return;

  console.log('[Service Worker] Syncing', queue.length, 'pending changes');

  const failed = [];

  for (const item of queue) {
    try {
      const response = await fetch(item.url, {
        method: item.method,
        headers: item.headers.reduce((obj, [key, val]) => {
          obj[key] = val;
          return obj;
        }, {}),
        body: item.body
      });

      if (!response.ok) {
        throw new Error(`Server returned ${response.status}`);
      }

      console.log('[Service Worker] Synced:', item.id);
    } catch (error) {
      console.error('[Service Worker] Failed to sync:', item.id, error);
      item.retries++;

      if (item.retries < 3) {
        failed.push(item);
      } else {
        // Max retries reached, notify client
        notifyClients({
          type: 'SYNC_FAILED',
          item: item,
          error: error.message
        });
      }
    }
  }

  // Save remaining items
  await saveSyncQueue(failed);

  // Notify clients of sync completion
  notifyClients({
    type: 'SYNC_COMPLETE',
    synced: queue.length - failed.length,
    failed: failed.length
  });
}
