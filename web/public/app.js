const HOST_WAIT_INTERVAL = 100 // [ms]

let basehost = undefined

function sleep(ms) {
    return new Promise((resolve) => setTimeout(resolve, ms))
}

async function api(
    endpoint,
    method = 'get',
    payload = undefined,
    host = undefined
) {
    let wait = false
    while (host === undefined) {
        host = basehost
        if (wait) {
            // console.log('Waiting for API connection')
            await sleep(HOST_WAIT_INTERVAL)
        } else wait = true
    }
    req = {
        method,
        headers: {
            'Content-Type': 'application/json',
        },
    }
    if (payload !== undefined) {
        req.body = JSON.stringify(payload)
    }
    return fetch(`${host}api/v0/${endpoint}`, req)
        .then((response) => response.json())
        .catch((response) => response)
}

async function find_host() {
    document.body.style.cursor = 'wait'
    host_list = [
        '',
        'http://localhost:3000/',
        'http://192.168.188.10:3000/',
        'http://192.168.188.20:3000/',
        'http://192.168.42.1:3000/',
    ]
    let connected = false
    for (host of host_list) {
        console.log(`Trying API host: ${host}`)
        r = await api('ping', 'get', undefined, host)

        if (r === 'Pong') {
            basehost = host
            console.log('Connected to host')
            connected = true
            break
        }
    }
    if (connected) {
        document.body.style.cursor = 'default'
        document.getElementById('host-info').innerText = basehost
        document.getElementById('loadscreen').classList.remove('is-active')
    }
}

function init_site() {
    find_host()
    // TODO error handling
    selectedProfile = ''
    profiles = {
        names: [],
        async init() {
            await this.update()
        },
        async update() {
            r = await api('profiles')
            this.names = r['entity']
        },
    }
    Alpine.store('profiles', profiles)

    volume = {
        level: 0,
        async init() {
            await this.update()
        },
        async update() {
            r = await api('system/volume')
            this.level = r['level']
        },
        async set(ev) {
            vol = parseInt(ev.target.value)
            if (this.level != vol) {
                this.level = vol
                await api('system/volume', 'post', { level: this.level })
            }
        },
    }

    intensity = {
        level: 0,
        async init() {
            await this.update()
        },
        async update() {
            r = await api('system/intensity')
            this.level = r['level']
        },
        async set(ev) {
            intensity = parseInt(ev.target.value)
            if (this.level != intensity) {
                this.level = intensity
                await api('system/intensity', 'post', { level: this.level })
            }
        },
    }

    Alpine.store('volume', volume)
    Alpine.store('intensity', intensity)

    selectedEffect = ''
    lights = {
        names: [],
        async init() {
            await this.update()
        },
        async update() {
            r = await api('lights')
            this.names = r['entity']
        },
    }

    sounds = {
        names: [],
        async init() {
            await this.update()
        },
        async update() {
            r = await api('sounds')
            this.names = r['entity']
        },
    }

    playlists = {
        lists: [
            {
                name: '',
                songs: [''],
                position: 0,
                chance: 50,
            },
        ],
        async init() {
            await this.update()
            await this.updateSongs()
            await this.updatePositions()
            await this.updateChances()
        },
        async update() {
            r = await api('music')
            for (const [index, playlist] of r['entity'].entries()) {
                this.lists[index] = {
                    name: playlist,
                    songs: [''],
                    position: 0,
                }
            }
        },
        async updateSongs(playlist = undefined) {
            for (const [index, list] of this.lists.entries()) {
                if (playlist !== undefined && list.name !== playlist) continue
                r = await api(`music/${list.name}`)
                this.lists[index].songs = r['entity']
            }
        },
        async updatePositions(playlist = undefined) {
            for (const [index, list] of this.lists.entries()) {
                if (playlist !== undefined && list.name !== playlist) continue
                r = await api(`music/${list.name}/position`)
                this.lists[index].position = r['position']
            }
        },
        async updateChances(playlist = undefined) {
            for (const [index, list] of this.lists.entries()) {
                if (playlist !== undefined && list.name !== playlist) continue
                r = await api(`music/${list.name}/chance`)
                this.lists[index].chance = r['chance']
            }
        },
    }

    playing = {
        info: {
            title: '',
            artist: '',
            image: '',
            playlist: '',
        },
        async init() {
            await this.update()
        },
        async update() {
            r = await api('music/playing')
            this.info = r
        },
    }

    Alpine.store('lights', lights)
    Alpine.store('sounds', sounds)
    Alpine.store('playlists', playlists)
    Alpine.store('playing', playing)

    subscribe_events()

    // mapboxgl.accessToken = 'pk.eyJ1IjoiZHVsbGkiLCJhIjoiY2t6bXJmZWlmMDJiajJ3cGRycThpZ3E1OSJ9.7vX_cgY5K65kdVstLM2WFg'
    // const map = new mapboxgl.Map({
    //     container: document.getElementById('map'), // container ID
    //     style: 'mapbox://styles/mapbox/light-v10', // style URL
    //     center: [8.88, 53.12], // starting position [lng, lat]
    //     zoom: 13, // starting zoom
    // })
}

async function subscribe_events(host = undefined) {
    let wait = false
    while (host === undefined) {
        host = basehost
        if (wait) {
            // console.log('Waiting for SSE connection')
            await sleep(HOST_WAIT_INTERVAL)
        } else wait = true
    }

    sse = new EventSource(`${host}sse?stream=events`)
    sse.onmessage = function (event) {
        data = JSON.parse(event.data)

        let all = data.origin == 'config' && data.type == 'changed'
        if (all || (data.origin == 'music' && data.type == 'playing')) {
            Alpine.store('playing').update()
        }
        if (all || (data.origin == 'music' && data.type == 'position')) {
            Alpine.store('playlists').updatePositions(data.name)
        }
        if (all || (data.origin == 'music' && data.type == 'shuffle')) {
            Alpine.store('playlists').updateSongs(data.name)
        }
        if (all || (data.origin == 'audio' && data.type == 'volume')) {
            Alpine.store('volume').update()
        }
        if (all || (data.origin == 'audio' && data.type == 'intensity')) {
            Alpine.store('playlists').updateChances()
            Alpine.store('intensity').update()
        }
    }
}

document.addEventListener('alpine:init', init_site)
