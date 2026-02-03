# SSO System - Flowcharts

## 1. Complete OAuth2 Authorization Flow

```mermaid
sequenceDiagram
    actor User
    participant ClientApp as SSO Client Application
    participant Browser
    participant SSOServer as SSO Server
    participant DB as Database
    participant Email as Email Service
    
    User->>ClientApp: 1. Access protected resource
    ClientApp->>ClientApp: 2. Check authentication
    
    alt Not Authenticated
        ClientApp->>Browser: 3. Redirect to SSO Server
        Note over ClientApp,Browser: /oauth/authorize?client_id=xxx&redirect_uri=xxx&response_type=code&state=xxx
        
        Browser->>SSOServer: 4. GET /oauth/authorize
        SSOServer->>SSOServer: 5. Validate client_id & redirect_uri
        
        alt Valid Client
            SSOServer->>Browser: 6. Show login page
            Browser->>User: 7. Display login form
            
            User->>Browser: 8. Enter username & password
            Browser->>SSOServer: 9. POST /auth/login
            
            SSOServer->>DB: 10. Verify credentials
            DB-->>SSOServer: 11. User data
            
            alt Invalid Credentials
                SSOServer->>DB: 12. Increment failed_login_attempts
                
                alt Max Attempts Reached
                    SSOServer->>DB: 13. Lock account
                    SSOServer->>Email: 14. Send lockout notification
                    SSOServer-->>Browser: 15. Account locked error
                else
                    SSOServer-->>Browser: 16. Invalid credentials error
                end
            else Valid Credentials
                SSOServer->>DB: 17. Reset failed_login_attempts
                SSOServer->>DB: 18. Check if 2FA enabled
                
                alt 2FA Enabled
                    SSOServer-->>Browser: 19. Show 2FA form
                    Browser->>User: 20. Display 2FA input
                    
                    User->>Browser: 21. Enter TOTP code
                    Browser->>SSOServer: 22. POST /auth/verify-2fa
                    
                    SSOServer->>DB: 23. Get TOTP secret
                    SSOServer->>SSOServer: 24. Validate TOTP code
                    
                    alt Invalid 2FA Code
                        SSOServer-->>Browser: 25. Invalid 2FA error
                    else Valid 2FA Code
                        SSOServer->>SSOServer: 26. Proceed to authorization
                    end
                else No 2FA
                    SSOServer->>SSOServer: 27. Proceed to authorization
                end
                
                SSOServer->>DB: 28. Generate authorization code
                SSOServer->>DB: 29. Store code (10 min expiry)
                SSOServer->>DB: 30. Create session
                SSOServer->>DB: 31. Log audit: login_success
                SSOServer->>DB: 32. Update last_login_at
                
                SSOServer->>Browser: 33. Redirect to client with code
                Note over SSOServer,Browser: redirect_uri?code=xxx&state=xxx
                
                Browser->>ClientApp: 34. GET callback with code
                
                ClientApp->>SSOServer: 35. POST /oauth/token
                Note over ClientApp,SSOServer: grant_type=authorization_code&code=xxx&client_id=xxx&client_secret=xxx
                
                SSOServer->>DB: 36. Validate code
                SSOServer->>DB: 37. Mark code as used
                SSOServer->>SSOServer: 38. Generate access_token (JWT)
                SSOServer->>SSOServer: 39. Generate refresh_token
                SSOServer->>DB: 40. Store refresh_token
                
                SSOServer-->>ClientApp: 41. Return tokens
                Note over SSOServer,ClientApp: {access_token, refresh_token, expires_in, token_type}
                
                ClientApp->>ClientApp: 42. Store tokens securely
                
                ClientApp->>SSOServer: 43. GET /oauth/userinfo
                Note over ClientApp,SSOServer: Authorization: Bearer {access_token}
                
                SSOServer->>SSOServer: 44. Validate access_token
                SSOServer->>DB: 45. Get user profile
                SSOServer-->>ClientApp: 46. Return user info
                
                ClientApp->>User: 47. Grant access to resource
            end
        else Invalid Client
            SSOServer-->>Browser: Error: Invalid client
        end
    else Already Authenticated
        ClientApp->>User: Show resource
    end
```

