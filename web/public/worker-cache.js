let cacheName = 'deichwave'
let filesToCache = [
    'index.html',
    'app.html',
    'app.js',
    'favicon.ico',
    'static/favicon.svg',
    'static/logo.svg',
    'static/touch-apple.png',
    'static/touch-google.png',
    'vendor/alpine.min.js',
    'vendor/bulma-pageloader.min.css',
    'vendor/bulma-prefers-dark.css',
    'vendor/bulma-slider.min.css',
    'vendor/bulma.min.css',
]

self.addEventListener('install', function (e) {
    console.log('Worker: Installing')
    e.waitUntil(
        (async () => {
            const cache = await caches.open(cacheName)
            console.log('Worker: Caching')
            await cache.addAll(filesToCache)
        })()
    )
})

self.addEventListener('fetch', function (e) {
    e.respondWith(
        (async () => {
            const cached = await caches.match(e.request)
            if (cached) {
                console.log(`Worker: Fetching ${e.request.url} from cache`)
                return cached
            }

            const fetched = await fetch(e.request)
            return fetched
        })()
    )
})
