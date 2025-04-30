const API_URL = 'http://localhost:8228/api';

function saveToken(token) {
  localStorage.setItem('token', token);
  document.querySelectorAll('.section').forEach(section => section.classList.add('hidden'));
  document.getElementById('calcSection').classList.remove('hidden');
  document.getElementById('createSection').classList.remove('hidden');
  document.getElementById('packagesSection').classList.remove('hidden');
  document.getElementById('logoutSection').classList.remove('hidden');
  getPackages();
}

function logout() {
  localStorage.removeItem('token');
  showToast('Вы вышли из аккаунта', 'success');
  document.querySelectorAll('.section').forEach(section => section.classList.add('hidden'));
  document.querySelectorAll('.section').forEach(section => {
    if (
      section.querySelector('#regEmail') ||
      section.querySelector('#loginEmail')
    ) {
      section.classList.remove('hidden');
    }
  });
}

function register() {
  const email = document.getElementById('regEmail').value;
  const password = document.getElementById('regPassword').value;

  fetch(`${API_URL}/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password })
  })
  .then(handleResponse)
  .then(data => {
    if (data.token) {
      saveToken(data.token);
      alert('Регистрация успешна!');
    }
  });
}

function login() {
  const email = document.getElementById('loginEmail').value;
  const password = document.getElementById('loginPassword').value;
  const button = event.target;
  toggleButtonLoading(button, true);

  fetch(`${API_URL}/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password })
  })
  .then(handleResponse)
  .then(data => {
    saveToken(data.token);
    showToast('Вход выполнен!', 'success');
  })
  .catch(err => showToast(`${err.message}`, 'error'))
  .finally(() => toggleButtonLoading(button, false));
}

function calculate() {
  const token = localStorage.getItem('token');
  const weight = parseFloat(document.getElementById('weight').value);
  const from = document.getElementById('from').value;
  const to = document.getElementById('to').value;
  const address = document.getElementById('address').value;

  fetch(`${API_URL}/calculate`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({ weight, from, to, address })
  })
  .then(handleResponse)
  .then(data => {
    document.getElementById('calcResult').innerHTML = `
      <div class="package-card">
        <p><strong>💰 Стоимость:</strong> ${data.cost} ${data.currency}</p>
        <p><strong>⏱ Время доставки:</strong> ${data.estimated_hours} ч</p>
      </div>
    `;
  })
  .catch(() => document.getElementById('calcResult').innerHTML = '');
}

function createPackage() {
  const token = localStorage.getItem('token');
  const weight = parseFloat(document.getElementById('weight').value);
  const from = document.getElementById('from').value;
  const to = document.getElementById('to').value;
  const address = document.getElementById('address').value;

  fetch(`${API_URL}/create`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({ weight, from, to, address })
  })
  .then(handleResponse)
  .then(data => {
    document.getElementById('createResult').innerHTML = `
      <div class="package-card">
        <p><strong>📦 ID:</strong> ${data.package_id}</p>
        <p><strong>📍 Статус:</strong> ${data.status}</p>
        <p><strong>💰 Стоимость:</strong> ${data.cost} ${data.currency}</p>
        <p><strong>⏱ Время доставки:</strong> ${data.estimated_hours} ч</p>
      </div>
    `;
  })
  .catch(() => document.getElementById('createResult').innerHTML = '');
}

function getPackages() {
  const token = localStorage.getItem('token');

  fetch(`${API_URL}/my/packages`, {
    method: 'GET',
    headers: { 'Authorization': `Bearer ${token}` }
  })
  .then(handleResponse)
  .then(data => {
    const container = document.getElementById('packagesResult');
    container.innerHTML = '';

    if (!data.length) {
      container.innerHTML = '<p>Нет посылок</p>';
      return;
    }

    data.forEach(pkg => {
      const card = document.createElement('div');
      const isPaid = pkg.payment_status === 'PAID';
      card.className = 'package-card ' + (isPaid ? 'paid' : 'unpaid');
      card.innerHTML = `
        <h4>📦 ${pkg.from} → ${pkg.to}</h4>
        <p><strong>Адрес:</strong> ${pkg.address}</p>
        <p><strong>Вес:</strong> ${pkg.weight} кг</p>
        <p><strong>Стоимость:</strong> ${pkg.cost} ${pkg.currency}</p>
        <p><strong>Ожидаемое время:</strong> ${pkg.estimated_hours} ч</p>
        <p><strong>Статус:</strong> ${pkg.status}</p>
        <p><strong>Оплата:</strong> ${isPaid ? 'Оплачено 💸' : 'Не оплачено ❌'}</p>
        ${!isPaid ? `<button onclick="payForPackage('${pkg.package_id}')">💳 Оплатить</button>` : ''}
      `;
      container.appendChild(card);
    });
  });
}

function payForPackage(packageId) {
  const token = localStorage.getItem('token');

  fetch(`${API_URL}/payment/${packageId}`, {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${token}` }
  })
  .then(async res => {
    const contentType = res.headers.get("content-type");

    if (!res.ok) {
      const errorText = contentType.includes("application/json") ? (await res.json()).message : await res.text();
      throw new Error(errorText || "Ошибка оплаты");
    }

    const message = contentType.includes("application/json") ? (await res.json()).message : await res.text();
    showToast(message, 'success');
    getPackages();
  })
  .catch(err => showToast(`❌ Не удалось оплатить: ${err.message}`, 'error'));
}

function handleResponse(res) {
  return res.json().then(data => {
    if (!res.ok) {
      const errorMsg = data.message || 'Произошла ошибка';
      showToast(`❌ ${errorMsg}`, 'error');
      throw new Error(errorMsg);
    }
    return data;
  }).catch(err => {
    throw err;
  });
}

function showToast(message, type = 'success') {
  const toastContainer = document.getElementById('toast-container');
  const toast = document.createElement('div');
  toast.className = `toast ${type}`;
  toast.textContent = message;
  toastContainer.appendChild(toast);
  setTimeout(() => toast.remove(), 3000);
}

function toggleButtonLoading(button, loading) {
  if (loading) button.classList.add('loading');
  else button.classList.remove('loading');
}

document.addEventListener('DOMContentLoaded', () => {
  const token = localStorage.getItem('token');
  if (token) {
    document.querySelectorAll('.section').forEach(section => section.classList.add('hidden'));
    document.getElementById('calcSection').classList.remove('hidden');
    document.getElementById('createSection').classList.remove('hidden');
    document.getElementById('packagesSection').classList.remove('hidden');
    document.getElementById('logoutSection').classList.remove('hidden');
    getPackages();
  }
});
