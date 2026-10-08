# 🍽️ Food Weight Calculator

[![Website](https://img.shields.io/badge/Live_Site-Online-brightgreen)](https://foodweight.online/)
[![Go](https://img.shields.io/badge/Go-1.27.1-blue)](https://golang.org/)

Food Weight Calculator is a simple web-based application that estimates the cooked weight of food based on its raw weight. It helps users make informed decisions about portion sizes and nutrition.

🔗 **Live Website:** [foodweight.online](https://foodweight.online/)
📂 **GitHub Repository:** [Kaido997/food-weight-calculator](https://github.com/Kaido997/food-weight-calculator)

## 🛠️ Features

- Convert raw food weight to cooked weight instantly.
- Support for various food types with different cooking weight loss factors.
- Responsive and lightweight web interface.
- Built using Go (without frameworks) for high performance and simplicity.

## 🏗️ Project Structure

```
food-weight-calculator/
│-- api/          # Handles API requests for weight calculations
│-- internal/     # Internal logic and database management
│-- services/     # Core business logic and calculations
│-- web/          # HTML, CSS, and JS frontend files
│-- main.go       # Entry point of the application
│-- go.mod        # Go dependencies
│-- Dockerfile    # Docker setup for deployment
│-- README.md     # Project documentation
```

## 🚀 Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/Kaido997/food-weight-calculator.git
   cd food-weight-calculator
   ```

2. **Run the application:**
   ```bash
   go run .
   ```
   or
   ```bash
   air
   ```

3. **Access the website:**
   Open `http://localhost:8080` in your web browser.

## 🌐 Usage

1. Enter the raw food weight (grams).
2. Choose the food type (e.g., chicken, beef, fish, vegetables).
3. View the estimated cooked weight instantly.

## 🤝 Contributing

Contributions are welcome! Feel free to open issues or submit pull requests on GitHub.

## 📧 Contact

For questions or suggestions, reach out via:

- GitHub: [Kaido997](https://github.com/Kaido997)

## Development

Requires Go 1.27.1 or newer. The app uses only the Go standard library and vanilla JavaScript; there are no module or frontend dependencies to install. Run from the repository root. Analytics storage is created automatically.

```sh
go test -race ./...
go vet ./...
go build -o /tmp/food-weight-calculator .
```

Set `PORT` to override port 8080. Analytics authentication uses `SECRET__ADMIN_PASSWORD`, the hexadecimal SHA-256 digest of the Authorization header value. Analytics are local to each instance; the current Fly configuration has no persistent volume.

## Shared food API

`internal/food.Service` is the application layer used directly by both the UI handlers and the public HTTP API. HTTP parsing and rendering stay outside it:

```go
store, err := database.OpenFoods("internal/database/foodtable.json")
// Handle err before creating the service.
foods := food.New(store)
weight, err := foods.Calculate("FOOD_RICE", 80, "g") // 200 grams
ids, err := foods.GetAll() // sorted food IDs
```

`Calculate(foodID string, rawWeight float64, unit string) (float64, error)` accepts `g`, `gr`, or `lbs` and returns cooked weight in the same unit. Weights must be finite and non-negative. Unknown foods and unsupported units return errors. `GetAll() ([]string, error)` returns IDs from the food database, independently of translations.

The public wrappers are:

- `POST /api/v1/calculate-cooked` with `{"food-type":"FOOD_RICE","quantity":80,"unit":"g"}` returns `{"cooked-weight":200.0}`. Existing v1 field names and one-decimal response precision are preserved; omitting `unit` defaults to grams.
- `GET /api/v1/foods` returns a sorted JSON array of food IDs.

### Replacing the food database

The JSON format remains an array of `{"name":"FOOD_RICE","factor":2.5}` entries. IDs must be nonblank and unique; factors must be positive finite numbers; the table must not be empty.

Replace `internal/database/foodtable.json` by writing a new file beside it and atomically renaming it over the original. The next service call reads the replacement without a restart. File contents are checked on every call and parsed only when changed, including replacements with identical size and modification time. Each call uses one complete validated snapshot.

Invalid or missing files return an error (HTTP 503); requests do not silently use stale factors. Correcting the file restores service automatically. Both language selectors use the current food IDs; a new food without a translated label displays its ID until a translation is supplied. Translation files themselves are loaded at startup.

## Interface and installation

The responsive split layout reuses `web/shell.html` and shared controls/tokens in `web/assets/app.css`. `theme.js` applies light/dark/system mode before CSS loads, and `shell.js` owns language selection, connection status, and installation. Food-specific form behavior stays in `calculator.js`. English and Italian UI messages live beside existing translations; `DESIGN.md` records the design and reuse contract.

The manifest, install icons, and root service worker support installation over HTTPS (or localhost). Browsers with an install prompt use the Install app button; other browsers receive home-screen instructions. The public shell opens offline after caching, but calculations require a connection. API responses, calculation results, and admin routes are never cached. Bump the cache version in `web/sw.js` when changing shell assets. New workers activate after existing app tabs close.

Validation adds no runtime dependencies:

```sh
node --test tests/ui.test.cjs
go test -race ./...
```

Browser rendering and real-device PWA installation still require a connected browser for verification.

## Deployment

Pushing `master` runs formatting checks, race-enabled Go tests, `go vet`, JavaScript tests, and a Docker build before deploying the existing `food-weight-calculator` Fly app. Pull requests run the same checks without deploying. The GitHub repository needs its existing `FLY_API_TOKEN` secret; admin access continues to use the Fly `SECRET__ADMIN_PASSWORD` secret.

The image embeds templates, scripts, styles, images, and fonts; food and translation JSON are copied into `/app/internal/database`. The analytics directory is created at runtime. Fly checks `/api/v1/foods` before routing traffic to verify that food data is available. The current app still uses instance-local analytics without a persistent volume.

For a manual release from a verified checkout:

```sh
flyctl config validate --strict
flyctl deploy --remote-only
```
