// services/storefront-web/src/instrumentation.ts
import { registerOTel } from '@vercel/otel'

export function register() {
    // Registers OpenTelemetry with the Next.js runtime
    // Automatically captures fetch requests and propagates W3C Trace Context headers
    // to the API Gateway and downstream Go microservices.
    registerOTel({
        serviceName: 'storefront-web',
    })
}
