// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/gardener/gardener/pkg/utils"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/gardener/gardener-extension-shoot-cert-service/pkg/apis/config"
)

// ValidateConfiguration validates the passed configuration instance.
func ValidateConfiguration(config *config.Configuration) field.ErrorList {
	allErrs := field.ErrorList{}

	if config.IssuerName == "" {
		allErrs = append(allErrs, field.Required(field.NewPath("issuerName"), "field is required"))
	}

	if config.DefaultRequestsPerDayQuota != nil && *config.DefaultRequestsPerDayQuota < 1 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("defaultRequestsPerDayQuota"), *config.DefaultRequestsPerDayQuota, "must be >= 1"))
	}

	allErrs = append(allErrs, validateACME(&config.ACME, field.NewPath("acme"))...)

	allErrs = append(allErrs, validatePrivateKeyDefaults(config.PrivateKeyDefaults, field.NewPath("privateKeyDefaults"))...)

	return allErrs
}

func validateACME(acme *config.ACME, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if _, err := url.ParseRequestURI(acme.Server); err != nil {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("server"), acme.Server, err.Error()))
	}

	if !utils.TestEmail(acme.Email) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("email"), acme.Email, "must be a valid mail address"))
	}

	if acme.PrecheckNameservers != nil {
		servers := strings.Split(*acme.PrecheckNameservers, ",")
		if len(servers) == 1 && len(servers[0]) == 0 {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("precheckNameservers"), *acme.PrecheckNameservers, "must contain at least one DNS server IP"))
		} else {
			for i, server := range servers {
				if net.ParseIP(server) == nil {
					allErrs = append(allErrs, field.Invalid(fldPath.Child("precheckNameservers"), *acme.PrecheckNameservers, fmt.Sprintf("invalid IP for %d. DNS server", i+1)))
				}
			}
		}
	}

	if acme.CACertificates != nil {
		s := strings.TrimSpace(*acme.CACertificates)
		if !strings.HasPrefix(s, "-----BEGIN CERTIFICATE-----") || !strings.HasSuffix(s, "-----END CERTIFICATE-----") {
			short := s
			if len(short) > 60 {
				short = s[:30] + "..." + s[len(s)-30:]
			}
			allErrs = append(allErrs, field.Invalid(fldPath.Child("caCertificates"), short, "invalid certificate(s), expected PEM format)"))
		}
	}
	return allErrs
}

func validatePrivateKeyDefaults(defaults *config.PrivateKeyDefaults, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if defaults == nil {
		return allErrs
	}

	if defaults.Algorithm != nil && *defaults.Algorithm != "RSA" && *defaults.Algorithm != "ECDSA" {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("algorithm"), *defaults.Algorithm, "algorithm must either be 'RSA' or 'ECDSA'"))
	}
	if defaults.SizeRSA != nil && *defaults.SizeRSA != 2048 && *defaults.SizeRSA != 3072 && *defaults.SizeRSA != 4096 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("sizeRSA"), *defaults.SizeRSA, "size for RSA algorithm must either be '2048' or '3072' or '4096"))
	}
	if defaults.SizeECDSA != nil && *defaults.SizeECDSA != 256 && *defaults.SizeECDSA != 384 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("sizeECDSA"), *defaults.SizeECDSA, "size for ECDSA algorithm must either be '256' or '384'"))
	}

	return allErrs
}
