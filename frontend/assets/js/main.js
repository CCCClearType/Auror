// ============================================================
// AurorNote — main.js
// 全域狀態與共用元件
// ============================================================

// ============================================================
// 頁面初始化
// ============================================================
document.documentElement.setAttribute('data-theme', 'dark');

document.addEventListener("DOMContentLoaded", () => {
    renderHeader();

    // 搜尋框事件（首頁）
    const searchForm = document.getElementById('search-form');
    if (searchForm) {
        searchForm.addEventListener('submit', (e) => {
            e.preventDefault();
            const keyword = document.getElementById('search-input').value.trim();
            if (keyword) {
                window.location.href = `/pages/store/search.html?q=${encodeURIComponent(keyword)}`;
            }
        });
    }
});

document.addEventListener('click', (event) => {
    const burger = event.target.closest('.navbar-burger');
    if (!burger) return;
    const target = document.getElementById(burger.dataset.target);
    if (!target) return;
    burger.classList.toggle('is-active');
    target.classList.toggle('is-active');
});

// ============================================================
// 渲染共用導覽列
// ============================================================
function renderHeader() {
    const header = document.getElementById('global-header');
    if (!header) return;

    // 登入頁與註冊頁只顯示 Logo
    const path = window.location.pathname;
    if (path.includes('login.html') || path.includes('register.html')) {
        header.innerHTML = `
            <nav class="navbar va-navbar" role="navigation" aria-label="main navigation">
                <div class="navbar-brand">
                    <a href="/" class="navbar-item logo">AurorNote</a>
                </div>
            </nav>`;
        return;
    }

    const currentRole = localStorage.getItem('userRole') || 'GUEST';
    const userDataStr = localStorage.getItem('currentUser');
    const username = userDataStr ? JSON.parse(userDataStr).username : '買家';

    let navItems = `
        <a class="navbar-item" href="/">商店首頁</a>
        <a class="navbar-item" href="/pages/is_ilearn_down.html" style="color: var(--accent-blue);">iLearn Status</a>
    `;

    if (currentRole !== 'GUEST') {
        navItems += `
            <a class="navbar-item" href="/pages/user/library">筆記庫</a>
            <a class="navbar-item" href="/pages/user/social">社群</a>
        `;
        if (currentRole === 'SELLER') {
            navItems += `<a class="navbar-item has-text-warning" href="/pages/dashboard/seller_dashboard">賣家中心</a>`;
        } else if (currentRole === 'CSR') {
            navItems += `<a class="navbar-item has-text-info" href="/pages/dashboard/csr_dashboard">客服中心</a>`;
        } else if (currentRole === 'ADMIN') {
            navItems += `<a class="navbar-item has-text-info" href="/pages/dashboard/csr_dashboard">客服中心</a>`;
            navItems += `<a class="navbar-item has-text-danger" href="/pages/dashboard/admin_dashboard">管理後台</a>`;
        }
    }

    let accountHtml = '';
    if (currentRole === 'GUEST') {
        accountHtml += `
            <div class="navbar-item">
                <div class="buttons">
                    <a href="/pages/auth/login" class="button is-primary"><strong>登入 / 註冊</strong></a>
                </div>
            </div>`;
    } else {
        accountHtml += `
            <a href="/pages/user/cart" class="navbar-item">購物車</a>
            <div class="navbar-item has-dropdown is-hoverable">
                <a class="navbar-link">${escapeHtml(username)}</a>
                <div class="navbar-dropdown is-right">
                    <a class="navbar-item" href="/pages/user/profile">基本資料</a>
                    <a class="navbar-item" href="/pages/user/history">購買紀錄</a>
                    <a class="navbar-item" href="/pages/user/refund_request">退款申請</a>
                    <a class="navbar-item" href="/pages/user/wishlist">願望清單</a>
                    <hr class="navbar-divider">
                    <a class="navbar-item" href="#" id="logout-btn">登出</a>
                </div>
            </div>
        `;
    }

    header.innerHTML = `
        <nav class="navbar va-navbar" role="navigation" aria-label="main navigation">
            <div class="navbar-brand">
                <a href="/" class="navbar-item logo">AurorNote</a>
                <a role="button" class="navbar-burger" aria-label="menu" aria-expanded="false" data-target="global-navbar-menu">
                    <span aria-hidden="true"></span>
                    <span aria-hidden="true"></span>
                    <span aria-hidden="true"></span>
                    <span aria-hidden="true"></span>
                </a>
            </div>
            <div id="global-navbar-menu" class="navbar-menu">
                <div class="navbar-start">${navItems}</div>
                <div class="navbar-end">${accountHtml}</div>
            </div>
        </nav>`;

    // 登出事件
    const logoutBtn = document.getElementById('logout-btn');
    if (logoutBtn) {
        logoutBtn.addEventListener('click', async (e) => {
            e.preventDefault();
            try {
                await apiLogout();
            } catch (_) {
                // 即使 API 失敗也清除本地狀態
                localStorage.removeItem('token');
                localStorage.removeItem('currentUser');
                localStorage.removeItem('userRole');
            }
            window.location.href = '/';
        });
    }
}

