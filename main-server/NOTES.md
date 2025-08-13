# Notes and Journal

## Features Or Capabilities

### User Management
Users can be of one of four types, Users, Company Admins, Workspace Admins and Super Admins.

1. Super Admins
   * Create, Read, Update or Delete Workspaces or Companies. Any workspace must be empty of companies before it can be deleted.
   * Any company must be empty of users and company admins before it can be deleted
   * Any actions with users including creating Super Admins
   * They can Delete other Super Admins but not themselves
   * Super Admins can reset passwords manually for anyone including themselves
   * New Super Admins have null workspaces and companies
2. Workspace Admin
   * Create, Read, Update Workspaces, Delete Companies
   * Can perform any user management at Workspace Admin level or lower
   * Workspaces contain companies
   * Workspace Admins cannot delete themselves or other Workspace Admins
   * They can disable other Workspace Admins
   * Workspace Admins can reset passwords manually for Users or Company Admins
   * New Workspace admins must be created with a workspace that exists
3. Company Admin
   * Create, Read, Update Companies
   * Each company is a part of a workspace
   * Can perform any user management at Company Admin level or lower
   * Companies contain users and Company Admins
   * A company is the main space where any given user performs actions (thus for audit tracking any action must be traced down to the company level)
   * Company Admins cannot delete themselves or other Company Admins
   * They can disable other Company Admins
   * All company admins are members of a company and a workspace
   * Company Admins can reset passwords manually for Users
   * New Company Admins must be created with a company and workspace that exists
4. Users
   * Users operate within a company and can perform any job that is permitted within said company
   * Users can Read or Update themselves but not any other users
   * There is no (as yet) restriction at a user-by-user basis for functionality or permission (perhaps there should be)
   * All users are members of a company and a workspace
   * New Users must be created with a company and workspace that exists

### Additional Requirements to Consider:
   Authentication & Security:
      * Password policies and reset functionality
      * Multi-factor authentication support
      * Session management and token expiration
      * Account lockout after failed attempts

   Audit & Compliance:
      * Audit logs for all admin actions (especially deletions/disables)
      * User activity tracking at company level (as you mentioned)
      * Data retention policies for deleted entities
   
   User Experience:
      * User invitation/onboarding flow
      * Email verification for new accounts
      * Bulk user operations (import/export)

   Business Logic:
      * Default workspace/company assignment for new users
      * User transfer between companies/workspaces
      * Handling of orphaned users when companies are deleted

### Permission Matrix:
Action	         Super Admin	   Workspace Admin	   Company Admin	   User
Manage Workspaces	✅	            ✅ (own)	            ❌	               ❌
Manage Companies	✅	            ✅ (in workspace)	   ✅ (own)	         ❌
Manage Users	   ✅	            ✅ (in workspace)	   ✅ (in company)	✅ (self only)
Delete Self	      ❌	            ❌	                  ❌	               ❌
Delete Same Level	✅	            ❌	                  ❌	               ❌

### Key Design Decisions:
Soft Deletes - Use deleted_at timestamps for audit trail
UUID Primary Keys - Better for distributed systems
Hierarchical Structure - Workspace → Company → User
Single User Table - With user_type enum rather than separate tables
Company-Scoped Auditing - All actions traceable to company level

### API Endpoints Structure
POST   /auth/login
POST   /auth/logout
GET    /users/me
PUT    /users/me
GET    /workspaces
POST   /workspaces
GET    /workspaces/{id}/companies
POST   /companies
GET    /companies/{id}/users
POST   /companies/{id}/users
PUT    /users/{id}/disable
DELETE /users/{id}

### Data Structures

