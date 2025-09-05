function(ctx) {
  local traits = ctx.identity.traits;

  identity: {
    id: ctx.identity.id,
    traits: {
      email: traits.email,
      display_name: traits.display_name,

      year_of_birth: traits.year_of_birth,

      [if std.objectHas(traits, 'birth_date') then 'birth_date']: traits.birth_date,

      [if std.objectHas(traits, 'bio') then 'bio']: traits.bio,
      [if std.objectHas(traits, 'location') then 'location']: traits.location,
      [if std.objectHas(traits, 'profile_image') then 'profile_image']: traits.profile_image,
    },
    schema_id: ctx.identity.schema_id,
    state: ctx.identity.state
  }
}
