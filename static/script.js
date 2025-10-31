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
    var mapDiv = document.getElementById('club-map');
    if (mapDiv) {
        var lat = parseFloat(mapDiv.dataset.lat);
        var lon = parseFloat(mapDiv.dataset.lon);
        var name = mapDiv.dataset.name;
        var map = L.map('club-map').setView([lat, lon], 13);
        L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
            maxZoom: 18,
            attribution: '© OpenStreetMap contributors'
        }).addTo(map);
        L.marker([lat, lon]).addTo(map)
            .bindPopup(name).openPopup();
    }

    mapDiv = document.getElementById('city-map');
    if (mapDiv) {
        // collect city data (all html elements with data-city)
        var cityData = [];
        document.querySelectorAll('[data-city]').forEach(function(cityEl) {
            // skip elements without coordinates
            if (!cityEl.dataset.lat || !cityEl.dataset.lon) return;
            cityData.push({
                url: cityEl.dataset.url,
                name: cityEl.dataset.name,
                clubs: cityEl.dataset.clubs,
                lat: cityEl.dataset.lat,
                lon: cityEl.dataset.lon
            });
        });
        const germany = [
            [50.913868, 5.603027],
            [55.329144, 8.041992],
            [50.999929, 15.227051],
            [47.034162, 10.217285]
        ];
        var map = L.map('city-map').fitBounds(germany)
        L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
            maxZoom: 18,
            attribution: '© OpenStreetMap contributors'
        }).addTo(map);
        cityData.forEach(function(city) {
            const clubText = city.clubs == 1 ? 'Club' : 'Clubs';
            L.marker([city.lat, city.lon]).addTo(map)
                .bindPopup('<a href="' + city.url + '">' + city.name + '</a> (' + city.clubs + ' ' + clubText + ')');
        });
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

        shareBtn.setAttribute('data-umami-event', 'share-click');
        shareBtn.setAttribute('data-umami-event-url', data.url);
        shareBtn.addEventListener('click', async (e) => {
            e.preventDefault();
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
