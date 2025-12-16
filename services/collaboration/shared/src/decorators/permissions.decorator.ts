import { SetMetadata } from '@nestjs/common';
import { PERMISSIONS_KEY, ROLES_KEY } from '../auth/permission.guard';

/**
 * Decorator to require specific permissions
 * Usage: @RequirePermissions('tenant:read', 'tenant:write')
 */
export const RequirePermissions = (...permissions: string[]) =>
  SetMetadata(PERMISSIONS_KEY, permissions);

/**
 * Decorator to require specific roles
 * Usage: @RequireRoles('admin', 'manager')
 */
export const RequireRoles = (...roles: string[]) =>
  SetMetadata(ROLES_KEY, roles);
