(() => {
    const messages = JSON.parse(document.getElementById('ui-messages').textContent);
    const mode = document.documentElement.dataset.themeMode;
    const themeInput = document.querySelector(`input[name="theme"][value="${mode}"]`);
    if (themeInput) themeInput.checked = true;
    document.getElementById('language').addEventListener('change', event => {
        const url = new URL(location.href);
        url.searchParams.set('lang', event.target.value);
        location.assign(url);
    });
    const status = document.getElementById('connection-status');
    const updateConnection = () => { status.hidden = navigator.onLine; };
    window.addEventListener('online', updateConnection);
    window.addEventListener('offline', updateConnection);
    updateConnection();

    const button = document.getElementById('install-button');
    const help = document.getElementById('install-help');
    let installPrompt;
    window.addEventListener('beforeinstallprompt', event => {
        event.preventDefault();
        installPrompt = event;
        help.textContent = messages.InstallReady;
    });
    button.addEventListener('click', async () => {
        if (installPrompt) {
            const prompt = installPrompt;
            installPrompt = null;
            button.disabled = true;
            try { await prompt.prompt(); await prompt.userChoice; }
            catch { help.hidden = false; }
            finally {
                button.disabled = false;
                help.textContent = messages.InstallHelp;
                button.setAttribute('aria-expanded', String(!help.hidden));
            }
        } else {
            help.hidden = !help.hidden;
            button.setAttribute('aria-expanded', String(!help.hidden));
        }
    });
    button.disabled = false;
    const installed = () => { button.hidden = true; help.hidden = true; };
    window.addEventListener('appinstalled', installed);
    if (matchMedia('(display-mode: standalone)').matches || navigator.standalone) installed();
    if ('serviceWorker' in navigator) {
        navigator.serviceWorker.register('/sw.js').catch(() => {
            // Installation is optional: a blocked worker must not affect calculations.
        });
    }
})();