## 2. Login Flow with 2FA

```mermaid
flowchart TD
    Start([User visits SSO Server]) --> CheckSession{Session<br/>exists?}
    
    CheckSession -->|Yes| CheckValid{Session<br/>valid?}
    CheckValid -->|Yes| Dashboard[Redirect to Dashboard]
    CheckValid -->|No| ShowLogin[Show Login Form]
    
    CheckSession -->|No| ShowLogin
    
    ShowLogin --> EnterCreds[User enters<br/>username & password]
    EnterCreds --> SubmitLogin[POST /auth/login]
    
    SubmitLogin --> ValidateCreds{Validate<br/>credentials}
    
    ValidateCreds -->|Invalid| IncrementFailed[Increment<br/>failed_login_attempts]
    IncrementFailed --> CheckLockout{Attempts >= 5?}
    
    CheckLockout -->|Yes| LockAccount[Lock account<br/>for 30 minutes]
    LockAccount --> SendEmail[Send lockout<br/>notification email]
    SendEmail --> ShowError1[Show account<br/>locked error]
    ShowError1 --> End1([End])
    
    CheckLockout -->|No| ShowError2[Show invalid<br/>credentials error]
    ShowError2 --> ShowLogin
    
    ValidateCreds -->|Valid| ResetAttempts[Reset<br/>failed_login_attempts]
    ResetAttempts --> Check2FA{2FA<br/>enabled?}
    
    Check2FA -->|No| CreateSession[Create session]
    CreateSession --> LogAudit[Log audit:<br/>login_success]
    LogAudit --> UpdateLastLogin[Update<br/>last_login_at]
    UpdateLastLogin --> Dashboard
    
    Check2FA -->|Yes| Show2FAForm[Show 2FA<br/>code input]
    Show2FAForm --> Enter2FA[User enters<br/>TOTP code]
    Enter2FA --> Submit2FA[POST /auth/verify-2fa]
    
    Submit2FA --> Validate2FA{Validate<br/>TOTP?}
    
    Validate2FA -->|Invalid| Increment2FAFailed[Increment 2FA<br/>failed attempts]
    Increment2FAFailed --> Check2FALimit{Attempts >= 3?}
    
    Check2FALimit -->|Yes| TempLock[Temporary lock<br/>15 minutes]
    TempLock --> ShowError3[Show too many<br/>attempts error]
    ShowError3 --> End2([End])
    
    Check2FALimit -->|No| ShowError4[Show invalid<br/>2FA code error]
    ShowError4 --> Show2FAForm
    
    Validate2FA -->|Valid| CreateSession
    
    Dashboard --> End3([End])
```

## 3. Forgot Password Flow

