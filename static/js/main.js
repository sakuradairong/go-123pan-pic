let apiToken = localStorage.getItem('pan_api_token') || '';

// 代理 fetch 请求注入凭证
const originalFetch = window.fetch;
window.fetch = async function () {
    let [resource, config] = arguments;
    if (!config) config = {};
    if (!config.headers) config.headers = {};
    
    if (apiToken) {
        config.headers['Authorization'] = `Bearer ${apiToken}`;
    }
    
    const response = await originalFetch(resource, config);
    if (response.status === 403) {
        showAuthModal();
    }
    return response;
};

// 身份验证逻辑
function showAuthModal() {
    document.getElementById('auth-modal').style.display = 'flex';
}

function saveAuth() {
    const val = document.getElementById('auth-input').value;
    if (val) {
        localStorage.setItem('pan_api_token', val);
        apiToken = val; 
        document.getElementById('auth-modal').style.display = 'none';
        showToast('身份验证成功', 'success');
        document.getElementById('auth-input').value = ''; 
        loadImages(); 
    } else {
        showToast('Token 不能为空', 'error');
    }
}

document.addEventListener('DOMContentLoaded', () => {
    initDragAndDrop();
    loadImages();
});

// Toast 全局提示
function showToast(message, type = 'info') {
    const container = document.getElementById('toast-container');
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.textContent = message;
    
    container.appendChild(toast);
    
    setTimeout(() => {
        toast.style.animation = 'slideOutRight 0.3s ease forwards';
        setTimeout(() => toast.remove(), 300);
    }, 3000);
}

// 拖拽与剪贴板处理
function initDragAndDrop() {
    const dropZone = document.getElementById('drop-zone');
    const fileInput = document.getElementById('file-input');

    dropZone.addEventListener('click', () => fileInput.click());

    fileInput.addEventListener('change', (e) => {
        handleFiles(e.target.files);
        e.target.value = '';
    });

    ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
        dropZone.addEventListener(eventName, preventDefaults, false);
    });

    ['dragenter', 'dragover'].forEach(eventName => {
        dropZone.addEventListener(eventName, () => dropZone.classList.add('drag-over'), false);
    });

    ['dragleave', 'drop'].forEach(eventName => {
        dropZone.addEventListener(eventName, () => dropZone.classList.remove('drag-over'), false);
    });

    dropZone.addEventListener('drop', (e) => {
        const dt = e.dataTransfer;
        const files = dt.files;
        handleFiles(files);
    }, false);

    // 全局剪贴板监听
    document.addEventListener('paste', (e) => {
        if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') return;

        const files = (e.clipboardData || window.clipboardData).files;
        if (files && files.length > 0) {
            e.preventDefault();
            handleFiles(files);
            showToast('正在上传剪贴板图片...', 'info');
        }
    });
}

function preventDefaults(e) {
    e.preventDefault();
    e.stopPropagation();
}

function handleFiles(files) {
    const validFiles = Array.from(files).filter(f => f.type.startsWith('image/'));
    
    if (validFiles.length === 0) {
        showToast('仅支持上传图片格式', 'error');
        return;
    }

    validFiles.forEach(uploadFile);
}

// 执行上传流程
async function uploadFile(file) {
    const queue = document.getElementById('upload-queue');
    const item = document.createElement('div');
    item.className = 'queue-item';
    const span = document.createElement('span');
    span.textContent = '正在上传: ' + file.name;
    item.appendChild(span);
    const loader = document.createElement('div');
    loader.className = 'loader';
    item.appendChild(loader);
    queue.appendChild(item);

    const formData = new FormData();
    formData.append('file', file);

    try {
        const response = await fetch('/api/upload', {
            method: 'POST',
            body: formData
        });

        const result = await response.json();
        
        item.style.animation = 'slideOutRight 0.3s forwards';
        setTimeout(() => item.remove(), 300);

        if (result.code === 0 || result.code === 200) {
            showToast(`上传成功: ${file.name}`, 'success');
            loadImages(); 
        } else {
            showToast(`上传失败: ${result.message || '未知错误'}`, 'error');
        }
    } catch (err) {
        item.remove();
        showToast(`网络错误: ${err.message}`, 'error');
    }
}

