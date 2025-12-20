local cjson = require "cjson.safe"

local JwtUserUuidHeader = {
  -- Priority lower than Kong's builtin jwt plugin (~1005) so we run after it.
  PRIORITY = 900,
  VERSION = "0.1",
}

-- decode_jwt_claims decodes the payload part of a JWT without verifying
-- the signature. We rely on Kong's jwt plugin to have already performed
-- signature and exp validation before this plugin runs.
local function decode_jwt_claims(token)
  if not token or token == "" then
    return nil
  end
  local header_b64, payload_b64 = token:match("([^%.]+)%.([^%.]+)%.[^%.]+")
  if not header_b64 or not payload_b64 then
    return nil
  end

  -- JWT 使用 base64url，需要先替换字符再解码。
  local function b64url_decode(input)
    input = input:gsub("-", "+"):gsub("_", "/")
    local pad = #input % 4
    if pad == 2 then
      input = input .. "=="
    elseif pad == 3 then
      input = input .. "="
    elseif pad ~= 0 then
      return nil
    end
    return ngx.decode_base64(input)
  end

  local payload_json = b64url_decode(payload_b64)
  if not payload_json then
    return nil
  end
  return cjson.decode(payload_json)
end

-- This plugin relies on the builtin jwt plugin having already:
--   1) Verified the JWT signature using the configured RSA public key.
--   2) Validated standard claims such as exp.
-- The verified token (including claims) is exposed via
-- kong.ctx.shared.authenticated_jwt_token. We only read the user_uuid-like
-- claim from there and forward it to upstream as X-User-UUID so that backend
-- services do not need to parse JWT again.
function JwtUserUuidHeader:access(conf)
  -- 1) 首选：从 jwt 插件注入的 authenticated_jwt_token 读取 claims。
  local shared = kong.ctx.shared
  local claims
  if shared then
    local token = shared.authenticated_jwt_token or shared.jwt_token
    if token and token.claims then
      claims = token.claims
    end
  end

  -- 2) 兜底：手工解析 Authorization 里的 JWT payload（不做验签）。
  if not claims then
    local auth_header = kong.request.get_header("authorization") or kong.request.get_header("Authorization")
    if auth_header and type(auth_header) == "string" then
      local prefix = auth_header:sub(1, 7)
      if prefix:lower() == "bearer " then
        local jwt = auth_header:sub(8)
        claims = decode_jwt_claims(jwt)
      end
    end
  end

  if not claims then
    return
  end

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
