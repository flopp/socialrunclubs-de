document.addEventListener('DOMContentLoaded', function() {
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
});
