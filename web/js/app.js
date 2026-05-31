/* ============================================
   模板项目前端 — 零依赖
   ============================================ */

const $userArea = document.getElementById('user-area');
const $btnLogin = document.getElementById('btn-login');
const $toast    = document.getElementById('toast');
const $main     = document.getElementById('main');

let toastTimer = null;

// ---------- Toast ----------
function showToast(msg, type) {
    type = type || 'info';
    $toast.textContent = msg;
    $toast.className = 'toast ' + type + ' show';
    clearTimeout(toastTimer);
    toastTimer = setTimeout(function () { $toast.classList.remove('show'); }, 2800);
}

// ---------- 退出 ----------
function logout() {
    window.location.href = '/logout';
}

// ---------- 初始化 ----------
(async function init() {
    // 检查登录状态
    try {
        var res = await fetch('/status', { credentials: 'include' });
        var data = await res.json();
        if (data.data && data.data.logged_in) {
            // 已登录 — 显示用户信息和退出
            $userArea.innerHTML =
                '<span>' + escHtml(data.data.username) + '</span>' +
                '<span class="role-badge">' + escHtml(data.data.role) + '</span>' +
                '<button class="btn btn-link btn-sm" onclick="logout()">退出</button>';
            $btnLogin.style.display = 'none';
        }
    } catch (_) {
        // ignore
    }
})();

// ---------- 安全转义 ----------
function escHtml(s) {
    return String(s).replace(/&/g, '&amp;').replace(/</g, '&lt;')
                    .replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}
