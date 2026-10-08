const { test } = require('node:test');
const assert = require('node:assert/strict');
const { readFileSync } = require('node:fs');
const vm = require('node:vm');

const read = path => readFileSync(path, 'utf8');
const translations = lang => JSON.parse(read(`internal/database/translations/${lang}.json`)).ui;
function element(value = '') {
    return {
        value, textContent: '', hidden: false, disabled: false, children: [], attributes: {}, listeners: {},
        setAttribute(key, value) { this.attributes[key] = value; },
        addEventListener(type, listener) { this.listeners[type] = listener; },
        append(child) { this.children.push(child); },
        replaceChildren(...children) { this.children = children; },
        focus() { this.focused = true; }
    };
}
function calculator(lang = 'en') {
    const ids = Object.fromEntries(['weight-form', 'ui-messages', 'result', 'form-error', 'reload-foods', 'weight-error', 'food-error', 'weight-help', 'food-help'].map(id => [id, element()]));
    const controls = Object.fromEntries(['button[type="submit"]', '[data-submit-label]', '.spinner', '.button-arrow'].map(key => [key, element()]));
    ids['ui-messages'].textContent = JSON.stringify(translations(lang));
    const form = ids['weight-form'];
    form.action = '/calculate';
    form.elements = { quantity: element('80'), 'food-type': element('FOOD_RICE'), unit: element('g') };
    form.elements['food-type'].selectedOptions = [{textContent: lang === 'it' ? 'Riso' : 'Rice'}];
    form.querySelector = key => controls[key];
    const requests = [];
    const timers = new Map();
    let timerID = 0;
    vm.runInNewContext(read('web/assets/calculator.js'), {
        document: {getElementById: id => ids[id], documentElement: {lang}, createElement: () => element()},
        Intl, Number, AbortController, URLSearchParams, TypeError, SyntaxError,
        FormData: class { *[Symbol.iterator]() { for (const [key, input] of Object.entries(form.elements)) yield [key, input.value]; } },
        setTimeout(fn) { timers.set(++timerID, fn); return timerID; },
        clearTimeout(id) { timers.delete(id); },
        fetch(url, options) { return new Promise((resolve, reject) => requests.push({url, options, resolve, reject})); }
    });
    return {ids, controls, form, requests, timers, submit: () => form.listeners.submit({preventDefault() {}})};
}
const response = (value = 200, status = 200) => ({ok: status === 200, status, json: async () => status === 400 ? {error: 'unknown-food'} : ({'cooked-weight': value})});

