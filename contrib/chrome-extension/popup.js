let currentData = null;
let currentHost = '';

function colorChip(hex) {
  const chip = document.createElement('div');
  chip.className = 'color-chip';
  chip.title = 'Нажми чтобы скопировать';

  const swatch = document.createElement('div');
  swatch.className = 'color-swatch';
  swatch.style.background = hex;

  const label = document.createElement('span');
  label.className = 'color-hex';
  label.textContent = hex;

  chip.appendChild(swatch);
  chip.appendChild(label);
  chip.addEventListener('click', () => {
    navigator.clipboard.writeText(hex);
    label.textContent = 'скопировано!';
    setTimeout(() => { label.textContent = hex; }, 1200);
  });
  return chip;
}

function renderColors(containerId, colors) {
  const el = document.getElementById(containerId);
  el.innerHTML = '';
  if (!colors.length) {
    el.innerHTML = '<span class="none">не найдено</span>';
    return;
  }
  colors.forEach(hex => el.appendChild(colorChip(hex)));
}

function buildCopyText(data, url) {
  const lines = [`Сайт: ${url}`, ''];

  if (data.logo && data.logo !== '__SVG_INLINE__') {
    lines.push(`Логотип: ${data.logo}`);
  } else if (data.logo === '__SVG_INLINE__') {
    lines.push('Логотип: SVG (встроенный)');
  } else if (data.faviconUrl) {
    lines.push(`Фавиконка: ${data.faviconUrl}`);
  }

  if (data.bgColors.length) {
    lines.push(`Цвет фона: ${data.bgColors.slice(0, 4).join(', ')}`);
  }
  if (data.textColors.length) {
    lines.push(`Цвет текста: ${data.textColors.slice(0, 4).join(', ')}`);
  }
  if (data.fonts.length) {
    lines.push(`Шрифты: ${data.fonts.join(', ')}`);
  }
  if (data.paddings.length) {
    lines.push(`Отступы: ${data.paddings.join(', ')}`);
  }

  return lines.join('\n');
}

// ---------- вывод mdview-темы из собранного стиля ----------

function hexToRgb(hex) {
  return [
    parseInt(hex.slice(1, 3), 16),
    parseInt(hex.slice(3, 5), 16),
    parseInt(hex.slice(5, 7), 16),
  ];
}

function luminance(hex) {
  const [r, g, b] = hexToRgb(hex).map(v => v / 255);
  return 0.299 * r + 0.587 * g + 0.114 * b;
}

function saturation(hex) {
  const [r, g, b] = hexToRgb(hex).map(v => v / 255);
  const max = Math.max(r, g, b), min = Math.min(r, g, b);
  return max === 0 ? 0 : (max - min) / max;
}

function blend(hexA, hexB, t) {
  const a = hexToRgb(hexA), b = hexToRgb(hexB);
  return '#' + a.map((v, i) => Math.round(v + (b[i] - v) * t)
    .toString(16).padStart(2, '0')).join('').toUpperCase();
}

function deriveMdviewTheme(data, host) {
  const bgColors = data.bgColors || [];
  const textColors = data.textColors || [];
  const all = [...bgColors, ...textColors];

  const bg = bgColors[0] || '#FFFFFF';
  const text = textColors[0] || (luminance(bg) > 0.5 ? '#1A1A1A' : '#FFFFFF');

  // акценты — самые насыщенные цвета палитры
  const accents = [...new Set(all)]
    .filter(h => saturation(h) > 0.35)
    .sort((a, b) => saturation(b) - saturation(a));
  const accent = accents[0] || '#7C88FC';
  const accent2 = accents.find(h => h !== accent) || '#FF8562';

  // поверхность — фоновый цвет, близкий по светлоте к основному, иначе лёгкий сдвиг к тексту
  const surface = bgColors.slice(1).find(c =>
    Math.abs(luminance(c) - luminance(bg)) < 0.3) || blend(bg, text, 0.06);
  const line = blend(bg, text, 0.18);

  return {
    format: 'mdview-theme',
    version: 1,
    name: host || 'imported style',
    site: host || '',
    logo: (data.logo && data.logo !== '__SVG_INLINE__') ? data.logo : (data.faviconUrl || ''),
    bg, text, accent, accent2, surface, line,
    fonts: data.fonts || [],
  };
}

// ---------- скачивание / импорт файлов ----------

function downloadJson(filename, obj) {
  const blob = new Blob([JSON.stringify(obj, null, 2)], { type: 'application/json' });
  const a = document.createElement('a');
  a.href = URL.createObjectURL(blob);
  a.download = filename;
  a.click();
  URL.revokeObjectURL(a.href);
}

function showHint(text) {
  const el = document.getElementById('toolbar-hint');
  el.textContent = text;
  el.style.display = 'block';
}

// ---------- рендер ----------