// ============================================================
// 渲染筆記卡片列表
// ============================================================
function renderNotes(notes) {
    const container = document.getElementById('note-list');
    if (!container) return;

    if (!notes) {
        // Skeleton Loading 骨架屏狀態
        container.innerHTML = '';
        for (let i = 0; i < 4; i++) {
            const skeleton = document.createElement('div');
            skeleton.className = 'note-list-card card';
            skeleton.innerHTML = `
                <div class="note-thumbnail skeleton"></div>
                <div class="note-list-info">
                    <div class="skeleton skeleton-title"></div>
                    <div class="skeleton skeleton-text"></div>
                    <div class="skeleton skeleton-text" style="width: 80%;"></div>
                    <div class="note-list-tags" style="margin-top: 10px;">
                        <span class="tag skeleton" style="width: 40px; height: 16px;"></span>
                        <span class="tag skeleton" style="width: 50px; height: 16px;"></span>
                    </div>
                </div>
            `;
            container.appendChild(skeleton);
        }
        return;
    }

    if (notes.length === 0) {
        container.innerHTML = `
            <div class="empty-state">
                <p>找不到符合條件的筆記</p>
                <a href="/" class="button is-light" style="margin-top:10px;">回首頁</a>
            </div>`;
        return;
    }

    container.innerHTML = '';

    const groupSelect = document.getElementById('filter-group');
    const groupBy = groupSelect ? groupSelect.value : '';

    const createNoteCard = (note) => {
        const card = document.createElement('div');
        card.className = 'note-list-card card';

        const groupedTags = {
            SUBJECT: [],
            SEMESTER: [],
            DEPARTMENT: [],
            COURSE_TYPE: [],
            TEACHER: [],
            GENERAL: []
        };
        (note.tags || []).forEach(t => {
            const type = t.tag_type || 'GENERAL';
            if (groupedTags[type]) {
                groupedTags[type].push(t);
            } else {
                groupedTags['GENERAL'].push(t);
            }
        });

        let tagsHtml = '';
        const order = ['SUBJECT', 'SEMESTER', 'DEPARTMENT', 'COURSE_TYPE', 'TEACHER', 'GENERAL'];
        order.forEach(type => {
            if (groupedTags[type].length > 0) {
                tagsHtml += `<div style="display: flex; gap: 4px; flex-wrap: wrap; margin-bottom: 6px;">`;
                groupedTags[type].forEach(t => {
                    const tagName = typeof t === 'string' ? t : (t.tag_name || t.name || t);
                    tagsHtml += `<span class="tag is-rounded type-${type}" style="border: none;">${escapeHtml(String(tagName))}</span>`;
                });
                tagsHtml += `</div>`;
            }
        });
        const priceHtml = note.price === 0
            ? `<span class="note-list-price free">免費</span>`
            : `<span class="note-list-price">NT$ ${note.price.toLocaleString()}</span>`;

        let coverHtml = `<span style="font-size:13px;color:#8f98a0;">[${escapeHtml(note.title)} 封面]</span>`;
        if (note.media && note.media.length > 0) {
            const thumb = note.media.find(m => m.media_type === 'thumbnail') || note.media.find(m => m.media_type === 'media');
            if (thumb) {
                coverHtml = `<img src="${thumb.file_url}" alt="cover" style="width:100%; height:100%; object-fit:cover; border-radius: 8px;">`;
            }
        }

        let classicBadgeHtml = '';
        if (note.is_classic) {
            classicBadgeHtml = `<div style="position:absolute; top:8px; right:8px; background:linear-gradient(45deg, #ffd700, #ff8c00); color:#000; font-weight:bold; padding:4px 8px; border-radius:12px; font-size:12px; box-shadow:0 0 10px rgba(255,215,0,0.5); z-index:10;"><i class="fas fa-crown"></i> 傳世經典</div>`;
        }

        card.innerHTML = `
            <div class="card-image note-thumbnail" style="position:relative;">
                ${classicBadgeHtml}
                ${coverHtml}
            </div>
            <div class="card-content note-list-info">
                <p class="title is-5 note-list-title">${escapeHtml(note.title)}</p>
                <p class="content note-list-desc">${escapeHtml(note.desc || '')}</p>
                <div class="note-list-tags tags">${tagsHtml}</div>
                <p>${priceHtml}</p>
            </div>
        `;

        card.addEventListener('click', () => {
            const noteId = note.note_id || note.id;
            window.location.href = `/pages/store/note_detail?id=${noteId}`;
        });
        return card;
    };

    // ── Group-by configuration ──────────────────────────────────
    const GROUP_META = {
        SEMESTER:    { label: '學期',  icon: 'fas fa-calendar-alt', color: '#66c0f4' },
        SUBJECT:     { label: '科目',  icon: 'fas fa-book',         color: '#a3e073' },
        DEPARTMENT:  { label: '系所',  icon: 'fas fa-university',   color: '#ffdd57' },
        COURSE_TYPE: { label: '屬性',  icon: 'fas fa-tags',         color: '#f14668' },
        TEACHER:     { label: '老師',  icon: 'fas fa-chalkboard-teacher', color: '#00d1b2' },
    };

    if (groupBy && GROUP_META[groupBy]) {
        const meta = GROUP_META[groupBy];
        const groups = {};
        const others = [];

        notes.forEach(note => {
            const matchTag = (note.tags || []).find(t => (t.tag_type || 'GENERAL') === groupBy);
            if (matchTag) {
                const name = typeof matchTag === 'string' ? matchTag : (matchTag.tag_name || matchTag.name || '');
                if (!groups[name]) groups[name] = [];
                groups[name].push(note);
            } else {
                others.push(note);
            }
        });

        // Sort group keys: semesters descending (e.g. 115-2 first), others ascending
        const sortedKeys = Object.keys(groups).sort((a, b) => {
            if (groupBy === 'SEMESTER') return b.localeCompare(a);
            return a.localeCompare(b);
        });

        sortedKeys.forEach(key => {
            const header = document.createElement('div');
            header.className = 'group-header';
            header.style.cssText = `
                grid-column: 1 / -1;
                display: flex;
                align-items: center;
                gap: 10px;
                padding: 12px 0 8px;
                margin-top: 16px;
                border-bottom: 2px solid ${meta.color}33;
            `;
            header.innerHTML = `
                <span style="
                    display: inline-flex;
                    align-items: center;
                    justify-content: center;
                    width: 32px;
                    height: 32px;
                    border-radius: 8px;
                    background: ${meta.color}22;
                    color: ${meta.color};
                    font-size: 14px;
                "><i class="${meta.icon}"></i></span>
                <span style="
                    font-size: 16px;
                    font-weight: 700;
                    color: ${meta.color};
                    letter-spacing: 0.5px;
                ">${meta.label}：${escapeHtml(key)}</span>
                <span class="tag is-rounded" style="
                    background: ${meta.color}22;
                    color: ${meta.color};
                    border: 1px solid ${meta.color}44;
                    font-size: 11px;
                    font-weight: 600;
                ">${groups[key].length} 篇</span>
            `;
            container.appendChild(header);

            groups[key].forEach(note => {
                container.appendChild(createNoteCard(note));
            });
        });

        if (others.length > 0) {
            const header = document.createElement('div');
            header.className = 'group-header';
            header.style.cssText = `
                grid-column: 1 / -1;
                display: flex;
                align-items: center;
                gap: 10px;
                padding: 12px 0 8px;
                margin-top: 16px;
                border-bottom: 2px solid rgba(255,255,255,0.08);
            `;
            header.innerHTML = `
                <span style="
                    display: inline-flex;
                    align-items: center;
                    justify-content: center;
                    width: 32px;
                    height: 32px;
                    border-radius: 8px;
                    background: rgba(255,255,255,0.06);
                    color: #888;
                    font-size: 14px;
                "><i class="fas fa-box-open"></i></span>
                <span style="
                    font-size: 16px;
                    font-weight: 700;
                    color: #888;
                    letter-spacing: 0.5px;
                ">未標記${meta.label}</span>
                <span class="tag is-rounded" style="
                    background: rgba(255,255,255,0.06);
                    color: #888;
                    border: 1px solid rgba(255,255,255,0.1);
                    font-size: 11px;
                    font-weight: 600;
                ">${others.length} 篇</span>
            `;
            container.appendChild(header);

            others.forEach(note => {
                container.appendChild(createNoteCard(note));
            });
        }
    } else {
        notes.forEach(note => {
            container.appendChild(createNoteCard(note));
        });
    }
}

