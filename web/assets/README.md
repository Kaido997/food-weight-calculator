# UI assets

`food-prep.png` was downloaded unchanged from https://foodweight.online/food.png on 2026-10-07, at the project owner's request. It is the image used on the existing Food Weight landing page.

`icon.svg` is the app's scale mark. Render the install icons without adding runtime dependencies:

```sh
rsvg-convert -w 192 -h 192 web/assets/icon.svg -o web/assets/icon-192.png
rsvg-convert -w 512 -h 512 web/assets/icon.svg -o web/assets/icon-512.png
```

`app.css` owns shared visual tokens and control recipes. `theme.js` initializes appearance before CSS; `shell.js` owns language, connectivity, and installation. `calculator.js` handles only the calculator workflow. `food-search.js` adds local substring and multiword filtering to the native food selector. See `DESIGN.md` for reuse and state contracts.

`libre-baskerville-latin.woff2` is Libre Baskerville Regular from Google Fonts, downloaded on 2026-10-07. The Latin subset covers the English and Italian headings. Its SIL Open Font License is included in `libre-baskerville-OFL.txt`. The font is self-hosted, preloaded, and cached with the PWA shell.
