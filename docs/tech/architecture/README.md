# Architecture

Start with the [overview](overview.md) — the components and the path a refresh
takes through them. Then the piece you are touching:

- [Backend](backend.md) — the Go packages, their boundaries, and the one
  dependency direction that is enforced by convention.
- [Frontend](frontend.md) — the Vue components, the two modules that hold the
  only global state there is, and where styling stops being a component's
  business.
- [Security model](security.md) — there is no login in front of the token, so
  three separate guards do the work a login would. This is the document to read
  before changing anything in the middle of `api.withMiddleware`.
