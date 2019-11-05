# spa

This folder contains a simple React GUI SPA that invokes the `/token/create-orphan` endpoint of `vault-token-issuer`

The static build output (i.e. `yarn build`) of this app (under `build/`) is embedded in `vault-token-issuer` via [vfsgen](https://github.com/shurcooL/vfsgen) via `go generate` as defined by the `go generate` directive comment in [vault-token-issuer.go](../vault-token-issuer.go)

## Setup

This is a React project. You should `npm install` from within this directory and then `yarn build`

## Debugging

To debug the SPA simply `npm start` in this directory then hit the ui @ http://localhost:3000

In `development` mode the `.env.development` file defines the ENV var `REACT_APP_VAULT_TOKEN_ISSUER_ROOT_URL` which points to the `vault-token-issuer` backend running in a separate process on `https://localhost:8443` (change as you please).

