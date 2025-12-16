export interface JwtPayload {
  /**
   * Subject (user ID from Keycloak)
   */
  sub: string;

  /**
   * Email address
   */
  email: string;

  /**
   * Tenant ID (from custom Keycloak mapper)
   */
  tenant_id?: string;

  /**
   * User roles
   */
  roles: string[];

  /**
   * Permissions/scopes
   */
  permissions?: string[];

  /**
   * Issued at timestamp
   */
  iat: number;

  /**
   * Expiration timestamp
   */
  exp: number;

  /**
   * Issuer (Keycloak realm URL)
   */
  iss: string;

  /**
   * Audience
   */
  aud: string | string[];
}
