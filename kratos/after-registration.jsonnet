function(ctx) {
  identity: {
    id: ctx.identity.id,
    traits: {
      email: ctx.identity.traits.email,
      display_name: ctx.identity.traits.display_name,
      year_of_birth: ctx.identity.traits.year_of_birth,
    },
    schema_id: ctx.identity.schema_id,
    state: ctx.identity.state
  }
}
