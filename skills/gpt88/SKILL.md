---
name: gpt88
description: "Connect AI clients and SDKs to GPT88's OpenAI-compatible API, authorize GPT88 Cloud clients, inspect live models and pricing, and manage user API keys safely. Use when the user asks to configure GPT88, GPT88 API, OpenAI SDK, Cursor, Codex, Claude Code, Cline, GPT88 Cloud login, model availability, pricing, usage, or API key management. Triggers include: \"接入 GPT88\", \"配置 GPT88\", \"GPT88 API Key\", and \"登录 GPT88 云端\". Do not use for Sub2API administrator operations."
---

# GPT88 Agent Onboarding

GPT88 exposes an OpenAI-compatible model gateway and a separate account/control
plane. Use this skill to connect a client end to end while keeping account tokens
and inference keys separate.

## Service map

Use the host named by the user when they provide one. Otherwise use these
production defaults:

```text
Account and control plane: https://agent.gpt88.cc
OpenAI-compatible API:     https://api.gpt88.cc/v1
API origin:                https://api.gpt88.cc
```

The API base already contains `/v1`. SDKs append paths such as `/models` or
`/chat/completions`; do not append `/v1` twice or put a full resource path in a
Base URL. For self-hosted or test deployments, use `GPT88_AGENT_ORIGIN` and
`GPT88_API_BASE_URL`. Cocode also accepts the legacy
`COCODE_GPT88_BASE_URL` override.

The first-party TUI names the account override `GPT88_ORIGIN`; desktop
compatibility code accepts `GPT88_AGENT_ORIGIN`. Both represent the same
account/control-plane origin and must point to an HTTPS origin outside local
development.

For shell-based flows, normalize the hosts once and reuse them:

```bash
export GPT88_ACCOUNT_BASE="${GPT88_ACCOUNT_BASE:-${GPT88_ORIGIN:-${GPT88_AGENT_ORIGIN:-https://agent.gpt88.cc}}}"
export GPT88_API_BASE="${GPT88_API_BASE:-${GPT88_API_BASE_URL:-${COCODE_GPT88_BASE_URL:-https://api.gpt88.cc/v1}}}"
export GPT88_API_ORIGIN="${GPT88_API_ORIGIN:-${GPT88_API_BASE%/}}"
GPT88_API_ORIGIN="${GPT88_API_ORIGIN%/v1}"
export GPT88_API_ORIGIN
```

There are two unrelated credential types:

- `GPT88_API_KEY` is the model gateway key. Use it with `/v1/*` inference
  requests and client configuration. The generated prefix is deployment-specific
  and must not be guessed.
- `GPT88_ACCESS_TOKEN` is an account/session credential from GPT88 login or a
  device flow. Use it only with account and catalog endpoints. It is not a
  replacement for an inference API key.

GPT88 Cloud in the first-party TUI/Desktop may use an access token plus a
non-secret cloud key ID or name reference. Generic SDKs and third-party clients
usually need the actual API key instead.

## Security rules

- Send a credential only to the configured GPT88 host. Keep account calls on the
  account origin and inference calls on the API origin. Reject unexpected
  redirects; `curl` examples must include `--max-redirs 0`.
- Use HTTPS outside local development. Prefer environment variables or a secret
  manager over command-line literals. Never put a full key, access token, refresh
  token, password, or device code in chat, logs, screenshots, source control, or
  an unprotected temporary file.
- Show at most a short masked prefix when confirming a credential. Do not print
  raw JSON from key-list or key-create responses because some deployments return
  the secret in `data.key`.
- Never use a web password, OAuth/session JWT, refresh token, or `device_code` as
  `GPT88_API_KEY`.
- The user completes recharge, checkout, payout, and other payment flows in the
  browser. Do not handle card data, wallet secrets, or payment credentials.
- Read-only requests can proceed when the user supplied the required credential.
  Ask for explicit confirmation immediately before creating, updating, deleting,
  revoking, redeeming, transferring, or otherwise changing keys, account data,
  quota, or money.

## Choose the workflow

1. For OpenAI SDK, Cursor, Codex, Cline, Claude Code, or another external
   client, use an existing GPT88 API key or create one through the account
   console. Start with the model-list probe below.
2. For GPT88 TUI/Desktop/Harness, use the product's built-in GPT88 login. Do
   not ask the user to paste a cloud key when the product supports account login.
3. Use the device flow in the next section only when the user explicitly needs
   an automated or custom client authorization flow.

## Obtain and validate an API key

The normal path is to sign in or register at the account console, open its API
Keys page, create a key for the target application, and copy the secret once.
Use a separate key per application where possible, and apply a group, expiry,
quota, or rate limit when the user requests one.

If the user explicitly asks the agent to create a key through the API, first
read the available groups when a group is needed, then state the exact name,
group, expiry, quota, and rate limits and obtain confirmation. The authenticated
endpoint is:

```bash
curl -sS --max-redirs 0 -X POST "$GPT88_ACCOUNT_BASE/api/v1/keys" \
  -H "Authorization: Bearer $GPT88_ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"my-project"}'
```