// 画廊视图获取
async function loadImages() {
    try {
        const response = await fetch('/api/images');
        if (response.status !== 200) {
            if (response.status === 403) {
                document.getElementById('stats-counter').textContent = '未授权验证';
            } else {
                showToast('获取图片列表失败 (' + response.status + ')', 'error');
                document.getElementById('stats-counter').textContent = '获取失败';
            }
        }

        const result = await response.json();
        
        if (result.code === 0 || result.code === 200) {
            renderGallery(result.data || []);
            document.getElementById('stats-counter').textContent = `共 ${(result.data || []).length} 张图片`;
        } else {
            showToast(`加载失败: ${result.message}`, 'error');
            document.getElementById('stats-counter').textContent = '获取失败';
        }
    } catch (err) {
        showToast('网络连接异常', 'error');
        document.getElementById('stats-counter').textContent = '离线';
    }
}

// 侧信道XSS安全防备渲染
function renderGallery(images) {
    const grid = document.getElementById('gallery-grid');
    grid.innerHTML = '';

    if (images.length === 0) {
        const emptyMsg = document.createElement('span');
        emptyMsg.style.color = 'var(--text-muted)';
        emptyMsg.style.fontSize = '14px';
        emptyMsg.textContent = '暂无图片';
        grid.appendChild(emptyMsg);
        return;
    }

    images.forEach(img => {
        const kbSize = (img.size / 1024).toFixed(1);
        const finalUrl = img.url || img.origin_url || '';
        const safeName = img.name || '';
        const createdAt = img.created_at || '';

        const card = document.createElement('div');
        card.className = 'img-card';
        card.setAttribute('role', 'figure');

        const imgEl = document.createElement('img');
        imgEl.src = finalUrl;
        imgEl.alt = safeName;
        imgEl.loading = 'lazy';
        imgEl.onerror = function () {
            this.src = "data:image/svg+xml,%3Csvg xmlns=%27http://www.w3.org/2000/svg%27 viewBox=%270 0 24 24%27 fill=%27%23ef4444%27%3E%3Cpath d=%27M12 22C6.477 22 2 17.523 2 12S6.477 2 12 2s10 4.477 10 10-4.477 10-10 10zm-1-7v2h2v-2h-2zm0-8v6h2V7h-2z%27/%3E%3C/svg%3E";
        };
        card.appendChild(imgEl);

        const overlay = document.createElement('div');
        overlay.className = 'card-overlay';

        const info = document.createElement('div');
        info.className = 'card-info';
        const nameStrong = document.createElement('strong');
        nameStrong.textContent = safeName;
        info.appendChild(nameStrong);
        info.appendChild(document.createElement('br'));
        let dateStr = kbSize + ' KB';
        if (createdAt) {
            const d = new Date(createdAt);
            dateStr += ' • ' + (isNaN(d.getTime()) ? createdAt : d.toLocaleDateString());
        }
        info.appendChild(document.createTextNode(dateStr));
        overlay.appendChild(info);

        const actions = document.createElement('div');
        actions.className = 'card-actions';

        const copyBtn = document.createElement('button');
        copyBtn.className = 'btn-action btn-copy';
        copyBtn.textContent = '复制链接';
        copyBtn.addEventListener('click', function () { copyUrl(finalUrl); });
        actions.appendChild(copyBtn);

        const delBtn = document.createElement('button');
        delBtn.className = 'btn-action btn-del';
        delBtn.textContent = '删除';
        delBtn.addEventListener('click', function () { deleteImage(img.id, safeName); });
        actions.appendChild(delBtn);

        overlay.appendChild(actions);
        card.appendChild(overlay);
        grid.appendChild(card);
    });
}

function copyUrl(url) {
    navigator.clipboard.writeText(url).then(() => {
        showToast('直链已复制到剪贴板', 'success');
    }).catch(err => {
        showToast('复制失败，请手动选取', 'error');
    });
}

async function deleteImage(id, name) {
    if (!confirm(`确定要永久删除图片 [${name}] 吗？`)) return;

    try {
        const response = await fetch(`/api/images/${id}`, { method: 'DELETE' });
        if (response.status !== 200) return; 

        const result = await response.json();
        
        if (result.code === 0 || result.code === 200) {
            showToast(`[${name}] 已删除`, 'success');
            loadImages();
        } else {
            showToast(`删除失败: ${result.message}`, 'error');
        }
    } catch (err) {
        showToast('删除网络请求失败', 'error');
    }
}
