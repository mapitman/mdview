function rgbToHex(rgb) {
  const m = rgb.match(/rgba?\((\d+),\s*(\d+),\s*(\d+)/);
  if (!m) return null;
  return '#' + [m[1], m[2], m[3]]
    .map(n => parseInt(n).toString(16).padStart(2, '0'))
    .join('').toUpperCase();
}

function getLuminance(hex) {
  const r = parseInt(hex.slice(1, 3), 16) / 255;
  const g = parseInt(hex.slice(3, 5), 16) / 255;
  const b = parseInt(hex.slice(5, 7), 16) / 255;
  return 0.299 * r + 0.587 * g + 0.114 * b;
}

function analyzeStyles() {
  // --- Logo ---
  let logo = null;

  const logoImgSelectors = [
    'a[href="/"] img', 'header img', 'nav img',
    'img[src*="logo" i]', 'img[alt*="logo" i]',
    'img[class*="logo" i]', 'img[id*="logo" i]',
    '[class*="logo" i] img', '[id*="logo" i] img',
    '.navbar-brand img', '.brand img',
  ];
  for (const sel of logoImgSelectors) {
    const el = document.querySelector(sel);
    if (el && el.src && !el.src.startsWith('data:')) { logo = el.src; break; }
  }

  // SVG logo
  if (!logo) {
    const svgEl = document.querySelector(
      '[class*="logo" i] svg, [id*="logo" i] svg, header svg, nav svg, .navbar-brand svg'
    );
    if (svgEl) logo = '__SVG_INLINE__';
  }

  // Favicon fallback
  const favicon = document.querySelector('link[rel~="icon"]');
  const faviconUrl = favicon ? favicon.href : null;

  // og:image
  const ogImage = document.querySelector('meta[property="og:image"]');
  const ogUrl = ogImage ? ogImage.content : null;

  // --- Colors & Fonts & Paddings ---
  const bgColorMap = new Map();   // hex -> count
  const textColorMap = new Map(); // hex -> count
  const fontMap = new Map();      // family -> count
  const paddingSet = new Set();

  const els = document.querySelectorAll(
    'body, header, nav, main, footer, section, article, aside, ' +
    'h1, h2, h3, h4, p, a, button, span, div, li, input, label, ' +
    '[class*="btn"], [class*="card"], [class*="hero"], [class*="banner"]'
  );

  els.forEach(el => {
    const s = window.getComputedStyle(el);

    // background-color
    const bg = s.backgroundColor;
    if (bg && bg !== 'rgba(0, 0, 0, 0)') {
      const hex = rgbToHex(bg);
      if (hex && hex !== '#000000' || bg.includes('0, 0, 0')) {
        bgColorMap.set(hex, (bgColorMap.get(hex) || 0) + 1);
      }
    }

    // color
    const col = s.color;
    if (col) {
      const hex = rgbToHex(col);
      if (hex) textColorMap.set(hex, (textColorMap.get(hex) || 0) + 1);
    }

    // font-family (first family only)
    const rawFont = s.fontFamily;
    if (rawFont) {
      const family = rawFont.split(',')[0].trim().replace(/['"]/g, '');
      if (family) fontMap.set(family, (fontMap.get(family) || 0) + 1);
    }

    // paddings
    ['paddingTop', 'paddingRight', 'paddingBottom', 'paddingLeft'].forEach(prop => {
      const val = parseFloat(s[prop]);
      if (val > 0 && val < 200) paddingSet.add(val);
    });
  });

  // Sort by frequency, dedupe, limit
  const sortByCount = (map) =>
    [...map.entries()].sort((a, b) => b[1] - a[1]).map(([k]) => k);

  const bgColors = sortByCount(bgColorMap).slice(0, 8);
  const textColors = sortByCount(textColorMap).slice(0, 8);

  // Unique sorted paddings
  const paddings = [...paddingSet]
    .sort((a, b) => a - b)
    .map(v => v + 'px')
    .slice(0, 12);

  const fonts = sortByCount(fontMap).slice(0, 6);

  return { logo, faviconUrl, ogUrl, bgColors, textColors, fonts, paddings };
}

analyzeStyles();
