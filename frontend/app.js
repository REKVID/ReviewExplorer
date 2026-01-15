const input = document.getElementById('school-search'),
    list = document.getElementById('suggestions'),
    view = document.getElementById('analytics-view'),
    title = document.getElementById('school-title'),
    text = document.getElementById('analytics-text'),
    searchBtn = document.getElementById('search-btn'),
    statusMsg = document.getElementById('status-msg');

let timer, charts = [], map, geoLayer;

DG.then(() => {
    map = DG.map('map', { center: [55.7558, 37.6173], zoom: 11 });

    fetch('http://localhost:8081/schools')
        .then(res => res.json())
        .then(schools => {
            if (!schools) return;
            const features = schools.map(s => ({
                type: "Feature",
                geometry: { type: "Point", coordinates: [s.lon, s.lat] },
                properties: { name: s.full_name, shortName: s.short_name }
            }));

            geoLayer = DG.geoJson({ type: "FeatureCollection", features }, {
                style: () => ({ color: "#4f46e5", weight: 2 }),
                onEachFeature: (feature, layer) => {
                    const name = feature.properties.name;
                    const short = feature.properties.shortName || "Школа";
                    layer.bindPopup(`<b>${short}</b><br><button onclick="startAnalysis('${name}')" style="margin-top:8px">Анализировать</button>`);

                    layer.on('click', () => {
                    });
                }
            }).addTo(map);
        });
});

input.oninput = (e) => {
    clearTimeout(timer);
    const q = e.target.value.trim();
    if (q.length < 2) return list.classList.add('hidden');
    timer = setTimeout(async () => {
        const res = await fetch(`http://localhost:8081/schools?q=${encodeURIComponent(q)}`);
        const data = await res.json();
        renderSuggestions(data || []);
    }, 300);
};

function renderSuggestions(data) {
    list.innerHTML = data.length ? '' : '<div class="suggestion-item">Не найдено</div>';
    data.slice(0, 5).forEach(s => {
        const div = document.createElement('div');
        div.className = 'suggestion-item';
        div.textContent = s.short_name || s.full_name;
        div.onclick = () => {
            input.value = s.full_name;
            list.classList.add('hidden');
            startAnalysis(s.full_name);

            if (s.lat && s.lon) {
                map.setView([s.lat, s.lon], 15);
                DG.popup().setLatLng([s.lat, s.lon]).setContent(s.short_name).openOn(map);
            }
        };
        list.appendChild(div);
    });
    list.classList.remove('hidden');
}

searchBtn.onclick = () => {
    const q = input.value.trim();
    if (q.length < 3) return alert('Введите название');
    startAnalysis(q);
};

function showReviews(title, list) {
    const overlay = document.getElementById('review-overlay');
    document.getElementById('overlay-title').textContent = title;
    document.getElementById('overlay-content').innerHTML = list.map(t => `<div style="padding:15px; border-bottom:1px solid #f1f5f9; line-height:1.5">${t}</div>`).join('');
    overlay.classList.remove('hidden');
}

async function startAnalysis(query) {
    if (!query) return;
    input.value = query;
    searchBtn.disabled = true;
    view.classList.remove('hidden');
    statusMsg.textContent = '⏳ Идет обработка данных...';
    charts.forEach(c => c.destroy());
    charts = [];

    const res = await fetch('http://localhost:8081/analyze', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ query })
    });
    const data = await res.json();

    if (res.ok) {
        title.textContent = data.school_name;
        statusMsg.textContent = '✅ Аналитика готова';
        text.innerHTML = `
            <div style="font-weight:850; font-size:1.2rem; margin-bottom:1.5rem">Сводка: ${data.stats.total} отзывов</div>
            <div style="display:grid; grid-template-columns:repeat(3,1fr); gap:15px">
                <div style="background:#f0fdf4; padding:15px; border-radius:15px; border:1px solid #dcfce7; color:#16a34a">
                    <div style="font-size:1.5rem; font-weight:900">${data.stats.positive}</div>
                    <small style="font-weight:700">ПОЗИТИВ</small>
                </div>
                <div style="background:#fef2f2; padding:15px; border-radius:15px; border:1px solid #fee2e2; color:#dc2626">
                    <div style="font-size:1.5rem; font-weight:900">${data.stats.negative}</div>
                    <small style="font-weight:700">НЕГАТИВ</small>
                </div>
                <div style="background:#f8fafc; padding:15px; border-radius:15px; border:1px solid #f1f5f9; color:#475569">
                    <div style="font-size:1.5rem; font-weight:900">${data.stats.neutral}</div>
                    <small style="font-weight:700">НЕЙТРАЛ</small>
                </div>
            </div>`;

        data.analytics.forEach((item, i) => {
            const isLine = item.type === 'line';
            const isHoriz = i === 2;

            const rawLabels = isLine ? item.payload.map(p => p.label) : Object.keys(item.payload);
            const labels = rawLabels.map(l => l.length > 12 ? l.split(' ') : l);

            const rawValues = isLine ? item.payload.map(p => p.value) : Object.values(item.payload);
            const values = rawValues.map(v => Array.isArray(v) ? v.length : v);

            const c = new Chart(document.getElementById(`chart-${i}`), {
                type: item.type,
                data: {
                    labels,
                    datasets: [{
                        data: values,
                        backgroundColor: isLine ? 'rgba(79, 70, 229, 0.15)' : 'rgba(79, 70, 229, 0.9)',
                        borderColor: '#4f46e5',
                        borderWidth: isLine ? 3 : 0,
                        borderRadius: 12,
                        fill: isLine,
                        tension: 0.4,
                        pointRadius: isLine ? 4 : 0
                    }]
                },
                options: {
                    indexAxis: isHoriz ? 'y' : 'x',
                    responsive: true,
                    maintainAspectRatio: true,
                    plugins: { legend: { display: false } },
                    onClick: (event, elements) => {
                        if (elements.length > 0) {
                            const idx = elements[0].index;
                            const theme = rawLabels[idx];
                            const source = rawValues[idx];
                            if (Array.isArray(source)) {
                                showReviews(`${theme} (${source.length} отзывов)`, source);
                            }
                        }
                    },
                    scales: {
                        y: { grid: { display: false }, ticks: { font: { weight: '700', size: 12, family: 'Inter' }, color: '#111827' } },
                        x: { grid: { display: false }, ticks: { font: { weight: '700', size: 12, family: 'Inter' }, color: '#111827' } }
                    }
                }
            });
            charts.push(c);
        });
    }
    searchBtn.disabled = false;
}
input.onkeypress = (e) => { if (e.key === 'Enter') searchBtn.click(); };
