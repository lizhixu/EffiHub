// 主题管理
const themeToggle = document.getElementById('themeToggle');
const themeIcon = themeToggle.querySelector('.theme-icon');
const themes = ['auto', 'light', 'dark'];
const themeIcons = { auto: '🖥️', light: '☀️', dark: '🌙' };

function getSystemTheme() {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

function applyTheme(theme) {
    const actualTheme = theme === 'auto' ? getSystemTheme() : theme;
    document.documentElement.setAttribute('data-theme', actualTheme);
    themeIcon.textContent = themeIcons[theme];
}

// 初始化主题
let currentTheme = localStorage.getItem('theme') || 'auto';
applyTheme(currentTheme);

// 监听系统主题变化
window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
    if (localStorage.getItem('theme') === 'auto') {
        applyTheme('auto');
    }
});

// 切换主题: auto -> light -> dark -> auto
themeToggle.addEventListener('click', () => {
    const idx = themes.indexOf(currentTheme);
    currentTheme = themes[(idx + 1) % themes.length];
    localStorage.setItem('theme', currentTheme);
    applyTheme(currentTheme);
});

const API = '/api';

async function loadData() {
    try {
        const [catRes, linkRes] = await Promise.all([
            fetch(`${API}/categories`),
            fetch(`${API}/links`)
        ]);
        if (!catRes.ok || !linkRes.ok) throw new Error('API 错误');
        const categories = await catRes.json();
        const links = await linkRes.json();
        if (!categories || categories.length === 0) { showEmpty(); return; }
        render(categories, links || []);
    } catch (e) { showError(); }
}

function render(categories, links) {
    const main = document.getElementById('mainContent');
    const sidebar = document.getElementById('sidebarNav');

    sidebar.innerHTML = categories.map((c, i) => `
        <a href="#${c.slug}" class="nav-item${i === 0 ? ' active' : ''}" data-category="${c.slug}">
            <img class="nav-icon" src="${c.icon}" alt="${c.name}" onerror="this.style.display='none'">
            <span>${c.name}</span>
        </a>
    `).join('');

    main.innerHTML = categories.map(cat => {
        const catLinks = links.filter(l => l.category_id === cat.id);
        return `
            <section class="category" id="${cat.slug}" data-category="${cat.slug}">
                <h2 class="category-title">${cat.name}</h2>
                <div class="links-grid">
                    ${catLinks.length ? catLinks.map(l => `
                        <a href="${l.url}" target="_blank" class="link-item">
                            <img class="link-icon" src="${l.icon}" alt="${l.name}" onerror="this.src='data:image/svg+xml,<svg xmlns=%22http://www.w3.org/2000/svg%22 viewBox=%220 0 24 24%22 fill=%22%23666%22><rect width=%2224%22 height=%2224%22 rx=%224%22/></svg>'">
                            <div class="link-info">
                                <span class="link-name">${l.name}</span>
                                <span class="link-desc">${l.desc || ''}</span>
                            </div>
                        </a>
                    `).join('') : '<p class="empty-tip">暂无链接</p>'}
                </div>
            </section>
        `;
    }).join('');
    initEvents();
}

function showEmpty() {
    document.getElementById('sidebarNav').innerHTML = '<div class="loading-placeholder">暂无分类</div>';
    document.getElementById('mainContent').innerHTML = `<div class="empty-state"><p>🚀 还没有数据</p><p>请先到 <a href="/admin.html">后台管理</a> 添加分类和链接</p></div>`;
}

function showError() {
    document.getElementById('sidebarNav').innerHTML = '<div class="loading-placeholder">加载失败</div>';
    document.getElementById('mainContent').innerHTML = `<div class="empty-state"><p>😢 数据加载失败</p><p>请检查后端服务是否启动</p></div>`;
}

function initEvents() {
    const navItems = document.querySelectorAll('.nav-item');
    const categories = document.querySelectorAll('.category');
    const searchInput = document.getElementById('searchInput');
    const main = document.getElementById('mainContent');

    navItems.forEach(item => {
        item.addEventListener('click', e => {
            e.preventDefault();
            navItems.forEach(n => n.classList.remove('active'));
            item.classList.add('active');
            document.getElementById(item.getAttribute('href').slice(1))?.scrollIntoView({ behavior: 'smooth', block: 'start' });
        });
    });

    searchInput.addEventListener('input', e => {
        const q = e.target.value.toLowerCase().trim();
        document.querySelectorAll('.link-item').forEach(item => {
            const name = item.querySelector('.link-name')?.textContent.toLowerCase() || '';
            const desc = item.querySelector('.link-desc')?.textContent.toLowerCase() || '';
            const match = !q || name.includes(q) || desc.includes(q);
            item.classList.toggle('hidden', !match);
            item.classList.toggle('highlight', match && q);
        });
        categories.forEach(cat => {
            cat.classList.toggle('hidden', !cat.querySelectorAll('.link-item:not(.hidden)').length && q);
        });
    });

    document.addEventListener('keydown', e => {
        if ((e.metaKey || e.ctrlKey) && e.key === 'k') { e.preventDefault(); searchInput.focus(); }
        if (e.key === 'Escape') { searchInput.blur(); searchInput.value = ''; searchInput.dispatchEvent(new Event('input')); }
    });

    main.addEventListener('scroll', () => {
        let current = '';
        categories.forEach(cat => { if (cat.getBoundingClientRect().top <= 100) current = cat.id; });
        if (current) navItems.forEach(n => n.classList.toggle('active', n.getAttribute('href') === '#' + current));
    });
}

loadData();