```mermaid
flowchart TD
    Start([User clicks<br/>Forgot Password]) --> ShowForm[Show email<br/>input form]
    
    ShowForm --> EnterEmail[User enters<br/>email address]
    EnterEmail --> Submit[POST /password/forgot]
    
    Submit --> CheckRateLimit{Rate limit<br/>check}
    
    CheckRateLimit -->|Exceeded| ShowError1[Show rate limit<br/>error message]
    ShowError1 --> End1([End])
    
    CheckRateLimit -->|OK| FindUser{User<br/>exists?}
    
    FindUser -->|No| ShowSuccess[Show generic<br/>success message]
    Note1[Security: Don't reveal<br/>if email exists]
    FindUser -.->|Security| Note1
    ShowSuccess --> End2([End])
    
    FindUser -->|Yes| CheckActive{Account<br/>active?}
    
    CheckActive -->|No| ShowSuccess
    
    CheckActive -->|Yes| GenerateToken[Generate secure<br/>reset token]
    GenerateToken --> StoreToken[Store token in DB<br/>with 1hr expiry]
    StoreToken --> CreateLink[Create reset link<br/>with token]
    CreateLink --> SendEmail[Send email with<br/>reset link]
    SendEmail --> ShowSuccess
    
    ShowSuccess --> UserChecksEmail[User checks email]
    UserChecksEmail --> ClickLink[User clicks<br/>reset link]
    ClickLink --> ValidateToken{Token<br/>valid?}
    
    ValidateToken -->|Expired| ShowError2[Show token<br/>expired error]
    ShowError2 --> OfferNewReset[Offer to send<br/>new reset link]
    OfferNewReset --> End3([End])
    
    ValidateToken -->|Used| ShowError3[Show token<br/>already used error]
    ShowError3 --> End4([End])
    
    ValidateToken -->|Invalid| ShowError4[Show invalid<br/>token error]
    ShowError4 --> End5([End])
    
    ValidateToken -->|Valid| ShowResetForm[Show reset<br/>password form]
    ShowResetForm --> EnterNewPass[User enters<br/>new password]
    EnterNewPass --> ValidatePassword{Meets password<br/>complexity?}
    
    ValidatePassword -->|No| ShowError5[Show password<br/>complexity error]
    ShowError5 --> ShowResetForm
    
    ValidatePassword -->|Yes| CheckHistory{Password in<br/>history?}
    
    CheckHistory -->|Yes| ShowError6[Show password<br/>reuse error]
    ShowError6 --> ShowResetForm
    
    CheckHistory -->|No| HashPassword[Hash new password]
    HashPassword --> UpdatePassword[Update password<br/>in database]
    UpdatePassword --> InvalidateToken[Invalidate<br/>reset token]
    InvalidateToken --> AddToHistory[Add to password<br/>history]
    AddToHistory --> InvalidateSessions[Invalidate all<br/>user sessions]
    InvalidateSessions --> LogAudit[Log audit:<br/>password_reset]
    LogAudit --> SendConfirmation[Send confirmation<br/>email]
    SendConfirmation --> ShowSuccess2[Show success<br/>message]
    ShowSuccess2 --> RedirectLogin[Redirect to<br/>login page]
    RedirectLogin --> End6([End])
```

## 4. Change Password Flow (Logged In User)

```mermaid
flowchart TD
    Start([User navigates to<br/>Change Password]) --> CheckAuth{User<br/>authenticated?}
    
    CheckAuth -->|No| RedirectLogin[Redirect to login]
    RedirectLogin --> End1([End])
    
    CheckAuth -->|Yes| ShowForm[Show change<br/>password form]
    
    ShowForm --> EnterPasswords[User enters:<br/>- Current password<br/>- New password<br/>- Confirm password]
    
    EnterPasswords --> Submit[POST /password/change]
    
    Submit --> ValidateCurrent{Validate<br/>current password}
    
    ValidateCurrent -->|Invalid| ShowError1[Show incorrect<br/>password error]
    ShowError1 --> ShowForm
    
    ValidateCurrent -->|Valid| CheckMatch{New password<br/>matches confirm?}
    
    CheckMatch -->|No| ShowError2[Show passwords<br/>don't match error]
    CheckError2 --> ShowForm
    
    CheckMatch -->|Yes| ValidateComplexity{Meets password<br/>complexity?}
    
    ValidateComplexity -->|No| ShowError3[Show complexity<br/>requirements error]
    ShowError3 --> ShowForm
    
    ValidateComplexity -->|Yes| CheckNotSame{Different from<br/>current password?}
    
    CheckNotSame -->|No| ShowError4[Show must be<br/>different error]
    ShowError4 --> ShowForm
    
    CheckNotSame -->|Yes| CheckHistory{Password in<br/>history?}
    
    CheckHistory -->|Yes| ShowError5[Show password<br/>reuse error]
    ShowError5 --> ShowForm
    
    CheckHistory -->|No| HashPassword[Hash new password]
    HashPassword --> UpdatePassword[Update password<br/>in database]
    UpdatePassword --> UpdateTimestamp[Update<br/>password_changed_at]
    UpdateTimestamp --> AddToHistory[Add old password<br/>to history]
    AddToHistory --> InvalidateSessions[Invalidate all<br/>other sessions]
    InvalidateSessions --> LogAudit[Log audit:<br/>password_changed]
    LogAudit --> SendEmail[Send email<br/>notification]
    SendEmail --> ShowSuccess[Show success<br/>message]
    ShowSuccess --> End2([End])
```

