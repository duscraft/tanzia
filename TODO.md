# Tanzia - Development Roadmap

## Completed

### Code Quality Improvements
- [x] Fix GitHub Actions CI (golangci-lint version, Go module caching)
- [x] Fix database schema (add is_premium column)
- [x] Fix missing return statements after http.Error
- [x] Fix resource leaks (close sql.Rows properly)
- [x] Replace log.Fatal with proper error handling
- [x] Replace panic() with error returns
- [x] Handle strconv parsing errors
- [x] Handle IsUserPremium errors
- [x] Handle template parsing errors
- [x] Fix Go naming conventions (redisURL, FreeTier constants)

### Premium Features
- [x] PDF export for quarterly/yearly reports (premium only)

---

## High Priority - Security

### Password Hashing Implementation
- [x] Add bcrypt dependency: `go get golang.org/x/crypto/bcrypt`
- [x] Create password hashing utility in `lib/helpers/password.go`
- [x] Update SignupHandler to hash passwords before storing
- [x] Update LoginHandler to compare hashed passwords with bcrypt.CompareHashAndPassword
- [x] Create database migration script for existing users (require password reset on first login)
- [x] Add password strength validation

### Session Security
- [x] Add CSRF protection for form submissions
- [x] Implement secure cookie settings (Secure, SameSite)
- [x] Add rate limiting for login attempts
- [x] Implement account lockout after failed attempts

---

## Payment Integration Roadmap

### Phase 1: Stripe Setup
- [x] Create Stripe account and get API keys
- [x] Add Stripe Go SDK: `go get github.com/stripe/stripe-go/v76`
- [x] Store Stripe keys in environment variables:
  - `STRIPE_SECRET_KEY`
  - `STRIPE_PUBLISHABLE_KEY`
  - `STRIPE_WEBHOOK_SECRET`
- [x] Create `lib/domains/stripe.go` for Stripe integration

### Phase 2: Database Schema Updates
- [x] Add stripe_customer_id column to users table
- [x] Database migration handles both new and existing installations

### Phase 3: Checkout Flow
- [x] Create checkout session endpoint `POST /subscribe`
- [x] Implement Stripe Checkout redirect
- [x] Create success page (redirects to dashboard with premium status)
- [x] Add pricing page with subscription options on homepage
- [x] Improve premium signup flow (redirect unauthenticated users to signup, auto-login after signup)

### Phase 4: Webhook Handler
- [x] Implement webhook endpoint `POST /webhooks/stripe`
- [x] Handle `checkout.session.completed` event
- [x] Handle `customer.subscription.updated` event
- [x] Handle `customer.subscription.deleted` event
- [x] Handle `invoice.payment_failed` event
- [x] Verify webhook signatures
- [x] Proper error handling (return 500 so Stripe retries on failure)

### Phase 5: Subscription Management
- [x] Create customer portal link endpoint (`GET /billing`)
- [x] Implement subscription cancellation (via Stripe Customer Portal)
- [ ] Add grace period handling
- [ ] Send email notifications for subscription events

### Phase 6: Premium Features Enforcement
- [x] Update IsUserPremium to check stripe_customer_id
- [x] Add premium badge to dashboard UI
- [x] Show upgrade prompts for free users at limits
- [x] Implement feature gates for premium-only features (PDF export, limits)

---

## Testing Improvements

### Unit Tests
- [ ] Add tests for authentication handlers (login, signup, logout)
- [x] Add tests for premium limiter functions
- [x] Add tests for CSRF protection
- [ ] Add tests for PDF export handler
- [ ] Add tests for dashboard data retrieval

### Integration Tests
- [ ] Add database integration tests with test containers
- [ ] Add end-to-end authentication flow tests
- [ ] Add subscription lifecycle tests

### Test Infrastructure
- [ ] Set up test database configuration
- [ ] Add CI test coverage reporting
- [ ] Target 60%+ code coverage

---

## Additional Features (Future)

### Data Management
- [ ] Add ability to edit/delete persons
- [ ] Add ability to edit/delete bills
- [ ] Add ability to edit/delete provisions
- [ ] Add data import from CSV/Excel
- [ ] Add data backup/export functionality

### Reporting
- [ ] Add quarterly report generation
- [ ] Add yearly report generation
- [ ] Add chart visualizations for balance over time
- [ ] Add email report delivery

### Multi-tenancy
- [ ] Support multiple coproprietes per user
- [ ] Add copropriete switching in dashboard
- [ ] Add copropriete-level settings

### Notifications
- [ ] Email notifications for new bills/provisions
- [ ] Payment reminder emails
- [ ] Balance threshold alerts

---

## Infrastructure

### Monitoring
- [ ] Add structured logging
- [ ] Add request tracing
- [ ] Add error reporting (Sentry or similar)
- [ ] Add performance monitoring

### Deployment
- [ ] Replace sshpass with SSH key authentication in CI
- [ ] Add Docker layer caching in CI
- [ ] Add staging environment
- [ ] Add database backup automation

---

## Notes

### Stripe Integration Resources
- Stripe Go SDK: https://github.com/stripe/stripe-go
- Checkout documentation: https://stripe.com/docs/checkout
- Webhooks guide: https://stripe.com/docs/webhooks
- Customer portal: https://stripe.com/docs/billing/subscriptions/integrating-customer-portal

### Pricing Suggestions
- Free tier: 5 persons, 5 bills, 10 provisions
- Premium tier: Unlimited + PDF export
- Suggested price: 4.99 EUR/month or 49 EUR/year
