import {
  Injectable,
  CanActivate,
  ExecutionContext,
  BadRequestException,
  Logger,
} from '@nestjs/common';
import { Request } from 'express';
import { JwtPayload } from '../types';

export interface RequestWithTenant extends Request {
  tenantId?: string;
  user?: JwtPayload;
}

@Injectable()
export class TenantGuard implements CanActivate {
  private readonly logger = new Logger(TenantGuard.name);

  canActivate(context: ExecutionContext): boolean {
    const request = context.switchToHttp().getRequest<RequestWithTenant>();
    const user = request.user as JwtPayload;

    // 1. Try JWT claim (primary)
    let tenantId = user?.tenant_id;

    // 2. Fallback to X-Tenant-ID header
    if (!tenantId) {
      tenantId = request.headers['x-tenant-id'] as string;
    }

    // 3. Fallback to subdomain extraction (optional)
    if (!tenantId && request.hostname) {
      const subdomain = request.hostname.split('.')[0];
      if (subdomain && subdomain !== 'api' && subdomain !== 'localhost') {
        tenantId = subdomain;
        this.logger.debug(`Extracted tenant from subdomain: ${tenantId}`);
      }
    }

    if (!tenantId) {
      throw new BadRequestException(
        'Tenant ID not found in JWT claims, X-Tenant-ID header, or subdomain',
      );
    }

    // Attach tenantId to request for downstream use
    request.tenantId = tenantId;

    this.logger.debug(
      `Request authenticated for tenant: ${tenantId}, user: ${user?.sub}`,
    );

    return true;
  }
}
