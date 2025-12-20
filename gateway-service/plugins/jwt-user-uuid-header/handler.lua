local JwtUserUuidHeader = {
  -- Priority lower than Kong's builtin jwt plugin (~1005) so we run after it.
  PRIORITY = 900,
  VERSION = "0.1",
}

-- This plugin relies on the builtin jwt plugin having already:
--   1) Verified the JWT signature using the configured RSA public key.
--   2) Validated standard claims such as exp.
-- The verified token (including claims) is exposed via
-- kong.ctx.shared.authenticated_jwt_token. We only read the user_uuid-like
-- claim from there and forward it to upstream as X-User-UUID so that backend
-- services do not need to parse JWT again.
function JwtUserUuidHeader:access(conf)
  local token = kong.ctx.shared.authenticated_jwt_token
  if not token or not token.claims then
    return
  end

  local claims = token.claims

  -- Try several common claim keys; adjust here if your JWT uses a different key.
  local user_uuid = claims.user_uuid or claims.UserUUID or claims.sub
  if not user_uuid or user_uuid == "" then
    return
  end

  -- Set/override upstream header; this means client-supplied X-User-UUID
  -- cannot override the value derived from the verified JWT.
  kong.service.request.set_header("X-User-UUID", user_uuid)
end

return JwtUserUuidHeader

