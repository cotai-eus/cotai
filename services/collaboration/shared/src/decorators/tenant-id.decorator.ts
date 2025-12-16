import { createParamDecorator, ExecutionContext } from '@nestjs/common';
import { RequestWithTenant } from '../auth/tenant.guard';

/**
 * Decorator to extract tenant ID from request
 * Usage: getTenant(@TenantId() tenantId: string)
 */
export const TenantId = createParamDecorator(
  (data: unknown, ctx: ExecutionContext): string => {
    const request = ctx.switchToHttp().getRequest<RequestWithTenant>();
    return request.tenantId || '';
  },
);
