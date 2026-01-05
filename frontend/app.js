const input = document.getElementById('school-search'),
    list = document.getElementById('suggestions'),
    view = document.getElementById('analytics-view'),
    title = document.getElementById('school-title'),
    text = document.getElementById('analytics-text'),
    searchBtn = document.getElementById('search-btn'),
    statusMsg = document.getElementById('status-msg');

let timer;

searchBtn.onclick = () => {
    const q = input.value.trim();
    if (q.length < 3) return alert('Введите название объекта для анализа');
    list.classList.add('hidden');
    startAnalysis(q);
};

input.onkeypress = (e) => {
    if (e.key === 'Enter') searchBtn.click();
};