```sql
-- Main users table
users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255), -- NULL if not set yet (invited users)
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    user_type user_type_enum NOT NULL, -- 'user', 'company_admin', 'workspace_admin', 'super_admin'
    company_id UUID REFERENCES companies(id),
    workspace_id UUID REFERENCES workspaces(id),
    is_active BOOLEAN DEFAULT true,
    email_verified BOOLEAN DEFAULT false,
    last_login_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP -- soft delete
);

-- User invitations
user_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL,
    invited_by_user_id UUID REFERENCES users(id) NOT NULL,
    user_type user_type_enum NOT NULL,
    company_id UUID REFERENCES companies(id),
    workspace_id UUID REFERENCES workspaces(id),
    token VARCHAR(255) UNIQUE NOT NULL, -- invitation token
    expires_at TIMESTAMP NOT NULL,
    accepted_at TIMESTAMP,
    created_user_id UUID REFERENCES users(id), -- Set when invitation is accepted
    created_at TIMESTAMP DEFAULT NOW(),
    
    INDEX idx_invitations_token (token),
    INDEX idx_invitations_email (email),
    INDEX idx_invitations_expires (expires_at)
);
-- Password reset tokens
password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) NOT NULL,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    
    INDEX idx_reset_tokens_token (token),
    INDEX idx_reset_tokens_user (user_id),
    INDEX idx_reset_tokens_expires (expires_at)
);

-- Email verification tokens
email_verification_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) NOT NULL,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    verified_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    
    INDEX idx_verification_tokens_token (token),
    INDEX idx_verification_tokens_user (user_id)
);

-- User sessions
user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) NOT NULL,
    token_hash VARCHAR(255) NOT NULL, -- hashed session token
    device_info JSONB, -- browser, OS, etc.
    ip_address INET,
    expires_at TIMESTAMP NOT NULL,
    last_activity_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    
    INDEX idx_sessions_token (token_hash),
    INDEX idx_sessions_user (user_id),
    INDEX idx_sessions_expires (expires_at)
);
-- MFA settings per user
user_mfa_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) NOT NULL UNIQUE,
    is_enabled BOOLEAN DEFAULT false,
    backup_codes_generated_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- TOTP (Time-based One-Time Password) like Google Authenticator
user_totp_secrets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) NOT NULL UNIQUE,
    secret_key VARCHAR(255) NOT NULL, -- encrypted TOTP secret
    is_verified BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    verified_at TIMESTAMP
);

-- Backup codes for MFA recovery
user_mfa_backup_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) NOT NULL,
    code_hash VARCHAR(255) NOT NULL, -- hashed backup code
    used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    
    INDEX idx_backup_codes_user (user_id)
);

-- MFA attempts tracking
user_mfa_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) NOT NULL,
    attempt_type VARCHAR(50) NOT NULL, -- 'totp', 'backup_code'
    success BOOLEAN NOT NULL,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    
    INDEX idx_mfa_attempts_user (user_id),
    INDEX idx_mfa_attempts_created (created_at)
);
-- Failed login attempts
failed_login_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL,
    ip_address INET,
    user_agent TEXT,
    failure_reason VARCHAR(100), -- 'invalid_password', 'account_locked', etc.
    created_at TIMESTAMP DEFAULT NOW(),
    
    INDEX idx_failed_logins_email (email),
    INDEX idx_failed_logins_ip (ip_address),
    INDEX idx_failed_logins_created (created_at)
);

-- Account lockouts
user_lockouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) NOT NULL,
    locked_until TIMESTAMP NOT NULL,
    reason VARCHAR(255), -- 'too_many_failed_attempts', 'admin_action'
    locked_by_user_id UUID REFERENCES users(id), -- NULL if automatic
    unlocked_at TIMESTAMP,
    unlocked_by_user_id UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    
    INDEX idx_lockouts_user (user_id),
    INDEX idx_lockouts_locked_until (locked_until)
);

-- Comprehensive audit log
audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id), -- Who performed the action
    action VARCHAR(100) NOT NULL, -- 'create_user', 'delete_company', etc.
    entity_type VARCHAR(50) NOT NULL, -- 'user', 'company', 'workspace'
    entity_id UUID NOT NULL, -- ID of the affected entity
    company_context_id UUID REFERENCES companies(id), -- For company-level tracking
    workspace_context_id UUID REFERENCES workspaces(id),
    old_values JSONB, -- What it was before
    new_values JSONB, -- What it became
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    
    INDEX idx_audit_user (user_id),
    INDEX idx_audit_entity (entity_type, entity_id),
    INDEX idx_audit_company (company_context_id),
    INDEX idx_audit_created (created_at)
);
-- Email notifications queue
email_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    email VARCHAR(255) NOT NULL, -- Might be different from user email
    template_name VARCHAR(100) NOT NULL, -- 'welcome', 'password_reset', etc.
    template_data JSONB, -- Variables for email template
    status VARCHAR(50) DEFAULT 'pending', -- 'pending', 'sent', 'failed'
    sent_at TIMESTAMP,
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    
    INDEX idx_notifications_status (status),
    INDEX idx_notifications_user (user_id),
    INDEX idx_notifications_created (created_at)
);
```




### Companies


### Workspaces