The response is normally an envelope with the key object under `data`. Save the
secret directly into the secure destination requested by the user, then discard
the raw response. Do not create a key merely to test connectivity.

Validate the key against the gateway before configuring a client:

```bash
export GPT88_API_KEY='paste-the-key-in-a-secret-aware-shell'

curl -sS --fail-with-body --max-redirs 0 \
  "$GPT88_API_BASE/models" \
  -H "Authorization: Bearer $GPT88_API_KEY"
```

If the client only supports an `x-api-key` header, use that header instead of
Bearer authentication only when its integration requires it. Do not send the
same secret to unrelated fallback hosts.

## Device authorization for custom clients

The compatibility device flow is hosted on the account origin. It returns a
short user code for the browser and a private device code for polling:

```bash
curl -sS --max-redirs 0 -X POST \
  "$GPT88_ACCOUNT_BASE/api/v1/auth/device/start" \
  -H 'Content-Type: application/json' \
  -d '{}'
```

The successful response is an envelope whose `data` contains:

```json
{
  "device_code": "<private>",
  "user_code": "ABCD2345",
  "verification_uri": "https://agent.gpt88.cc/auth/device",
  "verification_uri_complete": "https://agent.gpt88.cc/auth/device?code=ABCD2345",
  "expires_in": 600,
  "interval": 2
}
```

Show the user only `verification_uri_complete` (or the URI plus `user_code`) and
ask them to sign in and approve the matching code in their browser. Keep the
`device_code` private and poll no faster than the returned `interval`:

```bash
curl -sS --max-redirs 0 \
  "$GPT88_ACCOUNT_BASE/api/v1/auth/device/status?device_code=$DEVICE_CODE"
```

Interpret the response as follows:

- `status: "pending"`: wait the requested interval and poll again.
- `status: "approved"`: consume the result once and store the returned account
  token securely. The result can include `auth`, `access_token`, `base_url`, and
  `model`.
- `status: "expired"`, an invalid-code error, or a second read after approval:
  start a new flow. The approval result is one-shot and expires after about ten
  minutes in the current service.

This flow grants account authorization; it does not guarantee that an inference
API key is returned. If no `api_key` is present, use the console or the
confirmed `POST /api/v1/keys` flow to obtain one. Never pass the account JWT to
`$GPT88_API_BASE/models` as if it were a model key.

The first-party TUI uses a newer RFC 8628 surface,
`POST /v1/auth/device/authorizations` followed by
`POST /v1/auth/device/token`, and has its own client name, scopes, and storage
rules. Use the built-in TUI/Desktop login for that product. Do not improvise
those scopes or reuse a first-party refresh token in a generic client.

## Discover models and prices

Never hardcode a model list or price. Query the live public model square before
choosing a model:

```bash
# For a self-hosted API base, derive the public catalog origin unless it was
# explicitly overridden.
# `GPT88_API_ORIGIN` is derived from `GPT88_API_BASE` above unless overridden.
curl -sS --max-redirs 0 \
  "$GPT88_API_ORIGIN/api/v1/model-square/public"
```

The response is usually an envelope with groups, models, pricing, and group
倍率 fields. Present the current candidates that match the user's capability,
latency, or cost preference and let the user choose when there is more than one
reasonable option. Prices, availability, group multipliers, and billing rules
can change.

For an authenticated account's active cloud keys and their visible model
catalog, use the secret-free catalog endpoint:

```bash
curl -sS --max-redirs 0 \
  "$GPT88_ACCOUNT_BASE/api/v1/key-catalog" \
  -H "Authorization: Bearer $GPT88_ACCESS_TOKEN"
```

For the exact models accepted by a particular inference key, use:

```bash
curl -sS --max-redirs 0 \
  "$GPT88_API_BASE/models" \
  -H "Authorization: Bearer $GPT88_API_KEY"
```

Use the returned model `id` exactly. Do not infer aliases from a provider's
official model name, and do not assume every model supports tools, vision,
streaming, long context, or media generation.

## Configure clients

For OpenAI SDKs, Cursor, Codex, Cline, and other OpenAI-compatible clients:

```bash
export OPENAI_BASE_URL="https://api.gpt88.cc/v1"
export OPENAI_API_KEY="$GPT88_API_KEY"
export GPT88_MODEL='copy-an-id-from-the-live-model-list'
```

The first smoke test should be a short text request:

```bash
curl -sS --max-redirs 0 \
  "$OPENAI_BASE_URL/chat/completions" \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{"model":"'"$GPT88_MODEL"'","messages":[{"role":"user","content":"Reply with: connection ok"}],"max_tokens":16}'
```

The same API base also exposes the OpenAI Responses surface when the selected
client and model support it. Add streaming, tools, images, or long context only
after the plain-text request succeeds.

For Claude Code, use the Anthropic-compatible setting expected by the installed
version and keep the same GPT88 API key:

```bash
export ANTHROPIC_BASE_URL="https://api.gpt88.cc/v1"
# Current releases commonly use this name:
export ANTHROPIC_AUTH_TOKEN="$GPT88_API_KEY"
# Older releases may require ANTHROPIC_API_KEY instead. Use the variable
# documented by the installed Claude Code version, not both by default.
```

For Cocode/GPT88 TUI/Desktop, prefer the native account flow (`/login`, then
`/use gpt88`). Use BYOK and `GPT88_API_KEY` only when the user explicitly chose
their own key. Native clients may resolve a cloud key by ID/name and therefore
must not be forced to receive the cloud secret.

## Account, usage, and key management

Use `Authorization: Bearer $GPT88_ACCESS_TOKEN` for account endpoints. Start
with read-only calls:

```bash
AUTH_HEADER="Authorization: Bearer $GPT88_ACCESS_TOKEN"

curl -sS --max-redirs 0 "$GPT88_ACCOUNT_BASE/api/v1/auth/me" -H "$AUTH_HEADER"
curl -sS --max-redirs 0 "$GPT88_ACCOUNT_BASE/api/v1/user/profile" -H "$AUTH_HEADER"
curl -sS --max-redirs 0 "$GPT88_ACCOUNT_BASE/api/v1/user/platform-quotas" -H "$AUTH_HEADER"
curl -sS --max-redirs 0 "$GPT88_ACCOUNT_BASE/api/v1/usage/stats" -H "$AUTH_HEADER"
curl -sS --max-redirs 0 "$GPT88_ACCOUNT_BASE/api/v1/subscriptions/summary" -H "$AUTH_HEADER"
```

Useful authenticated resources include:

- `GET /api/v1/keys`: list user keys. Treat the response as sensitive because a
  deployment may include a secret field.
- `GET /api/v1/groups/available` and `GET /api/v1/groups/rates`: discover the
  groups and rates available to the account before creating a restricted key.
- `GET /api/v1/key-catalog`: inspect active keys and visible models without key
  secrets.
- `GET /api/v1/usage`, `/api/v1/usage/errors`, and `/api/v1/usage/stats`:
  investigate usage and failed requests.
- `GET /api/v1/subscriptions`, `/active`, `/progress`, and `/summary`:
  inspect subscription state.

Mutating endpoints include `POST/PUT/DELETE /api/v1/keys`, `POST /api/v1/redeem`,
and `POST /api/v1/user/aff/transfer`. Before any such request, state the exact
target and payload, get confirmation, execute once, and follow with a read-only
verification. For key creation, capture the one-time secret directly into the
requested secure store and show only a mask. For deletion, revocation, quota
changes, referral transfers, and redemption, confirmation is required even if
the request appears reversible.

Token refresh and revocation routes differ between the current account service
and older compatibility deployments. Use the route documented by the client
implementation, commonly `/v1/auth/token/refresh` and `/v1/auth/token/revoke`
for first-party clients, rather than blindly replaying both route families.

## Media and long-running requests

After text is working and the live catalog confirms capability, GPT88 may expose
OpenAI-compatible image and video routes such as:

- `POST /v1/images/generations` and `/v1/images/edits`;
- `POST /v1/images/generations/async` or `/v1/images/edits/async`, then
  `GET /v1/images/tasks/{task_id}`;
- `POST /v1/videos` or `/v1/videos/generations`, then the matching
  `GET /v1/videos/{request_id}` and content route.

Use the exact route and parameters supported by the selected model. Async jobs
can incur cost even when polling fails, so preserve the task ID and inspect the
account usage record instead of submitting duplicates.

## Troubleshooting

- `401` or `API_KEY_REQUIRED`: check the API host, `/v1` placement, Bearer
  syntax, and that the value is an inference API key rather than a login token.
- `403`: the key may be inactive, expired, restricted to another group, or the
  account token may lack the requested scope. Re-read the key catalog and
  account state.
- `404`: use `https://api.gpt88.cc/v1` as the Base URL and let the client append
  the resource path. Do not use the account console URL for inference or append
  `/chat/completions` to the Base URL.
- `400` or `model_not_found`: query the live model list and copy the exact
  `id`; model parameters and tool support are not universal.
- `402`, balance, quota, or group errors: inspect the account and key limits;
  the user must complete any recharge in the browser.
- `429`: reduce concurrency and use bounded exponential backoff. Do not create
  replacement keys in a retry loop.
- `5xx` or timeout: record timestamp, HTTP status, model, and request ID without
  the key; retry only within the client's bounded policy.
- Device flow expired or was consumed: start a fresh authorization and do not
  reuse the old device code.

## Completion checklist

Before declaring GPT88 connected, verify all of the following:

1. The account origin and API Base URL are the intended HTTPS hosts, with `/v1`
   present exactly once on the API base.
2. The client has an inference API key, not an account password or access token.
3. `GET /models` succeeds with that key and the selected model came from the
   live catalog.
4. One short text request succeeds before advanced features are enabled.
5. The final report contains only masked credential prefixes and no raw response
   containing secret fields.
