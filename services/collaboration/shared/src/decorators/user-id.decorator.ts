import { createParamDecorator, ExecutionContext } from '@nestjs/common';
import { JwtPayload } from '../types';

/**
 * Decorator to extract user ID from JWT payload
 * Usage: getUser(@UserId() userId: string)
 */
export const UserId = createParamDecorator(
  (data: unknown, ctx: ExecutionContext): string => {
    const request = ctx.switchToHttp().getRequest();
    const user = request.user as JwtPayload;
    return user?.sub || '';
  },
);
