# Mobile Social Login Implementation Guide

This guide explains how to implement social login with Google and Facebook for mobile applications using the S29 backend.

## Overview

The implementation includes:
- OAuth flow initiation for mobile apps
- Google and Facebook authentication via Kratos OIDC
- JWT token generation after successful OAuth
- Mobile-friendly redirect handling

## Required Environment Variables

Add these to your `.env` file or docker environment:

```bash
# Google OAuth
GOOGLE_OIDC_CLIENT_ID=your_google_client_id
GOOGLE_OIDC_CLIENT_SECRET=your_google_client_secret

# Facebook OAuth  
FACEBOOK_OIDC_CLIENT_ID=your_facebook_app_id
FACEBOOK_OIDC_CLIENT_SECRET=your_facebook_app_secret

# Kratos URLs
KRATOS_PUBLIC_URL=http://localhost:4433
KRATOS_ADMIN_URL=http://localhost:4434
```

## API Endpoints

### 1. Initiate OAuth Flow

**POST** `/api/v1/auth/oauth/init`

Request:
```json
{
  "provider": "google",  // or "facebook"
  "redirect_uri": "s29app://oauth/callback",  // optional, defaults to s29app://oauth/callback
  "state": "optional_state_parameter"
}
```

Response:
```json
{
  "success": true,
  "data": {
    "auth_url": "https://accounts.google.com/oauth/authorize?...",
    "state": "generated_state_parameter"
  }
}
```

### 2. Handle OAuth Callback (Redirect)

**GET** `/api/v1/auth/oauth/{provider}/callback?code=auth_code&state=state`

This endpoint automatically redirects to: `s29app://auth/success?access_token=...&token_type=Bearer&expires_in=86400`

### 3. Complete OAuth Login (JSON Response)

**POST** `/api/v1/auth/oauth/{provider}/complete`

Request:
```json
{
  "code": "authorization_code_from_oauth",
  "state": "state_parameter"
}
```

Response:
```json
{
  "success": true,
  "data": {
    "access_token": "jwt_token_here",
    "token_type": "Bearer",
    "expires_in": 86400,
    "user": {
      "id": "user_uuid",
      "kratos_identity_id": "kratos_uuid",
      "email": "user@example.com",
      "is_active": true
    }
  }
}
```

## Mobile App Integration

### Android Example

```kotlin
// Step 1: Initiate OAuth
val request = OAuthInitRequest(
    provider = "google",
    redirectUri = "s29app://oauth/callback"
)

// Call API to get auth URL
val response = apiClient.initiateOAuth(request)

// Step 2: Open browser with auth URL
val intent = Intent(Intent.ACTION_VIEW, Uri.parse(response.authUrl))
startActivity(intent)

// Step 3: Handle deep link callback
// In your Activity that handles s29app://auth/success
override fun onCreate(savedInstanceState: Bundle?) {
    super.onCreate(savedInstanceState)
    
    val uri = intent.data
    if (uri?.scheme == "s29app" && uri.host == "auth") {
        val accessToken = uri.getQueryParameter("access_token")
        val tokenType = uri.getQueryParameter("token_type")
        val expiresIn = uri.getQueryParameter("expires_in")
        
        // Save tokens and proceed to app
        saveTokens(accessToken, tokenType, expiresIn)
    }
}
```

### iOS Example

```swift
// Step 1: Initiate OAuth
let request = OAuthInitRequest(
    provider: "google",
    redirectUri: "s29app://oauth/callback"
)

// Call API to get auth URL
apiClient.initiateOAuth(request) { result in
    switch result {
    case .success(let response):
        // Step 2: Open Safari with auth URL
        let url = URL(string: response.authUrl)!
        UIApplication.shared.open(url)
    case .failure(let error):
        // Handle error
    }
}

// Step 3: Handle URL scheme in AppDelegate
func application(_ app: UIApplication, open url: URL, options: [UIApplication.OpenURLOptionsKey : Any] = [:]) -> Bool {
    if url.scheme == "s29app" && url.host == "auth" {
        let components = URLComponents(url: url, resolvingAgainstBaseURL: false)
        let accessToken = components?.queryItems?.first { $0.name == "access_token" }?.value
        let tokenType = components?.queryItems?.first { $0.name == "token_type" }?.value
        let expiresIn = components?.queryItems?.first { $0.name == "expires_in" }?.value
        
        // Save tokens and proceed to app
        saveTokens(accessToken: accessToken, tokenType: tokenType, expiresIn: expiresIn)
        return true
    }
    return false
}
```

## Setup Instructions

### 1. Google OAuth Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing one
3. Enable Google+ API
4. Go to Credentials → Create OAuth 2.0 Client ID
5. For mobile apps, select "iOS" or "Android" application type
6. Add your app's bundle ID (iOS) or package name and SHA-1 (Android)
7. For web redirect testing, also create a "Web application" client
8. Add redirect URIs:
   - `s29app://oauth/callback` 
   - `http://localhost:8080/api/v1/auth/oauth/callback`

### 2. Facebook OAuth Setup

1. Go to [Facebook Developers](https://developers.facebook.com/)
2. Create a new app
3. Add Facebook Login product
4. In Facebook Login settings:
   - Add `s29app://oauth/callback` to Valid OAuth Redirect URIs
   - Add `http://localhost:8080/api/v1/auth/oauth/callback` for testing
5. Get App ID and App Secret

### 3. Mobile App Configuration

#### Android

Add to `AndroidManifest.xml`:
```xml
<activity android:name=".AuthCallbackActivity">
    <intent-filter>
        <action android:name="android.intent.action.VIEW" />
        <category android:name="android.intent.category.DEFAULT" />
        <category android:name="android.intent.category.BROWSABLE" />
        <data android:scheme="s29app" android:host="auth" />
    </intent-filter>
</activity>
```

#### iOS

Add to `Info.plist`:
```xml
<key>CFBundleURLTypes</key>
<array>
    <dict>
        <key>CFBundleURLName</key>
        <string>s29app</string>
        <key>CFBundleURLSchemes</key>
        <array>
            <string>s29app</string>
        </array>
    </dict>
</array>
```

## Testing

1. Start your backend services:
   ```bash
   docker-compose up
   ```

2. Test OAuth initiation:
   ```bash
   curl -X POST http://localhost:8080/api/v1/auth/oauth/init \
     -H "Content-Type: application/json" \
     -d '{"provider": "google"}'
   ```

3. Open the returned `auth_url` in a browser to test the full flow

## Security Notes

- Always validate the `state` parameter to prevent CSRF attacks
- Use HTTPS in production
- Set appropriate redirect URI restrictions in OAuth provider settings
- Implement proper token storage and refresh logic in mobile apps
- Consider implementing token refresh endpoints for long-lived sessions

## Troubleshooting

### Common Issues

1. **Invalid redirect URI**: Ensure your mobile app's custom scheme is registered with OAuth providers
2. **CORS issues**: The backend is configured to allow CORS, but ensure your OAuth providers whitelist your domains
3. **Token validation errors**: Check that Kratos is properly configured and running
4. **Deep link not working**: Verify mobile app manifest/plist configuration

### Logs

Check Kratos logs for OAuth flow issues:
```bash
docker logs s29-kratos
```

Check backend logs:
```bash
docker logs s29-api
```
