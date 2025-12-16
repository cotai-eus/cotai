import {
  ExceptionFilter,
  Catch,
  ArgumentsHost,
  HttpException,
  HttpStatus,
  Logger,
} from '@nestjs/common';
import { Response } from 'express';
import {
  EntityNotFoundException,
  EntityAlreadyExistsException,
  ValidationException,
  UnauthorizedException,
  ForbiddenException,
  BusinessRuleViolationException,
  ConflictException,
} from './domain.exceptions';

export interface ErrorResponse {
  error: {
    code: string;
    message: string;
    details?: any;
  };
}

@Catch()
export class HttpExceptionFilter implements ExceptionFilter {
  private readonly logger = new Logger(HttpExceptionFilter.name);

  catch(exception: unknown, host: ArgumentsHost) {
    const ctx = host.switchToHttp();
    const response = ctx.getResponse<Response>();
    const request = ctx.getRequest();

    let status: number;
    let errorResponse: ErrorResponse;

    if (exception instanceof HttpException) {
      status = exception.getStatus();
      const exceptionResponse = exception.getResponse();

      errorResponse = {
        error: {
          code: exception.name,
          message: exception.message,
          details:
            typeof exceptionResponse === 'object'
              ? (exceptionResponse as any).message
              : exceptionResponse,
        },
      };
    } else if (exception instanceof EntityNotFoundException) {
      status = HttpStatus.NOT_FOUND;
      errorResponse = {
        error: {
          code: 'ENTITY_NOT_FOUND',
          message: exception.message,
        },
      };
    } else if (exception instanceof EntityAlreadyExistsException) {
      status = HttpStatus.CONFLICT;
      errorResponse = {
        error: {
          code: 'ENTITY_ALREADY_EXISTS',
          message: exception.message,
        },
      };
    } else if (exception instanceof ValidationException) {
      status = HttpStatus.BAD_REQUEST;
      errorResponse = {
        error: {
          code: 'VALIDATION_ERROR',
          message: exception.message,
          details: exception.errors,
        },
      };
    } else if (exception instanceof UnauthorizedException) {
      status = HttpStatus.UNAUTHORIZED;
      errorResponse = {
        error: {
          code: 'UNAUTHORIZED',
          message: exception.message,
        },
      };
    } else if (exception instanceof ForbiddenException) {
      status = HttpStatus.FORBIDDEN;
      errorResponse = {
        error: {
          code: 'FORBIDDEN',
          message: exception.message,
        },
      };
    } else if (exception instanceof BusinessRuleViolationException) {
      status = HttpStatus.UNPROCESSABLE_ENTITY;
      errorResponse = {
        error: {
          code: 'BUSINESS_RULE_VIOLATION',
          message: exception.message,
        },
      };
    } else if (exception instanceof ConflictException) {
      status = HttpStatus.CONFLICT;
      errorResponse = {
        error: {
          code: 'CONFLICT',
          message: exception.message,
        },
      };
    } else {
      // Unknown error
      status = HttpStatus.INTERNAL_SERVER_ERROR;
      errorResponse = {
        error: {
          code: 'INTERNAL_SERVER_ERROR',
          message: 'An unexpected error occurred',
        },
      };

      // Log unknown errors
      this.logger.error('Unhandled exception:', {
        exception,
        path: request.url,
        method: request.method,
        tenantId: request.tenantId,
        userId: request.user?.sub,
      });
    }

    response.status(status).json(errorResponse);
  }
}