// ============================================================
// 工具函式
// ============================================================
function escapeHtml(str) {
    if (typeof str !== 'string') return '';
    return str
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;');
}

function getCurrentUser() {
    const str = localStorage.getItem('currentUser');
    return str ? JSON.parse(str) : null;
}

function getCurrentRole() {
    return localStorage.getItem('userRole') || 'GUEST';
}

function getNoteCoverUrl(note) {
    const media = (note && note.media) || (note && note.Media) || [];
    const image = media.find(m => m.media_type === 'thumbnail') || media.find(m => m.media_type === 'media');
    return image ? image.file_url : '';
}

function renderMarkdown(markdown) {
    const source = String(markdown || '').trim();
    if (!source) return '<p class="has-text-grey">尚未填寫介紹</p>';
    if (!window.marked || !window.DOMPurify) {
        return `<p>${escapeHtml(source).replace(/\n/g, '<br>')}</p>`;
    }
    return DOMPurify.sanitize(marked.parse(source));
}

function requireLogin(redirectTo = '/pages/auth/login.html') {
    if (!localStorage.getItem('token')) {
        alert('請先登入！');
        window.location.href = redirectTo;
        return false;
    }
    return true;
}

// ============================================================
// 全域 Toast Notification 系統
// type: 'info', 'success', 'error'
// ============================================================
function showToast(message, type = 'info') {
    let container = document.getElementById('toast-container');
    if (!container) {
        container = document.createElement('div');
        container.id = 'toast-container';
        document.body.appendChild(container);
    }

    const toast = document.createElement('div');
    toast.className = `toast-message toast-${type}`;
    
    toast.innerHTML = `<span>${escapeHtml(message)}</span>`;
    container.appendChild(toast);

    // Animate in
    setTimeout(() => toast.classList.add('show'), 10);

    // Auto remove after 3s
    setTimeout(() => {
        toast.classList.remove('show');
        setTimeout(() => toast.remove(), 400); // Wait for transition
    }, 3000);
}
