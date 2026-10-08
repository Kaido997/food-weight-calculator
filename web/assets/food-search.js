(() => {
    const search = document.getElementById('food-search');
    const select = document.getElementById('foodType');
    const clear = document.getElementById('clear-food-search');
    const hint = document.getElementById('food-help');
    const messages = JSON.parse(document.getElementById('ui-messages').textContent);
    const normal = value => value.normalize('NFD').replace(/\p{M}/gu, '').toLocaleLowerCase(document.documentElement.lang).trim().replace(/\s+/g, ' ');
    const foods = Array.from(select.options, option => ({value: option.value, label: option.textContent, text: normal(option.textContent)}));
    const defaultHint = hint.textContent;
    const collator = new Intl.Collator(document.documentElement.lang);

    function filter() {
        const query = normal(search.value);
        const terms = query.split(' ');
        const selected = select.value;
        // Prefer complete words over partial matches. No fuzzy engine or network calls.
        function rank(text) {
            if (text === query) return 0;
            if (text.startsWith(query + ' ')) return 1;
            if ((' ' + text + ' ').includes(' ' + query + ' ')) return 2;
            if (text.startsWith(query)) return 3;
            if (text.includes(query)) return 4;
            return 5;
        }
        const matches = query ? foods.filter(food => terms.every(term => food.text.includes(term)))
            .sort((a, b) => rank(a.text) - rank(b.text) || a.text.length - b.text.length || collator.compare(a.label, b.label)) : foods;
        select.replaceChildren(...matches.map(food => new Option(food.label, food.value)));
        if (matches.some(food => food.value === selected)) select.value = selected;
        if (!matches.length) select.append(new Option(messages.NoFoods, ''));
        clear.hidden = !search.value;
        hint.textContent = query ? (matches.length ? messages.FoundFoods.replace('{count}', matches.length) : messages.NoFoods) : defaultHint;
        select.dispatchEvent(new Event('change', {bubbles: true}));
    }
    search.addEventListener('input', event => { if (!event.isComposing) filter(); });
    search.addEventListener('compositionend', filter);
    search.addEventListener('keydown', event => {
        if (event.key === 'Enter' && !event.isComposing) {
            event.preventDefault();
            filter();
            select.focus();
        }
    });
    clear.addEventListener('click', () => {
        search.value = '';
        filter();
        search.focus();
    });
    clear.disabled = false;
    document.getElementById('food-search-box').hidden = false;
})();
