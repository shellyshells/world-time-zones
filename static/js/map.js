// Initialize map with custom CRS to enable wrapping
const map = L.map('map', {
    center: [20, 0],
    zoom: 2,
    maxBounds: [[-90, -Infinity], [90, Infinity]],
    minZoom: 1,
    worldCopyJump: true
});

// Add the tile layer with noWrap: false to enable infinite scrolling
L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
    noWrap: false,
    bounds: [[-90, -Infinity], [90, Infinity]]
}).addTo(map);

// Function to calculate terminator coordinates with corrected solar position
function calculateTerminator() {
    const now = new Date();
    const julianDate = (now.getTime() / 86400000) + 2440587.5;
    const century = (julianDate - 2451545) / 36525;
    
    // Calculate solar position with time correction
    const GMST = (280.46061837 + 360.98564736629 * (julianDate - 2451545.0)) % 360;
    const meanLongitude = (280.460 + 36000.77 * century) % 360;
    const meanAnomaly = (357.528 + 35999.05 * century) % 360;
    const equationOfCenter = 1.915 * Math.sin(meanAnomaly * Math.PI / 180) 
                           + 0.02 * Math.sin(2 * meanAnomaly * Math.PI / 180);
    const eclipticLongitude = meanLongitude + equationOfCenter;
    const obliquity = 23.439 - 0.013 * century;
    
    const declination = Math.asin(Math.sin(obliquity * Math.PI / 180) 
                      * Math.sin(eclipticLongitude * Math.PI / 180)) * 180 / Math.PI;
    const rightAscension = Math.atan2(Math.cos(obliquity * Math.PI / 180) 
                         * Math.sin(eclipticLongitude * Math.PI / 180), 
                         Math.cos(eclipticLongitude * Math.PI / 180)) * 180 / Math.PI;
    
    const adjustedRA = rightAscension - GMST;
    
    const terminatorCoords = [];
    for (let lon = -180; lon <= 180; lon += 2) {
        const localHourAngle = lon - adjustedRA;
        const latRad = Math.atan(-Math.cos(localHourAngle * Math.PI / 180) 
                     / Math.tan(declination * Math.PI / 180));
        const lat = latRad * 180 / Math.PI;
        terminatorCoords.push([lat, lon]);
    }
    
    return { 
        coords: terminatorCoords, 
        rightAscension: adjustedRA,
        isDay: (lon) => {
            const localHourAngle = lon - adjustedRA;
            return Math.abs(localHourAngle) < 90;
        }
    };
}

// Animation state variables
let animationFrameId;
let lastDrawTime = 0;
let nightPolygon;
let terminatorVisible = false;

// Function to draw terminator and night shading
function drawTerminator(timestamp) {
    const elapsed = timestamp - lastDrawTime;
    
    if (elapsed > 100) {
        if (window.terminatorLine) {
            map.removeLayer(window.terminatorLine);
        }
        if (nightPolygon) {
            map.removeLayer(nightPolygon);
        }

        const { coords, rightAscension, isDay } = calculateTerminator();
        
        window.terminatorLine = L.polyline(coords, {
            color: '#FFA500',
            weight: 2,
            opacity: 0.7,
            smoothFactor: 1.5,
            dashArray: '5, 5'
        }).addTo(map);
        
        const nightCoords = [
            ...coords,
            [90, coords[coords.length - 1][1]],
            [90, -180],
            [-90, -180],
            [-90, coords[0][1]],
            coords[0]
        ];

        const centerLon = (coords[0][1] + coords[coords.length - 1][1]) / 2;
        if (isDay(centerLon)) {
            nightCoords.reverse();
        }

        nightPolygon = L.polygon(nightCoords, {
            color: '#000',
            weight: 0,
            fillColor: '#000',
            fillOpacity: 0.2,
            smoothFactor: 1.5
        }).addTo(map);

        lastDrawTime = timestamp;
    }
    
    if (terminatorVisible) {
        animationFrameId = requestAnimationFrame(drawTerminator);
    }
}

// Function to get color based on UTC offset
function getColorForOffset(offset) {
    const normalized = (parseFloat(offset) + 12) / 24;
    return `hsl(${normalized * 240}, 70%, 50%)`;
}

// Function to get local time for a timezone
function getLocalTime(offset) {
    const now = new Date();
    const utc = now.getTime() + (now.getTimezoneOffset() * 60000);
    return new Date(utc + (3600000 * offset)).toLocaleTimeString();
}

