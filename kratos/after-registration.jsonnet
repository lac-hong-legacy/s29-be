function(ctx) {
  identity: {
    id: ctx.identity.id,
    traits: {
      email: ctx.identity.traits.email,
      display_name: ctx.identity.traits.display_name,
      // Handle optional fields that might be present
      [if std.objectHas(ctx.identity.traits, 'first_name') then 'first_name']: ctx.identity.traits.first_name,
      [if std.objectHas(ctx.identity.traits, 'last_name') then 'last_name']: ctx.identity.traits.last_name,
      // user_type is required, but provide default if missing
      user_type: if std.objectHas(ctx.identity.traits, 'user_type') && 
                     (ctx.identity.traits.user_type == 'listener' || ctx.identity.traits.user_type == 'artist')
                  then ctx.identity.traits.user_type 
                  else 'listener',
      // artist_name is required when user_type is 'artist'
      [if std.objectHas(ctx.identity.traits, 'user_type') && 
          ctx.identity.traits.user_type == 'artist' && 
          std.objectHas(ctx.identity.traits, 'artist_name') then 'artist_name']: 
        ctx.identity.traits.artist_name,
      // Handle other optional fields
      [if std.objectHas(ctx.identity.traits, 'bio') then 'bio']: ctx.identity.traits.bio,
      [if std.objectHas(ctx.identity.traits, 'location') then 'location']: ctx.identity.traits.location,
      [if std.objectHas(ctx.identity.traits, 'profile_image') then 'profile_image']: ctx.identity.traits.profile_image,
      [if std.objectHas(ctx.identity.traits, 'preferences') then 'preferences']: ctx.identity.traits.preferences
    },
    schema_id: ctx.identity.schema_id,
    state: ctx.identity.state
  }
}