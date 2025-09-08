local claims = {
  email_verified: false,
} + std.extVar('claims');

{
  identity: {
    traits: {
      email: claims.email,
      username: std.strReplace(claims.email, '@', '_at_'),
      first_name: if std.objectHas(claims, 'first_name') then claims.first_name else '',
      last_name: if std.objectHas(claims, 'last_name') then claims.last_name else '',
      profile_image: if std.objectHas(claims, 'picture') && std.objectHas(claims.picture, 'data') && std.objectHas(claims.picture.data, 'url') then claims.picture.data.url else '',
      age: 18,
    },
  },
}