test('English and Italian have the same complete UI message keys', () => {
    assert.deepEqual(Object.keys(translations('en')).sort(), Object.keys(translations('it')).sort());
    for (const lang of ['en', 'it']) for (const value of Object.values(translations(lang))) assert.ok(value.trim());
});
test('validation focuses the invalid input without sending a request', async () => {
    const app = calculator();
    app.form.elements.quantity.value = '-2';
    await app.submit();
    assert.equal(app.requests.length, 0);
    assert.equal(app.form.elements.quantity.focused, true);
    assert.equal(app.form.elements.quantity.attributes['aria-invalid'], 'true');
    assert.equal(app.ids['weight-error'].textContent, translations('en').InvalidWeight);
});
test('successful calculation sends selected units and renders localized numbers', async () => {
    const app = calculator('it');
    app.form.elements.unit.value = 'lbs';
    const pending = app.submit();
    assert.equal(app.controls['button[type="submit"]'].disabled, true);
    assert.equal(app.requests[0].options.body.get('unit'), 'lbs');
    app.requests[0].resolve(response(200.25));
    await pending;
    assert.equal(app.ids.result.children[0].textContent, '200,25');
    assert.equal(app.ids.result.children[0].children[0].textContent, 'lbs');
    assert.equal(app.controls['button[type="submit"]'].disabled, false);
    assert.equal(app.timers.size, 0);
});
test('duplicate submits are blocked and edits cancel stale results', async () => {
    const app = calculator();
    const first = app.submit();
    await app.submit();
    assert.equal(app.requests.length, 1);
    app.form.elements.quantity.value = '40';
    app.form.listeners.input();
    assert.equal(app.requests[0].options.signal.aborted, true);
    const second = app.submit();
    app.requests[0].resolve(response(200));
    await first;
    assert.equal(app.controls['button[type="submit"]'].disabled, true);
    app.requests[1].resolve(response(100));
    await second;
    assert.equal(app.ids.result.children[0].textContent, '100');
    app.form.listeners.change();
    assert.equal(app.ids.result.children[0].textContent, translations('en').ResultEmpty);
});
test('server errors preserve values and a retry can succeed', async () => {
    const app = calculator();
    const pending = app.submit();
    app.requests[0].resolve(response(0, 503));
    await pending;
    assert.equal(app.ids['form-error'].textContent, translations('en').Unavailable);
    assert.equal(app.form.elements.quantity.value, '80');
    const retry = app.submit();
    app.requests[1].resolve(response());
    await retry;
    assert.equal(app.ids['form-error'].textContent, '');
    assert.equal(app.ids.result.children[0].textContent, '200');
});
test('timeout, offline failure, changed food, and IME composition have recovery states', async () => {
    const app = calculator();
    app.form.listeners.compositionstart();
    await app.submit();
    assert.equal(app.requests.length, 0);
    app.form.listeners.compositionend();
    let pending = app.submit();
    for (const timer of app.timers.values()) timer();
    assert.equal(app.requests[0].options.signal.aborted, true);
    app.requests[0].reject({name: 'AbortError'});
    await pending;
    assert.equal(app.ids['form-error'].textContent, translations('en').Timeout);
    pending = app.submit();
    app.requests[1].reject(new TypeError('network failed'));
    await pending;
    assert.equal(app.ids['form-error'].textContent, translations('en').RequestError);
    pending = app.submit();
    app.requests[2].resolve(response(0, 400));
    await pending;
    assert.equal(app.ids['reload-foods'].hidden, false);
    assert.equal(app.ids['form-error'].textContent, translations('en').ChangedFood);
});
test('theme follows system only in system mode and survives blocked storage', () => {
    const media = {matches: true, addEventListener(_, fn) { this.change = fn; }};
    const root = {dataset: {}};
    let change;
    vm.runInNewContext(read('web/assets/theme.js'), {
        matchMedia: () => media,
        localStorage: {getItem() { throw new Error('blocked'); }, setItem() { throw new Error('blocked'); }},
        document: {documentElement: root, querySelector: () => element(), addEventListener(_, fn) { change = fn; }}
    });
    assert.equal(root.dataset.theme, 'dark');
    change({target: {matches: () => true, value: 'light'}});
    media.change();
    assert.equal(root.dataset.theme, 'light');
    change({target: {matches: () => true, value: 'system'}});
    assert.equal(root.dataset.theme, 'dark');
    media.matches = false;
    media.change();
    assert.equal(root.dataset.theme, 'light');
});
test('worker caches public shell only and falls back offline without intercepting calculations', async () => {
    const listeners = {};
    const entries = new Map();
    let online = true;
    vm.runInNewContext(read('web/sw.js'), {
        URL, Response,
        self: {location: {origin: 'https://example.test'}, addEventListener(type, fn) { listeners[type] = fn; }},
        caches: {open: async () => ({put: async (request, response) => entries.set(request.url, response), match: async request => entries.get(request.url)})},
        fetch: async () => { if (!online) throw new TypeError('offline'); return new Response('shell'); }
    });
    const request = (path, method = 'GET') => {
        let pending;
        listeners.fetch({request: {url: 'https://example.test' + path, method}, respondWith(value) { pending = value; }});
        return pending;
    };
    assert.equal(request('/calculate', 'POST'), undefined);
    assert.equal(request('/api/v1/foods'), undefined);
    assert.equal(request('/admin/analytics'), undefined);
    assert.equal(await (await request('/?lang=it')).text(), 'shell');
    online = false;
    assert.equal(await (await request('/?lang=it')).text(), 'shell');
    assert.equal((await request('/?lang=en')).status, 503);
});

test('install controls expose fallback instructions and use the native install event when available', async () => {
    const ids = Object.fromEntries(['ui-messages', 'language', 'connection-status', 'install-button', 'install-help'].map(id => [id, element()]));
    ids['ui-messages'].textContent = JSON.stringify(translations('en'));
    ids['install-help'].hidden = true;
    const events = {};
    const assignments = [];
    const navigator = {onLine: true};
    const location = {href: 'https://example.test/?lang=en', assign(url) { assignments.push(url.href); }};
    vm.runInNewContext(read('web/assets/shell.js'), {
        document: {getElementById: id => ids[id], documentElement: {dataset: {themeMode: 'system'}}, querySelector: () => element()},
        window: {addEventListener(type, fn) { events[type] = fn; }},
        navigator, location, URL, matchMedia: () => ({matches: false})
    });
    await ids['install-button'].listeners.click();
    assert.equal(ids['install-help'].hidden, false);
    assert.equal(ids['install-button'].attributes['aria-expanded'], 'true');
    let prompted = false;
    events.beforeinstallprompt({preventDefault() {}, prompt: async () => { prompted = true; }, userChoice: Promise.resolve({outcome: 'dismissed'})});
    await ids['install-button'].listeners.click();
    assert.equal(prompted, true);
    assert.equal(ids['install-button'].disabled, false);
    navigator.onLine = false;
    events.offline();
    assert.equal(ids['connection-status'].hidden, false);
    ids.language.listeners.change({target: {value: 'it'}});
    assert.equal(assignments[0], 'https://example.test/?lang=it');
    events.appinstalled();
    assert.equal(ids['install-button'].hidden, true);
});

