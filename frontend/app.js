const input = document.getElementById('school-search'),
    list = document.getElementById('suggestions'),
    view = document.getElementById('analytics-view'),
    title = document.getElementById('school-title'),
    text = document.getElementById('analytics-text'),
    searchBtn = document.getElementById('search-btn'),
    statusMsg = document.getElementById('status-msg');

let timer;

input.oninput = (e) => {
    clearTimeout(timer);
    const q = e.target.value.trim();
    if (q.length < 2) return list.classList.add('hidden');

    timer = setTimeout(async () => {
        try {
            const res = await fetch(`http://localhost:8081/schools?q=${encodeURIComponent(q)}`);
            const data = await res.json();
            renderSuggestions(data ? data.slice(0, 5) : []);
        } catch (err) {
            console.error(err);
        }
    }, 300);
};

function renderSuggestions(data) {
    list.innerHTML = data.length ? '' : '<div class="suggestion-item">Объект не найден в базе</div>';
    data.forEach(s => {
        const div = document.createElement('div');
        div.className = 'suggestion-item';
        div.textContent = s.full_name;
        div.onclick = () => {
            input.value = s.full_name;
            list.classList.add('hidden');
            startAnalysis(s.full_name);
        };
        list.appendChild(div);
    });
    list.classList.remove('hidden');
}

document.onclick = (e) => {
    if (!e.target.closest('.search-wrapper')) {
        list.classList.add('hidden');
    }
};

searchBtn.onclick = () => {
    const q = input.value.trim();
    if (q.length < 3) return alert('Введите название объекта для анализа');
    list.classList.add('hidden');
    startAnalysis(q);
};

async function startAnalysis(query) {
    searchBtn.disabled = true;
    view.classList.remove('hidden');
    title.textContent = query;
    statusMsg.textContent = '⏳ Поиск в базе и сбор отзывов...';
    text.textContent = 'Это может занять от 30 до 60 секунд, если отзывов еще нет в нашей системе.';

    try {
        const res = await fetch('http://localhost:8081/analyze', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ query: query })
        });

        const data = await res.json();

        if (res.ok) {
            title.textContent = data.school_name;
            statusMsg.textContent = '✅ Анализ успешно завершен';

            let html = `
                <div style="font-size: 1.2rem; margin-bottom: 1.5rem; font-weight: 600;">
                    Общее количество отзывов: ${data.stats.total}
                </div>
                <div style="display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 1rem; text-align: center; margin-bottom: 2rem;">
                    <div style="padding: 1rem; background: #ecfdf5; border-radius: 12px;">
                        <div style="color: #059669; font-size: 1.5rem; font-weight: 800;">${data.stats.positive}</div>
                        <div style="color: #065f46; font-size: 0.9rem;">Положительных</div>
                    </div>
                    <div style="padding: 1rem; background: #fef2f2; border-radius: 12px;">
                        <div style="color: #dc2626; font-size: 1.5rem; font-weight: 800;">${data.stats.negative}</div>
                        <div style="color: #991b1b; font-size: 0.9rem;">Отрицательных</div>
                    </div>
                    <div style="padding: 1rem; background: #f8fafc; border-radius: 12px;">
                        <div style="color: #64748b; font-size: 1.5rem; font-weight: 800;">${data.stats.neutral}</div>
                        <div style="color: #334155; font-size: 0.9rem;">Нейтральных</div>
                    </div>
                </div>
            `;

            if (data.analytics) {
                data.analytics.forEach(item => {
                    html += `<h3 style="margin: 1.5rem 0 0.5rem; font-size: 1rem;">${item.name}</h3>`;
                    const entries = Object.entries(item.payload);
                    const max = Math.max(...entries.map(e => e[1])) || 1;

                    entries.forEach(([label, value]) => {
                        const width = (value / max) * 100;
                        html += `
                            <div style="margin-bottom: 0.5rem;">
                                <div style="display: flex; justify-content: space-between; font-size: 0.8rem; margin-bottom: 2px;">
                                    <span>${label}</span>
                                    <span>${typeof value === 'number' ? value.toFixed(1) : value}</span>
                                </div>
                                <div style="background: #f1f5f9; border-radius: 4px; height: 8px;">
                                    <div style="background: var(--primary); width: ${width}%; height: 100%; border-radius: 4px; transition: width 1s;"></div>
                                </div>
                            </div>
                        `;
                    });
                });
            }
            text.innerHTML = html;
        } else {
            statusMsg.textContent = 'Внимание';
            text.textContent = data.message || 'Объект не найден';
        }
    } catch (err) {
        statusMsg.textContent = 'Ошибка системы';
        text.textContent = 'Не удалось получить ответ от сервера';
        console.error(err);
    } finally {
        searchBtn.disabled = false;
    }
}
input.onkeypress = (e) => {
    if (e.key === 'Enter') searchBtn.click();
};
