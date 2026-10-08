// Runs before CSS to avoid a light flash when a saved dark preference exists.
(() => {
    const media = matchMedia('(prefers-color-scheme: dark)');
    let mode = 'system';
    try { mode = localStorage.getItem('foodweight.theme') || mode; } catch {}
    if (!['light', 'dark', 'system'].includes(mode)) mode = 'system';
    function apply(value) {
        mode = value;
        const theme = value === 'system' ? (media.matches ? 'dark' : 'light') : value;
        document.documentElement.dataset.theme = theme;
        document.documentElement.dataset.themeMode = value;
        document.querySelector('meta[name="theme-color"]')?.setAttribute('content', theme === 'dark' ? '#14231f' : '#f6f8f5');
    }
    apply(mode);
    media.addEventListener('change', () => apply(mode));
    document.addEventListener('change', event => {
        if (event.target.matches('input[name="theme"]')) {
            apply(event.target.value);
            try { localStorage.setItem('foodweight.theme', mode); } catch {}
        }
    });
})();
