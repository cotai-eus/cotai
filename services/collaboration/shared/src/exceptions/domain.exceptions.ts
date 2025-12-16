import { HttpException, HttpStatus } from '@nestjs/common';

/**
 * Base domain exception
 */
export class DomainException extends Error {
  constructor(message: string) {
    super(message);
    this.name = this.constructor.name;
    Error.captureStackTrace(this, this.constructor);
  }
}

/**
 * Entity not found
 */
export class EntityNotFoundException extends DomainException {
  constructor(entityName: string, identifier: string) {
    super(`${entityName} with identifier '${identifier}' not found`);
  }
}

/**
 * Entity already exists
 */
export class EntityAlreadyExistsException extends DomainException {
  constructor(entityName: string, identifier: string) {
    super(`${entityName} with identifier '${identifier}' already exists`);
  }
}

/**
 * Validation error
 */
export class ValidationException extends DomainException {
  constructor(message: string, public readonly errors?: Record<string, string[]>) {
    super(message);
  }
}

/**
 * Unauthorized access
 */
export class UnauthorizedException extends DomainException {
  constructor(message: string = 'Unauthorized access') {
    super(message);
  }
}

/**
 * Forbidden access
 */
export class ForbiddenException extends DomainException {
  constructor(message: string = 'Forbidden access') {
    super(message);
  }
}

/**
 * Business rule violation
 */
export class BusinessRuleViolationException extends DomainException {
  constructor(message: string) {
    super(message);
  }
}

/**
 * Conflict error (e.g., optimistic locking)
 */
export class ConflictException extends DomainException {
  constructor(message: string) {
    super(message);
  }
}

/**
 * Helper functions to check exception types
 */
export function isNotFoundException(error: unknown): error is EntityNotFoundException {
  return error instanceof EntityNotFoundException;
}

export function isAlreadyExistsException(error: unknown): error is EntityAlreadyExistsException {
  return error instanceof EntityAlreadyExistsException;
}

export function isValidationException(error: unknown): error is ValidationException {
  return error instanceof ValidationException;
}

export function isBusinessRuleViolation(
  error: unknown,
): error is BusinessRuleViolationException {
  return error instanceof BusinessRuleViolationException;
}
