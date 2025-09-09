local claims = {
  email_verified: false,
} + std.extVar('claims');

{
  identity: {
    traits: {
      email: claims.email,
      username: std.strReplace(claims.email, '@', '_at_'),
      first_name: if std.objectHas(claims, 'given_name') then claims.given_name else '',
      last_name: if std.objectHas(claims, 'family_name') then claims.family_name else '',
      profile_image: if std.objectHas(claims, 'picture') then claims.picture else '',
      age: 18,
    },
  },
}
