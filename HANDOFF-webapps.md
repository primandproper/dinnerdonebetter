# Handoff: web apps on `@primandproper/platform-client`

Temporary. Delete this file in the PR that lands the work it describes.

## Where things stand

The consumer and admin web apps have been moved off the retired `AuthService`. Every call
they make now goes through `@primandproper/platform-client`.

- **Branch**: `webapps-on-platform-client`, cut from `main` at `643b29d97` (#1443).
  - Its two commits are the `platform-client-prototype` branch, replayed onto `main`.
  - **The rest of the work is uncommitted in the working tree.** Check `git status` before doing
    anything else, and don't switch branches or stash it away.
- **Client**: [primandproper/platform-client-ts#40](https://github.com/primandproper/platform-client-ts/pull/40)
  re-pins the client to platform-go v14.2.0. It adds passkey sign-in, `Session.switchAccount`,
  and the v14.2.0 refusal reasons. The apps need all of it.
- **The blocker**: the apps install the client from a local tarball:
  `file:../../../platform-client-ts/primandproper-platform-client-0.0.0.tgz`.
  - It is referenced from `frontend/consumer/package.json`, `frontend/admin/package.json` and
    `frontend/packages/api-client/package.json`.
  - CI can't resolve it, so the frontend workflows this branch re-enables fail until a published
    version replaces it ([platform-client-ts#29](https://github.com/primandproper/platform-client-ts/issues/29)).

## What's done on this branch

### Consumer

- Sign-in uses `signIn` and sign-out uses `signOut`.
- Passkey sign-in uses `beginPasskeySignIn` and `passkeySignIn`, through `PasskeysService`.
  The login page handles the second-factor retry.
- Passkey enrollment, the list and removal use `PasskeysService`.
- `/account/sessions` uses `ListSignIns`, `EndSignIn` and `EndOtherSignIns`.
- Reset links use `requestPasswordReset`, `verifyPasswordResetToken` and `completePasswordReset`.
  The link is checked before the form renders.
- `verifyEmailAddress`, `GetSelf` and `GetPrincipal` replace the old self and active-account
  calls.
- `UpdateUsername` now says why a change was refused.

### Admin

- The raw `locals.accessToken` is replaced by a `Session` per request over the cookie, as in the
  consumer. The 37 page files take `locals.session`.
- Operator sign-in uses `adminSignIn`, and passkey sign-in works as it does in the consumer.
- `/users/[id]/sessions` uses `SignInAdministrationService`.

### `frontend/packages/api-client`

- `admin-clients.ts` and `createPlatformClient` are deleted, along with the `auth/*` exports.
- It exports `InternalOperationsService`.

### Repo

- The root `build`, `lint`, `test` and `format` targets run the frontend again. iOS stays out.
- The four `frontend_*` workflows have their `pull_request` triggers back.
- `PROTO_TS_HANDWRITTEN` in the `Makefile` is updated.
- `docs/auth-flow.md` and `docs/identity.md` describe the new flow.

## Once a version of the client is published

1. **Merge #40 and publish.** Confirm with `npm view @primandproper/platform-client versions`.
2. **Point the three `package.json` files at the version.** Pin it exactly, as the repo pins
   everything (the `pin:check` lint step enforces this).
   - `frontend/consumer/package.json`
   - `frontend/admin/package.json`
   - `frontend/packages/api-client/package.json`
   - If it went to GitHub Packages rather than public npm, add the `@primandproper` registry and
     an auth token to `frontend/.npmrc`. The frontend Dockerfiles and the CI workflows need the
     same token.
3. **Reinstall**: `cd frontend && npm install`. This rewrites the lockfile entry; the tarball's
   `integrity` was deleted from it by hand.
4. **Check everything** from `frontend/`:
   - `npm run check` must show 0 errors. 33 and 8 warnings are the baseline from `main`.
   - `npm run lint -w consumer -w admin`
   - `npm test`
   - `npm run format:check`
   - `npm run build`
   - Once committed, also run the full `npm run lint`. Its pin check diffs the working tree, so
     it always fails on uncommitted work.
   - From the repo root, run `make proto_typescript` and confirm it leaves no diff under
     `frontend/packages/api-client`.
5. **If the published version differs from #40's head**, rerun the local verification below,
   at least sign-in and the account pages.
6. **Commit in two parts and open the PR**:
   - the consumer/admin migration;
   - "restore the frontend to the build and CI".

   Delete this file in the same PR.

## Local verification recipe

1. **Backend**: from `backend/`, run `go run ./cmd/localdev/server`. It needs Docker for its
   Postgres. gRPC is on `:8001`.
   - Seeded logins:
     - `admin_user` / `admin_pass`, an operator;
     - `member_user_1` / `member_pass_1`;
     - `member_user_2` / `member_pass_2`.
   - All three have a proven TOTP secret of base32 all-zeros. A current code is
     `python3 -c "import base64,hmac,hashlib,struct,time;k=base64.b32decode('A'*103+'=');h=hmac.new(k,struct.pack('>Q',int(time.time())//30),hashlib.sha1).digest();o=h[-1]&15;print('%06d'%((struct.unpack('>I',h[o:o+4])[0]&0x7fffffff)%1000000))"`.
2. **Each app**: from `frontend/consumer` or `frontend/admin`, run
   `GRPC_API_SERVER_URL=localhost:8001 DEVELOPING_LOCALLY=true COOKIE_ENCRYPTION_KEY=$(openssl rand -base64 32) npx vite dev --port <port>`.
3. **Passkeys** only verify from an origin the relying party allows: `http://localhost` or
   `http://localhost:8888` (`backend/internal/branding`). Run the app under test on `8888`.
4. **Form actions** with curl need `Origin: <app url>`, `accept: application/json` and a
   form-encoded body.
5. **Passkey ceremonies**: Playwright with a CDP virtual authenticator works.
   - Run `npx playwright install chromium`.
   - Use `WebAuthn.addVirtualAuthenticator` with `ctap2`, `internal`, `hasResidentKey`,
     `hasUserVerification` and `isUserVerified`.
   - Wait for hydration before clicking.
   - A key without user verification can't do this server's usernameless login in Chrome. Its
     `BeginLogin` never names credentials, so R19's code-then-tap-again branch can't be driven
     from a browser.

## Left for later

- **Terraform** still provisions OAuth2 client secrets for both web apps, and neither app reads
  them now:
  - `ADMIN_WEBAPP_OAUTH2_CLIENT_*` and `CONSUMER_WEBAPP_OAUTH2_CLIENT_*`;
  - in `backend/deploy/environments/prod/terraform/` and
    `docs/required-secrets-and-variables.md`.
  - Removing them also means removing the Terraform Cloud workspace variables.
- **Admin's meal-planning calls** still take untyped request objects (`loose` in
  `frontend/admin/src/lib/grpc/clients.ts`), as they did before.
- **More than one replica** of either app needs a `SharedExchangeCoordinator`. Today each process
  uses the in-memory one.
- **The iOS app** is still on the retired `AuthService` and out of the build. `platform-client-swift`
  is its equivalent of this work.
