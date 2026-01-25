package models

// RentData holds the rent information collected from user
type RentData struct {
	TotalBill int64  // Total rent amount
	Electric  int64  // Electric amount
	Water     int64  // Water amount
	OtherFees int64  // Calculated: TotalBill - Electric - Water
	Payer     string // Who paid the rent (e.g., @ng0cth1nh)

	// Per-member shares (calculated based on weights)
	MemberShares []MemberShare
}

// MemberShare holds the share for each member
type MemberShare struct {
	Username      string
	ElectricShare int64
	WaterShare    int64
	OtherShare    int64
	TotalShare    int64
}

// CalculateOtherFees calculates and sets the OtherFees field
func (r *RentData) CalculateOtherFees() {
	r.OtherFees = r.TotalBill - r.Electric - r.Water
}

// CalculateMemberShares calculates per-member shares based on weights
// Electric and Water: split by weight
// OtherFees: split equally
func (r *RentData) CalculateMemberShares(members []Member) {
	if len(members) == 0 {
		return
	}

	// Calculate total weight
	totalWeight := 0
	for _, m := range members {
		totalWeight += m.Weight
	}

	// Calculate shares
	r.MemberShares = make([]MemberShare, len(members))
	for i, m := range members {
		// Electric and Water: split by weight
		electricShare := (r.Electric * int64(m.Weight)) / int64(totalWeight)
		waterShare := (r.Water * int64(m.Weight)) / int64(totalWeight)
		// Other fees: split equally
		otherShare := r.OtherFees / int64(len(members))

		r.MemberShares[i] = MemberShare{
			Username:      m.Username,
			ElectricShare: electricShare,
			WaterShare:    waterShare,
			OtherShare:    otherShare,
			TotalShare:    electricShare + waterShare + otherShare,
		}
	}
}
