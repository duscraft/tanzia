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
- [ ] Add bcrypt dependency: `go get golang.org/x/crypto/bcrypt`
- [ ] Create password hashing utility in `lib/helpers/password.go`
- [ ] Update SignupHandler to hash passwords before storing
- [ ] Update LoginHandler to compare hashed passwords with bcrypt.CompareHashAndPassword
- [ ] Create database migration script for existing users (require password reset on first login)
- [ ] Add password strength validation

### Session Security
- [ ] Add CSRF protection for form submissions
- [ ] Implement secure cookie settings (Secure, SameSite)
- [ ] Add rate limiting for login attempts
- [ ] Implement account lockout after failed attempts

---

## Payment Integration Roadmap

### Phase 1: Stripe Setup
- [ ] Create Stripe account and get API keys
- [ ] Add Stripe Go SDK: `go get github.com/stripe/stripe-go/v76`
- [ ] Store Stripe keys in environment variables:
  - `STRIPE_SECRET_KEY`
  - `STRIPE_PUBLISHABLE_KEY`
  - `STRIPE_WEBHOOK_SECRET`
- [ ] Create `lib/helpers/stripe.go` for Stripe integration

### Phase 2: Database Schema Updates
```sql
ALTER TABLE users ADD COLUMN stripe_customer_id TEXT;
ALTER TABLE users ADD COLUMN subscription_id TEXT;
ALTER TABLE users ADD COLUMN subscription_status TEXT DEFAULT 'none';
ALTER TABLE users ADD COLUMN subscription_end_date TIMESTAMP;
```

### Phase 3: Checkout Flow
- [ ] Create checkout session endpoint `POST /api/checkout`
- [ ] Implement Stripe Checkout redirect
- [ ] Create success/cancel pages
- [ ] Add pricing page with subscription options

### Phase 4: Webhook Handler
- [ ] Implement webhook endpoint `POST /api/webhooks/stripe`
- [ ] Handle `checkout.session.completed` event
- [ ] Handle `customer.subscription.updated` event
- [ ] Handle `customer.subscription.deleted` event
- [ ] Handle `invoice.payment_failed` event
- [ ] Verify webhook signatures

### Phase 5: Subscription Management
- [ ] Create customer portal link endpoint
- [ ] Implement subscription cancellation
- [ ] Add grace period handling
- [ ] Send email notifications for subscription events

### Phase 6: Premium Features Enforcement
- [ ] Update IsUserPremium to check subscription_status and subscription_end_date
- [ ] Add premium badge to dashboard UI
- [ ] Show upgrade prompts for free users at limits
- [ ] Implement feature gates for premium-only features

---

## Testing Improvements

### Unit Tests
- [ ] Add tests for authentication handlers (login, signup, logout)
- [ ] Add tests for premium limiter functions
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