// Update the timezone info display
function updateTimezoneInfo(tzid) {
    const selectedTimezone = document.getElementById('selected-timezone');
    selectedTimezone.textContent = tzid;
}

let timezoneBorders = null;
let timezonesVisible = false;

function loadTimezoneBorders() {
    fetch('/api/timezone-borders')
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.json();
        })
        .then(data => {
            if (timezoneBorders) {
                map.removeLayer(timezoneBorders);
            }

            timezoneBorders = L.geoJSON(data, {
                style: function(feature) {
                    return {
                        color: '#333',
                        weight: 1,
                        fillColor: getColorForOffset(feature.properties.offset),
                        fillOpacity: 0.3
                    };
                },
                onEachFeature: function(feature, layer) {
                    layer.on('click', function(e) {
                        updateTimezoneInfo(feature.properties.tzid);
                    });

                    layer.on('mouseover', function(e) {
                        this.setStyle({
                            fillOpacity: 0.6
                        });
                    });
                    layer.on('mouseout', function(e) {
                        this.setStyle({
                            fillOpacity: 0.3
                        });
                    });
                }
            }).addTo(map);
        })
        .catch(error => {
            console.error('Error loading timezone data:', error);
            alert('Failed to load timezone data. Please try again later.');
        });
}

function toggleTimezones() {
    timezonesVisible = !timezonesVisible;
    const button = document.getElementById('timezone-toggle');
    button.classList.toggle('active');
    const label = button.querySelector('.label');
    label.textContent = timezonesVisible ? 'Hide Time Zones' : 'Show Time Zones';
    
    if (timezonesVisible) {
        if (!timezoneBorders) {
            loadTimezoneBorders();
        } else {
            map.addLayer(timezoneBorders);
        }
    } else if (timezoneBorders) {
        map.removeLayer(timezoneBorders);
    }
}

function toggleTerminator() {
    terminatorVisible = !terminatorVisible;
    const button = document.getElementById('terminator-toggle');
    button.classList.toggle('active');
    const label = button.querySelector('.label');
    label.textContent = terminatorVisible ? 'Hide Day/Night' : 'Show Day/Night';
    
    if (terminatorVisible) {
        lastDrawTime = 0;
        animationFrameId = requestAnimationFrame(drawTerminator);
    } else {
        if (animationFrameId) {
            cancelAnimationFrame(animationFrameId);
        }
        if (window.terminatorLine) {
            map.removeLayer(window.terminatorLine);
        }
        if (nightPolygon) {
            map.removeLayer(nightPolygon);
        }
    }
}

let infoVisible = false;
const infoBanner = document.querySelector('.info-banner');

function toggleInfoBanner(event) {
    // Prevent the click from propagating to the document
    event.stopPropagation();
    
    infoVisible = !infoVisible;
    const button = document.getElementById('info-toggle');
    button.classList.toggle('active');
    infoBanner.style.display = infoVisible ? 'block' : 'none';
    const label = button.querySelector('.label');
    label.textContent = infoVisible ? 'Close Info' : 'About Time Zones';
}

document.getElementById('timezone-toggle').addEventListener('click', toggleTimezones);
document.getElementById('terminator-toggle').addEventListener('click', toggleTerminator);
document.getElementById('info-toggle').addEventListener('click', toggleInfoBanner);

document.addEventListener('click', (e) => {
    if (infoVisible && !infoBanner.contains(e.target) && !e.target.closest('#info-toggle')) {
        infoVisible = false;
        const button = document.getElementById('info-toggle');
        button.classList.remove('active');
        infoBanner.style.display = 'none';
        const label = button.querySelector('.label');
        label.textContent = 'About Time Zones';
    }
});

// Fullscreen functionality
const mapContainer = document.querySelector('.map-container');
const fullscreenButton = document.getElementById('fullscreen-toggle');
let isFullscreen = false;

function toggleFullscreen() {
    isFullscreen = !isFullscreen;
    mapContainer.classList.toggle('map-fullscreen');
    fullscreenButton.textContent = isFullscreen ? '⤓' : '⤢';
    fullscreenButton.setAttribute('aria-label', 
        isFullscreen ? 'Exit fullscreen' : 'Enter fullscreen'
    );
    map.invalidateSize();
}

// Handle ESC key to exit fullscreen
document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape' && isFullscreen) {
        toggleFullscreen();
    }
});

fullscreenButton.addEventListener('click', toggleFullscreen);