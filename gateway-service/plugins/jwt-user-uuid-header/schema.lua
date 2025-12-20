local typedefs = require "kong.db.schema.typedefs"

return {
  name = "jwt-user-uuid-header",
  fields = {
    { consumer  = typedefs.no_consumer },
    { protocols = typedefs.protocols_http },
    {
      config = {
        type = "record",
        fields = {
          -- No configurable fields for now; this plugin has a fixed behaviour:
          -- read user_uuid-like claim from an already verified JWT and
          -- forward it as X-User-UUID.
        },
      },
    },
  },
}

