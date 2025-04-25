const API_URL = 'http://localhost:8228/api';

function saveToken(token) {
  localStorage.setItem('token', token);

  document.querySelectorAll('.section').forEach(section => section.classList.add('hidden'));
  document.getElementById('calcSection').classList.remove('hidden');
  document.getElementById('createSection').classList.remove('hidden');
  document.getElementById('packagesSection').classList.remove('hidden');
}

function register() {
  const email = document.getElementById('regEmail').value;
  const password = document.getElementById('regPassword').value;

  fetch(`${API_URL}/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password })
  }).then(res => res.json()).then(data => {
    if (data.token) {
      saveToken(data.token);
      alert('Регистрация успешна!');
    }
  });
}

function login() {
  const email = document.getElementById('loginEmail').value;
  const password = document.getElementById('loginPassword').value;

  fetch(`${API_URL}/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password })
  }).then(res => res.json()).then(data => {
    if (data.token) {
      saveToken(data.token);
      alert('Вход выполнен!');
    }
  });
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
  .then(res => res.json())
  .then(data => {
    document.getElementById('calcResult').innerHTML = `
      <div class="package-card">
        <p><strong>💰 Стоимость:</strong> ${data.cost} ${data.currency}</p>
        <p><strong>⏱ Время доставки:</strong> ${data.estimated_hours} ч</p>
      </div>
    `;
  });
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
  .then(res => res.json())
  .then(data => {
    document.getElementById('createResult').innerHTML = `
      <div class="package-card">
        <p><strong>📦 ID:</strong> ${data.package_id}</p>
        <p><strong>📍 Статус:</strong> ${data.status}</p>
        <p><strong>💰 Стоимость:</strong> ${data.cost} ${data.currency}</p>
        <p><strong>⏱ Время доставки:</strong> ${data.estimated_hours} ч</p>
      </div>
    `;
  });
}

function getPackages() {
  const token = localStorage.getItem('token');

  fetch(`${API_URL}/my/packages`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${token}`
    }
  }).then(res => res.json()).then(data => {
    const container = document.getElementById('packagesResult');
    container.innerHTML = '';

    if (!data.length) {
      container.innerHTML = '<p>Нет посылок</p>';
      return;
    }

    data.forEach(pkg => {
      const card = document.createElement('div');
      card.className = 'package-card';
      card.innerHTML = `
        <h4>📦 ${pkg.from} → ${pkg.to}</h4>
        <p><strong>Адрес:</strong> ${pkg.address}</p>
        <p><strong>Вес:</strong> ${pkg.weight} кг</p>
        <p><strong>Стоимость:</strong> ${pkg.cost} ${pkg.currency}</p>
        <p><strong>Ожидаемое время:</strong> ${pkg.estimated_hours} ч</p>
        <p><strong>Статус:</strong> ${pkg.status}</p>
      `;
      container.appendChild(card);
    });
  });
}