## 5. Two-Factor Authentication Setup Flow

```mermaid
flowchart TD
    Start([User enables 2FA<br/>in settings]) --> CheckAuth{User<br/>authenticated?}
    
    CheckAuth -->|No| RedirectLogin[Redirect to login]
    RedirectLogin --> End1([End])
    
    CheckAuth -->|Yes| Check2FAStatus{2FA already<br/>enabled?}
    
    Check2FAStatus -->|Yes| ShowDisable[Show disable<br/>2FA option]
    ShowDisable --> End2([End])
    
    Check2FAStatus -->|No| InitiateSetup[POST /2fa/setup]
    
    InitiateSetup --> GenerateSecret[Generate random<br/>TOTP secret]
    GenerateSecret --> CreateQR[Create QR code<br/>for secret]
    CreateQR --> EncryptSecret[Encrypt secret<br/>with AES-256]
    EncryptSecret --> StoreTemp[Store in DB<br/>enabled=false]
    StoreTemp --> ShowQR[Display QR code<br/>and manual entry key]
    
    ShowQR --> UserScans[User scans QR with<br/>authenticator app]
    UserScans --> EnterCode[User enters<br/>verification code]
    EnterCode --> Submit[POST /2fa/verify]
    
    Submit --> ValidateCode{Validate<br/>TOTP code?}
    
    ValidateCode -->|Invalid| IncrementAttempts[Increment<br/>failed attempts]
    IncrementAttempts --> CheckAttempts{Attempts<br/>>= 3?}
    
    CheckAttempts -->|Yes| CancelSetup[Cancel 2FA setup]
    CancelSetup --> DeleteTemp[Delete temp data]
    DeleteTemp --> ShowError1[Show too many<br/>attempts error]
    ShowError1 --> End3([End])
    
    CheckAttempts -->|No| ShowError2[Show invalid<br/>code error]
    ShowError2 --> ShowQR
    
    ValidateCode -->|Valid| Enable2FA[Set enabled=true<br/>in database]
    Enable2FA --> SetTimestamp[Set enabled_at<br/>timestamp]
    SetTimestamp --> GenerateBackup[Generate 10<br/>backup codes]
    GenerateBackup --> EncryptBackup[Encrypt backup<br/>codes]
    EncryptBackup --> StoreBackup[Store encrypted<br/>backup codes]
    StoreBackup --> LogAudit[Log audit:<br/>2fa_enabled]
    LogAudit --> SendEmail[Send confirmation<br/>email]
    SendEmail --> ShowBackupCodes[Display backup codes<br/>for user to save]
    
    ShowBackupCodes --> UserConfirms[User confirms<br/>saved backup codes]
    UserConfirms --> ShowSuccess[Show success<br/>message]
    ShowSuccess --> End4([End])
```

## 6. Token Refresh Flow

