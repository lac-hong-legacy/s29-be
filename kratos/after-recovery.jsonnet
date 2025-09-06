function(ctx) {
  identity: {
    id: ctx.identity.id,
    traits: {
      email: ctx.identity.traits.email,
      username: ctx.identity.traits.username,
      age: ctx.identity.traits.age,

    },
    schema_id: ctx.identity.schema_id,
    state: ctx.identity.state
  },
  recovery_info: {
    flow_id: ctx.flow.id,
    recovery_method: "code",
    flow_type: ctx.flow.type
  }
}