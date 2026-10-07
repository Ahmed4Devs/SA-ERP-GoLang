# User administration and authorization scope

The **Authorization** page provides user creation, search/pagination, activation,
suspension, password reset, role assignment, and assignment history in Arabic and
English. User identities belong to the dedicated customer installation. Only a
global `iam.users.view` or `iam.users.manage` grant permits administration;
selecting an entity cannot upgrade an entity administrator into a global one.
Role administration similarly requires global `authz.view` / `authz.manage`.

## API

- `GET /api/v1/auth/users?search=&limit=50&offset=0`
- `POST /api/v1/auth/users`
- `POST /api/v1/auth/users/{id}/status`
- `POST /api/v1/auth/users/{id}/password`
- `GET /api/v1/authz/assignments?user_id=...`

Status and password commands require the current `expected_version`. A concurrent
change returns HTTP 409; reload before retrying. Passwords never appear in user
responses. Unknown request fields and extra JSON values are rejected. Commands
write a security event in the same transaction; failure of the audit insert rolls
back user creation, status changes, and password changes.

A user cannot deactivate their own account. The last active global authorization
administrator cannot be deactivated or lose the final administration grant.
Permission checks query the database on every request, so committed revocation
is effective across API instances without waiting for a local cache to expire.

## Sessions and MFA

Password reset and account suspension/deactivation invalidate stored sessions
and pending MFA login challenges. Session/challenge issuance takes the same user
row lock as administrative changes and checks the user version observed during
password/MFA verification. An old password check running concurrently with reset
cannot create a session afterward; a completed MFA proof also becomes invalid if
reset occurs before session issuance. Logout-all advances this version too.

MFA activation also advances the authentication version so earlier password-only
verification cannot issue a session afterward. Concurrent MFA guesses share one
locked five-attempt budget; failed attempts commit before the next guess runs.
Logout endpoints report database failures instead of claiming successful revocation.

Password reset preserves enrolled MFA factors. The user must authenticate with
the replacement password and their existing factor/recovery code. Recovery codes
are not returned by user administration.

## Branch scope and upgrades

A branch grant only authorizes an explicit request for that branch and its legal
entity. It grants neither entity-wide nor global access. Endpoints that operate
at legal-entity scope require an entity/global grant, even if the user has a
branch grant. The current UI lists those entity capabilities accordingly.

Migration 330 requires branch assignments to include the matching legal entity
and enforces the pair with a composite foreign key. It intentionally fails on
legacy invalid assignments. Review and correct the authorization data before
upgrading; do not transform a branch grant into a broader entity grant.

Migration 340 applies the same ownership rule to journal branches, departments,
and cost centers. Review invalid historical dimensions with the responsible
accountant; the migration does not rewrite posted journals. Refresh runtime
grants after applying migrations, as described in `database-roles.md`.

Migration 350 locks both journal parents while editing/reassigning lines, so a
concurrent posting cannot slip past the immutability check. A reversal may add
its metadata, but cannot rewrite the original posted amounts or source.
