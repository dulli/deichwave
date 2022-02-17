function api(endpoint, method = 'get', payload = undefined) {
    req = {
        method,
        headers: {
            'Content-Type': 'application/json',
        },
    }
    if (payload !== undefined) {
        req.body = JSON.stringify(payload)
    }
    return fetch(`api/v0/${endpoint}`, req).then((response) => response.json())
}

document.addEventListener('alpine:init', () => {
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

    sse = new EventSource('sse?stream=events')
    sse.onmessage = function (event) {
        data = JSON.parse(event.data)
        // console.log(data)

        if (data.origin == 'music' && data.type == 'playing') {
            Alpine.store('playing').update()
            return
        }
        if (data.origin == 'music' && data.type == 'position') {
            Alpine.store('playlists').updatePositions(data.name)
            return
        }
        if (data.origin == 'music' && data.type == 'shuffle') {
            Alpine.store('playlists').updateSongs(data.name)
            return
        }
        if (data.origin == 'audio' && data.type == 'volume') {
            Alpine.store('volume').update()
            return
        }
        if (data.origin == 'audio' && data.type == 'intensity') {
            Alpine.store('playlists').updateChances()
            Alpine.store('intensity').update()
            return
        }
    }

    // mapboxgl.accessToken = 'pk.eyJ1IjoiZHVsbGkiLCJhIjoiY2t6bXJmZWlmMDJiajJ3cGRycThpZ3E1OSJ9.7vX_cgY5K65kdVstLM2WFg'
    // const map = new mapboxgl.Map({
    //     container: document.getElementById('map'), // container ID
    //     style: 'mapbox://styles/mapbox/light-v10', // style URL
    //     center: [8.88, 53.12], // starting position [lng, lat]
    //     zoom: 13, // starting zoom
    // })
})
