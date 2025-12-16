import {
  Injectable,
  NestInterceptor,
  ExecutionContext,
  CallHandler,
  Logger,
} from '@nestjs/common';
import { Observable } from 'rxjs';
import { tap } from 'rxjs/operators';
import { initTracer, JaegerTracer, TracingConfig, TracingOptions } from 'jaeger-client';

export interface JaegerConfig {
  serviceName: string;
  agentHost?: string;
  agentPort?: number;
  samplerType?: string;
  samplerParam?: number;
}

@Injectable()
export class TracingInterceptor implements NestInterceptor {
  private readonly logger = new Logger(TracingInterceptor.name);
  private tracer?: JaegerTracer;

  constructor(config: JaegerConfig) {
    // Only initialize if Jaeger is configured
    if (config.agentHost) {
      const tracingConfig: TracingConfig = {
        serviceName: config.serviceName,
        sampler: {
          type: config.samplerType || 'probabilistic',
          param: config.samplerParam || 0.1,
        },
        reporter: {
          agentHost: config.agentHost,
          agentPort: config.agentPort || 6831,
          logSpans: false,
        },
      };

      const tracingOptions: TracingOptions = {
        logger: {
          info: (msg: string) => this.logger.debug(msg),
          error: (msg: string) => this.logger.error(msg),
        },
      };

      this.tracer = initTracer(tracingConfig, tracingOptions);
      this.logger.log('Jaeger tracing initialized');
    } else {
      this.logger.warn('Jaeger not configured, tracing disabled');
    }
  }

  intercept(context: ExecutionContext, next: CallHandler): Observable<any> {
    if (!this.tracer) {
      return next.handle();
    }

    const request = context.switchToHttp().getRequest();
    const operationName = `${request.method} ${request.route?.path || request.url}`;

    const span = this.tracer.startSpan(operationName);

    // Set standard tags
    span.setTag('http.method', request.method);
    span.setTag('http.url', request.url);
    span.setTag('http.route', request.route?.path);

    // Set tenant context if available
    if (request.tenantId) {
      span.setTag('tenant.id', request.tenantId);
    }

    // Set user context if available
    if (request.user?.sub) {
      span.setTag('user.id', request.user.sub);
    }

    const startTime = Date.now();

    return next.handle().pipe(
      tap({
        next: () => {
          const duration = Date.now() - startTime;
          span.setTag('http.status_code', 200);
          span.setTag('duration_ms', duration);
          span.finish();
        },
        error: (error) => {
          const duration = Date.now() - startTime;
          span.setTag('error', true);
          span.setTag('http.status_code', error.status || 500);
          span.setTag('duration_ms', duration);
          span.log({
            event: 'error',
            message: error.message,
            stack: error.stack,
          });
          span.finish();
        },
      }),
    );
  }

  /**
   * Get tracer instance for manual span creation
   */
  getTracer(): JaegerTracer | undefined {
    return this.tracer;
  }
}
