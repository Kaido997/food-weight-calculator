(() => {
    const form = document.getElementById('weight-form');
    const messages = JSON.parse(document.getElementById('ui-messages').textContent);
    const weight = form.elements.quantity;
    const food = form.elements['food-type'];
    const submit = form.querySelector('button[type="submit"]');
    const submitLabel = form.querySelector('[data-submit-label]');
    const spinner = form.querySelector('.spinner');
    const arrow = form.querySelector('.button-arrow');
    const result = document.getElementById('result');
    const error = document.getElementById('form-error');
    const reload = document.getElementById('reload-foods');
    const number = new Intl.NumberFormat(document.documentElement.lang, { maximumFractionDigits: 2 });
    let activeRequest;
    let composing = false;
    let hasResult = false;

    function busy(value) {
        form.setAttribute('aria-busy', String(value));
        submit.disabled = value;
        spinner.hidden = !value;
        arrow.hidden = value;
        submitLabel.textContent = value ? messages.Calculating : messages.Calculate;
    }
    function fieldError(input, id, message = '') {
        const hint = document.getElementById(id);
        hint.textContent = message;
        hint.hidden = !message;
        document.getElementById(id.replace("-error", "-help")).hidden = Boolean(message);
        input.setAttribute('aria-invalid', String(Boolean(message)));
    }
    function emptyResult() {
        if (!hasResult) return;
        hasResult = false;
        const hint = document.createElement('p');
        hint.className = 'result-empty';
        hint.textContent = messages.ResultEmpty;
        result.replaceChildren(hint);
    }
    function invalidate() {
        // A result belongs to the submitted values, never to subsequent edits.
        activeRequest?.abort();
        activeRequest = null;
        busy(false);
        error.textContent = '';
        reload.hidden = true;
        fieldError(weight, 'weight-error');
        fieldError(food, 'food-error');
        emptyResult();
    }
    form.addEventListener('input', invalidate);
    form.addEventListener('change', invalidate);
    form.addEventListener('compositionstart', () => { composing = true; });
    form.addEventListener('compositionend', () => { composing = false; });
    form.addEventListener('submit', async event => {
        event.preventDefault();
        if (composing || activeRequest) return;
        error.textContent = '';
        reload.hidden = true;
        const quantity = Number(weight.value);
        let invalid;
        if (!food.value) {
            fieldError(food, 'food-error', messages.InvalidFood);
            invalid = food;
        }
        if (!weight.value.trim() || !Number.isFinite(quantity) || quantity < 0) {
            fieldError(weight, 'weight-error', messages.InvalidWeight);
            invalid ||= weight;
        }
        if (invalid) { invalid.focus(); return; }
        const controller = new AbortController();
        activeRequest = controller;
        const timeout = setTimeout(() => controller.abort(), 12000);
        const unit = form.elements.unit.value;
        const foodLabel = food.selectedOptions[0].textContent;
        busy(true);
        emptyResult();
        try {
            const response = await fetch(form.action, {
                method: 'POST',
                headers: { Accept: 'application/json' },
                body: new URLSearchParams(new FormData(form)),
                signal: controller.signal
            });
            if (activeRequest !== controller) return;
            if (!response.ok) {
                if (response.status === 400) {
                    const details = await response.json();
                    if (activeRequest !== controller) return;
                    if (details.error === 'unknown-food') {
                        reload.hidden = false;
                        throw new Error(messages.ChangedFood);
                    }
                    fieldError(weight, 'weight-error', messages.InvalidWeight);
                    weight.focus();
                    throw new Error(messages.InvalidWeight);
                }
                throw new Error(response.status === 503 ? messages.Unavailable : messages.RequestError);
            }
            const data = await response.json();
            if (activeRequest !== controller) return;
            if (!Number.isFinite(data['cooked-weight'])) throw new Error(messages.RequestError);
            const value = document.createElement('p');
            value.className = 'result-value';
            value.textContent = number.format(data['cooked-weight']);
            const suffix = document.createElement('span');
            suffix.className = 'result-unit';
            suffix.textContent = unit;
            value.append(suffix);
            const context = document.createElement('p');
            context.className = 'result-context';
            context.textContent = `${foodLabel} · ${messages.ResultFor}: ${number.format(quantity)} ${unit}`;
            result.replaceChildren(value, context);
            hasResult = true;
        } catch (failure) {
            if (activeRequest !== controller) return;
            error.textContent = failure.name === 'AbortError' ? messages.Timeout
                : failure instanceof TypeError || failure instanceof SyntaxError ? messages.RequestError : failure.message;
        } finally {
            clearTimeout(timeout);
            if (activeRequest === controller) { activeRequest = null; busy(false); }
        }
    });
})();
