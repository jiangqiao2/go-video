# Repository Guidelines

## Project Structure & Module Organization
- Microservices: `user-service`, `upload-service`, `transcode-service` (Go 1.24) follow DDD folders (`app`, `ddd/adapter|application|domain|infrastructure`, `pkg` utilities, `configs/*.yaml`).
- Gateway: `gateway-service` ships Kong declarative config (`kong.yml`, `.env.*`, `gen-kong.sh`); no Go code here.
- Frontend: `frontend` is React 18 + Vite + TypeScript with `src/pages`, `components`, `services` (axios APIs), `store` (zustand), `types`, `utils`.
- Contracts & data: gRPC definitions live in `proto/` (module `go-vedio-1/proto`); DB schema/seeds in `scripts/mysql`. PEM keys and sample YAMLs are for local dev only.

## Build, Test, and Development Commands
- Bootstrap DBs: `mysql -h 127.0.0.1 -P 3306 -u root -p < scripts/mysql/init_all.sql`.
- Backend dev: from a service root run `go mod tidy && go run main.go` (override with `CONFIG_PATH` to point at a specific YAML). Health checks at `/health`.
- Backend tests: `go test ./...` per service; add unit tests near domain/service code when you introduce logic.
- Frontend: `cd frontend && npm install && npm run dev` for local, `npm run build` for production bundle, `npm run lint` before PRs.
- Docker/publish (optional): `TAG=dev ./build_push.sh` builds all images; gateway config comes from `cd gateway-service && ENV_FILE=.env.prod ./gen-kong.sh`.

## Coding Style & Naming Conventions
- Run `gofmt`/`go fmt ./...`; keep packages lower_snake, exported Go identifiers PascalCase. Use structured logging via `pkg/logger` with field maps and context/timeouts for I/O.
- Register new HTTP/gRPC adapters via `init` plugins in `ddd/adapter/*` so `manager` picks them up; keep DTOs and request validation close to adapters/application layer.
- Frontend: TypeScript + functional components, colocate API clients in `src/services`, state in `store`, and shared styles in `index.css`. Prefer explicit types over `any`.

## Testing Guidelines
- Favor table-driven Go tests for domain/service functions; avoid hitting real Redis/MinIO—mock interfaces or use in-memory fakes. Document fixtures/config keys required for new features.
- No frontend test suite yet; at minimum run `npm run lint` + `npm run build` and note manual checks (e.g., upload flow, playback) in PRs.

## Commit & Pull Request Guidelines
- Use Conventional Commit format seen in history: `feat(scope): message`, `refactor(area): ...`, `style: ...`; scope is usually the service or feature (e.g., `upload`, `avatar`).
- PRs should list impacted services, configs (`configs/*.yaml`, `.env.*`), DB migrations (`scripts/mysql`), and test/manual results. Attach UI screenshots/GIFs for frontend work. Keep changes focused and avoid committing generated binaries.

## Security & Configuration Tips
- Do not commit real secrets; sample YAMLs, `.env.*`, and PEM keys are for development only. Point `CONFIG_PATH`/`GATEWAY_ENV` to private files in deployments.
- Verify object storage buckets and etcd endpoints before enabling service registry; keep CORS origins in configs aligned with the frontend host.
