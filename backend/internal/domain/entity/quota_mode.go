package entity

import "errors"

type QuotaMode string

const (
	QuotaModeReal    QuotaMode = "real"
	QuotaModeInherit QuotaMode = "inherit"
	QuotaModeVirtual QuotaMode = "virtual"
)

func (m QuotaMode) Valid() error {
	switch m {
	case QuotaModeReal, QuotaModeInherit, QuotaModeVirtual:
		return nil
	default:
		return errors.New("invalid quota_mode: must be real, inherit, or virtual")
	}
}
