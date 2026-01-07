package domains

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/duscraft/tanzia/lib/helpers"
	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/billingportal/session"
	checkoutsession "github.com/stripe/stripe-go/v84/checkout/session"
	"github.com/stripe/stripe-go/v84/webhook"
)

func init() {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
}

func CreateCheckoutSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := GetAuthenticatedUserID(w, r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	db, err := helpers.GetConnectionManager().GetConnection("postgres")
	if err != nil {
		log.Printf("Database connection error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var email string
	var stripeCustomerID sql.NullString
	err = db.QueryRow("SELECT email, stripe_customer_id FROM users WHERE id = $1", userID).Scan(&email, &stripeCustomerID)
	if err != nil {
		log.Printf("Error fetching user: %v", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	domain := os.Getenv("DOMAIN")
	if domain == "" {
		domain = "http://localhost:8080"
	}

	priceID := os.Getenv("STRIPE_PRICE_ID")
	if priceID == "" {
		log.Printf("STRIPE_PRICE_ID not configured")
		http.Error(w, "Payment not configured", http.StatusInternalServerError)
		return
	}

	params := &stripe.CheckoutSessionParams{
		SuccessURL: stripe.String(domain + "/dashboard?payment=success"),
		CancelURL:  stripe.String(domain + "/dashboard?payment=cancelled"),
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
	}

	if stripeCustomerID.Valid && stripeCustomerID.String != "" {
		params.Customer = stripe.String(stripeCustomerID.String)
	} else {
		params.CustomerEmail = stripe.String(email)
	}

	s, err := checkoutsession.New(params)
	if err != nil {
		log.Printf("Error creating checkout session: %v", err)
		http.Error(w, "Failed to create checkout session", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, s.URL, http.StatusSeeOther)
}

func CustomerPortalHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := GetAuthenticatedUserID(w, r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	db, err := helpers.GetConnectionManager().GetConnection("postgres")
	if err != nil {
		log.Printf("Database connection error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var stripeCustomerID sql.NullString
	err = db.QueryRow("SELECT stripe_customer_id FROM users WHERE id = $1", userID).Scan(&stripeCustomerID)
	if err != nil {
		log.Printf("Error fetching user: %v", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if !stripeCustomerID.Valid || stripeCustomerID.String == "" {
		http.Error(w, "No billing information available", http.StatusBadRequest)
		return
	}

	domain := os.Getenv("DOMAIN")
	if domain == "" {
		domain = "http://localhost:8080"
	}

	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(stripeCustomerID.String),
		ReturnURL: stripe.String(domain + "/dashboard"),
	}

	portalSession, err := session.New(params)
	if err != nil {
		log.Printf("Error creating portal session: %v", err)
		http.Error(w, "Failed to create portal session", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, portalSession.URL, http.StatusSeeOther)
}

func StripeWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading webhook body: %v", err)
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if webhookSecret == "" {
		log.Printf("STRIPE_WEBHOOK_SECRET is not configured")
		http.Error(w, "Webhook not configured", http.StatusInternalServerError)
		return
	}
	signatureHeader := r.Header.Get("Stripe-Signature")

	event, err := webhook.ConstructEvent(payload, signatureHeader, webhookSecret)
	if err != nil {
		log.Printf("Webhook signature verification failed: %v", err)
		http.Error(w, "Invalid signature", http.StatusBadRequest)
		return
	}

	db, err := helpers.GetConnectionManager().GetConnection("postgres")
	if err != nil {
		log.Printf("Database connection error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var handlerErr error
	switch event.Type {
	case "checkout.session.completed":
		handlerErr = handleCheckoutCompleted(db, event)
	case "customer.subscription.updated", "customer.subscription.created", "customer.subscription.resumed":
		handlerErr = handleSubscriptionUpdated(db, event)
	case "customer.subscription.deleted":
		handlerErr = handleSubscriptionDeleted(db, event)
	default:
		log.Printf("Unhandled webhook event type: %s", event.Type)
	}

	if handlerErr != nil {
		log.Printf("Webhook handler error: %v", handlerErr)
		http.Error(w, "Processing failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleCheckoutCompleted(db *sql.DB, event stripe.Event) error {
	var checkoutSession stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &checkoutSession); err != nil {
		log.Printf("Error parsing checkout.session.completed: %v", err)
		return err
	}

	if checkoutSession.Customer == nil || checkoutSession.CustomerDetails == nil {
		log.Printf("Checkout session missing customer data")
		return fmt.Errorf("checkout session missing customer data")
	}

	customerID := checkoutSession.Customer.ID
	customerEmail := checkoutSession.CustomerDetails.Email

	log.Printf("Checkout completed for customer %s (%s)", customerID, customerEmail)

	_, err := db.Exec(
		"UPDATE users SET is_premium = TRUE, stripe_customer_id = $1 WHERE email = $2",
		customerID, customerEmail,
	)
	if err != nil {
		log.Printf("Error updating user premium status: %v", err)
		return err
	}

	log.Printf("User %s upgraded to premium", customerEmail)
	return nil
}

func handleSubscriptionUpdated(db *sql.DB, event stripe.Event) error {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		log.Printf("Error parsing customer.subscription.updated: %v", err)
		return err
	}

	if subscription.Customer == nil {
		log.Printf("Subscription event missing customer data")
		return fmt.Errorf("subscription event missing customer data")
	}

	customerID := subscription.Customer.ID
	status := subscription.Status

	log.Printf("Subscription updated for customer %s: status=%s", customerID, status)

	isPremium := status == stripe.SubscriptionStatusActive || status == stripe.SubscriptionStatusTrialing

	_, err := db.Exec(
		"UPDATE users SET is_premium = $1 WHERE stripe_customer_id = $2",
		isPremium, customerID,
	)
	if err != nil {
		log.Printf("Error updating subscription status: %v", err)
		return err
	}
	return nil
}

func handleSubscriptionDeleted(db *sql.DB, event stripe.Event) error {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		log.Printf("Error parsing customer.subscription.deleted: %v", err)
		return err
	}

	if subscription.Customer == nil {
		log.Printf("Subscription deleted event missing customer data")
		return fmt.Errorf("subscription deleted event missing customer data")
	}

	customerID := subscription.Customer.ID

	log.Printf("Subscription deleted for customer %s", customerID)

	_, err := db.Exec(
		"UPDATE users SET is_premium = FALSE WHERE stripe_customer_id = $1",
		customerID,
	)
	if err != nil {
		log.Printf("Error revoking premium status: %v", err)
		return err
	}

	log.Printf("Premium status revoked for customer %s", customerID)
	return nil
}

func GetStripePublishableKey() string {
	return os.Getenv("STRIPE_PUBLISHABLE_KEY")
}
