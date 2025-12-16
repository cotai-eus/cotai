import { Injectable, LoggerService as NestLoggerService } from '@nestjs/common';
import * as winston from 'winston';

export interface LoggerConfig {
  serviceName: string;
  environment: string;
  logLevel?: string;
}

@Injectable()
export class LoggerService implements NestLoggerService {
  private logger: winston.Logger;
  private serviceName: string;

  constructor(config: LoggerConfig) {
    this.serviceName = config.serviceName;

    const format = winston.format.combine(
      winston.format.timestamp({ format: 'YYYY-MM-DD HH:mm:ss' }),
      winston.format.errors({ stack: true }),
      winston.format.splat(),
      winston.format.json(),
    );

    this.logger = winston.createLogger({
      level: config.logLevel || process.env.LOG_LEVEL || 'info',
      format,
      defaultMeta: {
        service: config.serviceName,
        environment: config.environment,
      },
      transports: [
        new winston.transports.Console({
          format:
            config.environment === 'development'
              ? winston.format.combine(
                  winston.format.colorize(),
                  winston.format.simple(),
                )
              : format,
        }),
      ],
    });
  }

  log(message: string, context?: string, meta?: any) {
    this.logger.info(message, { context: context || this.serviceName, ...meta });
  }

  error(message: string, trace?: string, context?: string, meta?: any) {
    this.logger.error(message, {
      trace,
      context: context || this.serviceName,
      ...meta,
    });
  }

  warn(message: string, context?: string, meta?: any) {
    this.logger.warn(message, { context: context || this.serviceName, ...meta });
  }

  debug(message: string, context?: string, meta?: any) {
    this.logger.debug(message, { context: context || this.serviceName, ...meta });
  }

  verbose(message: string, context?: string, meta?: any) {
    this.logger.verbose(message, { context: context || this.serviceName, ...meta });
  }

  /**
   * Log with tenant and user context
   */
  logWithContext(
    level: 'info' | 'error' | 'warn' | 'debug',
    message: string,
    context: {
      tenantId?: string;
      userId?: string;
      correlationId?: string;
      [key: string]: any;
    },
  ) {
    this.logger.log(level, message, {
      ...context,
      service: this.serviceName,
    });
  }
}