```mermaid
flowchart TD
    Start([Client makes API request]) --> CheckToken{Access token<br/>valid?}
    
    CheckToken -->|Valid| MakeRequest[Proceed with<br/>API request]
    MakeRequest --> End1([End])
    
    CheckToken -->|Expired| CheckRefresh{Refresh token<br/>exists?}
    
    CheckRefresh -->|No| RedirectLogin[Redirect to<br/>SSO Server login]
    RedirectLogin --> End2([End])
    
    CheckRefresh -->|Yes| SendRefresh[POST /oauth/token<br/>grant_type=refresh_token]
    
    SendRefresh --> ValidateRefresh{Validate<br/>refresh token?}
    
    ValidateRefresh -->|Invalid| ClearTokens[Clear stored<br/>tokens]
    ClearTokens --> RedirectLogin
    
    ValidateRefresh -->|Revoked| ClearTokens
    
    ValidateRefresh -->|Expired| ClearTokens
    
    ValidateRefresh -->|Valid| CheckUser{User account<br/>active?}
    
    CheckUser -->|No| RevokeToken[Revoke refresh<br/>token]
    RevokeToken --> ClearTokens
    
    CheckUser -->|Yes| GenerateNew[Generate new<br/>access token]
    GenerateNew --> RotateRefresh[Generate new<br/>refresh token]
    RotateRefresh --> RevokeOld[Revoke old<br/>refresh token]
    RevokeOld --> StoreNew[Store new<br/>refresh token]
    StoreNew --> ReturnTokens[Return new tokens<br/>to client]
    ReturnTokens --> UpdateClient[Client stores<br/>new tokens]
    UpdateClient --> RetryRequest[Retry original<br/>request]
    RetryRequest --> End3([End])
```

## 7. Session Management Flow

```mermaid
flowchart TD
    Start([User makes request]) --> CheckCookie{Session cookie<br/>present?}
    
    CheckCookie -->|No| Unauthenticated[Return<br/>unauthenticated]
    Unauthenticated --> End1([End])
    
    CheckCookie -->|Yes| LookupSession[Lookup session<br/>in database]
    
    LookupSession --> SessionExists{Session<br/>exists?}
    
    SessionExists -->|No| DeleteCookie[Clear session<br/>cookie]
    DeleteCookie --> Unauthenticated
    
    SessionExists -->|Yes| CheckExpiry{Session<br/>expired?}
    
    CheckExpiry -->|Yes| DeleteSession[Delete session<br/>from DB]
    DeleteSession --> DeleteCookie
    
    CheckExpiry -->|No| CheckUser{User account<br/>active?}
    
    CheckUser -->|No| DeleteSession
    
    CheckUser -->|Yes| CheckInactivity{Inactive ><br/>timeout?}
    
    CheckInactivity -->|Yes| DeleteSession
    
    CheckInactivity -->|No| UpdateActivity[Update<br/>last_activity_at]
    UpdateActivity --> ExtendExpiry[Extend session<br/>expiry (sliding)]
    ExtendExpiry --> LoadUser[Load user data<br/>with permissions]
    LoadUser --> Authenticated[User authenticated]
    Authenticated --> End2([End])
```

## 8. Account Lockout and Recovery Flow

