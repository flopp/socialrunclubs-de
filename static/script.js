document.addEventListener('DOMContentLoaded', function() {
    // FILTER
    var filterInput = document.getElementById('filter-input');
    if (filterInput) {
        filterInput.addEventListener('input', function() {
            var filter = filterInput.value.toLowerCase();
            document.querySelectorAll('[data-search]').forEach(function(el) {
                if (filter === '' || el.dataset.search.includes(filter)) {
                    el.style.display = '';
                } else {
                    el.style.display = 'none';
                }
            });
        });
    }

    // MAPS
    const germany = [
        [50.913868, 5.603027],
        [55.329144, 8.041992],
        [50.999929, 15.227051],
        [47.034162, 10.217285]
    ];
    var baseLayer = L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        maxZoom: 18,
        attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
    });
    
    var mapDiv = document.getElementById('club-map');
    if (mapDiv) {
        var lat = parseFloat(mapDiv.dataset.lat);
        var lon = parseFloat(mapDiv.dataset.lon);
        var name = mapDiv.dataset.name;
        var map = L.map('club-map').setView([lat, lon], 13);
        baseLayer.addTo(map);
        L.marker([lat, lon]).addTo(map)
            .bindPopup(name).openPopup();
    }

    if (document.getElementById('city-map')) {
        var markers = L.layerGroup();
        document.querySelectorAll('[data-search]').forEach(function(cityEl) {
            var city = {
                url: cityEl.dataset.url,
                name: cityEl.dataset.name,
                clubs: cityEl.dataset.clubs,
                lat: cityEl.dataset.lat,
                lon: cityEl.dataset.lon
            };
            const clubText = city.clubs == 1 ? 'Club' : 'Clubs';
            var marker = L.marker([city.lat, city.lon])
                .bindPopup('<a href="' + city.url + '">' + city.name + '</a> (' + city.clubs + ' ' + clubText + ')');
            markers.addLayer(marker);
        });
        var map = L.map('city-map', {gestureHandling: true, layers: [baseLayer, markers]}).fitBounds(germany);
    }

    if (document.getElementById('clubs-map')) {
        //var markers = L.markerClusterGroup();
        var markers = L.layerGroup();
        document.querySelectorAll('[data-search]').forEach(function(clubEl) {
            var club = {
                name: clubEl.dataset.name,
                lat: parseFloat(clubEl.dataset.lat),
                lon: parseFloat(clubEl.dataset.lon),
                url: clubEl.dataset.url
            };
            var marker = L.marker([club.lat, club.lon])
                .bindPopup('<a href="' + club.url + '">' + club.name + '</a>');
            markers.addLayer(marker);
        });
        var map = L.map('clubs-map', {gestureHandling: true, layers: [baseLayer, markers]}).fitBounds(germany);
    }

    if (document.getElementById('cluster-map')) {
        var markers = L.markerClusterGroup();

        clusterData.forEach(function(club) {
            var marker = L.marker([club[0], club[1]])
                .bindPopup('<a href="' + club[3] + '">' + club[2] + '</a>');
            markers.addLayer(marker);
        });
        var map = L.map('cluster-map', {gestureHandling: true, layers: [baseLayer, markers]}).fitBounds(germany);
    }

    // SHARE
    document.querySelectorAll('[data-share]').forEach(function(shareBtn) {
        var url = shareBtn.dataset.url || window.location.href;
        var title = shareBtn.dataset.title || document.title;
        var data = {
            title: title,
            url: url + "?utm_source=share_button"
        };

        if (navigator.canShare === undefined || !navigator.canShare(data)) {
            // hide button
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
