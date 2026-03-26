document.addEventListener('DOMContentLoaded', function() {
    // CONSTANTS
    const MAX_CITY_SEARCH_RESULTS = 5;
    const GERMANY_BOUNDS = [
        [50.913868, 5.8],
        [55.1, 8.041992],
        [50.999929, 15.0],
        [47.3, 10.217285]
    ];

    // UTILITY FUNCTIONS
    const createElement = (tag, options = {}) => {
        const element = document.createElement(tag);
        Object.assign(element, options);
        return element;
    };

    const getClubText = (count) => count === 1 ? 'Club' : 'Clubs';
    const getSocialRunClubText = (count) => count === 1 ? 'Social Run Club' : 'Social Run Clubs';

    const fixLeafletButtons = (div) => {
        div.querySelectorAll('[role="button"]').forEach((btn) => {
            btn.removeAttribute('role');
        });
    };

    // FILTER
    const filterInput = document.getElementById('filter-input');
    if (filterInput) {
        filterInput.addEventListener('input', function() {
            const filter = filterInput.value.toLowerCase();
            document.querySelectorAll('[data-search]').forEach(function(el) {
                el.style.display = filter === '' || el.dataset.search.includes(filter) ? '' : 'none';
            });
        });
    }

    // CITY SEARCH
    const citySearchInput = document.getElementById('city-search-input');
    if (citySearchInput) {
        const citySearchResults = document.getElementById('city-search-results');

        const addNotFoundResult = () => {
            const liEl = createElement('li', { textContent: 'Keine passende Stadt gefunden' });
            citySearchResults.appendChild(liEl);
        };

        const addMoreResultsHint = (count) => {
            const liEl = createElement('li', { textContent: `... und ${count} weitere Städte` });
            citySearchResults.appendChild(liEl);
        };

        const addResult = (city) => {
            const [cityName, cityUrl, cityClubs] = city;
            const clubText = getSocialRunClubText(cityClubs);
            const liEl = document.createElement('li');
            const aEl = createElement('a', {
                href: cityUrl,
                textContent: `${cityName} (${cityClubs} ${clubText})`,
                className: 'city-search-result'
            });
            liEl.appendChild(aEl);
            citySearchResults.appendChild(liEl);
        };

        citySearchInput.addEventListener('input', function() {
            const filter = citySearchInput.value.toLowerCase().trim();
            if (filter === '') {
                citySearchResults.style.display = 'none';
                citySearchResults.innerHTML = '';
                return;
            }

            const results = cityData.filter(city => city[0].toLowerCase().includes(filter));
            citySearchResults.innerHTML = '';

            if (results.length === 0) {
                addNotFoundResult();
            } else {
                const displayResults = results.length > MAX_CITY_SEARCH_RESULTS
                    ? results.slice(0, MAX_CITY_SEARCH_RESULTS)
                    : results;
                displayResults.forEach(addResult);

                if (results.length > MAX_CITY_SEARCH_RESULTS) {
                    addMoreResultsHint(results.length - MAX_CITY_SEARCH_RESULTS);
                }
            }
            citySearchResults.style.display = 'block';
        });
    }

    // MAPS
    const baseLayer = L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        maxZoom: 18,
        attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
    });

    const initializeMap = (mapId, options = {}) => {
        const map = L.map(mapId, { gestureHandling: true, ...options });
        baseLayer.addTo(map);
        return map;
    };

    const createMarker = (lat, lon, popupContent) => {
        return L.marker([lat, lon]).bindPopup(popupContent);
    };

    // Club Map
    const clubMapDiv = document.getElementById('club-map');
    if (clubMapDiv) {
        const lat = parseFloat(clubMapDiv.dataset.lat);
        const lon = parseFloat(clubMapDiv.dataset.lon);
        const name = clubMapDiv.dataset.name;
        const cityName = clubMapDiv.dataset.cityname;
        const hasNoLocation = clubMapDiv.dataset.nolocation !== undefined;

        let popup = `<b>${name}</b>`;
        if (hasNoLocation) {
            popup += `<br><i>irgendwo in ${cityName}</i><br>(Wir kennen den genauen Ort leider nicht 😞)`;
        } else {
            popup += `<br><i>${cityName}</i>`;
        }

        const map = initializeMap('club-map').setView([lat, lon], 13);
        createMarker(lat, lon, popup).addTo(map).openPopup();
        fixLeafletButtons(clubMapDiv);
    }

    // City Map
    const cityMapDiv = document.getElementById('city-map');
    if (cityMapDiv) {
        const lat = parseFloat(cityMapDiv.dataset.lat);
        const lon = parseFloat(cityMapDiv.dataset.lon);
        const name = cityMapDiv.dataset.name;

        const map = initializeMap('city-map');
        createMarker(lat, lon, name).addTo(map).openPopup();
        map.fitBounds(GERMANY_BOUNDS);
        fixLeafletButtons(cityMapDiv);
    }

    // Cities Map
    const citiesMapDiv = document.getElementById('cities-map');
    if (citiesMapDiv) {
        const markers = L.layerGroup();
        document.querySelectorAll('[data-search]').forEach(function(cityEl) {
            const city = {
                url: cityEl.dataset.url,
                name: cityEl.dataset.name,
                clubs: cityEl.dataset.clubs,
                lat: cityEl.dataset.lat,
                lon: cityEl.dataset.lon
            };
            const clubText = getClubText(city.clubs);
            const popup = `<a href="${city.url}">${city.name}</a> (${city.clubs} ${clubText})`;
            const marker = createMarker(city.lat, city.lon, popup);
            markers.addLayer(marker);
        });
        const map = initializeMap('cities-map', { layers: [baseLayer, markers] }).fitBounds(GERMANY_BOUNDS);
        fixLeafletButtons(citiesMapDiv);
    }

    // Clubs Map
    const clubsMapDiv = document.getElementById('clubs-map');
    if (clubsMapDiv) {
        const markers = L.layerGroup();
        document.querySelectorAll('[data-search]').forEach(function(clubEl) {
            const club = {
                name: clubEl.dataset.name,
                lat: parseFloat(clubEl.dataset.lat),
                lon: parseFloat(clubEl.dataset.lon),
                url: clubEl.dataset.url
            };
            const popup = `<a href="${club.url}">${club.name}</a>`;
            const marker = createMarker(club.lat, club.lon, popup);
            markers.addLayer(marker);
        });
        const map = initializeMap('clubs-map', { layers: [baseLayer, markers] }).fitBounds(GERMANY_BOUNDS);
        fixLeafletButtons(clubsMapDiv);
    }

    // Cluster Map
    const clusterMapDiv = document.getElementById('cluster-map');
    if (clusterMapDiv) {
        const markers = L.markerClusterGroup();
        clusterData.forEach(function(club) {
            const popup = `<a href="${club[4]}">${club[2]}</a><br>in ${club[3]}`;
            const marker = createMarker(club[0], club[1], popup);
            markers.addLayer(marker);
        });
        const map = initializeMap('cluster-map', { layers: [baseLayer, markers] }).fitBounds(GERMANY_BOUNDS);
        fixLeafletButtons(clusterMapDiv);
    }

    // SHARE
    document.querySelectorAll('[data-share]').forEach(function(shareBtn) {
        const url = shareBtn.dataset.url || window.location.href;
        const title = shareBtn.dataset.title || document.title;
        const data = {
            title: title,
            url: url + "?utm_source=share_button"
        };

        if (navigator.canShare === undefined || !navigator.canShare(data)) {
            shareBtn.style.display = 'none';
            return;
        }

        shareBtn.addEventListener('click', async (e) => {
            e.preventDefault();
            if (window.umami !== undefined) {
                window.umami.track('share-click', {url: data.url});
            }
            try {
                await navigator.share(data);
            } catch (error) {
                console.error("Error sharing:", error);
            }
        });
    });

    // UMAMI
    document.querySelectorAll("a[target=_blank]").forEach((a) => {
        if (a.getAttribute("data-umami-event") === null) {
            a.setAttribute('data-umami-event', 'outbound-link-click');
        }
        a.setAttribute('data-umami-event-url', a.href);
    });

    if (location.hash === '#disable-umami') {
        localStorage.setItem('umami.disabled', 'true');
        alert('Umami is now DISABLED in this browser.');
    }
    if (location.hash === '#enable-umami') {
        localStorage.removeItem('umami.disabled');
        alert('Umami is now ENABLED in this browser.');
    }
});
