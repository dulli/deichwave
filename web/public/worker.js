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
    e.waitUntil(
        caches.open(cacheName).then(function (cache) {
            return cache.addAll(filesToCache)
        })
    )
})

self.addEventListener('fetch', function (e) {
    e.respondWith(
        caches.match(e.request).then(function (response) {
            return response || fetch(e.request)
        })
    )
})