test('both themes meet text and form-outline contrast requirements', () => {
    const css = read('web/assets/app.css');
    const parse = selector => Object.fromEntries([...css.match(selector)[1].matchAll(/--([\w-]+):\s*(#[\da-f]{6});/g)].map(match => [match[1], match[2]]));
    const light = parse(/:root \{([\s\S]*?)\}/);
    const dark = {...light, ...parse(/:root\[data-theme="dark"\] \{([\s\S]*?)\}/)};
    const luminance = hex => {
        const [r, g, b] = hex.slice(1).match(/../g).map(h => parseInt(h, 16) / 255).map(v => v <= .04045 ? v / 12.92 : ((v + .055) / 1.055) ** 2.4);
        return r * .2126 + g * .7152 + b * .0722;
    };
    const ratio = (a, b) => (Math.max(luminance(a), luminance(b)) + .05) / (Math.min(luminance(a), luminance(b)) + .05);
    for (const theme of [light, dark]) {
        for (const background of ['bg', 'surface', 'result']) {
            assert.ok(ratio(theme.text, theme[background]) >= 4.5);
            assert.ok(ratio(theme.muted, theme[background]) >= 4.5);
        }
        assert.ok(ratio(theme['on-accent'], theme.accent) >= 4.5);
        assert.ok(ratio(theme['control-border'], theme.surface) >= 3);
        assert.ok(ratio(theme.danger, theme.bg) >= 4.5);
    }
});

function foodSearch(options, selected = options[0][0], lang = 'en') {
    const ids = Object.fromEntries(['food-search', 'foodType', 'clear-food-search', 'food-help', 'ui-messages', 'food-search-box'].map(id => [id, element()]));
    ids['ui-messages'].textContent = JSON.stringify(translations(lang));
    ids['food-help'].textContent = 'Choose a food';
    const select = ids.foodType;
    select.options = options.map(([value, textContent]) => ({value, textContent}));
    select.value = selected;
    select.replaceChildren = function (...options) { this.options = options; this.value = options[0]?.value || ''; };
    select.append = function (option) { this.options.push(option); };
    select.dispatchEvent = event => { select.lastEvent = event; };
    vm.runInNewContext(read('web/assets/food-search.js'), {
        document: {getElementById: id => ids[id], documentElement: {lang}},
        Intl, Event,
        Option: function (textContent, value) { this.textContent = textContent; this.value = value; }
    });
    return {ids, search(query, composing = false) {
        ids['food-search'].value = query;
        ids['food-search'].listeners.input({isComposing: composing});
        return Array.from(select.options, option => option.value);
    }};
}
test('food search matches inside names and ranks whole-word matches ahead of fragments', () => {
    const app = foodSearch([
        ['rice', 'Rice'], ['long', 'Long semolina pasta'], ['short', 'Short semolina pasta'],
        ['prefix', 'Pasta fresca'], ['exact', 'Pasta'], ['fragment', 'Pastafarian food']
    ]);
    assert.deepEqual(app.search('pasta'), ['exact', 'prefix', 'long', 'short', 'fragment']);
    assert.equal(app.ids.foodType.value, 'exact');
    assert.deepEqual(app.search('long pasta'), ['long']);
    assert.equal(app.ids.foodType.value, 'long');
    assert.equal(app.ids.foodType.lastEvent.type, 'change');
    assert.equal(app.ids.foodType.lastEvent.bubbles, true);
});
test('food search ignores case, accents and excess whitespace; clear restores the list', () => {
    const app = foodSearch([['a', 'Caffè'], ['b', 'Pasta di semola lunga'], ['c', 'Riso']], 'c', 'it');
    assert.deepEqual(app.search('  CAFFE  '), ['a']);
    assert.deepEqual(app.search('semola   lunga'), ['b']);
    assert.deepEqual(app.search('no such food'), ['']);
    assert.equal(app.ids.foodType.options[0].textContent, translations('it').NoFoods);
    app.ids['clear-food-search'].listeners.click();
    assert.deepEqual(Array.from(app.ids.foodType.options, o => o.value), ['a', 'b', 'c']);
    assert.equal(app.ids['food-search'].focused, true);
    assert.equal(app.ids['clear-food-search'].hidden, true);
});
test('food search preserves a matching selection and waits for IME composition', () => {
    const app = foodSearch([['a', 'Short semolina pasta'], ['b', 'Long semolina pasta']], 'a');
    app.search('pasta');
    assert.equal(app.ids.foodType.value, 'a');
    app.search('long', true);
    assert.equal(app.ids.foodType.options.length, 2);
    app.ids['food-search'].listeners.compositionend();
    assert.equal(app.ids.foodType.options.length, 1);
    let prevented = false;
    app.ids['food-search'].listeners.keydown({key: 'Enter', isComposing: false, preventDefault() { prevented = true; }});
    assert.equal(prevented, true);
    assert.equal(app.ids.foodType.focused, true);
});
