const $ = (id) => document.getElementById(id);
const API = 'http://localhost:8081';

const input = $('school-search');
const list = $('suggestions');
const view = $('analytics-view');
const title = $('school-title');
const text = $('analytics-text');
const searchBtn = $('search-btn');
const statusMsg = $('status-msg');

let timer, charts = [], map;

const COLORS = {
    primary: '#4f46e5',
    negative: '#64748b',
    primaryBg: 'rgba(79, 70, 229, 0.14)',
    negativeBg: 'rgba(100, 116, 139, 0.14)'
};

const FONT = { weight: '700', size: 12, family: 'Inter' };
const TICKS = { font: FONT, color: '#111827' };

if (typeof particlesJS !== 'undefined') {
    particlesJS('particles-js', {
        particles: {
            number: { value: 50 },
            color: { value: COLORS.primary },
            shape: { type: 'circle' },
            opacity: { value: 0.8 },
            size: { value: 2 },
            line_linked: { enable: true, distance: 150, color: COLORS.primary, opacity: 0.4, width: 1.5 },
            move: { enable: true, speed: 3 }
        }
    });
}

const apiGet = (path) => fetch(`${API}${path}`).then(r => r.json());
const apiPost = (path, body) => fetch(`${API}${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body)
}).then(async (r) => ({ ok: r.ok, data: await r.json() }));

if (typeof DG !== 'undefined' && $('map')) {
    DG.then(async () => {
        map = DG.map('map', { center: [55.7558, 37.6173], zoom: 11 });
        const schools = await apiGet('/schools');
        if (!schools) return;

        const features = schools.map(s => ({
            type: "Feature",
            geometry: { type: "Point", coordinates: [s.lon, s.lat] },
            properties: { name: s.full_name, shortName: s.short_name, hasReviews: (s.review_count || 0) > 0 }
        }));

        DG.geoJson({ type: "FeatureCollection", features }, {
            pointToLayer: (feature, latlng) => {
                const has = !!feature.properties.hasReviews;
                const color = has ? COLORS.primary : '#94a3b8';
                const fill = has ? COLORS.primaryBg : 'rgba(148, 163, 184, 0.14)';
                return DG.circleMarker(latlng, { radius: 6, color, weight: 2, fillColor: fill, fillOpacity: 1 });
            },
            onEachFeature: (feature, layer) => {
                const name = feature.properties.name;
                const short = feature.properties.shortName || "Школа";
                layer.bindPopup(`<b>${short}</b><br><button onclick="startAnalysis('${name}')" style="margin-top:8px">Анализировать</button>`);
            }
        }).addTo(map);
    });
}

input?.addEventListener('input', (e) => {
    clearTimeout(timer);
    const q = e.target.value.trim();
    if (q.length < 2) return list?.classList.add('hidden');
    timer = setTimeout(async () => renderSuggestions((await apiGet(`/schools?q=${encodeURIComponent(q)}`)) || []), 300);
});

input?.addEventListener('keypress', (e) => { if (e.key === 'Enter') searchBtn?.click(); });

function renderSuggestions(data) {
    if (!list) return;
    list.innerHTML = data.length ? '' : '<div class="suggestion-item">Не найдено</div>';
    data.slice(0, 5).forEach(s => {
        const div = document.createElement('div');
        div.className = 'suggestion-item';
        div.textContent = s.short_name || s.full_name;
        div.style.borderLeft = (s.review_count || 0) > 0 ? `4px solid ${COLORS.primary}` : '4px solid #cbd5e1';
        div.style.paddingLeft = '21px';
        div.onclick = () => {
            input && (input.value = s.full_name);
            list.classList.add('hidden');
            startAnalysis(s.full_name);
            if (s.lat && s.lon && map) {
                map.setView([s.lat, s.lon], 15);
                DG.popup().setLatLng([s.lat, s.lon]).setContent(s.short_name).openOn(map);
            }
        };
        list.appendChild(div);
    });
    list.classList.remove('hidden');
}

searchBtn?.addEventListener('click', () => {
    const q = input?.value.trim() || '';
    if (q.length < 3) return alert('Введите название');
    startAnalysis(q);
});

function showReviews(title, list) {
    const main = $('main-container');
    const side = $('side-reviews-panel');
    $('side-panel-title') && ($('side-panel-title').textContent = title);
    $('side-panel-content') && ($('side-panel-content').innerHTML = list.map(t => `<div style="padding:15px; border-bottom:1px solid #f1f5f9; line-height:1.5">${t}</div>`).join(''));
    main?.classList.add('reviews-active-layout');
    side?.classList.remove('hidden');
}

function closeReviews() {
    $('main-container')?.classList.remove('reviews-active-layout');
    $('side-reviews-panel')?.classList.add('hidden');
}

async function startAnalysis(query) {
    if (!query) return;
    input && (input.value = query);
    searchBtn && (searchBtn.disabled = true);
    view?.classList.remove('hidden');
    statusMsg && (statusMsg.textContent = '⏳ Идет обработка данных...');

    charts.forEach(c => c.destroy());
    charts = [];

    const { ok, data } = await apiPost('/analyze', { query }).catch(() => ({ ok: false, data: null }));
    if (!ok || !data) {
        statusMsg && (statusMsg.textContent = '❌ Ошибка при загрузке данных');
        searchBtn && (searchBtn.disabled = false);
        return;
    }

    title && (title.textContent = data.school_name);
    statusMsg && (statusMsg.textContent = '✅ Аналитика готова');
    if (text) {
        text.innerHTML = `
            <div style="font-weight:850; font-size:1.2rem; margin-bottom:1.5rem">Сводка: ${data.stats.total} отзывов</div>
            <div style="display:grid; grid-template-columns:repeat(3,1fr); gap:15px">
                <div style="background:${COLORS.primaryBg}; padding:15px; border-radius:15px; border:1px solid rgba(79, 70, 229, 0.25); color:${COLORS.primary}; opacity:0.95">
                    <div style="font-size:1.5rem; font-weight:900">${data.stats.positive}</div>
                    <small style="font-weight:700">ПОЗИТИВ</small>
                </div>
                <div style="background:${COLORS.negativeBg}; padding:15px; border-radius:15px; border:1px solid rgba(100, 116, 139, 0.25); color:${COLORS.negative}; opacity:0.95">
                    <div style="font-size:1.5rem; font-weight:900">${data.stats.negative}</div>
                    <small style="font-weight:700">НЕГАТИВ</small>
                </div>
                <div style="background:#f8fafc; padding:15px; border-radius:15px; border:1px solid #f1f5f9; color:#475569; opacity:0.9">
                    <div style="font-size:1.5rem; font-weight:900">${data.stats.neutral}</div>
                    <small style="font-weight:700">НЕЙТРАЛ</small>
                </div>
            </div>`;
    }

    const getCount = (v) => Array.isArray(v) ? v.length : (v && typeof v === 'object' && typeof v.count === 'number') ? v.count : 0;
    const getExamples = (v) => Array.isArray(v) ? v : (v && typeof v === 'object' && Array.isArray(v.examples)) ? v.examples : [];

    const mkStacked = (el, labels, posValues, negValues, indexAxis) => new Chart(el, {
        type: 'bar',
        data: {
            labels,
            datasets: [
                { label: 'Позитив', data: posValues, backgroundColor: 'rgba(79, 70, 229, 0.9)', borderColor: COLORS.primary, borderWidth: 1 },
                { label: 'Негатив', data: negValues, backgroundColor: 'rgba(100, 116, 139, 0.9)', borderColor: COLORS.negative, borderWidth: 1 }
            ]
        },
        options: {
            indexAxis,
            responsive: true,
            maintainAspectRatio: true,
            plugins: { legend: { display: true, position: 'top', labels: { font: FONT, color: '#111827' } } },
            scales: { x: { stacked: true, grid: { display: false }, ticks: TICKS }, y: { stacked: true, grid: { display: false }, ticks: TICKS } }
        }
    });

    const mkBar = (el, labels, values, indexAxis, onClick) => new Chart(el, {
        type: 'bar',
        data: { labels, datasets: [{ data: values, backgroundColor: 'rgba(79, 70, 229, 0.9)', borderColor: COLORS.primary, borderWidth: 0, borderRadius: 12 }] },
        options: {
            indexAxis,
            responsive: true,
            maintainAspectRatio: true,
            plugins: { legend: { display: false } },
            onClick,
            scales: { y: { grid: { display: false }, ticks: TICKS }, x: { grid: { display: false }, ticks: TICKS } }
        }
    });

    (data.analytics || []).forEach((item, i) => {
        const el = $(`chart-${i}`);
        if (!el) return;
        if (item.type === 'stackedBar') {
            const isSeason = item.name === 'Сезонность активности';
            const labels = isSeason ? item.payload.map(p => p.label) : item.payload.map(p => p.category);
            const posValues = item.payload.map(p => p.pos);
            const negValues = item.payload.map(p => p.neg);
            charts.push(mkStacked(el, labels, posValues, negValues, isSeason ? 'x' : 'y'));
            return;
        }

        const rawLabels = Object.keys(item.payload || {});
        const rawValues = Object.values(item.payload || {});
        const labels = rawLabels.map(l => (typeof l === 'string' && l.length > 12) ? l.split(' ') : l);
        const values = rawValues.map(getCount);
        const isHoriz = (i === 0 || i === 1);
        charts.push(mkBar(el, labels, values, isHoriz ? 'y' : 'x', (_e, elements) => {
            if (!elements?.length) return;
            const idx = elements[0].index;
            const theme = rawLabels[idx];
            const source = rawValues[idx];
            showReviews(`${theme} (${getCount(source)} отзывов)`, getExamples(source));
        }));
    });

    searchBtn && (searchBtn.disabled = false);
}
