const HOST_WAIT_INTERVAL = 100 // [ms]
const BATTERY_POLL_INTERVAL = 5 * 60 * 1000 // [min * s * ms]

let basehost = undefined
let polling = false

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
    document.getElementById('loadscreen').classList.add('is-active')
    host_list = [
        '',
        'http://bbycr:3000/',
        'http://192.168.42.1:3000/',
        'http://pi:3000/',
        'http://localhost:3000/',
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
        r = await api('info/Meta', 'get', undefined, host)

        document.body.style.cursor = 'default'
        document.title = r.Name
        document.getElementById('host-name').innerText = r.Name
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

    webio = {
        switches: [],
        async init() {
            await this.update()
        },
        async update() {
            r = await api('info/webio')
            this.switches = r['Switches']
        },
    }

    Alpine.store('lights', lights)
    Alpine.store('sounds', sounds)
    Alpine.store('playlists', playlists)
    Alpine.store('playing', playing)
    Alpine.store('webio', webio)

    subscribe_events()

    // mapboxgl.accessToken = ''
    // const map = new mapboxgl.Map({
    //     container: document.getElementById('map'), // container ID
    //     style: 'mapbox://styles/mapbox/light-v10', // style URL
    //     center: [8.88, 53.12], // starting position [lng, lat]
    //     zoom: 13, // starting zoom
    // })
}

async function poll_battery(host = undefined) {
    polling = true
    let bms = (await api('shell/bms-check', 'post', undefined, host)).trim()
    let bs = document.getElementById('battery-status')
    let bl = document.getElementById('battery-level')
    let bd = document.getElementById('battery-details')

    async function refresh_battery() {
        if (!polling) return
        r = JSON.parse(await api('shell/bms-read', 'post', undefined, host))
        if (bms == 'jbd') bl.innerText = `${r['PercentCapacity']}%`
        bd.innerText = JSON.stringify(r)
        setTimeout(refresh_battery, BATTERY_POLL_INTERVAL)
    }

    if (bms == 'none' || bms == 'command could not be found')
        bs.classList.add('is-hidden')
    else refresh_battery()
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
        console.log(data)

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
    sse.onerror = function () {
        window.location = window.location
    }

    poll_battery(host)
}

document.addEventListener('alpine:init', init_site)
