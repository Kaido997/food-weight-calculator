# Repository Guidelines

# Intro
Food weight calculatro is a handy calculator as the name says, about calculating the weight of cooked food knowing the raw weghit this tool turns out to be useful while on a food plan and ou have X grams of cooked rice but you need to ate 80gr of raw rice.

# Goals
- keep the project simple without overcomplicating things
- prioritize low latency fast resposiveness of the ui and ease of use ui

# Quality Rules
- use go and vanilla javascript with html and css, do not add more dependecies unless strictly necessary
- Keep the implementation small, sharp, easy to understand. Try to write elegant code in a state of grace. Don't settle for the first thing that comes to mind, try to find the most minimal and better working design. Don't introduce slop: very fragile code that just patches specific cases, dead code, useless code and code ways more complicated of how it should be.
- Comment important inference code where the model mechanics, cache lifetime, memory policy, or API orchestration are not obvious from the local code.
- Prefer comments beside the implementation over separate design documents.
- Keep comments instructive and compact: explain why a shape, ordering, cache boundary, or memory choice exists.
- Keep public APIs narrow. CLI/server code should not know tensor internals.
- Do not add permanent semantic variants behind flags. Diagnostic switches are fine when they validate the one release path.

## Project Structure & Module Organization

- `main.go` initializes data, embeds `web/*.html`, and registers page routes using Go’s standard `net/http` package.
- `api/handler/` contains the versioned HTTP endpoints; `internal/food/` owns the shared calculation and food-list service.
- `internal/database/` manages food factors, translations, and file-backed analytics. Food definitions live in `foodtable.json`; English and Italian translations live in `translations/`.
- `services/auth_service/` handles analytics authentication.
- `web/` contains Go HTML templates with reusable shell templates; `web/assets/` holds shared CSS, vanilla JavaScript, images, and the self-hosted font. There is no frontend build pipeline.

## Build, Test, and Development Commands

Use Go 1.27.1 or newer, as specified by `go.mod`. Run commands from the repository root because data paths depend on the working directory.
Add and make test for each new and old feature.

- `go run .` — start the application at `http://localhost:8080`; override the port with `PORT`.
- `air` — optionally run with automatic rebuilds using `.air.toml`; install Air separately.
- `go build -o /tmp/food-weight-calculator .` — compile the application without adding a binary to the repository.
- `go test -race ./...` — run all Go tests with race detection.
- `node --test tests/ui.test.cjs` — run JavaScript interaction and theme checks (Node 22 in CI).
- `go vet ./...` — check for common Go mistakes.

## Coding Style & Naming Conventions

Format changed Go files with `gofmt -w`; use standard Go tabs, exported `PascalCase` names, and unexported `camelCase` names. Match the existing four-space indentation in HTML/CSS. No dedicated lint configuration exists. Preserve API field names such as `food-type`, and keep food keys consistent across the factor table and both translation files.

## Testing Guidelines

Go and JavaScript tests cover the service, routes, dynamic JSON reloads, authentication, UI interactions, and PWA assets. No coverage threshold is set. Add standard-library `testing` tests in adjacent `*_test.go` files, using `TestXxx` names and `net/http/httptest` for handlers. Cover conversion factors, missing foods, and malformed requests. Isolate analytics writes with temporary directories. For UI changes, manually check calculations and both language selections.

## Commit & Pull Request Guidelines

History uses short, informal subjects such as `added analytics page`; no Conventional Commits convention is established. Write concise, action-oriented subjects. PRs should describe the behavior change, link relevant issues, report validation, and include screenshots for UI changes. Pushes to `master` trigger Fly.io deployment.

## Configuration & Secrets

Keep credentials and generated analytics out of commits. `SECRET__ADMIN_PASSWORD` expects a hexadecimal SHA-256 digest of the supplied authorization value. Avoid logging credentials when changing authentication code.