function render(data, host, imported) {
  currentData = data;
  currentHost = host;

  document.getElementById('domain').innerHTML = imported
    ? `${host}<span class="badge-imported">импорт</span>`
    : host;
  document.getElementById('status').style.display = 'none';
  document.getElementById('content').style.display = 'block';

  // Logo
  const logoEl = document.getElementById('logo-content');
  logoEl.innerHTML = '';
  if (data.logo && data.logo !== '__SVG_INLINE__') {
    const img = document.createElement('img');
    img.className = 'logo-preview';
    img.src = data.logo;
    img.onerror = () => img.remove();

    const link = document.createElement('a');
    link.className = 'logo-url';
    link.href = data.logo;
    link.target = '_blank';
    link.textContent = data.logo;

    const row = document.createElement('div');
    row.className = 'logo-row';
    row.appendChild(img);
    row.appendChild(link);
    logoEl.appendChild(row);
  } else if (data.logo === '__SVG_INLINE__') {
    logoEl.innerHTML = '<span style="color:#aaa;font-size:12px">SVG логотип (встроен в HTML)</span>';
    if (data.faviconUrl) {
      logoEl.innerHTML += `<br><a class="logo-url" href="${data.faviconUrl}" target="_blank">${data.faviconUrl}</a>`;
    }
  } else if (data.faviconUrl) {
    const link = document.createElement('a');
    link.className = 'logo-url';
    link.href = data.faviconUrl;
    link.target = '_blank';
    link.textContent = data.faviconUrl;
    const note = document.createElement('span');
    note.style.cssText = 'font-size:11px;color:#555;display:block;margin-bottom:4px';
    note.textContent = 'Логотип не найден, фавиконка:';
    logoEl.appendChild(note);
    logoEl.appendChild(link);
  } else {
    logoEl.innerHTML = '<span class="none">логотип не найден</span>';
  }

  renderColors('bg-colors', data.bgColors);
  renderColors('text-colors', data.textColors);

  // Fonts
  const fontsEl = document.getElementById('fonts');
  fontsEl.innerHTML = '';
  if (!data.fonts.length) {
    fontsEl.innerHTML = '<span class="none">не найдено</span>';
  } else {
    data.fonts.forEach(font => {
      const item = document.createElement('div');
      item.className = 'font-item';
      item.innerHTML = `<span class="font-name">${font}</span><span class="font-sample" style="font-family:'${font}'">Aa Бб</span>`;
      fontsEl.appendChild(item);
    });
  }

  // Paddings
  const padEl = document.getElementById('paddings');
  padEl.innerHTML = '';
  if (!data.paddings.length) {
    padEl.innerHTML = '<span class="none">не найдено</span>';
  } else {
    data.paddings.forEach(p => {
      const chip = document.createElement('span');
      chip.className = 'pad-chip';
      chip.textContent = p;
      padEl.appendChild(chip);
    });
  }

  document.getElementById('copy-text').value = buildCopyText(data, host);
}

// ---------- кнопки ----------

document.getElementById('copy-btn').addEventListener('click', () => {
  navigator.clipboard.writeText(document.getElementById('copy-text').value);
  const btn = document.getElementById('copy-btn');
  btn.textContent = 'Скопировано!';
  btn.classList.add('copied');
  setTimeout(() => {
    btn.textContent = 'Скопировать';
    btn.classList.remove('copied');
  }, 1500);
});

document.getElementById('export-btn').addEventListener('click', () => {
  if (!currentData) { showHint('Нет данных для экспорта — открой обычную страницу.'); return; }
  const safeHost = (currentHost || 'style').replace(/[^\w.-]/g, '_');
  downloadJson(`style-${safeHost}.json`, {
    format: 'style-inspector',
    version: 1,
    site: currentHost,
    exportedAt: new Date().toISOString(),
    data: currentData,
  });
});

document.getElementById('import-btn').addEventListener('click', () => {
  document.getElementById('import-file').click();
});

document.getElementById('import-file').addEventListener('change', async (e) => {
  const file = e.target.files[0];
  if (!file) return;
  try {
    const parsed = JSON.parse(await file.text());
    // принимаем и обёртку экспорта, и «голые» данные
    const data = parsed.data || parsed;
    if (!Array.isArray(data.bgColors) || !Array.isArray(data.textColors)) {
      throw new Error('bad format');
    }
    data.fonts = data.fonts || [];
    data.paddings = data.paddings || [];
    render(data, parsed.site || file.name, true);
    showHint(`Импортирован стиль: ${file.name}`);
  } catch {
    showHint('Не удалось прочитать файл — это не экспорт Style Inspector.');
  }
  e.target.value = '';
});

document.getElementById('mdview-btn').addEventListener('click', () => {
  if (!currentData) { showHint('Нет данных — открой обычную страницу или импортируй стиль.'); return; }
  const theme = deriveMdviewTheme(currentData, currentHost);
  downloadJson('mdview-theme.json', theme);
  showHint('Тема скачана. Внедрить: mdview --install-theme ~/Загрузки/mdview-theme.json');
});

// ---------- запуск ----------

async function run() {
  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
  const host = tab.url ? new URL(tab.url).hostname : '';
  document.getElementById('domain').textContent = host;

  let data;
  try {
    const results = await chrome.scripting.executeScript({
      target: { tabId: tab.id },
      files: ['content.js'],
    });
    data = results[0].result;
  } catch (e) {
    document.getElementById('status').textContent =
      'Нет доступа к этой странице (системная страница Chrome). Можно импортировать сохранённый стиль.';
    return;
  }

  render(data, host, false);
}

run();
