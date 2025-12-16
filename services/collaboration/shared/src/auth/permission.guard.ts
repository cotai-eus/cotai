import {
  Injectable,
  CanActivate,
  ExecutionContext,
  ForbiddenException,
  Logger,
} from '@nestjs/common';
import { Reflector } from '@nestjs/core';
import { JwtPayload } from '../types';

export const PERMISSIONS_KEY = 'permissions';
export const ROLES_KEY = 'roles';

@Injectable()
export class PermissionGuard implements CanActivate {
  private readonly logger = new Logger(PermissionGuard.name);

  constructor(private reflector: Reflector) {}

  canActivate(context: ExecutionContext): boolean {
    const requiredPermissions = this.reflector.getAllAndOverride<string[]>(
      PERMISSIONS_KEY,
      [context.getHandler(), context.getClass()],
    );

    const requiredRoles = this.reflector.getAllAndOverride<string[]>(
      ROLES_KEY,
      [context.getHandler(), context.getClass()],
    );

    // If no permissions or roles required, allow access
    if (!requiredPermissions && !requiredRoles) {
      return true;
    }

    const request = context.switchToHttp().getRequest();
    const user = request.user as JwtPayload;

    if (!user) {
      throw new ForbiddenException('User not authenticated');
    }

    // Check permissions
    if (requiredPermissions && requiredPermissions.length > 0) {
      const hasPermission = requiredPermissions.every((permission) =>
        user.permissions?.includes(permission),
      );

      if (!hasPermission) {
        this.logger.warn(
          `User ${user.sub} missing required permissions: ${requiredPermissions.join(', ')}`,
        );
        throw new ForbiddenException('Insufficient permissions');
      }
    }

    // Check roles
    if (requiredRoles && requiredRoles.length > 0) {
      const hasRole = requiredRoles.some((role) => user.roles?.includes(role));

      if (!hasRole) {
        this.logger.warn(
          `User ${user.sub} missing required roles: ${requiredRoles.join(', ')}`,
        );
        throw new ForbiddenException('Insufficient role');
      }
    }

    return true;
  }
}