```mermaid
flowchart TD
    Start([Failed login attempt]) --> IncrementCounter[Increment<br/>failed_login_attempts]
    
    IncrementCounter --> CheckCount{Attempts<br/>>= 5?}
    
    CheckCount -->|No| UpdateCounter[Update counter<br/>in database]
    UpdateCounter --> ShowError[Show invalid<br/>credentials error]
    ShowError --> End1([End])
    
    CheckCount -->|Yes| LockAccount[Set is_locked=true]
    LockAccount --> SetUnlockTime[Set locked_until<br/>timestamp (+30 min)]
    SetUnlockTime --> LogLockout[Log audit:<br/>account_locked]
    LogLockout --> SendEmail[Send lockout<br/>notification email]
    SendEmail --> ShowLocked[Show account<br/>locked message]
    ShowLocked --> End2([End])
    
    Start2([User tries to login<br/>on locked account]) --> CheckLocked{Account<br/>locked?}
    
    CheckLocked -->|No| ProceedLogin[Proceed with<br/>normal login]
    ProceedLogin --> End3([End])
    
    CheckLocked -->|Yes| CheckUnlockTime{Current time ><br/>locked_until?}
    
    CheckUnlockTime -->|No| ShowStillLocked[Show account<br/>locked message]
    ShowStillLocked --> ShowContactSupport[Show contact<br/>support option]
    ShowContactSupport --> End4([End])
    
    CheckUnlockTime -->|Yes| UnlockAccount[Set is_locked=false]
    UnlockAccount --> ResetCounter[Reset failed_login_attempts=0]
    ResetCounter --> ClearUnlockTime[Clear locked_until]
    ClearUnlockTime --> LogUnlock[Log audit:<br/>account_unlocked]
    LogUnlock --> ProceedLogin
    
    Start3([Admin unlocks account]) --> AdminAction[Admin clicks<br/>unlock in dashboard]
    AdminAction --> ManualUnlock[POST /api/users/:id/unlock]
    ManualUnlock --> SetUnlocked[Set is_locked=false]
    SetUnlocked --> ResetAdminCounter[Reset failed_login_attempts=0]
    ResetAdminCounter --> ClearAdminTime[Clear locked_until]
    ClearAdminTime --> LogAdminUnlock[Log audit:<br/>admin_unlocked_account]
    LogAdminUnlock --> NotifyUser[Send email:<br/>account unlocked]
    NotifyUser --> End5([End])
```

## 9. Password Expiry Warning Flow

```mermaid
flowchart TD
    Start([Daily scheduled job]) --> GetUsers[Get all active users]
    
    GetUsers --> ForEachUser{For each user}
    
    ForEachUser --> CalculateAge[Calculate password age<br/>current_date - password_changed_at]
    
    CalculateAge --> GetPolicy[Get password policy<br/>from config]
    
    GetPolicy --> CheckExpiry{Age >= expiry_days?}
    
    CheckExpiry -->|Yes| CheckGrace{Within grace<br/>period?}
    
    CheckGrace -->|Yes| SendWarning[Send urgent<br/>expiry warning]
    SendWarning --> FlagExpiring[Flag account as<br/>expiring soon]
    FlagExpiring --> NextUser[Next user]
    
    CheckGrace -->|No| ForceExpiry[Force password<br/>change on next login]
    ForceExpiry --> SendExpired[Send password<br/>expired email]
    SendExpired --> NextUser
    
    CheckExpiry -->|No| CheckWarning{Age >= expiry_days<br/>- warning_days?}
    
    CheckWarning -->|Yes| CheckSent{Warning already<br/>sent?}
    
    CheckSent -->|No| SendFirstWarning[Send first<br/>warning email]
    SendFirstWarning --> MarkSent[Mark warning sent]
    MarkSent --> NextUser
    
    CheckSent -->|Yes| NextUser
    
    CheckWarning -->|No| NextUser
    
    NextUser --> ForEachUser
    
    ForEachUser -->|Done| End([End job])
    
    Start2([User with expired<br/>password logs in]) --> NormalAuth[Complete normal<br/>authentication]
    
    NormalAuth --> CheckPasswordAge{Password<br/>expired?}
    
    CheckPasswordAge -->|No| AllowLogin[Allow login]
    AllowLogin --> End2([End])
    
    CheckPasswordAge -->|Yes| ForceChange[Redirect to<br/>force change password]
    ForceChange --> ShowChangeForm[Show password<br/>change form]
    ShowChangeForm --> UserChanges[User changes<br/>password]
    UserChanges --> ValidateNew{Validate new<br/>password?}
    
    ValidateNew -->|No| ShowErrors[Show validation<br/>errors]
    ShowErrors --> ShowChangeForm
    
    ValidateNew -->|Yes| UpdatePassword[Update password]
    UpdatePassword --> ResetTimestamp[Update<br/>password_changed_at]
    ResetTimestamp --> AllowAccess[Allow access to<br/>application]
    AllowAccess --> End3([End])
```

---

**Document Version**: 1.0  
**Last Updated**: 2026-01-26  
**Diagram Format**: Mermaid
