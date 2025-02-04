package usecase

import (
	"fmt"
	"net"
	"strings"
)

// emailValidatorImpl implements EmailValidator interface
type emailValidatorImpl struct {
	trustedDomains map[string]struct{} // Whitelist of trusted email domains
}

// NewEmailValidator creates a new instance of EmailValidator
func NewEmailValidator(trustedDomains []string) EmailValidator {
	domains := make(map[string]struct{}, len(trustedDomains))
	for _, domain := range trustedDomains {
		domains[domain] = struct{}{}
	}

	return &emailValidatorImpl{
		trustedDomains: domains,
	}
}

func (v *emailValidatorImpl) ValidateDKIM(email string) error {
	// TODO: Implement DKIM signature verification using a DKIM library
	// For now, doing basic header check
	if !strings.Contains(email, "DKIM-Signature:") {
		return fmt.Errorf("missing DKIM signature")
	}
	return nil
}

func (v *emailValidatorImpl) ValidateSPF(email string) error {
	// TODO: Implement proper SPF record checking
	// For example:
	// 1. Extract sender domain from email
	// 2. Lookup SPF record for domain
	// 3. Validate sending IP against SPF policy

	// Basic example:
	domain := v.extractDomain(email)
	if domain == "" {
		return fmt.Errorf("could not extract domain from email")
	}

	_, err := net.LookupTXT(domain)
	if err != nil {
		return fmt.Errorf("SPF record lookup failed: %v", err)
	}

	return nil
}

func (v *emailValidatorImpl) ValidateSender(email string) error {
	domain := v.extractDomain(email)
	if domain == "" {
		return fmt.Errorf("could not extract domain from email")
	}

	// Check if domain is in trusted list
	if _, trusted := v.trustedDomains[domain]; !trusted {
		return fmt.Errorf("sender domain %s is not trusted", domain)
	}

	return nil
}

func (v *emailValidatorImpl) extractDomain(email string) string {
	// Basic domain extraction from email
	// In production, use a proper email parsing library
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}

// AddTrustedDomain adds a domain to the trusted domains list
func (v *emailValidatorImpl) AddTrustedDomain(domain string) {
	v.trustedDomains[domain] = struct{}{}
}

// RemoveTrustedDomain removes a domain from the trusted domains list
func (v *emailValidatorImpl) RemoveTrustedDomain(domain string) {
	delete(v.trustedDomains, domain)
}
